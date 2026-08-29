package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web"
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

	// DatabasePath is exposed so a test can assert on the file itself — that it is
	// a real file rather than :memory:, and that it is gone after the run.
	DatabasePath string

	// Client carries a cookie jar, so a request made after a login keeps the
	// session cookie exactly as a browser would. F1-B04 adds the login helper that
	// makes use of it; until then nothing sets a cookie.
	Client *http.Client
}

// newTestApp starts an application on its own database.
//
// Every call gets a fresh temp directory, so tests share nothing and may run in
// parallel. Cleanup is registered here rather than returned: a teardown a test can
// forget is a teardown a test will forget.
func newTestApp(t *testing.T) *testApp {
	t.Helper()

	return newTestAppWithRoutes(t, nil)
}

// newTestAppWithRoutes appends extra routes to the production router, for a test
// that needs a handler no real endpoint provides yet — one that panics, say.
//
// The routes are added to the real mux, so they run behind the real middleware
// chain; that is the point, since the middleware is what is under test.
func newTestAppWithRoutes(t *testing.T, register func(chi.Router)) *testApp {
	t.Helper()

	// A real file rather than :memory: — an in-memory database gives every
	// connection its own private schema, which would make the two pools invisible
	// to each other and hide exactly the bugs these tests exist to catch.
	databasePath := filepath.Join(t.TempDir(), "wedding-test.db")

	config := configuration.Config{DatabasePath: databasePath}

	database, err := configuration.OpenDatabase(config)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, database.Close())
	})

	// Same order as main: migrate before anything serves a request.
	require.NoError(t, persistence.Migrate(context.Background(), database.Write, discardingLogger().Logger))

	router := web.NewRouter(discardingLogger(), database)
	if register != nil {
		register(router)
	}

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &testApp{
		t:            t,
		URL:          server.URL,
		Database:     database,
		DatabasePath: databasePath,
		Client:       &http.Client{Jar: jar},
	}
}

// discardingLogger keeps a passing run readable. A failing assertion says more than a
// request log line would, and the panic-recovery test would otherwise print a stack
// trace that reads like a failure.
func discardingLogger() *httplog.Logger {
	return &httplog.Logger{
		Logger:  slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError})),
		Options: httplog.Options{Concise: true},
	}
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

func (app *testApp) request(method, path string, payload any) *testResponse {
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
