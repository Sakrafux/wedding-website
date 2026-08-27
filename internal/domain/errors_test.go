package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

func TestErrorTextIsTheCodeAndTheCause(t *testing.T) {
	assert.Equal(t, "unknown_login_code", domain.NewError(domain.CodeUnknownLoginCode).Error())

	wrapped := domain.WrapError(domain.CodeUnknownLoginCode, errors.New("database is locked"))
	assert.Equal(t, "unknown_login_code: database is locked", wrapped.Error())
}

// TestCauseStaysReachable pins the Unwrap contract: infrastructure needs errors.Is
// on the cause — a retry decision on SQLITE_BUSY, say — without the domain having
// to expose the cause as a field.
func TestCauseStaysReachable(t *testing.T) {
	cause := errors.New("database is locked")

	assert.ErrorIs(t, domain.WrapError(domain.CodeUnknownLoginCode, cause), cause)
	assert.NoError(t, errors.Unwrap(domain.NewError(domain.CodeUnknownLoginCode)))
}

// TestCodeSurvivesWrapping is what RespondError depends on: the application layer
// may add context with fmt.Errorf and the code must still be found.
func TestCodeSurvivesWrapping(t *testing.T) {
	domainError, found := errors.AsType[domain.Error](errors.Join(errors.New("context"), domain.NewError(domain.CodeUnknownLoginCode)))

	assert.True(t, found)
	assert.Equal(t, domain.CodeUnknownLoginCode, domainError.Code)
}
