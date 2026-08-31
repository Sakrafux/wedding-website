package domain

import (
	"crypto/rand"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// codeAlphabet is the alphabet household login codes are drawn from and the only
// one they are accepted in.
//
// Exactly four characters are missing, and they look like typos otherwise: 0, O,
// 1 and I. Both members of each confusable pair are dropped rather than one,
// because keeping either still leaves a guest deciding which of two glyphs they
// are looking at on a printed card. L, U and V are kept: uppercase L is not 1 in
// the invite's typeface, and dropping more would cost the property below.
//
// 32 characters is not a coincidence and not free to change: it must stay a
// divisor of 256, see GenerateCode.
const codeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// codeLength of 6 gives 32^6 ≈ 1.07 billion codes for ~60 households. Short enough
// to type on a phone, vast enough that guessing one is not a threat — see
// specification/06-privacy-security.md for the arithmetic against the rate limit.
const codeLength = 6

// ErrMalformedCode reports a login code that cannot be a code at all — wrong
// length, or a character outside codeAlphabet.
//
// Deliberately a plain sentinel and not a domain.Error: "malformed" is never
// reported to a guest. The login endpoint answers the same generic failure for a
// malformed code as for an unknown one, so a wire ErrorCode for this
// would be a code no response may ever carry. Callers that do want to tell the
// two apart — an admin entering a code by hand — branch on this sentinel.
var ErrMalformedCode = errors.New("login code is not six characters from the code alphabet")

// GenerateCode returns a fresh household login code in stored form.
//
// It does not check the code against the database: uniqueness is the UNIQUE index
// on household.code, so the caller retries on a rejected insert rather than asking
// first and racing between the question and the answer — see
// persistence.HouseholdStore.withGeneratedCode.
//
// No error return, because crypto/rand.Read cannot fail — since Go 1.24 it panics
// if the system entropy source is broken, which is not a condition this app could
// do anything sensible about anyway.
func GenerateCode() string {
	random := make([]byte, codeLength)
	_, _ = rand.Read(random)

	code := make([]byte, codeLength)
	for index, value := range random {
		// Modulo is unbiased here only because len(codeAlphabet) is 32 and 32
		// divides 256 exactly, so every alphabet character is reachable from
		// exactly 8 byte values. Change the alphabet's length and this silently
		// starts favouring its first characters — rejection-sample instead.
		code[index] = codeAlphabet[int(value)%len(codeAlphabet)]
	}
	return string(code)
}

// NormalizeCode returns the canonical stored form of whatever a guest typed:
// upper case, no whitespace, no dashes.
//
// It normalizes only, and does not judge the result — run ValidateCode on the
// output. Splitting the two keeps "what the guest meant" separate from "is that a
// code at all", which is what lets the login endpoint report one failure for a
// malformed code and an unknown one alike.
//
// Whitespace covers the non-breaking space, because that is what copying the code
// out of the invitation PDF produces.
//
// Dashes are stripped even though nothing prints one any more: the code is printed
// as six ungrouped characters (a group separator was dropped as unnecessary at this
// length, and it was the source of every awkward case in the input field). Guests
// still type dashes out of habit and word processors still turn a typed hyphen into
// an en or em dash, so accepting all three costs one character class and spares
// somebody a rejected code they typed correctly.
//
// Lookalike folding (0→O, 1→I) is deliberately *not* done. It was rejected because
// it widens the accepted set — two different strings would resolve to one code —
// to solve a confusion the alphabet already makes impossible: no code contains an
// O or an I to be confused with.
func NormalizeCode(input string) string {
	var normalized strings.Builder
	normalized.Grow(len(input))

	for _, character := range input {
		if unicode.IsSpace(character) || unicode.Is(unicode.Dash, character) {
			continue
		}
		normalized.WriteRune(unicode.ToUpper(character))
	}
	return normalized.String()
}

// ValidateCode reports whether code, already normalized, has the shape of a login
// code, returning ErrMalformedCode if it does not.
//
// Shape only — a well-formed code no household holds is not this function's
// problem. Checking it before touching the database means a mistyped code costs no
// query, and keeps the login path from treating arbitrary guest input as a lookup
// key.
func ValidateCode(code string) error {
	if utf8.RuneCountInString(code) != codeLength {
		return ErrMalformedCode
	}

	for _, character := range code {
		if !strings.ContainsRune(codeAlphabet, character) {
			return ErrMalformedCode
		}
	}
	return nil
}
