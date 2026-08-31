package web

import (
	"github.com/Sakrafux/wedding-website/internal/application/auth"
	"github.com/Sakrafux/wedding-website/internal/application/exports"
	"github.com/Sakrafux/wedding-website/internal/application/households"
	"github.com/Sakrafux/wedding-website/internal/application/rsvp"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// Dependencies is everything the router needs from outside itself, named rather
// than positional.
//
// It exists so that adding a use case is one field here and one line at the call
// site, instead of a fifth and sixth anonymous argument to NewRouter whose order
// nothing but the compiler remembers. Handlers and middleware are still built
// inside NewRouter from these — what a route needs is the router's business; what
// the application can do is not.
//
// A plain struct on purpose: no interfaces, no container, no DI framework. There is
// exactly one implementation of each field and exactly one process assembling them,
// so anything more would be indirection with no second case to justify it.
type Dependencies struct {
	Config configuration.Config
	// Database is here for the readiness probe, which pings both pools. It is not a
	// licence for handlers to query: SQL lives in persistence, and the stores reach
	// the router only through the use cases below.
	Database *configuration.Database
	Auth     *auth.UseCase
	// Households is the admin guest list: households, their members, and login code
	// reissue. Admin-only, and mounted only under /api/admin.
	Households *households.UseCase
	// Exports serves the two admin CSV downloads. Admin-only: codes.csv is the whole
	// login-code list.
	Exports *exports.UseCase
	// RSVP serves both the guests' own answer and the admin's answer on their behalf.
	// The one use case behind two pairs of routes — see F3-B06.
	RSVP *rsvp.UseCase
}
