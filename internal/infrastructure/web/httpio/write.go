package httpio

import (
	"encoding/json"
	"net/http"

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

// WriteError answers with the API's error envelope. See dto.ErrorResponse.
//
// code is a stable English identifier the frontend may branch on; message is
// German and shown to the guest verbatim, so callers must keep internal detail
// such as SQL or file paths out of it.
func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	WriteJSON(w, r, status, dto.ErrorResponse{
		Error: dto.ErrorBody{Code: code, Message: message},
	})
}
