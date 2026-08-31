package application

import "errors"

// ErrNotFound reports that the row a request addressed does not exist.
//
// Translated from persistence.ErrNotFound at the use case boundary so that a
// handler can branch on a miss without importing the persistence package — "no such
// household" is a fact about our data, not about the store that looked for it.
var ErrNotFound = errors.New("no such row")
