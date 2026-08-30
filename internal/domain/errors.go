package domain

// ErrorCode identifies a failure the API reports. Values are stable, English and
// snake_case, because they are part of the wire contract: the frontend branches on
// them, so renaming one is a breaking change.
type ErrorCode string

// Every code the API can return is declared here rather than next to the rule that
// raises it. One list is worth the small break with "keep it near the code": it can
// be read against the message table in httpio to see every failure a guest can
// meet, and a duplicate code is visible instead of hiding in another file.
const (
	// CodeUnknownLoginCode is a login attempt with a code no household has. A
	// malformed code reports the same value on purpose; see ErrMalformedCode.
	CodeUnknownLoginCode ErrorCode = "unknown_login_code"
	// CodeUnauthenticated is a request to a guarded route with no usable session:
	// no cookie, an expired one, or — on an admin route — a household session. All
	// three are one code, because the caller's next step is the same in each case
	// and distinguishing them only tells an unauthenticated client what exists.
	CodeUnauthenticated ErrorCode = "unauthenticated"
	// CodeRateLimited is a login endpoint refusing a caller who has spent their
	// failure budget. Always temporary: there is no lockout in this application.
	CodeRateLimited ErrorCode = "rate_limited"
	// CodeInvalidCredentials is a failed admin login. One code for a wrong
	// username and a wrong password alike — the distinction would confirm a valid
	// username to whoever is guessing.
	CodeInvalidCredentials ErrorCode = "invalid_credentials"
)

// Error is a failure a caller is expected to handle, identified by its code.
//
// It deliberately carries no user-facing text and no HTTP status. The German
// message and the status are a web concern, mapped once in httpio — a domain that
// carried German prose would put UI copy behind the business rules, and a domain
// error whose Message is echoed to the client is how a login code or an SQL string
// ends up in a response body.
//
// A value type, not a pointer: a domain error is data, and comparing or copying one
// should not depend on identity.
type Error struct {
	// Code is what the client sees and what httpio maps to a status and a message.
	Code ErrorCode
	// cause is the underlying failure, unexported so it cannot be serialized by
	// accident. It reaches the log and never the response.
	cause error
}

// NewError returns an Error for an expected condition with no underlying cause —
// a rule that says no, such as an unknown login code or a passed RSVP deadline.
func NewError(code ErrorCode) error {
	return Error{Code: code}
}

// WrapError returns an Error carrying cause for the log.
//
// Reserve it for a failure that is genuinely unexpected at this code path. An
// expected condition should carry no cause: httpio logs a wrapped cause as a
// warning, so wrapping the routine "no such row" behind an unknown login code
// would turn every mistyped code into a log warning.
func WrapError(code ErrorCode, cause error) error {
	return Error{Code: code, cause: cause}
}

// Error implements error. The text is for logs only, never for a response.
func (e Error) Error() string {
	if e.cause == nil {
		return string(e.Code)
	}
	return string(e.Code) + ": " + e.cause.Error()
}

// Unwrap exposes the cause to errors.Is and errors.As.
func (e Error) Unwrap() error {
	return e.cause
}
