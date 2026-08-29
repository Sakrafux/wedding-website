package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/handler"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
	frontend "github.com/Sakrafux/wedding-website/web"
)

// NewRouter builds the application router: global middleware, robots.txt, the /api
// tree with its JSON fallbacks, and the embedded SPA on everything else.
//
// Later stories widen the parameter list with handler dependencies. main and the
// integration tests both construct the router through this one function, so the
// tests exercise the real middleware chain rather than an approximation of it.
func NewRouter(logger *httplog.Logger, database *configuration.Database, config configuration.Config) *chi.Mux {
	system := handler.NewSystem(database)

	router := chi.NewRouter()
	registerMiddleware(router, logger, config)

	router.Get("/robots.txt", system.Robots)

	router.Route("/api", func(api chi.Router) {
		api.Get("/health", system.Health)
		api.Get("/ready", system.Ready)

		api.NotFound(system.APINotFound)
		api.MethodNotAllowed(system.APIMethodNotAllowed)
	})

	// Registered last and as the catch-all, so /api keeps its own JSON 404: an
	// unknown API path must never fall through to the SPA and answer HTML.
	router.NotFound(newStaticHandler(frontend.Bundle()).ServeHTTP)

	return router
}

// registerMiddleware installs the global chain. Order matters: the request ID must
// exist before the logger records it, and the recoverer must sit inside the logger
// so a panic still produces a request log line.
func registerMiddleware(router *chi.Mux, logger *httplog.Logger, config configuration.Config) {
	// First in the chain, so that everything after it — routing, logging, the SPA
	// handler — sees the same path whether or not the reverse proxy stripped the
	// public prefix. A no-op when the app is served at the root.
	router.Use(middleware.StripPublicBasePath(config.PublicBasePath))

	router.Use(middleware.RequestID)

	// Outside the recoverer, so that the 500 an unrecovered panic produces still
	// carries the headers.
	router.Use(middleware.SecurityHeaders)

	// chi's middleware.RealIP is deliberately not used: it believes X-Forwarded-For
	// from any source, which would make the login rate limit trivially bypassable.
	// F1-B05 resolves the client IP against TRUSTED_PROXY_CIDRS instead.

	// httplog.Handler, not httplog.RequestLogger: the latter is a chain that bundles
	// chi's own RequestID and Recoverer, and both are replaced here — chi's RequestID
	// would overwrite our speakable ID with its "host/pid-000123" format, and chi's
	// Recoverer answers with an empty body instead of the error envelope.
	router.Use(httplog.Handler(logger))

	router.Use(middleware.Recoverer)
}
