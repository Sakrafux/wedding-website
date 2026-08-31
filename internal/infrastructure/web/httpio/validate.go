package httpio

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

// validate is the process-wide validator.
//
// One instance, because it caches the struct reflection it does — building a fresh
// one per request would redo that work on every call for no benefit. It is
// goroutine-safe by design of the library.
var validate = newValidator()

func newValidator() *validator.Validate {
	instance := validator.New(validator.WithRequiredStructEnabled())

	// Errors are reported under the JSON field name, not the Go one: the frontend
	// renders each message next to the control whose name it knows, and that name is
	// the one on the wire. Without this, `DisplayName` would arrive at a form that
	// only knows `display_name`.
	instance.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})

	return instance
}

// Validate checks a decoded request body against its `validate` tags and returns a
// ValidationError keyed by JSON field name.
//
// Every request DTO with per-field rules goes through here, so the German wording of
// "this field is required" is written once. A rule with no message below still
// produces a field entry — a generic sentence next to the right control beats a
// correct sentence at the top of the form.
func Validate(body any) error {
	// The leaf name, which is the whole key for a flat body: `display_name`, `age`.
	return validateKeyedBy(body, func(fieldError validator.FieldError) string {
		return fieldError.Field()
	})
}

// ValidatePaths is Validate for a body with a nested list, keying each error by its
// path inside the body — "rsvp_note", "members[0].attending" — after passing that path
// through rewrite.
//
// It exists because the leaf name is ambiguous the moment a body carries a list: eight
// members would report eight errors under the single key `attending`, and seven of
// them would be lost. rewrite is what turns the index into whatever the endpoint's
// contract keys by — the RSVP endpoints key by member id, because the frontend renders
// cards by id and an index breaks the moment the list is filtered.
func ValidatePaths(body any, rewrite func(path string) string) error {
	return validateKeyedBy(body, func(fieldError validator.FieldError) string {
		// Namespace is "RSVPSaveRequest.members[0].attending" — with the JSON names,
		// because of RegisterTagNameFunc above. The leading struct name is Go's and no
		// client has ever heard of it.
		_, path, _ := strings.Cut(fieldError.Namespace(), ".")
		return rewrite(path)
	})
}

// validateKeyedBy runs the validator and reports every violation under the key that
// key returns for it.
func validateKeyedBy(body any, key func(fieldError validator.FieldError) string) error {
	err := validate.Struct(body)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		// Not a struct: a programming error at the call site, not something a client
		// did. Reported as a 500 rather than as "check the marked fields".
		return fmt.Errorf("validating request body: %w", err)
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return fmt.Errorf("validating request body: %w", err)
	}

	fields := make(map[string]string, len(fieldErrors))
	for _, fieldError := range fieldErrors {
		fields[key(fieldError)] = validationMessage(fieldError)
	}
	return ValidationError{Fields: fields}
}

// validationMessage is the German sentence for one violated rule.
//
// Kept in one switch rather than as a message per field: the rules are generic
// ("required", "max"), and eighty per-field sentences would be eighty strings to
// proof-read where five do the job. A field that genuinely needs its own wording
// gets it at the endpoint, not here.
func validationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "Bitte fülle dieses Feld aus."
	case "min":
		if isNumeric(fieldError) {
			return fmt.Sprintf("Bitte gib mindestens %s an.", fieldError.Param())
		}
		return fmt.Sprintf("Bitte gib mindestens %s Zeichen ein.", fieldError.Param())
	case "max":
		if isNumeric(fieldError) {
			return fmt.Sprintf("Bitte gib höchstens %s an.", fieldError.Param())
		}
		return fmt.Sprintf("Bitte gib höchstens %s Zeichen ein.", fieldError.Param())
	case "gte":
		return fmt.Sprintf("Der Wert darf nicht kleiner als %s sein.", fieldError.Param())
	case "lte":
		return fmt.Sprintf("Der Wert darf höchstens %s sein.", fieldError.Param())
	case "oneof":
		// A guest cannot produce this through the UI — the controls only offer the
		// allowed values — so the wording is for the one case that does happen: a
		// hand-made request, or a frontend bug.
		return "Dieser Wert ist hier nicht erlaubt."
	default:
		return "Bitte prüfe dieses Feld."
	}
}

// isNumeric reports whether the field being complained about holds a number, so that
// `min` and `max` do not tell somebody to type more "characters" into an integer.
func isNumeric(fieldError validator.FieldError) bool {
	switch fieldError.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// AgeValidationError translates the domain's age rules into a field error on `age`,
// and passes anything else through unchanged.
//
// The rule lives in the domain (domain.ResolveAge) because it is a business fact;
// the field name and the German sentence live here because they are the shape of a
// request and the wording of a response. A driver error from the column's CHECK
// constraint would be neither — it is not a message a form can put next to a field,
// which is why the pairing is enforced before the write.
func AgeValidationError(err error) error {
	return AgeValidationErrorAt("age", err)
}

// AgeValidationErrorAt is AgeValidationError under a chosen field key, for a body that
// carries more than one person's age: the RSVP endpoints key theirs
// `members.<id>.age`, so the message lands on the right card.
func AgeValidationErrorAt(field string, err error) error {
	switch {
	case errors.Is(err, domain.ErrAgeOnAdult):
		return ValidationError{Fields: map[string]string{
			field: "Ein Alter speichern wir nur für Kinder.",
		}}
	case errors.Is(err, domain.ErrAgeOutOfRange):
		return ValidationError{Fields: map[string]string{
			field: "Bitte gib ein Alter zwischen 0 und 17 Jahren an.",
		}}
	default:
		return err
	}
}
