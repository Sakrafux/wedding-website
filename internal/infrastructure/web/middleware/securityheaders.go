package middleware

import "net/http"

// contentSecurityPolicy is deliberately strict because it costs nothing here:
// everything is served from one origin — no CDN, no external fonts, no analytics —
// so there is no legitimate source to allow beyond 'self'.
//
// `img-src` additionally allows `data:` for the small inline SVG icons a bundler
// emits. There is no 'unsafe-inline' and no 'unsafe-eval', and neither may be added:
// a policy loosened once, under deploy pressure, is never tightened again. If the
// frontend needs one of them, the frontend is what changes.
//
// `frame-ancestors 'none'` forbids embedding, `base-uri 'none'` stops an injected
// <base> from re-pointing every relative URL, and `form-action 'self'` keeps a form
// from posting guest answers to another origin.
const contentSecurityPolicy = "default-src 'self'; " +
	"img-src 'self' data:; " +
	"style-src 'self'; " +
	"font-src 'self'; " +
	"script-src 'self'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'none'; " +
	"form-action 'self'"

// securityHeaders is the full set applied to every response, API and SPA alike.
//
// HSTS is absent on purpose: TLS terminates at the reverse proxy, the Go process
// speaks plain HTTP, and Strict-Transport-Security is the proxy's to set.
var securityHeaders = map[string]string{
	"Content-Security-Policy": contentSecurityPolicy,

	// Stops a browser from guessing a content type — an uploaded photo must never be
	// sniffed into script.
	"X-Content-Type-Options": "nosniff",

	// Nothing external is ever linked, so a referrer has nobody to inform and could
	// only leak a path.
	"Referrer-Policy": "no-referrer",

	// Redundant with frame-ancestors for current browsers, kept for the older ones.
	"X-Frame-Options": "DENY",

	// No feature of this app needs any of these.
	"Permissions-Policy": "geolocation=(), microphone=(), camera=()",

	// "Nothing is indexed" is a product principle. The header covers responses a
	// crawler reaches without reading robots.txt, and outranks it — robots.txt asks
	// a crawler not to fetch, this tells it not to publish what it did fetch.
	"X-Robots-Tag": "noindex, nofollow",
}

// SecurityHeaders sets the browser hardening headers on every response.
//
// The headers are written before the handler runs, since a handler that calls
// WriteHeader flushes them and any later mutation is silently lost. It sets rather
// than appends: a duplicated CSP is intersected by the browser, which is how a
// header set in two places becomes a page that fails to load.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		for name, value := range securityHeaders {
			header.Set(name, value)
		}

		next.ServeHTTP(w, r)
	})
}
