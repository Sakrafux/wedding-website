package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/application"
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
func NewRouter(logger *httplog.Logger, config configuration.Config, database *configuration.Database, auth *application.Auth) *chi.Mux {
	system := handler.NewSystem(database)
	authHandler := handler.NewAuth(auth, config.SessionCookieSecure)
	sessions := middleware.NewSessionGate(auth, config.SessionCookieSecure)

	// One limiter per endpoint, not one shared: a guest fumbling their code must
	// never consume the admin's budget, or the other way round.
	guestLoginLimiter := middleware.NewGuestLoginLimiter()
	adminLoginLimiter := middleware.NewAdminLoginLimiter()

	router := chi.NewRouter()
	registerMiddleware(router, logger)

	router.Get("/robots.txt", system.Robots)

	router.Route("/api", func(api chi.Router) {
		// Before anything that wants to know who is calling: the client address is
		// the rate limiter's key and the audit trail's record, and resolving it in
		// one place is what keeps the trusted-proxy rule from being bypassed by a
		// handler that reads the header itself.
		api.Use(middleware.ClientIP(config.TrustedProxyCIDRs))

		// Resolving the session for the whole tree, including the probes and the
		// login endpoints, is what lets an unauthenticated request be a normal
		// request rather than a special case. Nothing here refuses anybody; the
		// Require gates below do that.
		api.Use(sessions.Resolve)

		api.Get("/health", system.Health)
		api.Get("/ready", system.Ready)

		api.With(guestLoginLimiter.LimitLoginFailures).Post("/auth/login", authHandler.LogIn)
		api.With(adminLoginLimiter.LimitLoginFailures).Post("/auth/admin/login", authHandler.AdminLogIn)
		api.Post("/auth/logout", authHandler.LogOut)

		// Everything a logged-in household may reach.
		api.Group(func(household chi.Router) {
			household.Use(middleware.RequireHousehold)

			household.Get("/me", authHandler.Me)
		})

		// The admin subtree is mounted with its gate before it has a single route,
		// and deliberately so: this is where every admin-only rule in the
		// application rests, budget above all, and a subtree that is created at the
		// same moment as the first endpoint under it is a subtree somebody can
		// create without one. Anything under /api/admin is therefore refused to a
		// household session already; F5, F6 and F8 add the routes behind it.
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.RequireAdmin)

			// The catch-all is what makes the gate above real, and removing it
			// would be silent: chi builds a sub-router's middleware chain only when
			// a route is registered on it, and a sub-router with none serves its
			// NotFound handler directly — middleware and all — so an empty guarded
			// subtree answers 404 to everyone instead of 401 to strangers. Keep at
			// least one route here.
			admin.Handle("/*", http.HandlerFunc(system.APINotFound))
		})

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
func registerMiddleware(router *chi.Mux, logger *httplog.Logger) {
	router.Use(middleware.RequestID)

	// Outside the recoverer, so that the 500 an unrecovered panic produces still
	// carries the headers.
	router.Use(middleware.SecurityHeaders)

	// chi's middleware.RealIP is deliberately not used: it believes X-Forwarded-For
	// from any source, which would make the login rate limit trivially bypassable.
	// middleware.ClientIP, mounted on the /api tree, resolves the address against
	// TRUSTED_PROXY_CIDRS instead.

	// httplog.Handler, not httplog.RequestLogger: the latter is a chain that bundles
	// chi's own RequestID and Recoverer, and both are replaced here — chi's RequestID
	// would overwrite our speakable ID with its "host/pid-000123" format, and chi's
	// Recoverer answers with an empty body instead of the error envelope.
	router.Use(httplog.Handler(logger))

	router.Use(middleware.Recoverer)
}
