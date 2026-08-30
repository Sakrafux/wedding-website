package integration

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The test server is reached on the loopback address, so trusting it is what lets
// a test act as several different clients by setting X-Forwarded-For. In
// production the same mechanism trusts the container network and nothing else.
const testProxyCIDR = "127.0.0.0/8"

// Ten failures is the budget; the eleventh attempt is refused. The refusal is a
// normal error envelope, so the frontend renders it like any other failure.
func TestGuestLoginIsLimitedAfterTenFailures(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	for attempt := range 10 {
		response := app.logIn("ZZZZZZ")
		require.Equalf(t, http.StatusUnauthorized, response.Status, "attempt %d should be a normal rejection", attempt+1)
	}

	response := app.logIn("ZZZZZZ")

	assert.Equal(t, http.StatusTooManyRequests, response.Status)
	assert.Equal(t, "rate_limited", response.errorEnvelope().Code)
	assert.Contains(t, response.errorEnvelope().Message, "ruf uns einfach an")

	retryAfter, err := strconv.Atoi(response.Header.Get("Retry-After"))
	require.NoError(t, err, "Retry-After must be a number of seconds")
	assert.Positive(t, retryAfter)
	assert.LessOrEqual(t, retryAfter, 3600)
}

// Once the budget is spent, even the correct code is refused — the limiter runs
// before the handler. This is the cost of the design, and it is bounded: the
// window expires on its own, which the limiter's unit tests prove deterministically
// rather than by making this suite wait an hour.
func TestRateLimitedGuestIsRefusedEvenWithTheRightCode(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	for range 10 {
		require.Equal(t, http.StatusUnauthorized, app.logIn("ZZZZZZ").Status)
	}

	assert.Equal(t, http.StatusTooManyRequests, app.logIn("ABC234").Status)
}

// A household on a shared connection, or a family passing one phone around, must
// not spend the budget by logging in successfully.
func TestSuccessfulLoginsDoNotConsumeTheBudget(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	for attempt := range 20 {
		require.Equalf(t, http.StatusOK, app.logIn("ABC234").Status, "login %d should still be allowed", attempt+1)
	}
}

func TestRateLimitIsPerClientAddress(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, withTrustedProxies(testProxyCIDR))
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	for range 10 {
		require.Equal(t, http.StatusUnauthorized, app.logInFrom("ZZZZZZ", "203.0.113.7").Status)
	}
	require.Equal(t, http.StatusTooManyRequests, app.logInFrom("ZZZZZZ", "203.0.113.7").Status)

	// A different guest, unaffected — and able to log in.
	assert.Equal(t, http.StatusOK, app.logInFrom("ABC234", "203.0.113.99").Status)
}

// The whole point of resolving the address against TRUSTED_PROXY_CIDRS: without
// it, anybody could spend a fresh budget on every attempt by changing a header.
func TestForgedForwardedHeaderCannotEscapeTheLimit(t *testing.T) {
	t.Parallel()

	// No trusted proxies, so X-Forwarded-For is ignored entirely and every request
	// keys on the loopback peer regardless of what it claims.
	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	for attempt := range 10 {
		response := app.logInFrom("ZZZZZZ", "203.0.113."+strconv.Itoa(attempt))
		require.Equal(t, http.StatusUnauthorized, response.Status)
	}

	assert.Equal(t, http.StatusTooManyRequests, app.logInFrom("ZZZZZZ", "203.0.113.200").Status)
}

// Stricter, because the admin door is the only one where guessing pays.
func TestAdminLoginIsLimitedAfterFiveFailures(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	for attempt := range 5 {
		response := app.postJSON("/api/auth/admin/login", map[string]string{"user": "admin", "password": "wrong"})
		require.Equalf(t, http.StatusUnauthorized, response.Status, "attempt %d", attempt+1)
	}

	response := app.postJSON("/api/auth/admin/login", map[string]string{"user": "admin", "password": "wrong"})

	assert.Equal(t, http.StatusTooManyRequests, response.Status)
	assert.Equal(t, "rate_limited", response.errorEnvelope().Code)
}

// Two budgets, not one: a guest fumbling their code must not lock the admin out of
// their own site, and the admin's stricter limit must not leak onto the guests.
func TestGuestAndAdminBudgetsAreSeparate(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	for range 5 {
		require.Equal(t, http.StatusUnauthorized,
			app.postJSON("/api/auth/admin/login", map[string]string{"user": "admin", "password": "wrong"}).Status)
	}
	require.Equal(t, http.StatusTooManyRequests,
		app.postJSON("/api/auth/admin/login", map[string]string{"user": "admin", "password": "wrong"}).Status)

	assert.Equal(t, http.StatusOK, app.logIn("ABC234").Status, "the guest door is untouched")
}
