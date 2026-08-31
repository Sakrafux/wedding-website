package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/security"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

// testApp is a running instance of the whole application: a migrated database on a
// fresh temp file, the production router behind the production middleware, and an
// HTTP client that keeps cookies. Everything a test needs is reachable from here, so
// setting one up is a single line and tearing it down is not the test's problem.
type testApp struct {
	t *testing.T

	// URL is the base address of the test server, without a trailing slash.
	URL string

	Database *configuration.Database

	// Router is the real mux. Exposed so a test can enumerate the registered
	// routes rather than restate them — which is what makes the admin-gate suite
	// fail when somebody adds a route and forgets the guard.
	Router *chi.Mux

	// DatabasePath is exposed so a test can assert on the file itself — that it is
	// a real file rather than :memory:, and that it is gone after the run.
	DatabasePath string

	// Logs is everything the application logged during the test — request lines and
	// the application's own entries alike. Assertable, because a couple of decisions
	// in this app are recorded *only* as a log line: a CSV export, for instance, is
	// deliberately not an audit row.
	Logs *testLog

	// Client carries a cookie jar, so a request made after a login keeps the
	// session cookie exactly as a browser would — which is what makes logIn below
	// enough to authenticate every later request in a test.
	Client *http.Client
}

// testAppOption tweaks how the application under test is built. Defaults are what
// a test that does not care should get, so an option is only ever present in the
// tests that are actually about that setting.
type testAppOption func(*testAppSpec)

type testAppSpec struct {
	registerExtraRoutes func(chi.Router)
	cookieSecure        bool
	trustedProxies      []netip.Prefix
	generateCode        func() string
}

// The admin credentials every test app is configured with. Constants rather than
// fixture options: there is exactly one admin, and a test that wants to fail the
// login can simply send something else.
const (
	testAdminUser     = "test-admin"
	testAdminPassword = "test-admin-password"
)

// withExtraRoutes appends routes to the production router, for a test that needs a
// handler no real endpoint provides — one that panics, say.
//
// The routes are added to the real mux, so they run behind the real middleware
// chain; that is the point, since the middleware is what is under test.
func withExtraRoutes(register func(chi.Router)) testAppOption {
	return func(spec *testAppSpec) { spec.registerExtraRoutes = register }
}

// withCodeGenerator replaces the login-code generator, for the collision-retry
// tests. With 32^6 codes and a handful in use, the retry path cannot be reached any
// other way — and an untested retry is one that may not work when it is finally
// needed.
func withCodeGenerator(generate func() string) testAppOption {
	return func(spec *testAppSpec) { spec.generateCode = generate }
}

// withSecureCookies turns SESSION_COOKIE_SECURE on, as production has it.
//
// Off by default, and that default is not laziness: the test server speaks plain
// HTTP, and Go's cookie jar — correctly — refuses to send a Secure cookie back
// over http. A secure-by-default harness would therefore make every authenticated
// test fail for a reason that has nothing to do with what it is testing. The one
// test that cares asserts on the Set-Cookie header instead of on the round trip.
func withSecureCookies() testAppOption {
	return func(spec *testAppSpec) { spec.cookieSecure = true }
}

// withTrustedProxies configures TRUSTED_PROXY_CIDRS. Empty by default, which is
// what a test that does not care should get: with no trusted proxy the client
// address is the direct peer, and X-Forwarded-For is ignored entirely.
func withTrustedProxies(cidrs ...string) testAppOption {
	return func(spec *testAppSpec) {
		for _, cidr := range cidrs {
			spec.trustedProxies = append(spec.trustedProxies, netip.MustParsePrefix(cidr))
		}
	}
}

// newTestApp starts an application on its own database.
//
// Every call gets a fresh temp directory, so tests share nothing and may run in
// parallel. Cleanup is registered here rather than returned: a teardown a test can
// forget is a teardown a test will forget.
func newTestApp(t *testing.T, options ...testAppOption) *testApp {
	t.Helper()

	var spec testAppSpec
	for _, option := range options {
		option(&spec)
	}

	// A real file rather than :memory: — an in-memory database gives every
	// connection its own private schema, which would make the two pools invisible
	// to each other and hide exactly the bugs these tests exist to catch.
	databasePath := filepath.Join(t.TempDir(), "wedding-test.db")

	logs := &testLog{}
	logger := testLogger(logs)

	config := configuration.Config{
		DatabasePath:        databasePath,
		SessionCookieSecure: spec.cookieSecure,
		TrustedProxyCIDRs:   spec.trustedProxies,
		AdminUser:           testAdminUser,
		AdminPassword:       testAdminPassword,
	}

	database, err := configuration.OpenDatabase(config)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, database.Close())
	})

	// Same order as main: migrate before anything serves a request.
	require.NoError(t, persistence.Migrate(context.Background(), database.Write, logger.Logger))

	// Wired exactly as main wires it, so the tests exercise the real chain of
	// handler, use case and store rather than a shorter one assembled for them.
	sessions := persistence.NewSessionStore(database)
	householdStore := persistence.NewHouseholdStore(database)
	if spec.generateCode != nil {
		householdStore = householdStore.WithCodeGenerator(spec.generateCode)
	}
	guestStore := persistence.NewGuestStore(database)
	auditStore := persistence.NewAuditStore(database)

	auth := application.NewAuth(
		sessions,
		householdStore,
		persistence.NewSettingStore(database),
		auditStore,
		security.AdminCredentials{User: config.AdminUser, Password: config.AdminPassword},
		logger.Logger,
	)

	router := web.NewRouter(logger, web.Dependencies{
		Config:     config,
		Database:   database,
		Auth:       auth,
		Households: application.NewHouseholds(householdStore, guestStore, sessions, auditStore, logger.Logger),
		Exports:    application.NewExports(householdStore, persistence.NewExportStore(database)),
	})
	if spec.registerExtraRoutes != nil {
		spec.registerExtraRoutes(router)
	}

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &testApp{
		t:            t,
		URL:          server.URL,
		Database:     database,
		Router:       router,
		DatabasePath: databasePath,
		Logs:         logs,
		Client:       &http.Client{Jar: jar},
	}
}

// onANewDevice returns the same application reached through a fresh cookie jar.
//
// It is what lets one test hold two sessions at once — two households logged in on
// two phones, or an admin alongside a household. A single jar cannot: a login revokes
// whatever session the request already carried, which is deliberate (one cookie, one
// subject) and would otherwise quietly undo the setup of any such test.
func (app *testApp) onANewDevice() *testApp {
	app.t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(app.t, err)

	device := *app
	device.Client = &http.Client{Jar: jar}
	return &device
}

// countHouseholdSessions is how the revocation tests assert on the table rather than
// on what the endpoint said it did.
func (app *testApp) countHouseholdSessions(householdID int64) int {
	app.t.Helper()

	var count int
	require.NoError(app.t, app.Database.Read.Get(&count,
		`SELECT COUNT(*) FROM session WHERE subject_type = 'household' AND subject_id = ?`, householdID))
	return count
}

// testLogger sends the application's logs to a buffer instead of to stdout.
//
// A buffer rather than io.Discard so a test can assert on a line the application
// records nowhere else, and rather than the test's own output because a passing run
// should be readable — the panic-recovery test would otherwise print a stack trace
// that looks exactly like a failure.
func testLogger(out io.Writer) *httplog.Logger {
	return &httplog.Logger{
		Logger:  slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo})),
		Options: httplog.Options{Concise: true},
	}
}

// testLog is a buffer safe to read while the server is still writing to it: the
// handlers run in the httptest server's own goroutines, so an unguarded bytes.Buffer
// would be a data race the race detector finds sooner or later.
type testLog struct {
	mutex   sync.Mutex
	entries bytes.Buffer
}

func (log *testLog) Write(entry []byte) (int, error) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	return log.entries.Write(entry)
}

// String is everything logged so far, as one blob of JSON lines.
func (log *testLog) String() string {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	return log.entries.String()
}

// testResponse is a response already read into memory, so a test can assert on the
// body more than once and never has to close anything.
type testResponse struct {
	t *testing.T

	Status      int
	ContentType string
	Body        string
	Header      http.Header
}

func (app *testApp) get(path string) *testResponse {
	app.t.Helper()

	return app.request(http.MethodGet, path, nil)
}

// postJSON marshals payload as the request body. A string or []byte payload is sent
// as-is, which is how a test sends deliberately malformed JSON.
func (app *testApp) postJSON(path string, payload any) *testResponse {
	app.t.Helper()

	return app.request(http.MethodPost, path, payload)
}

// post sends a POST with no body, for the endpoints that take none.
func (app *testApp) post(path string) *testResponse {
	app.t.Helper()

	return app.request(http.MethodPost, path, nil)
}

// patchJSON and deleteRequest exist for the admin CRUD routes; both go through
// request, so a body is encoded and a session cookie carried exactly as for a POST.
func (app *testApp) patchJSON(path string, payload any) *testResponse {
	app.t.Helper()

	return app.request(http.MethodPatch, path, payload)
}

func (app *testApp) deleteRequest(path string) *testResponse {
	app.t.Helper()

	return app.request(http.MethodDelete, path, nil)
}

func (app *testApp) request(method, path string, payload any) *testResponse {
	app.t.Helper()

	return app.requestWith(method, path, payload, nil)
}

// requestWith adds request headers. Its one caller today is the rate-limit suite,
// which needs several client addresses against one server.
func (app *testApp) requestWith(method, path string, payload any, headers map[string]string) *testResponse {
	app.t.Helper()

	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(encodePayload(app.t, payload))
	}

	request, err := http.NewRequest(method, app.URL+path, body)
	require.NoError(app.t, err)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}

	response, err := app.Client.Do(request)
	require.NoErrorf(app.t, err, "%s %s", method, path)
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	require.NoErrorf(app.t, err, "reading body of %s %s", method, path)

	return &testResponse{
		t:           app.t,
		Status:      response.StatusCode,
		ContentType: response.Header.Get("Content-Type"),
		Body:        strings.TrimSpace(string(raw)),
		Header:      response.Header,
	}
}

func encodePayload(t *testing.T, payload any) []byte {
	t.Helper()

	switch typed := payload.(type) {
	case string:
		return []byte(typed)
	case []byte:
		return typed
	default:
		encoded, err := json.Marshal(payload)
		require.NoError(t, err)
		return encoded
	}
}

// logIn redeems a household code through the real endpoint, leaving the session
// cookie in the jar so every later request in the test is authenticated.
//
// It goes through HTTP rather than creating a session in the database, so that a
// test about anything downstream of login still fails if login itself breaks.
func (app *testApp) logIn(code string) *testResponse {
	app.t.Helper()

	return app.postJSON("/api/auth/login", map[string]string{"code": code})
}

// logInFrom attempts a household login as if it came from clientIP.
//
// Only effective on an app configured with withTrustedProxies covering the
// loopback address the test server is reached on — which is exactly the property
// the trusted-proxy rule is there to enforce.
func (app *testApp) logInFrom(code, clientIP string) *testResponse {
	app.t.Helper()

	return app.requestWith(http.MethodPost, "/api/auth/login",
		map[string]string{"code": code}, map[string]string{"X-Forwarded-For": clientIP})
}

// logInAsAdmin redeems the configured admin credentials through the real endpoint.
func (app *testApp) logInAsAdmin() *testResponse {
	app.t.Helper()

	return app.postJSON("/api/auth/admin/login", map[string]string{"user": testAdminUser, "password": testAdminPassword})
}

// householdStore and guestStore are the production stores over the test database,
// for asserting on what an endpoint actually wrote rather than on what it reported
// having written.
func (app *testApp) householdStore() *persistence.HouseholdStore {
	return persistence.NewHouseholdStore(app.Database)
}

func (app *testApp) guestStore() *persistence.GuestStore {
	return persistence.NewGuestStore(app.Database)
}

// storedCode reads a household's login code straight from the table. Used wherever a
// test has to know whether a code changed, since no response a test asserts on is
// allowed to be trusted for that.
func (app *testApp) storedCode(householdID int64) string {
	app.t.Helper()

	var code string
	require.NoError(app.t, app.Database.Read.Get(&code, `SELECT code FROM household WHERE id = ?`, householdID))
	return code
}

// sessionStore is the production store over the test database, for the tests that
// are about the store itself and for seeding a session no endpoint can create yet.
func (app *testApp) sessionStore() *persistence.SessionStore {
	return persistence.NewSessionStore(app.Database)
}

// putSessionCookie places a raw token in the cookie jar by hand, for a test that
// needs the client to present a cookie no endpoint would ever issue — a leftover
// value, or a session seeded straight into the store with a chosen expiry.
func (app *testApp) putSessionCookie(token string) {
	app.t.Helper()

	serverURL, err := url.Parse(app.URL)
	require.NoError(app.t, err)

	app.Client.Jar.SetCookies(serverURL, []*http.Cookie{{
		Name:  middleware.SessionCookieName,
		Value: token,
		Path:  "/",
	}})
}

// auditRows returns the audit log as it stands, newest last. Tests assert against
// the table itself, because the whole value of an append-only log is that it says
// what happened rather than what an endpoint reports having happened.
func (app *testApp) auditRows() []auditRow {
	app.t.Helper()

	var rows []auditRow
	require.NoError(app.t, app.Database.Read.Select(&rows,
		`SELECT at, actor_type, actor_id, entity, entity_id, action, before, after FROM audit_log ORDER BY id`))
	return rows
}

// auditPayloads is the whole log's before/after JSON concatenated, for asserting that
// a value — a login code, a password — appears nowhere in it under any key.
func (app *testApp) auditPayloads() string {
	app.t.Helper()

	var payloads string
	require.NoError(app.t, app.Database.Read.Get(&payloads,
		`SELECT COALESCE(GROUP_CONCAT(COALESCE(before, '') || '|' || COALESCE(after, '')), '') FROM audit_log`))
	return payloads
}

// auditRow mirrors the audit_log columns. Kept in the test package rather than
// imported, so a column rename shows up here as a failure.
type auditRow struct {
	At        string         `db:"at"`
	ActorType string         `db:"actor_type"`
	ActorID   sql.NullInt64  `db:"actor_id"`
	Entity    string         `db:"entity"`
	EntityID  int64          `db:"entity_id"`
	Action    string         `db:"action"`
	Before    sql.NullString `db:"before"`
	After     sql.NullString `db:"after"`
}

// countSessions is how the revocation and purge tests assert on the table rather
// than on the API's opinion of it.
func (app *testApp) countSessions() int {
	app.t.Helper()

	var count int
	require.NoError(app.t, app.Database.Read.Get(&count, `SELECT COUNT(*) FROM session`))
	return count
}

// setCookie returns the cookie the response set, or nil. Used to assert on the
// attributes — HttpOnly, SameSite, Max-Age — that the jar itself does not expose.
func (response *testResponse) setCookie(name string) *http.Cookie {
	response.t.Helper()

	for _, cookie := range (&http.Response{Header: response.Header}).Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// decodeJSON unmarshals the body into target, failing the test if it is not JSON.
func (response *testResponse) decodeJSON(target any) {
	response.t.Helper()

	require.Truef(response.t, strings.HasPrefix(response.ContentType, "application/json"),
		"expected a JSON response, got content type %q", response.ContentType)
	require.NoErrorf(response.t, json.Unmarshal([]byte(response.Body), target), "body: %s", response.Body)
}

// errorEnvelope decodes the error envelope from E0-06.
func (response *testResponse) errorEnvelope() errorBody {
	response.t.Helper()

	var envelope struct {
		Error errorBody `json:"error"`
	}
	response.decodeJSON(&envelope)

	return envelope.Error
}

// errorBody mirrors dto.ErrorBody instead of importing it, so a rename in the DTO
// shows up here as a failing wire-format assertion rather than compiling silently.
type errorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Fields    map[string]string `json:"fields"`
}

// assertNoLeak fails if the response carries data no guest may ever see. Every
// guest-facing test calls it.
//
// This is the mechanical half of the DTO privacy rule from 04-architecture: the rule
// says "never serialize domain structs", and a rule enforced by discipline alone is
// one that breaks during a late-night change. The extra forbiddenValues are for the
// values a test knows are secret — a household's login code, say — which catches a
// leak under a field name nobody thought to forbid.
func (response *testResponse) assertNoLeak(forbiddenValues ...string) {
	response.t.Helper()

	if leak := findLeak(response.Body); leak != "" {
		response.t.Errorf("response leaks private data: %s\nbody: %s", leak, response.Body)
	}

	for _, value := range forbiddenValues {
		if value != "" && strings.Contains(response.Body, value) {
			response.t.Errorf("response leaks the secret value %q\nbody: %s", value, response.Body)
		}
	}
}

// forbiddenFields are the JSON keys that must never appear in a guest-facing
// response: the household login code, our private note, and every column of
// budget_item that carries a number or a name we negotiated.
var forbiddenFields = map[string]bool{
	"code":           true,
	"admin_note":     true,
	"planned_cents":  true,
	"actual_cents":   true,
	"paid_cents":     true,
	"external_cents": true,
	"per_head_cents": true,
	"paid_by":        true,
	"vendor":         true,
}

// findLeak returns a description of the first forbidden field in a JSON body, or the
// empty string if there is none.
//
// Split out from assertNoLeak so the leak detector itself is testable without a fake
// *testing.T — a privacy check nobody has watched fail is a privacy check that may
// not work.
//
// A body that is not JSON — the SPA shell, robots.txt — is not scanned: there are no
// field names in it, and the value scan in assertNoLeak covers what it can.
func findLeak(body string) string {
	var decoded any
	if json.Unmarshal([]byte(body), &decoded) != nil {
		return ""
	}

	return findForbiddenField(decoded, "$")
}

func findForbiddenField(node any, path string) string {
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			childPath := path + "." + key
			// The one exception: the error envelope's own "code" is the machine
			// -readable error kind ("not_found"), not a household login code.
			if forbiddenFields[key] && childPath != "$.error.code" {
				return "forbidden field " + childPath
			}
			if leak := findForbiddenField(value, childPath); leak != "" {
				return leak
			}
		}
	case []any:
		for _, value := range typed {
			if leak := findForbiddenField(value, path+"[]"); leak != "" {
				return leak
			}
		}
	}

	return ""
}
