// Package handler holds the HTTP request handlers, one file per resource.
//
// A handler parses the request, delegates to the application layer and formats the
// response through httpio. It contains no business rules and no SQL, and it never
// serializes a domain struct — every body is an explicit type from dto.
//
// Handlers live here rather than in web so that web stays what it is: the wiring
// point. Everything a handler needs is passed to its constructor; nothing reaches
// for a package-level dependency.
package handler
