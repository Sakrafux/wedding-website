// Package application holds the use cases: it orchestrates domain rules and the
// persistence stores into the operations the API offers (log in, submit an RSVP,
// assign a seat, roll up a budget).
//
// It never imports the web package and contains no SQL. Stores are passed in as
// concrete types — there are no port interfaces, because there will be exactly one
// database and one web adapter (see specification/04-architecture.md).
package application
