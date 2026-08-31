package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrNotFound reports that a query matched no row.
//
// Stores translate database/sql's ErrNoRows into this before returning, so the
// application layer can branch on a miss without importing database/sql — "no such
// household" is a fact about our data, not about the driver we happen to use.
var ErrNotFound = errors.New("no matching row")

// requireOneRow turns "the statement matched nothing" into ErrNotFound.
//
// Every UPDATE and DELETE in this package addresses one row by id, so zero affected
// rows means the row is gone — which the caller has to answer as a 404 rather than
// as a silent success. operation names the statement for the wrapped driver error.
func requireOneRow(result sql.Result, operation string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// isUniqueViolation reports whether err is SQLite rejecting a duplicate key.
//
// Matched on the message rather than on a driver error type: modernc.org/sqlite
// exposes its result codes only through its own error struct, and type-asserting
// against it would couple this package to the driver for a single string
// comparison. The one caller is login-code assignment, where the UNIQUE index on
// household.code is the sole authority on uniqueness and a rejected insert means
// "try another code".
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
