package persistence

import "errors"

// ErrNotFound reports that a query matched no row.
//
// Stores translate database/sql's ErrNoRows into this before returning, so the
// application layer can branch on a miss without importing database/sql — "no such
// household" is a fact about our data, not about the driver we happen to use.
var ErrNotFound = errors.New("no matching row")
