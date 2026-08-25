// Package middleware holds the app-specific HTTP middleware: household session
// resolution, the admin-only gate, login rate limiting and trusted-proxy client IP
// resolution.
//
// Generic concerns (request ID, panic recovery) come from chi's own middleware
// package. Only what encodes a decision of this app lives here.
//
// Rejections are written through httpio so that a 401 from the session gate has
// the same shape as any handler's error.
//
// Authorization is enforced here and nowhere else in the request path: a hidden
// nav link in the frontend is not a security control.
package middleware
