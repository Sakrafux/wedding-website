// Package domain holds the entities, enums and invariants of the wedding app:
// households, guests, RSVP scope, seating and budget rollups.
//
// It imports no other internal package, and nothing here knows about HTTP or SQL.
// Business rules are pure functions over plain structs so they can be tested
// without a server and without a database.
//
// All enum values are English. German exists only as display labels in the frontend.
package domain
