// Package middleware holds the app-specific HTTP middleware: the request ID,
// security headers, panic recovery, household session resolution, the admin-only
// gate, login rate limiting and trusted-proxy client IP resolution.
//
// Request ID and panic recovery exist here rather than coming from chi because both
// of chi's versions fail a requirement of this app: its ID format cannot be read
// out over the phone, and its recoverer answers with an empty body instead of the
// error envelope. Everything else generic still comes from chi.
//
// Rejections are written through httpio so that a 401 from the session gate has
// the same shape as any handler's error.
//
// Authorization is enforced here and nowhere else in the request path: a hidden
// nav link in the frontend is not a security control.
package middleware
