package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A fixed instant: every assertion here is about elapsed time, and a test that
// consulted the real clock would be a test that occasionally fails at midnight.
var limiterNow = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

func TestRateLimiterAllowsUpToTheLimit(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(3, time.Hour)

	for attempt := range 3 {
		_, isLimited := limiter.RetryAfter("10.0.0.1", limiterNow)
		require.Falsef(t, isLimited, "attempt %d should be allowed", attempt+1)

		limiter.RecordFailure("10.0.0.1", limiterNow)
	}

	retryAfter, isLimited := limiter.RetryAfter("10.0.0.1", limiterNow)
	assert.True(t, isLimited)
	assert.Equal(t, time.Hour, retryAfter)
}

// A limit, not a lockout. This is the assertion the threat model cares about most:
// there must be no path that keeps a household out permanently.
func TestRateLimiterWindowExpiresOnItsOwn(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(2, time.Hour)
	limiter.RecordFailure("10.0.0.1", limiterNow)
	limiter.RecordFailure("10.0.0.1", limiterNow)

	_, isLimited := limiter.RetryAfter("10.0.0.1", limiterNow.Add(59*time.Minute))
	require.True(t, isLimited, "still inside the window")

	_, isLimited = limiter.RetryAfter("10.0.0.1", limiterNow.Add(time.Hour+time.Second))
	assert.False(t, isLimited, "the window must expire without anyone intervening")
}

// The window slides: the wait is until the *oldest* attempt ages out, so somebody
// who spread their attempts over an hour is let back in gradually.
func TestRateLimiterReportsTheWaitUntilTheOldestAttemptAgesOut(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(2, time.Hour)
	limiter.RecordFailure("10.0.0.1", limiterNow)
	limiter.RecordFailure("10.0.0.1", limiterNow.Add(30*time.Minute))

	retryAfter, isLimited := limiter.RetryAfter("10.0.0.1", limiterNow.Add(45*time.Minute))

	require.True(t, isLimited)
	assert.Equal(t, 15*time.Minute, retryAfter)
}

// Retrying while over budget must not push the window forward, or a client
// hammering the endpoint would never be let back in — the lockout the threat model
// forbids, arrived at by accident.
func TestRateLimiterDoesNotExtendTheWaitOnFurtherAttempts(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(2, time.Hour)
	limiter.RecordFailure("10.0.0.1", limiterNow)
	limiter.RecordFailure("10.0.0.1", limiterNow)

	for minute := range 30 {
		limiter.RecordFailure("10.0.0.1", limiterNow.Add(time.Duration(minute)*time.Minute))
	}

	_, isLimited := limiter.RetryAfter("10.0.0.1", limiterNow.Add(time.Hour+time.Second))
	assert.False(t, isLimited)
}

func TestRateLimiterKeysAreIndependent(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(1, time.Hour)
	limiter.RecordFailure("10.0.0.1", limiterNow)

	_, isLimited := limiter.RetryAfter("10.0.0.1", limiterNow)
	assert.True(t, isLimited)

	_, isLimited = limiter.RetryAfter("10.0.0.2", limiterNow)
	assert.False(t, isLimited, "one guest's fumbling must not affect another's")
}

// A wait under a second would otherwise be reported as Retry-After: 0, which
// invites an immediate retry that is certain to fail.
func TestRateLimiterNeverReportsAZeroWait(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(1, time.Hour)
	limiter.RecordFailure("10.0.0.1", limiterNow)

	retryAfter, isLimited := limiter.RetryAfter("10.0.0.1", limiterNow.Add(time.Hour-100*time.Millisecond))

	require.True(t, isLimited)
	assert.Equal(t, time.Second, retryAfter)
}

// Keys from a stream of changing addresses must not accumulate: the map is the
// only thing here that could grow without bound.
func TestRateLimiterEvictsIdleKeys(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(5, time.Hour)
	for address := range 100 {
		limiter.RecordFailure(string(rune(address)), limiterNow)
	}
	require.Len(t, limiter.failures, 100)

	// One attempt well after the window, which is what triggers the sweep.
	limiter.RecordFailure("10.0.0.1", limiterNow.Add(2*time.Hour))

	assert.Len(t, limiter.failures, 1, "only the live key should remain")
}
