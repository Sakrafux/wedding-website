package application

import (
	"errors"
	"fmt"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
)

// ErrNotFound reports that the row a request addressed does not exist.
//
// Translated from persistence.ErrNotFound at the use case boundary so that a
// handler can branch on a miss without importing the persistence package — "no such
// household" is a fact about our data, not about the store that looked for it.
var ErrNotFound = errors.New("no such row")

// TranslateNotFound maps the store's miss onto this layer's own sentinel, and
// passes everything else through untouched.
//
// It lives in the parent package rather than in one use case because every use case
// below needs it and none of them may import each other: a second copy would be a
// second chance for a 404 to surface as a 500.
func TranslateNotFound(err error) error {
	if errors.Is(err, persistence.ErrNotFound) {
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return err
}
