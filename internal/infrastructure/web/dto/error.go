package dto

// ErrorResponse is the single error shape of the whole API.
//
// One shape for every failure means the frontend has exactly one error path, and
// every failure a guest can see carries an ID they can read out over the phone.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries a machine-readable code, a German message and the request ID.
type ErrorBody struct {
	// Code is a stable, English, snake_case identifier the frontend may branch on.
	Code string `json:"code"`
	// Message is German, informal ("du"), and safe to show a guest verbatim —
	// so it never contains internal detail such as SQL, Go error strings or paths.
	Message string `json:"message"`
	// RequestID is the same ID as the X-Request-Id header and the request's log
	// line. Repeated in the body because that is the part a guest can see, screenshot
	// and read aloud; a header is invisible to them.
	RequestID string `json:"request_id"`
	// Fields maps input name to a German message for that input, and is present
	// only for validation failures. Keys are the JSON field names of the request
	// body, so the frontend can render each message next to its own control.
	Fields map[string]string `json:"fields,omitempty"`
}
