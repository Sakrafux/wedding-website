package httpio

import (
	"errors"
	"net/http"

	"github.com/go-chi/httplog/v2"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
)

// Transport-level codes: a request that never reached a business rule.
//
// They are declared here rather than in domain's code list because routing and
// readiness are not business facts — the domain has no opinion about a mistyped URL.
// The type is still domain.ErrorCode so that every failure in the app travels as one
// kind of value, and so errorResponses below stays the single list of everything a
// guest can be shown.
const (
	codeNotFound         domain.ErrorCode = "not_found"
	codeMethodNotAllowed domain.ErrorCode = "method_not_allowed"
	codeNotReady         domain.ErrorCode = "not_ready"
	codeValidationFailed domain.ErrorCode = "validation_failed"
	codeInternal         domain.ErrorCode = "internal_error"
)

// Sentinels for the transport-level failures, so a caller reports one by passing a
// value to RespondError instead of choosing a status and a message of its own.
var (
	// ErrNotFound is an unmatched path under /api.
	ErrNotFound = domain.NewError(codeNotFound)
	// ErrMethodNotAllowed is a known /api path called with the wrong method.
	ErrMethodNotAllowed = domain.NewError(codeMethodNotAllowed)
	// ErrNotReady is a dependency the process cannot reach; see the readiness probe.
	ErrNotReady = domain.NewError(codeNotReady)
	// ErrInternal is for a caller that has already logged the cause itself — panic
	// recovery, in particular. Any other error reaching RespondError produces the
	// same response anyway, so reach for this one only to avoid a duplicate log line.
	ErrInternal = domain.NewError(codeInternal)
)

// ValidationError reports rejected inputs, keyed by the JSON field name of the
// request body so the frontend can render each message next to its own control.
//
// It is a web type, not a domain one: the field names it carries are the shape of a
// request body, which the domain neither knows nor should. DecodeJSON reports a body
// that is not usable JSON through it with no fields at all; Validate in validate.go
// fills the fields in from validator/v10.
type ValidationError struct {
	Fields map[string]string
}

// Error implements error. For logs only, never for a response.
func (e ValidationError) Error() string {
	return string(codeValidationFailed)
}

// errorResponse is the HTTP face of one domain.ErrorCode.
type errorResponse struct {
	status int
	// message is German, informal ("du") and shown to the guest verbatim.
	message string
}

// errorResponses maps every error code the API can return to its status and German
// message.
//
// This table is the audit list for the whole API: read it top to bottom and you have
// seen every sentence a guest can be shown, which is not true of messages scattered
// across handlers. A code missing from here fails closed to a generic 500 rather
// than inventing a message, so forgetting an entry is a visible bug and never a leak.
//
// Statuses live here and not in the domain because "unknown login code" is a
// business fact, while "401" is a decision about a transport the domain is not
// allowed to know about.
var errorResponses = map[domain.ErrorCode]errorResponse{
	codeNotFound: {
		http.StatusNotFound,
		"Diese Adresse gibt es nicht.",
	},
	codeMethodNotAllowed: {
		http.StatusMethodNotAllowed,
		"Diese Anfrage ist hier nicht erlaubt.",
	},
	codeNotReady: {
		http.StatusServiceUnavailable,
		"Der Dienst ist gerade nicht bereit. Bitte versuche es in einem Moment noch einmal.",
	},
	codeValidationFailed: {
		http.StatusBadRequest,
		// Stays general: the detail belongs at the input, not at the top of the form.
		// 400 rather than 422 because the frontend decides how to render from the
		// presence of fields, so a second "invalid input" status consumes nothing.
		"Bitte prüfe die markierten Felder.",
	},
	codeInternal: {
		http.StatusInternalServerError,
		// Deliberately vague: anything more specific would leak what broke, and there
		// is nothing a guest could do with the detail.
		"Da ist etwas schiefgegangen. Bitte versuche es später noch einmal.",
	},
	domain.CodeUnknownLoginCode: {
		http.StatusUnauthorized,
		// Phrased as a typo rather than a rejection: the overwhelmingly likely cause
		// is a mistyped character, not someone without an invitation. The sentence
		// about capitalisation is there because it is the first thing a guest
		// retries, and retrying the same code in capitals wastes an attempt against
		// the rate limit for nothing.
		"Diesen Code kennen wir nicht. Schau bitte noch mal auf deine Karte — Groß- und Kleinschreibung ist egal.",
	},
	domain.CodeInvalidCredentials: {
		http.StatusUnauthorized,
		// Says nothing about which half was wrong, and nothing about whether the
		// username exists. The only person who sees this is the one who set the
		// password, so there is no usability cost to being terse.
		"Anmeldung fehlgeschlagen.",
	},
	domain.CodeRateLimited: {
		http.StatusTooManyRequests,
		// Ends with a way out that does not involve the website. The person most
		// likely to see this sentence is a guest who has mistyped the same code ten
		// times, not an attacker, and they need to know the evening is not lost.
		"Zu viele Versuche. Bitte warte ein paar Minuten und probier es dann noch einmal. Wenn es weiter nicht klappt, ruf uns einfach an.",
	},
	domain.CodeRSVPClosed: {
		http.StatusConflict,
		// Named as a state, with the way out stated: the person reading this is a guest
		// whose plans changed, and the phone call is the actual remedy. 409 rather than
		// 403 — nothing is wrong with who they are.
		"Die Rückmeldefrist ist vorbei. Wenn sich etwas geändert hat, ruf uns bitte kurz an.",
	},
	domain.CodeMemberSetMismatch: {
		http.StatusConflict,
		// The stale-tab case: the household list changed while this form was open. The
		// only honest move is to show the new list, so the sentence asks for exactly
		// that and does not pretend the answer was saved.
		"Die Liste der Personen hat sich geändert. Bitte lade die Seite neu.",
	},
	domain.CodePlusOneNotAllowed: {
		http.StatusConflict,
		// Reads as an offer rather than a refusal, because the answer is yes: we will
		// enter anybody, we just want to hear the headcount. Says nothing about which
		// rule was hit — the guest's next step is the same either way, and naming the
		// rule would describe a household this caller cannot see. 409 rather than 403:
		// nothing is wrong with who they are.
		"Weitere Personen tragen wir gern für euch ein — ruf uns bitte kurz an: +43 650 9408100.",
	},
	domain.CodeCannotRemoveMember: {
		http.StatusConflict,
		// Names the actual remedy: the guest's goal is to say somebody is not coming,
		// and the form can do exactly that.
		"Diese Person haben wir eingetragen. Wenn sie nicht kommt, wähl bitte «Kommt nicht» aus.",
	},
	domain.CodeUnauthenticated: {
		http.StatusUnauthorized,
		// One sentence for "no session", "expired session" and "wrong kind of
		// session" alike. The frontend knows which login screen to send the caller
		// to from the route; the message only has to say that they need one.
		"Bitte melde dich an.",
	},
}

// RespondError answers with the API's error envelope, and is the only way the app
// reports a failure — a handler that picks its own status and message is how one
// endpoint ends up leaking a database string.
//
// A domain.Error is translated through errorResponses, a ValidationError adds the
// per-field messages. Anything else — an unmapped code, a driver error, a bug — is
// logged and answered with the generic 500, because an error this layer does not
// recognise cannot be described to a guest and its text may carry SQL, a file path
// or a login code.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	if validationError, isValidationError := errors.AsType[ValidationError](err); isValidationError {
		writeMapped(w, r, codeValidationFailed, validationError.Fields)
		return
	}

	domainError, isDomainError := errors.AsType[domain.Error](err)
	if !isDomainError {
		httplog.LogEntry(r.Context()).Error("unhandled error", "error", err)
		writeMapped(w, r, codeInternal, nil)
		return
	}

	// A cause means the domain wrapped something it did not expect. The response is
	// still the mapped one — the guest's situation has not changed — but the cause
	// would otherwise be lost, since httplog records only status and path.
	if cause := domainError.Unwrap(); cause != nil {
		httplog.LogEntry(r.Context()).Warn("domain error with cause", "code", domainError.Code, "error", cause)
	}

	writeMapped(w, r, domainError.Code, nil)
}

// writeMapped writes the envelope for code, falling back to the generic 500 when the
// code has no entry in errorResponses.
//
// The fallback is what makes an incomplete table safe: the guest gets a sentence
// that is true but says nothing, and the log names the code so it is obvious what to
// add. It cannot recurse, because codeInternal is always mapped.
func writeMapped(w http.ResponseWriter, r *http.Request, code domain.ErrorCode, fields map[string]string) {
	response, isMapped := errorResponses[code]
	if !isMapped {
		httplog.LogEntry(r.Context()).Error("error code has no response mapping", "code", code)
		writeMapped(w, r, codeInternal, nil)
		return
	}

	writeErrorBody(w, r, response.status, dto.ErrorBody{
		Code:    string(code),
		Message: response.message,
		Fields:  fields,
	})
}
