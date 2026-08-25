// Package httpio writes JSON responses and the API's single error envelope.
//
// It is its own package rather than a set of helpers inside web because the
// middleware also has to reject requests in the standard error shape, and web
// imports middleware to build the router — helpers living in web would make that
// an import cycle. Keeping the writers here means a 401 from the session gate and
// a 404 from a handler are the same shape by construction.
//
// It depends on dto and nothing else in this project.
package httpio
