// Assembly of the application: which stores exist, which use cases are built from
// them, and what the router is handed. Kept apart from main.go, which owns the
// process lifecycle — configuration, logging, signals, listening and draining.
//
// The two grow at different rates. Startup order is finished and rarely changes;
// this file gains a block per epic. Splitting them means the file you read to
// understand *when* things happen is not the file you edit to add *what* exists.
//
// Both are package main on purpose. cmd is the composition root and the only place
// allowed to know every layer at once — put this in package web and web would import
// persistence, at which point nothing but discipline stops a handler from building a
// store and running SQL.

package main

import (
	"log/slog"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/security"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web"
)

// wire builds every store and use case, and returns them as the router's
// dependencies.
//
// One function so that assembling the application is one thing to read and one
// place to extend: a new use case is a store, a constructor call and a field, all
// visible together. It deliberately does not open the database or start anything —
// run owns those lifetimes, because it is the function with the defers.
//
// The session store is returned alongside, because main's purge loop must sweep the
// same store the request path writes to. It is not a Dependencies field: the router
// reaches sessions through Auth, and a second door onto them invites a handler to
// use it.
func wire(config configuration.Config, database *configuration.Database, logger *slog.Logger) (web.Dependencies, *persistence.SessionStore) {
	sessions := persistence.NewSessionStore(database)
	householdStore := persistence.NewHouseholdStore(database)
	guestStore := persistence.NewGuestStore(database)
	auditStore := persistence.NewAuditStore(database)

	auth := application.NewAuth(
		sessions,
		householdStore,
		persistence.NewSettingStore(database),
		auditStore,
		security.AdminCredentials{User: config.AdminUser, Password: config.AdminPassword},
		logger,
	)

	households := application.NewHouseholds(householdStore, guestStore, sessions, auditStore, logger)

	return web.Dependencies{
		Config:     config,
		Database:   database,
		Auth:       auth,
		Households: households,
	}, sessions
}
