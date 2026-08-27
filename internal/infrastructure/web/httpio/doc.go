// Package httpio writes JSON responses and the API's single error envelope.
//
// It has exactly two entry points: WriteJSON for a success body, RespondError for
// every failure. There is deliberately no exported writer that takes a status, a
// code or a message — those three come only from the errorResponses table in
// respond.go, so that table is the complete list of what a guest can be shown and no
// handler can invent a fourth wording or leak a database string.
//
// It is its own package rather than a set of helpers inside web because the
// middleware also has to reject requests in the standard error shape, and web
// imports middleware to build the router — helpers living in web would make that
// an import cycle. Keeping the writers here means a 401 from the session gate and
// a 404 from a handler are the same shape by construction.
//
// It depends on dto and domain and nothing else in this project. domain because
// RespondError owns the mapping from domain error codes to statuses and German
// messages; the dependency goes this way only, and never back.
package httpio
