package web

import (
	"github.com/go-chi/chi/v5"
	// Aliased: this file will also import our own middleware package.
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/handler"
)

// NewRouter builds the application router: global middleware, the /api tree and
// its JSON fallbacks.
//
// Later stories widen the parameter list with handler dependencies. main and the
// integration tests both construct the router through this one function, so the
// tests exercise the real middleware chain rather than an approximation of it.
func NewRouter(logger *httplog.Logger) *chi.Mux {
	router := chi.NewRouter()
	registerMiddleware(router, logger)

	router.Route("/api", func(api chi.Router) {
		api.Get("/health", handler.Health)

		api.NotFound(handler.APINotFound)
		api.MethodNotAllowed(handler.APIMethodNotAllowed)
	})

	return router
}

// registerMiddleware installs the global chain. Order matters: the request ID must
// exist before the logger records it, and the recoverer must sit inside the logger
// so a panic still produces a request log line.
func registerMiddleware(router *chi.Mux, logger *httplog.Logger) {
	router.Use(chimiddleware.RequestID)

	// chi's middleware.RealIP is deliberately not used: it believes X-Forwarded-For
	// from any source, which would make the login rate limit trivially bypassable.
	// F1-B05 resolves the client IP against TRUSTED_PROXY_CIDRS instead.

	router.Use(httplog.RequestLogger(logger))

	// Placeholder recovery so a panic cannot kill the process. E0-06 replaces this
	// with one that emits the JSON error envelope and the request ID.
	router.Use(chimiddleware.Recoverer)
}
