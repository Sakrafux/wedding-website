// Package httpio_test drives the writers through the real middleware chain rather
// than calling them with a bare context, because the guarantee under test — one ID
// in the header, the body and the log line — only exists once RequestID and httplog
// are wired together.
package httpio_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

// TestRequestIDIsTheSameInHeaderBodyAndLog is the whole point of the request ID: a
// guest reads the ID out of an error message, and it has to find the log line.
func TestRequestIDIsTheSameInHeaderBodyAndLog(t *testing.T) {
	var logOutput bytes.Buffer
	response := serve(t, &logOutput, func(w http.ResponseWriter, r *http.Request) {
		httpio.RespondError(w, r, httpio.ErrNotFound)
	})

	body := decodeError(t, response)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.NotEmpty(t, body.Error.RequestID)
	assert.Equal(t, body.Error.RequestID, response.Header().Get(middleware.RequestIDHeader))
	assert.Contains(t, logOutput.String(), body.Error.RequestID)
}

// TestRequestIDIsSetOnSuccessResponses covers the reason the header is not limited
// to failures: a guest reporting a page that merely looks wrong is still traceable.
func TestRequestIDIsSetOnSuccessResponses(t *testing.T) {
	response := serve(t, nil, func(w http.ResponseWriter, r *http.Request) {
		httpio.WriteJSON(w, r, http.StatusOK, dto.HealthResponse{Status: "ok"})
	})

	assert.Len(t, response.Header().Get(middleware.RequestIDHeader), 7)
}

// TestClientSuppliedRequestIDIsIgnored guards against a guest-controllable header
// shaping our log keys.
func TestClientSuppliedRequestIDIsIgnored(t *testing.T) {
	var logOutput bytes.Buffer
	response := serveRequest(t, &logOutput, requestWithHeader(middleware.RequestIDHeader, "injected-by-client"), func(w http.ResponseWriter, r *http.Request) {
		httpio.RespondError(w, r, httpio.ErrNotFound)
	})

	assert.NotEqual(t, "injected-by-client", response.Header().Get(middleware.RequestIDHeader))
	assert.NotContains(t, logOutput.String(), "injected-by-client")
	assert.NotContains(t, decodeError(t, response).Error.RequestID, "injected")
}

// TestFieldsAreOmittedWithoutValidationErrors pins the omitempty: a non-validation
// failure must not carry an empty fields object the frontend would have to special-case.
func TestFieldsAreOmittedWithoutValidationErrors(t *testing.T) {
	response := serve(t, nil, func(w http.ResponseWriter, r *http.Request) {
		httpio.RespondError(w, r, httpio.ErrNotFound)
	})

	assert.NotContains(t, response.Body.String(), "fields")
}

// serve runs handler behind the production middleware chain. logOutput may be nil
// when the test does not inspect the log.
func serve(t *testing.T, logOutput *bytes.Buffer, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	return serveRequest(t, logOutput, httptest.NewRequest(http.MethodGet, "/probe", nil), handler)
}

func serveRequest(t *testing.T, logOutput *bytes.Buffer, request *http.Request, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	// Writer discards unless the test asked for the output. Level Info because the
	// per-request line httplog emits — the one carrying the request ID — is at Info.
	var writer *bytes.Buffer = logOutput
	if writer == nil {
		writer = &bytes.Buffer{}
	}
	logger := &httplog.Logger{
		Logger:  slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})),
		Options: httplog.Options{Concise: true},
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(httplog.Handler(logger))
	router.Use(middleware.Recoverer)
	router.Get("/probe", handler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

func requestWithHeader(name, value string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/probe", nil)
	request.Header.Set(name, value)
	return request
}

func decodeError(t *testing.T, response *httptest.ResponseRecorder) dto.ErrorResponse {
	t.Helper()

	var body dto.ErrorResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))

	return body
}
