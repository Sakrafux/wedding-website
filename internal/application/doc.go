// Package application is the root of the use case layer: it holds what every use
// case below shares, which today is the "no such row" sentinel and its translation
// from persistence.
//
// The use cases themselves live one level down, one package each — application/auth,
// application/households, application/exports, application/rsvp. They were split out
// when the second use case arrived (F3-B02): a single package meant every type
// carried its topic in its own name (application.NewHouseholds, application.Auth),
// and the package name was doing none of the work. Each subpackage exposes one
// UseCase type, so a call site reads rsvp.Load rather than application.LoadRSVP.
//
// Nothing here imports the web package and nothing here contains SQL. Stores are
// passed in as concrete types — there are no port interfaces, because there will be
// exactly one database and one web adapter (see specification/04-architecture.md).
// Subpackages do not import each other: a use case that needs another one's work is
// a use case that has not been drawn correctly.
package application
