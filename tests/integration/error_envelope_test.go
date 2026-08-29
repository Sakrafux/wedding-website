package integration

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// requestIDHeader is spelled out rather than taken from the middleware package: the
// header name is part of the API contract, and a test that imports the constant
// would follow a rename instead of failing on it.
const requestIDHeader = "X-Request-Id"

func TestUnknownRouteReturnsEnvelopeWithRequestID(t *testing.T) {
	app := newTestApp(t)

	response := app.get("/api/does-not-exist")
	body := response.errorEnvelope()

	assert.Equal(t, http.StatusNotFound, response.Status)
	assert.Equal(t, "not_found", body.Code)
	assert.NotEmpty(t, body.RequestID)
	assert.Equal(t, body.RequestID, response.Header.Get(requestIDHeader))
	// An error envelope is guest-facing like any other response, and it is the one
	// place where a "code" field is legitimate — so it is worth pinning that the
	// leak detector passes it.
	response.assertNoLeak()
}

func TestSuccessResponseCarriesRequestIDHeader(t *testing.T) {
	app := newTestApp(t)

	response := app.get("/api/health")

	assert.NotEmpty(t, response.Header.Get(requestIDHeader))
}

// TestPanicReturnsCleanEnvelope is the guard that matters most here: a panic must
// reach the guest as a German sentence, never as a stack trace or a Go type name.
func TestPanicReturnsCleanEnvelope(t *testing.T) {
	app := newTestAppWithRoutes(t, func(router chi.Router) {
		router.Get("/api/panic", func(http.ResponseWriter, *http.Request) {
			panic("deliberate test panic")
		})
	})

	response := app.get("/api/panic")
	body := response.errorEnvelope()

	assert.Equal(t, http.StatusInternalServerError, response.Status)
	assert.Equal(t, "internal_error", body.Code)
	assert.NotEmpty(t, body.RequestID)

	for _, leak := range []string{"deliberate test panic", "goroutine", "runtime.", ".go:", "wedding-website"} {
		assert.NotContains(t, body.Message, leak, "error message must not leak internals")
	}
}

func TestValidationErrorReturnsFieldsKeyedByInputName(t *testing.T) {
	app := newTestAppWithRoutes(t, func(router chi.Router) {
		// Stands in for a real endpoint until F1-B04 has one that validates a body.
		router.Get("/api/invalid", func(w http.ResponseWriter, r *http.Request) {
			httpio.RespondError(w, r, httpio.ValidationError{Fields: map[string]string{"code": "Der Code ist zu kurz."}})
		})
	})

	response := app.get("/api/invalid")
	body := response.errorEnvelope()

	assert.Equal(t, http.StatusBadRequest, response.Status)
	assert.Equal(t, "validation_failed", body.Code)
	assert.Equal(t, map[string]string{"code": "Der Code ist zu kurz."}, body.Fields)
}
