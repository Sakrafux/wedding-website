package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/security"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

// sessionExpiry reads a session's stored expiry, which is how the refresh-throttle
// test observes whether a write happened at all.
func (app *testApp) sessionExpiry(t *testing.T) time.Time {
	t.Helper()

	var stored string
	require.NoError(t, app.Database.Read.Get(&stored, `SELECT expires_at FROM session`))

	parsed, err := time.Parse(time.RFC3339, stored)
	require.NoError(t, err)
	return parsed
}

// ageSession moves a session's timestamps into the past, standing in for a guest
// who last opened the site some time ago. Cheaper and far more reliable than
// making the test wait a day.
//
// Both columns move together, as they would have if the session really were that
// old. Ageing last_seen_at alone would leave expires_at where it is, and since the
// stored format has second precision the refreshed expiry would land in the same
// second as the original — a test that could not tell a write from a no-op.
//
// The arithmetic is SQLite's, so the rewritten values keep exactly the format the
// schema's own defaults produce.
func (app *testApp) ageSession(t *testing.T, by string) {
	t.Helper()

	_, err := app.Database.Write.Exec(`
		UPDATE session SET
			last_seen_at = strftime('%Y-%m-%dT%H:%M:%SZ', last_seen_at, ?),
			expires_at   = strftime('%Y-%m-%dT%H:%M:%SZ', expires_at, ?)`, by, by)
	require.NoError(t, err)
}

// An anonymous request is a normal request, not an error. The login endpoints and
// the probes sit behind the same middleware and must work without a session.
func TestAnonymousRequestsReachPublicRoutes(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	assert.Equal(t, http.StatusOK, app.get("/api/health").Status)
	assert.Equal(t, http.StatusOK, app.get("/api/ready").Status)
}

// A cookie value that is not a token at all must not reach the error path: it is
// the shape of every stale cookie left over from a previous deployment.
func TestGarbageSessionCookieIsTreatedAsAnonymous(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	app.putSessionCookie("this is not a session token")

	guarded := app.get("/api/me")
	assert.Equal(t, http.StatusUnauthorized, guarded.Status)
	assert.Equal(t, "unauthenticated", guarded.errorEnvelope().Code)

	// The useless cookie is cleared on the way, so the browser stops presenting it
	// on every request for the next year. Asserted on the first request the cookie
	// is sent with: by the second one the jar has already dropped it.
	cleared := guarded.setCookie(middleware.SessionCookieName)
	require.NotNil(t, cleared)
	assert.Equal(t, -1, cleared.MaxAge)

	// A public route is unaffected by any of it.
	assert.Equal(t, http.StatusOK, app.get("/api/health").Status)
}

func TestHouseholdSessionPassesTheHouseholdGate(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	assert.Equal(t, http.StatusOK, app.get("/api/me").Status)
}

// The boundary that matters: admin-only data must be genuinely unreachable from a
// guest session, not merely hidden by a frontend that does not render the link.
func TestHouseholdSessionIsRefusedByTheAdminGate(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	response := app.get("/api/admin/budget")

	assert.Equal(t, http.StatusUnauthorized, response.Status)
	assert.Equal(t, "unauthenticated", response.errorEnvelope().Code)
}

func TestAnonymousRequestIsRefusedByTheAdminGate(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	assert.Equal(t, http.StatusUnauthorized, app.get("/api/admin/budget").Status)
}

// The admin passes the gate and then meets an ordinary 404, because no route is
// mounted behind it yet (F1-B07). The distinction is the assertion: 404 proves the
// gate let them through, where 401 would prove it did not.
func TestAdminSessionPassesTheAdminGate(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	app.putSessionCookie(app.createAdminSession(t))

	response := app.get("/api/admin/budget")

	assert.Equal(t, http.StatusNotFound, response.Status)
	assert.Equal(t, "not_found", response.errorEnvelope().Code)
}

// An admin is not a household: they have no members, no RSVP and no seat, so a
// household route has nothing to answer them with.
func TestAdminSessionIsRefusedByTheHouseholdGate(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	app.putSessionCookie(app.createAdminSession(t))

	response := app.get("/api/me")

	assert.Equal(t, http.StatusUnauthorized, response.Status)
	assert.Equal(t, "unauthenticated", response.errorEnvelope().Code)
}

func TestExpiredSessionIsRefused(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	token := security.NewSessionToken()
	expired := newSessionFor(household.ID, token, time.Now().Add(-time.Minute), time.Now().Add(-2*time.Hour))
	require.NoError(t, app.sessionStore().Create(context.Background(), expired))
	app.putSessionCookie(token)

	assert.Equal(t, http.StatusUnauthorized, app.get("/api/me").Status)
	assert.Equal(t, 0, app.countSessions(), "an expired session should not be left in the table")
}

// A household deleted while somebody was logged in leaves a session pointing at
// nothing. Resolving it as anonymous — and deleting it — is what stops every query
// downstream from having to defend against a dangling id.
func TestSessionOfADeletedHouseholdIsTreatedAsAnonymous(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	_, err := app.Database.Write.Exec(`DELETE FROM household WHERE id = ?`, household.ID)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, app.get("/api/me").Status)
	assert.Equal(t, 0, app.countSessions(), "the orphaned session should be gone")
}

// Rolling refresh is what makes "log in once, ever" true, and the throttle is what
// keeps it from costing a write on every request against a single-writer database.
func TestSessionRefreshesAtMostOnceADay(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	afterLogin := app.sessionExpiry(t)

	require.Equal(t, http.StatusOK, app.get("/api/me").Status)
	assert.Equal(t, afterLogin, app.sessionExpiry(t), "a second request the same day must not write")

	app.ageSession(t, "-25 hours")
	aged := app.sessionExpiry(t)

	require.Equal(t, http.StatusOK, app.get("/api/me").Status)
	refreshed := app.sessionExpiry(t)
	assert.True(t, refreshed.After(aged), "a request a day later should roll the session forward")
	assert.WithinDuration(t, afterLogin, refreshed, time.Minute, "and rolled to a full lifetime from now")

	// And having just rolled, it settles down again.
	require.Equal(t, http.StatusOK, app.get("/api/me").Status)
	assert.Equal(t, refreshed, app.sessionExpiry(t))
}

// An admin session that rolled on use would never end. Asserted through the real
// middleware, because that is where the mistake would be made.
//
// Aged by seven hours rather than the twenty-five the household test uses: an
// eight-hour session cannot survive to the refresh threshold, which is the point —
// the only way an admin session ever rolls is if the subject-type check is dropped,
// and then it would roll here.
func TestAdminSessionDoesNotRoll(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	app.putSessionCookie(app.createAdminSession(t))

	app.ageSession(t, "-7 hours")
	aged := app.sessionExpiry(t)

	require.Equal(t, http.StatusNotFound, app.get("/api/admin/budget").Status)

	assert.Equal(t, aged, app.sessionExpiry(t), "an admin session that rolled would never end")
}

// createAdminSession writes an admin session straight through the store and
// returns its token. The admin login endpoint is F1-B07; until it exists this is
// how a test gets an admin cookie.
func (app *testApp) createAdminSession(t *testing.T) string {
	t.Helper()

	token := security.NewSessionToken()
	session := domain.NewSession(security.HashSessionToken(token), domain.SubjectTypeAdmin, 0, time.Now(), "Mozilla/5.0 (Test)", "10.0.0.1")
	require.NoError(t, app.sessionStore().Create(context.Background(), session))

	return token
}
