package integration

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

func TestAdminLoginIssuesAShortLivedSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	response := app.logInAsAdmin()
	require.Equal(t, http.StatusOK, response.Status)

	var body struct {
		SubjectType string `json:"subject_type"`
	}
	response.decodeJSON(&body)
	assert.Equal(t, "admin", body.SubjectType)

	cookie := response.setCookie(middleware.SessionCookieName)
	require.NotNil(t, cookie, "the admin gets the same cookie as a household")
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)

	// Eight hours, not a year. The admin session reaches the budget and every
	// household's data, so an unlocked laptop must stop being a standing key.
	assert.InDelta(t, (8 * time.Hour).Seconds(), cookie.MaxAge, 60)

	var subjectType string
	require.NoError(t, app.Database.Read.Get(&subjectType, `SELECT subject_type FROM session`))
	assert.Equal(t, "admin", subjectType)
}

// Wrong username and wrong password must be one failure. Telling them apart
// confirms a valid username to whoever is guessing, and the one person who knows
// the real username does not need the hint.
func TestAdminLoginRejectsBothHalvesIdentically(t *testing.T) {
	t.Parallel()

	attempts := map[string]map[string]string{
		"wrong password": {"user": testAdminUser, "password": "not-the-password"},
		"wrong user":     {"user": "someone-else", "password": testAdminPassword},
		"both wrong":     {"user": "someone-else", "password": "not-the-password"},
		"empty":          {"user": "", "password": ""},
	}

	for name, credentials := range attempts {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(t)

			response := app.postJSON("/api/auth/admin/login", credentials)
			body := response.errorEnvelope()

			assert.Equal(t, http.StatusUnauthorized, response.Status)
			assert.Equal(t, "invalid_credentials", body.Code)
			assert.Equal(t, "Anmeldung fehlgeschlagen.", body.Message)
			assert.Nil(t, response.setCookie(middleware.SessionCookieName))
			assert.Equal(t, 0, app.countSessions())
		})
	}
}

// The password must not reach the audit log, the same way a household code must
// not. It is the one credential in the system with no expiry at all.
func TestAdminPasswordNeverReachesTheAuditLog(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)
	require.Equal(t, http.StatusUnauthorized,
		app.postJSON("/api/auth/admin/login", map[string]string{"user": testAdminUser, "password": "wrong"}).Status)

	var wholeTable string
	require.NoError(t, app.Database.Read.Get(&wholeTable,
		`SELECT COALESCE(GROUP_CONCAT(COALESCE(before, '') || '|' || COALESCE(after, '')), '') FROM audit_log`))

	assert.NotContains(t, wholeTable, testAdminPassword)
	assert.NotContains(t, wholeTable, "wrong")
}

// The hard boundary of the whole product, asserted against the routes that
// actually exist rather than against a list somebody has to remember to update.
// A route added under /api/admin without the guard fails here.
func TestEveryAdminRouteRefusesAHouseholdSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	routes := adminRoutes(t, app)
	require.NotEmpty(t, routes, "the admin subtree should have at least the catch-all")

	for _, route := range routes {
		response := app.requestWith(route.method, route.path, nil, nil)

		assert.Equalf(t, http.StatusUnauthorized, response.Status, "%s %s", route.method, route.path)

		// HEAD carries no body by definition, so there is no envelope to read on
		// that one; the status is the whole answer.
		if route.method != http.MethodHead {
			assert.Equalf(t, "unauthenticated", response.errorEnvelope().Code, "%s %s", route.method, route.path)
		}
	}
}

func TestEveryAdminRouteRefusesAnAnonymousRequest(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	for _, route := range adminRoutes(t, app) {
		response := app.requestWith(route.method, route.path, nil, nil)

		assert.Equalf(t, http.StatusUnauthorized, response.Status, "%s %s", route.method, route.path)
	}
}

// And the gate lets the admin through — otherwise the two tests above would pass
// just as well against a subtree that refuses everybody.
func TestAdminSessionReachesTheAdminSubtree(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)

	for _, route := range adminRoutes(t, app) {
		response := app.requestWith(route.method, route.path, nil, nil)

		assert.NotEqualf(t, http.StatusUnauthorized, response.Status,
			"the admin must pass the gate on %s %s", route.method, route.path)
	}
}

// The probe the admin frontend's guard is built on: it must say yes for an admin
// session and 401 for everything else, since that is the only thing that tells a
// returning admin their cookie is still good.
func TestAdminMeIdentifiesAnAdminSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)

	response := app.get("/api/admin/me")
	require.Equal(t, http.StatusOK, response.Status)

	var body struct {
		SubjectType string `json:"subject_type"`
	}
	response.decodeJSON(&body)
	assert.Equal(t, "admin", body.SubjectType)
	response.assertNoLeak(testAdminUser, testAdminPassword)
}

// One cookie, one subject: an admin logging in on a device that still holds a
// household session must not end up with both, or subject_type stops being the
// whole answer to "who is this".
func TestAdminLoginReplacesAHouseholdSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)

	assert.Equal(t, 1, app.countSessions(), "the household session must be revoked, not merely replaced in the jar")
	assert.Equal(t, http.StatusUnauthorized, app.get("/api/me").Status, "the admin is not a household")
}

func TestHouseholdLoginReplacesAnAdminSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	assert.Equal(t, 1, app.countSessions())
	assert.Equal(t, household.ID, app.get("/api/me").bootstrap().Household.ID)
	assert.Equal(t, http.StatusUnauthorized, app.get("/api/admin/budget").Status)
}

type registeredRoute struct {
	method string
	path   string
}

// adminRoutes walks the real router for everything mounted under /api/admin.
//
// Wildcards are replaced with a concrete segment so the path can actually be
// requested; what is under test is the gate, and the gate runs before routing
// resolves to a handler or to a 404.
func adminRoutes(t *testing.T, app *testApp) []registeredRoute {
	t.Helper()

	var routes []registeredRoute
	require.NoError(t, chi.Walk(app.Router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if !strings.HasPrefix(route, "/api/admin") {
			return nil
		}
		routes = append(routes, registeredRoute{method: method, path: strings.ReplaceAll(route, "/*", "/budget")})
		return nil
	}))

	return routes
}
