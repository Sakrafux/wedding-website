package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// requestIDHeader is spelled out rather than taken from the middleware package: the
// header name is part of the API contract, and a test that imports the constant
// would follow a rename instead of failing on it.
const requestIDHeader = "X-Request-Id"

func TestUnknownRouteReturnsEnvelopeWithRequestID(t *testing.T) {
	server := newTestServer(t)

	response, body := getEnvelope(t, server.URL, "/api/does-not-exist")

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Equal(t, "not_found", body.Code)
	assert.NotEmpty(t, body.RequestID)
	assert.Equal(t, body.RequestID, response.Header.Get(requestIDHeader))
}

func TestSuccessResponseCarriesRequestIDHeader(t *testing.T) {
	server := newTestServer(t)

	response, err := http.Get(server.URL + "/api/health")
	require.NoError(t, err)
	defer response.Body.Close()

	assert.NotEmpty(t, response.Header.Get(requestIDHeader))
}

// TestPanicReturnsCleanEnvelope is the guard that matters most here: a panic must
// reach the guest as a German sentence, never as a stack trace or a Go type name.
func TestPanicReturnsCleanEnvelope(t *testing.T) {
	server := newTestServerWithRoutes(t, newTestDatabase(t), func(router chi.Router) {
		router.Get("/api/panic", func(http.ResponseWriter, *http.Request) {
			panic("deliberate test panic")
		})
	})

	response, body := getEnvelope(t, server.URL, "/api/panic")

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Equal(t, "internal_error", body.Code)
	assert.NotEmpty(t, body.RequestID)

	for _, leak := range []string{"deliberate test panic", "goroutine", "runtime.", ".go:", "wedding-website"} {
		assert.NotContains(t, body.Message, leak, "error message must not leak internals")
	}
}

func TestValidationErrorReturnsFieldsKeyedByInputName(t *testing.T) {
	server := newTestServerWithRoutes(t, newTestDatabase(t), func(router chi.Router) {
		// Stands in for a real endpoint until F1-B04 has one that validates a body.
		router.Get("/api/invalid", func(w http.ResponseWriter, r *http.Request) {
			httpio.RespondError(w, r, httpio.ValidationError{Fields: map[string]string{"code": "Der Code ist zu kurz."}})
		})
	})

	response, body := getEnvelope(t, server.URL, "/api/invalid")

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Equal(t, "validation_failed", body.Code)
	assert.Equal(t, map[string]string{"code": "Der Code ist zu kurz."}, body.Fields)
}

// errorBody mirrors dto.ErrorBody instead of importing it, so a rename in the DTO
// shows up here as a failing wire-format assertion rather than compiling silently.
type errorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Fields    map[string]string `json:"fields"`
}

func getEnvelope(t *testing.T, baseURL, path string) (*http.Response, errorBody) {
	t.Helper()

	response, err := http.Get(baseURL + path)
	require.NoError(t, err)
	defer response.Body.Close()

	require.True(t, strings.HasPrefix(response.Header.Get("Content-Type"), "application/json"))

	var envelope struct {
		Error errorBody `json:"error"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))

	return response, envelope.Error
}
