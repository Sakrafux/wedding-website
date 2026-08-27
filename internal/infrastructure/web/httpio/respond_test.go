package httpio_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

func TestRespondErrorMapsDomainCodeToStatusAndGermanMessage(t *testing.T) {
	response := serve(t, nil, func(w http.ResponseWriter, r *http.Request) {
		httpio.RespondError(w, r, domain.NewError(domain.CodeUnknownLoginCode))
	})

	body := decodeError(t, response)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, string(domain.CodeUnknownLoginCode), body.Error.Code)
	assert.Contains(t, body.Error.Message, "Einladung")
}

// TestRespondErrorFindsWrappedDomainError covers the application layer adding
// context with fmt.Errorf: the code must still be found down the chain.
func TestRespondErrorFindsWrappedDomainError(t *testing.T) {
	response := serve(t, nil, func(w http.ResponseWriter, r *http.Request) {
		wrapped := fmt.Errorf("resolving household: %w", domain.NewError(domain.CodeUnknownLoginCode))
		httpio.RespondError(w, r, wrapped)
	})

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, string(domain.CodeUnknownLoginCode), decodeError(t, response).Error.Code)
}

func TestRespondErrorKeysValidationMessagesByField(t *testing.T) {
	response := serve(t, nil, func(w http.ResponseWriter, r *http.Request) {
		httpio.RespondError(w, r, httpio.ValidationError{Fields: map[string]string{"code": "Der Code ist zu kurz."}})
	})

	body := decodeError(t, response)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_failed", body.Error.Code)
	assert.Equal(t, map[string]string{"code": "Der Code ist zu kurz."}, body.Error.Fields)
}

// TestTransportSentinelsAreMapped pins the statuses of the failures that never reach
// a business rule; they share the one table with the domain codes.
func TestTransportSentinelsAreMapped(t *testing.T) {
	for _, testCase := range []struct {
		err    error
		status int
		code   string
	}{
		{httpio.ErrNotFound, http.StatusNotFound, "not_found"},
		{httpio.ErrMethodNotAllowed, http.StatusMethodNotAllowed, "method_not_allowed"},
		{httpio.ErrNotReady, http.StatusServiceUnavailable, "not_ready"},
		{httpio.ErrInternal, http.StatusInternalServerError, "internal_error"},
	} {
		t.Run(testCase.code, func(t *testing.T) {
			response := serve(t, nil, func(w http.ResponseWriter, r *http.Request) {
				httpio.RespondError(w, r, testCase.err)
			})

			assert.Equal(t, testCase.status, response.Code)
			assert.Equal(t, testCase.code, decodeError(t, response).Error.Code)
		})
	}
}

// TestRespondErrorFailsClosedOnUnmappedCode is the reason the table is allowed to
// be incomplete: a code with no message must never improvise one.
func TestRespondErrorFailsClosedOnUnmappedCode(t *testing.T) {
	var logOutput bytes.Buffer
	response := serve(t, &logOutput, func(w http.ResponseWriter, r *http.Request) {
		httpio.RespondError(w, r, domain.NewError("no_such_mapping"))
	})

	body := decodeError(t, response)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "internal_error", body.Error.Code)
	assert.NotContains(t, response.Body.String(), "no_such_mapping")
	assert.Contains(t, logOutput.String(), "no_such_mapping")
}

// TestRespondErrorHidesNonDomainErrors is the leak guard: a driver error carries
// the database path, and its text must not reach the response under any code path.
func TestRespondErrorHidesNonDomainErrors(t *testing.T) {
	var logOutput bytes.Buffer
	response := serve(t, &logOutput, func(w http.ResponseWriter, r *http.Request) {
		httpio.RespondError(w, r, errors.New("no such table: household in /data/wedding.db"))
	})

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.NotContains(t, response.Body.String(), "household")
	assert.NotContains(t, response.Body.String(), "wedding.db")
	assert.Contains(t, logOutput.String(), "wedding.db")
}

// TestRespondErrorLogsTheWrappedCause proves the cause reaches the log while the
// mapped response is unchanged — the point of domain.WrapError.
func TestRespondErrorLogsTheWrappedCause(t *testing.T) {
	var logOutput bytes.Buffer
	response := serve(t, &logOutput, func(w http.ResponseWriter, r *http.Request) {
		cause := errors.New("database is locked")
		httpio.RespondError(w, r, domain.WrapError(domain.CodeUnknownLoginCode, cause))
	})

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.NotContains(t, response.Body.String(), "database is locked")
	assert.Contains(t, logOutput.String(), "database is locked")
}
