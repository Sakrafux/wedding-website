package middleware

import (
	"context"
	"math/rand/v2"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// RequestIDHeader is the response header carrying the request ID. Set on every
// response, not only on failures, so a guest who describes a page that "looks
// wrong" can still be tied to a log line.
const RequestIDHeader = "X-Request-Id"

// requestIDAlphabet is base32 without the glyphs that are misread when spoken or
// hand-copied: no 0/O, 1/I/L, U/V. Identical in purpose to the login-code alphabet
// (F1-B01), which defines its own — a shared constant would couple the printed
// invite format to a logging detail, and either may change alone.
const requestIDAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// requestIDLength of 7 gives 32^7 ≈ 3.4e10 values. Collisions do not matter: the
// ID only has to be unique among the handful of requests near one phone call, and
// every character is one more thing to read aloud.
const requestIDLength = 7

// RequestID assigns each request a short, speakable ID, exposes it in the
// X-Request-Id response header and stores it in the context.
//
// It replaces chi's own RequestID middleware, whose "hostname/pid-000123" format
// is unusable over the phone, but writes to chi's context key on purpose: httplog
// reads that key, so the log line and the error body carry the same ID without
// further wiring.
//
// An incoming X-Request-Id is deliberately ignored rather than adopted. The header
// is guest-controllable, so trusting it would let a client shape our log keys and
// break the format the support flow depends on.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()

		w.Header().Set(RequestIDHeader, requestID)

		ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID returns a random requestIDLength-character ID.
//
// math/rand rather than crypto/rand: the ID is a log correlation handle, never a
// secret and never an authorization token, so unpredictability buys nothing and
// the pseudo-random source is cheaper on every single request.
func newRequestID() string {
	id := make([]byte, requestIDLength)
	for i := range id {
		id[i] = requestIDAlphabet[rand.IntN(len(requestIDAlphabet))]
	}
	return string(id)
}
