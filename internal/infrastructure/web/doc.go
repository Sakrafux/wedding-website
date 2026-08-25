// Package web is the wiring point of the HTTP adapter: it assembles the chi
// router from the global middleware chain, the handlers and — from E0-09 — the
// embedded SPA.
//
// The adapter is split so that each part has one job and the import graph stays
// acyclic:
//
//   - handler — request handlers, one file per resource
//   - middleware — session resolution, admin gate, rate limiting
//   - httpio — JSON and error-envelope writers, shared by the two above
//   - dto — explicit request and response types
//
// Nothing in this tree holds a business rule or a line of SQL, and no domain
// struct is ever serialized: response bodies are dto types, because
// household.code and admin_note must never reach a guest.
package web
