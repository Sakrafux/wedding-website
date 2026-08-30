package httpio

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// maxRequestBodyBytes caps a JSON request body at 1 MiB — orders of magnitude more
// than the largest form this app has, and small enough that a malicious or looping
// client cannot make the process allocate its way out of memory. Photo uploads do
// not come through here; F10 sets its own, larger limit.
const maxRequestBodyBytes = 1 << 20

// jsonContentType is required on every request that carries a body.
const jsonContentType = "application/json"

// DecodeJSON reads the request body into target, or returns a ValidationError.
//
// Requiring Content-Type: application/json is a CSRF control, not pedantry. A
// cross-site HTML form can only send urlencoded, multipart or text/plain bodies, so
// this check — together with SameSite=Lax on the session cookie — is what keeps a
// third-party page from posting to the API on a guest's behalf. It matters most on
// login, which is the one mutation that needs no existing cookie and would
// otherwise let an attacker's page log a guest into the attacker's household.
//
// Unknown fields are rejected. The frontend is built into the same binary that
// serves it, so a field the server does not know is never version skew — it is a
// typo, and silently ignoring it would mean an answer a guest gave was quietly
// dropped.
//
// Every failure reports the same ValidationError with no per-field messages: at
// this point nothing has been parsed, so there is no field to blame. Endpoints add
// their own field rules on top.
func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	contentType := r.Header.Get("Content-Type")
	// The header may carry parameters, as in "application/json; charset=utf-8".
	if mediaType, _, _ := strings.Cut(contentType, ";"); !strings.EqualFold(strings.TrimSpace(mediaType), jsonContentType) {
		return ValidationError{}
	}

	// The ResponseWriter is passed so that an over-long body also closes the
	// connection rather than letting the client keep streaming into a request we
	// have already given up on.
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return ValidationError{}
	}

	// A second value in the body means the client sent something other than the one
	// object this endpoint accepts, and the part that was ignored may be the part
	// that mattered.
	if err := decoder.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		return ValidationError{}
	}
	return nil
}
