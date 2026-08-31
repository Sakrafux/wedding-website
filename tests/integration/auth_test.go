package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

// bootstrapBody mirrors dto.BootstrapResponse rather than importing it, so a
// renamed JSON tag shows up here as a failing wire-format assertion instead of
// compiling silently — the frontend reads these names, not the Go field names.
type bootstrapBody struct {
	Household struct {
		ID          int64  `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"household"`
	Members []struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Kind      string `json:"kind"`
		Origin    string `json:"origin"`
	} `json:"members"`
	Flags struct {
		RSVPOpen         bool `json:"rsvp_open"`
		SeatingPublished bool `json:"seating_published"`
		GalleryVisible   bool `json:"gallery_visible"`
		UploadsOpen      bool `json:"uploads_open"`
	} `json:"flags"`
	RSVPDeadline time.Time `json:"rsvp_deadline"`
}

func (response *testResponse) bootstrap() bootstrapBody {
	response.t.Helper()

	var body bootstrapBody
	response.decodeJSON(&body)
	return body
}

func TestLoginWithAValidCodeReturnsTheHousehold(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write,
		withCode("ABC234"),
		withDisplayName("Familie Müller"),
		withAdult("Anna", "Müller"),
		withChild("Emil", "Müller", 4),
	)

	response := app.logIn("ABC234")
	require.Equal(t, http.StatusOK, response.Status)

	body := response.bootstrap()

	assert.Equal(t, household.ID, body.Household.ID)
	assert.Equal(t, "Familie Müller", body.Household.DisplayName)

	require.Len(t, body.Members, 2)
	assert.Equal(t, "Anna", body.Members[0].FirstName)
	assert.Equal(t, "adult", body.Members[0].Kind)
	assert.Equal(t, "seeded", body.Members[0].Origin)
	assert.Equal(t, "Emil", body.Members[1].FirstName)
	assert.Equal(t, "child", body.Members[1].Kind)

	// Seeded by migration 0001: all three gates start closed, and the deadline is
	// far enough away that the RSVP form is open.
	assert.True(t, body.Flags.RSVPOpen)
	assert.False(t, body.Flags.SeatingPublished)
	assert.False(t, body.Flags.GalleryVisible)
	assert.False(t, body.Flags.UploadsOpen)
	assert.Equal(t, time.Date(2027, 5, 17, 21, 59, 59, 0, time.UTC), body.RSVPDeadline.UTC())
}

func TestLoginSetsAHardenedSessionCookie(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	cookie := app.logIn("ABC234").setCookie(middleware.SessionCookieName)
	require.NotNil(t, cookie, "login must set the session cookie")

	assert.True(t, cookie.HttpOnly, "no script needs the token")
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.NotEmpty(t, cookie.Value)

	// A year, matching the session row. The tolerance absorbs the moment between
	// creating the session and writing the header.
	assert.InDelta(t, (365 * 24 * time.Hour).Seconds(), cookie.MaxAge, 60)
}

// SESSION_COOKIE_SECURE exists only so local development over plain HTTP works.
// Both directions are asserted, because the failure that matters — Secure quietly
// off in production — is invisible until somebody reads a cookie off the wire.
func TestSecureCookieAttributeFollowsTheConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("off for local development", func(t *testing.T) {
		t.Parallel()

		app := newTestApp(t)
		seedHousehold(t, app.Database.Write, withCode("ABC234"))

		assert.False(t, app.logIn("ABC234").setCookie(middleware.SessionCookieName).Secure)
	})

	t.Run("on as production has it", func(t *testing.T) {
		t.Parallel()

		app := newTestApp(t, withSecureCookies())
		seedHousehold(t, app.Database.Write, withCode("ABC234"))

		assert.True(t, app.logIn("ABC234").setCookie(middleware.SessionCookieName).Secure)
	})
}

// Everything a guest might plausibly type, given a card printed as ABC234 and a
// phone that capitalises whatever it likes.
func TestLoginAcceptsEveryFormOfTheSameCode(t *testing.T) {
	t.Parallel()

	submitted := []string{"ABC234", "abc234", "ABC-234", "abc-234", "ABC 234", " abc 234 ", "abc–234"}

	for _, code := range submitted {
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(t)
			household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

			response := app.logIn(code)

			require.Equal(t, http.StatusOK, response.Status)
			assert.Equal(t, household.ID, response.bootstrap().Household.ID)
		})
	}
}

// A code that does not exist and a code that could not exist must be
// indistinguishable. Telling the two apart would turn guessing from hopeless into
// merely slow, and the guest cannot act on the difference anyway.
func TestLoginRejectsUnknownAndMalformedCodesIdentically(t *testing.T) {
	t.Parallel()

	rejected := map[string]string{
		"unknown but well formed": "ZZZZZZ",
		"too short":               "ABC",
		"too long":                "ABC2345",
		"excluded glyph":          "ABC23O",
		"empty":                   "",
		"not a code at all":       "<script>",
	}

	for name, code := range rejected {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(t)
			seedHousehold(t, app.Database.Write, withCode("ABC234"))

			response := app.logIn(code)
			body := response.errorEnvelope()

			assert.Equal(t, http.StatusUnauthorized, response.Status)
			assert.Equal(t, "unknown_login_code", body.Code)
			assert.Contains(t, body.Message, "Karte")
			assert.Nil(t, response.setCookie(middleware.SessionCookieName), "a failed login must set no cookie")
			assert.Equal(t, 0, app.countSessions())
		})
	}
}

// The CSRF control from 06-privacy-security: a cross-site HTML form cannot send
// application/json, and login is the one mutation that needs no existing cookie.
func TestLoginRequiresAJSONContentType(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	request, err := http.NewRequest(http.MethodPost, app.URL+"/api/auth/login", nil)
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Equal(t, 0, app.countSessions())
}

// "Did they even see it?" is the question the nudge list answers, and this column
// is the only thing that can answer it.
func TestLoginRecordsLastLogin(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	var beforeLogin *string
	require.NoError(t, app.Database.Read.Get(&beforeLogin, `SELECT last_login_at FROM household WHERE id = ?`, household.ID))
	require.Nil(t, beforeLogin, "a seeded household has never logged in")

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	var afterLogin *string
	require.NoError(t, app.Database.Read.Get(&afterLogin, `SELECT last_login_at FROM household WHERE id = ?`, household.ID))
	require.NotNil(t, afterLogin)

	recorded, err := time.Parse(time.RFC3339, *afterLogin)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), recorded, time.Minute)
}

// The leak this whole DTO discipline exists to prevent. Asserted on both bodies,
// because they are produced by the same mapper and a future change to it would
// otherwise only be caught on whichever endpoint someone remembered.
func TestBootstrapResponsesNeverCarryTheCodeOrTheAdminNote(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write,
		withCode("ABC234"),
		withAdminNote("Streiten sich mit den Nachbarn"),
		withGuests(2),
	)

	login := app.logIn("ABC234")
	require.Equal(t, http.StatusOK, login.Status)
	login.assertNoLeak("ABC234", "ABC-234", "Streiten sich mit den Nachbarn")

	me := app.get("/api/me")
	require.Equal(t, http.StatusOK, me.Status)
	me.assertNoLeak("ABC234", "ABC-234", "Streiten sich mit den Nachbarn")
}

func TestMeRequiresASession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	response := app.get("/api/me")

	assert.Equal(t, http.StatusUnauthorized, response.Status)
	assert.Equal(t, "unauthenticated", response.errorEnvelope().Code)
}

// /api/me must answer the same body as login, so that a returning guest and a
// just-logged-in guest render from the same data.
func TestMeReturnsTheSameBodyAsLogin(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"), withDisplayName("Familie Müller"), withGuests(3))

	login := app.logIn("ABC234")
	require.Equal(t, http.StatusOK, login.Status)

	me := app.get("/api/me")
	require.Equal(t, http.StatusOK, me.Status)

	assert.Equal(t, login.bootstrap(), me.bootstrap())
}

func TestLogoutRevokesTheSessionAndClearsTheCookie(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	require.Equal(t, 1, app.countSessions())

	response := app.post("/api/auth/logout")

	assert.Equal(t, http.StatusNoContent, response.Status)
	assert.Equal(t, 0, app.countSessions(), "revocation must be real, not just a cleared cookie")

	cleared := response.setCookie(middleware.SessionCookieName)
	require.NotNil(t, cleared)
	assert.Equal(t, -1, cleared.MaxAge)
	assert.Empty(t, cleared.Value)

	assert.Equal(t, http.StatusUnauthorized, app.get("/api/me").Status)
}

// Logging out twice, or without ever logging in, is not a mistake worth reporting:
// the caller wanted to end up unauthenticated and they are.
func TestLogoutIsIdempotentAndWorksWhenAnonymous(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	assert.Equal(t, http.StatusNoContent, app.post("/api/auth/logout").Status)

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	assert.Equal(t, http.StatusNoContent, app.post("/api/auth/logout").Status)
	assert.Equal(t, http.StatusNoContent, app.post("/api/auth/logout").Status)
}

// Logging in while holding a cookie that names nothing — the state everybody is in
// after a `DELETE FROM session`, and after any session simply expires.
//
// Worth its own test because two Set-Cookie headers are written on this one
// response: the resolve middleware clears the dead cookie, then the handler sets
// the new one. If they ever came out in the other order, login would appear to
// succeed and the guest would stay logged out.
func TestLoginWorksWhileHoldingAStaleCookie(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	// Every session revoked out from under the browser, which still has the cookie.
	_, err := app.Database.Write.Exec(`DELETE FROM session`)
	require.NoError(t, err)

	response := app.logIn("ABC234")
	require.Equal(t, http.StatusOK, response.Status)
	assert.Equal(t, household.ID, response.bootstrap().Household.ID)

	// The cookie the jar kept is the new one, so the next request is authenticated.
	me := app.get("/api/me")
	assert.Equal(t, http.StatusOK, me.Status)
	assert.Equal(t, household.ID, me.bootstrap().Household.ID)
}

// The shared-tablet case: somebody logs in with the wrong household's card, then
// with the right one. They must end up on the household the code names, and the
// session they no longer own must not linger for a year.
func TestLoggingInAgainReplacesTheExistingSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	first := seedHousehold(t, app.Database.Write, withCode("ABC234"), withDisplayName("Familie Müller"))
	second := seedHousehold(t, app.Database.Write, withCode("DEF567"), withDisplayName("Familie Schmidt"))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	require.Equal(t, first.ID, app.get("/api/me").bootstrap().Household.ID)

	response := app.logIn("DEF567")
	require.Equal(t, http.StatusOK, response.Status)

	assert.Equal(t, second.ID, response.bootstrap().Household.ID)
	assert.Equal(t, second.ID, app.get("/api/me").bootstrap().Household.ID)
	assert.Equal(t, 1, app.countSessions(), "the previous session must be revoked, not merely forgotten")
}
