package dto

// ErrorResponse is the single error shape of the whole API.
//
// One shape for every failure means the frontend has exactly one error path.
// E0-06 grows the mapping from domain errors into this type; for now only routing
// failures produce it.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries a machine-readable code and a German message.
type ErrorBody struct {
	// Code is a stable, English, snake_case identifier the frontend may branch on.
	Code string `json:"code"`
	// Message is German, informal ("du"), and safe to show a guest verbatim —
	// so it never contains internal detail such as SQL or file paths.
	Message string `json:"message"`
}
