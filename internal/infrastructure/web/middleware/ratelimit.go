package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// Login attempt budgets, per client IP per hour.
//
// Only *failures* count. A household on a shared connection, or a family passing
// one phone around, must not spend the budget by logging in successfully.
//
// The guest limit is generous because the thing it guards is already hopeless:
// with 32^6 codes and ~60 in use, a blind guess hits with probability 5.6e-8, so
// ten tries an hour is centuries per IP. It exists to bound the pointless, not to
// stop a plausible attack. The admin limit is stricter because that door is the one
// where guessing pays, and the person behind it knows their own password.
const (
	guestLoginFailureLimit = 10
	adminLoginFailureLimit = 5
	loginFailureWindow     = time.Hour
)

// evictionInterval is how often idle keys are dropped. Tied to the window: a key
// whose attempts have all expired carries no information, and sweeping more often
// than the window would just be work.
const evictionInterval = loginFailureWindow

// RateLimiter counts recent failures per key in a sliding window.
//
// In memory, with no Redis and no table. One process, ~60 households, and losing
// the counts on restart is not merely acceptable but faintly desirable: a restart
// should not leave a guest who mistyped their code still locked out of retrying.
type RateLimiter struct {
	limit  int
	window time.Duration

	mutex sync.Mutex
	// failures holds the timestamps of recent failures per key. Bounded by limit
	// per key, so the slice never needs to be anything cleverer.
	failures map[string][]time.Time
	// lastEviction drives the sweep. Doing it here, on a request, rather than in a
	// goroutine keeps the whole limiter free of lifecycle: nothing to start and
	// nothing to stop.
	lastEviction time.Time
}

// NewGuestLoginLimiter and NewAdminLoginLimiter keep the two budgets and their
// reasoning in this file, so the router wires a named policy rather than a pair of
// numbers whose justification lives somewhere else.
func NewGuestLoginLimiter() *RateLimiter {
	return NewRateLimiter(guestLoginFailureLimit, loginFailureWindow)
}

func NewAdminLoginLimiter() *RateLimiter {
	return NewRateLimiter(adminLoginFailureLimit, loginFailureWindow)
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:        limit,
		window:       window,
		failures:     make(map[string][]time.Time),
		lastEviction: time.Now(),
	}
}

// RetryAfter reports how long the key must wait, and false when it may proceed.
//
// The wait is until the oldest failure in the window ages out, so the limit always
// expires on its own. There is no code path that blocks a key permanently, and per
// the threat model that is the point: locking out a confused seventy-five-year-old
// is a worse outcome than the attack being prevented.
func (limiter *RateLimiter) RetryAfter(key string, now time.Time) (time.Duration, bool) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	recent := limiter.pruned(key, now)
	if len(recent) < limiter.limit {
		return 0, false
	}

	// Rounded up, so a wait of 100ms is reported as 1 second rather than 0 — a
	// Retry-After of 0 invites an immediate retry that is certain to fail. Only
	// a remainder rounds up; a whole number of seconds is already right.
	wait := recent[0].Add(limiter.window).Sub(now)
	if rounded := wait.Truncate(time.Second); rounded < wait {
		return rounded + time.Second, true
	}
	return wait, true
}

// RecordFailure charges one attempt against the key.
func (limiter *RateLimiter) RecordFailure(key string, now time.Time) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	limiter.evict(now)

	recent := limiter.pruned(key, now)
	// Capped at the limit: once a key is over budget, further attempts do not push
	// the window forward, so the wait cannot be extended indefinitely by retrying.
	// Without this a client hammering the endpoint would never be let back in.
	if len(recent) < limiter.limit {
		limiter.failures[key] = append(recent, now)
	}
}

// pruned drops the key's failures that have aged out of the window, storing the
// result so the trimming is not repeated on the next call.
func (limiter *RateLimiter) pruned(key string, now time.Time) []time.Time {
	cutoff := now.Add(-limiter.window)

	recent := limiter.failures[key]
	for len(recent) > 0 && recent[0].Before(cutoff) {
		recent = recent[1:]
	}

	if len(recent) == 0 {
		delete(limiter.failures, key)
	} else {
		limiter.failures[key] = recent
	}
	return recent
}

// evict drops keys whose failures have all aged out, so a stream of requests from
// changing addresses cannot grow the map without bound. Called under the lock.
func (limiter *RateLimiter) evict(now time.Time) {
	if now.Sub(limiter.lastEviction) < evictionInterval {
		return
	}
	limiter.lastEviction = now

	for key := range limiter.failures {
		limiter.pruned(key, now)
	}
}

// LimitLoginFailures refuses a caller who has spent their budget, and charges the
// ones whose attempt comes back unauthorized.
//
// Failure is read off the response status rather than reported by the handler.
// That keeps the endpoint itself unaware it is being limited, and it cannot forget
// to report: any 401 out of a login endpoint is a failed attempt by definition.
func (limiter *RateLimiter) LimitLoginFailures(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := ClientIPFromContext(r.Context())

		if retryAfter, isLimited := limiter.RetryAfter(key, time.Now()); isLimited {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			httpio.RespondError(w, r, domain.NewError(domain.CodeRateLimited))
			return
		}

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		if recorder.status == http.StatusUnauthorized {
			limiter.RecordFailure(key, time.Now())
		}
	})
}

// statusRecorder remembers the status a handler wrote.
//
// Deliberately minimal, and not chi's WrapResponseWriter: the only thing needed
// here is the status code, and none of the endpoints behind this middleware
// stream, flush or hijack.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}
