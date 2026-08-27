package httpio

import (
	"encoding/json"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
)

// WriteJSON serializes body with the given status code.
//
// The request is a parameter only so an encoding failure can be logged against
// its request ID. Such a failure cannot be reported to the client — the status
// line is already on the wire — so the response is left truncated and the log is
// the only trace. In practice it means a body that is not encodable, which is a
// programming error rather than a runtime condition.
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		httplog.LogEntry(r.Context()).Error("response encoding failed", "error", err)
	}
}

// writeErrorBody stamps the request ID onto an error body and writes it.
//
// Unexported on purpose: every error response goes through RespondError, so that the
// code, the status and the German message can only come from the one table in
// respond.go. A public writer taking a code and a message would let a handler invent
// either, which is how one endpoint ends up leaking a database string.
//
// The ID is read from chi's context key, which middleware.RequestID writes and
// httplog also reads — that shared key is what makes the body, the X-Request-Id
// header and the log line agree. It is empty when the middleware did not run, as
// in a test calling a handler directly; the envelope then simply has no ID, which
// beats failing the response over a support convenience.
func writeErrorBody(w http.ResponseWriter, r *http.Request, status int, body dto.ErrorBody) {
	body.RequestID = chimiddleware.GetReqID(r.Context())

	WriteJSON(w, r, status, dto.ErrorResponse{Error: body})
}
