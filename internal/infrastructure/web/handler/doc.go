// Package handler holds the HTTP request handlers.
//
// A handler parses the request, delegates to the application layer and formats the
// response through httpio. It contains no business rules and no SQL, and it never
// serializes a domain struct — every body is an explicit type from dto.
//
// One file per resource group, each holding a struct, its constructor and its
// endpoints as methods — System here, later Auth, RSVP, Admin. Deliberately not one
// struct for the whole API: that would end up holding every store and service and
// let any endpoint reach any of them, whereas per-resource types keep "which
// handlers can touch the budget" answerable from the type. Endpoints with no
// dependencies are methods too, so construction and route registration have exactly
// one shape.
//
// Handlers live here rather than in web so that web stays what it is: the wiring
// point. Everything a handler needs is passed to its constructor; nothing reaches
// for a package-level dependency.
package handler
