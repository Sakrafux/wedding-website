package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generatedCodeSampleSize is large enough that a skewed generator shows up in the
// distribution check below, and small enough to stay a millisecond-scale test.
const generatedCodeSampleSize = 10000

func TestGenerateCodeProducesWellFormedCodes(t *testing.T) {
	t.Parallel()

	for range 1000 {
		code := GenerateCode()

		require.Len(t, code, codeLength)
		require.NoErrorf(t, ValidateCode(code), "generated code %q is not valid", code)
	}
}

// The alphabet already excludes these, so the test is really about the alphabet
// constant: it fails the moment someone "completes" it back to full base32, which
// is the change that would put an O on a printed card next to a 0.
func TestGenerateCodeNeverProducesAmbiguousGlyphs(t *testing.T) {
	t.Parallel()

	const ambiguous = "01IO"

	for range 1000 {
		code := GenerateCode()

		assert.Falsef(t, strings.ContainsAny(code, ambiguous),
			"generated code %q contains an ambiguous glyph", code)
	}
}

// A biased generator is invisible in a well-formed-output test and would quietly
// shrink the keyspace, so the distribution is asserted rather than assumed. The
// bound is loose on purpose: this catches a broken generator, not a subtly
// imperfect one, and a tight bound would flake.
func TestGenerateCodeUsesTheWholeAlphabet(t *testing.T) {
	t.Parallel()

	counts := make(map[rune]int, len(codeAlphabet))
	for range generatedCodeSampleSize {
		for _, character := range GenerateCode() {
			counts[character]++
		}
	}

	require.Len(t, counts, len(codeAlphabet), "not every alphabet character was generated")

	expected := generatedCodeSampleSize * codeLength / len(codeAlphabet)
	for _, character := range codeAlphabet {
		assert.InEpsilonf(t, expected, counts[character], 0.25,
			"character %q appeared %d times, expected roughly %d", character, counts[character], expected)
	}
}

func TestNormalizeCode(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    string
		expected string
	}{
		"already canonical":  {"ABC234", "ABC234"},
		"lower case":         {"abc234", "ABC234"},
		"printed form":       {"ABC-234", "ABC234"},
		"lower printed form": {"abc-234", "ABC234"},
		"spaced":             {"ABC 234", "ABC234"},
		"surrounding space":  {" abc234 ", "ABC234"},
		"lower and spaced":   {"abc 234", "ABC234"},
		// What a word processor makes of a typed hyphen, and what copying out of
		// the invitation PDF produces. Both reach us from a guest's clipboard.
		"en dash":            {"ABC–234", "ABC234"},
		"em dash":            {"ABC—234", "ABC234"},
		"non-breaking space": {"ABC 234", "ABC234"},
		"tab and newline":    {"ABC\t234\n", "ABC234"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			normalized := NormalizeCode(test.input)

			assert.Equal(t, test.expected, normalized)
			require.NoError(t, ValidateCode(normalized))
			assert.Equal(t, "ABC-234", FormatCode(normalized))
		})
	}
}

// Normalizing does not validate, so an input with nothing code-like in it comes
// back as an empty string rather than an error. It is ValidateCode that says no.
func TestNormalizeCodeOfEmptyInputIsMalformed(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", NormalizeCode(" - "))
	assert.ErrorIs(t, ValidateCode(NormalizeCode("")), ErrMalformedCode)
}

func TestValidateCodeRejectsMalformedCodes(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"too short":            "ABC23",
		"too long":             "ABC2345",
		"empty":                "",
		"excluded zero":        "ABC230",
		"excluded one":         "ABC231",
		"excluded letter O":    "ABC23O",
		"excluded letter I":    "ABC23I",
		"lower case":           "abc234",
		"still carries a dash": "ABC-23",
		"punctuation":          "ABC23!",
		// Six runes, but not six bytes: a length check on len() alone would let
		// this through and hand a non-ASCII string to the lookup.
		"multi-byte runes": "ÄBC234",
	}

	for name, code := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, ValidateCode(code), ErrMalformedCode)
		})
	}
}

func TestFormatCode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ABC-234", FormatCode("ABC234"))
	assert.Equal(t, "234-ABC", FormatCode("234ABC"))
}

// Formatting is presentation, and presentation must not be able to crash a
// request. Anything of the wrong length comes back untouched.
func TestFormatCodeLeavesUnexpectedLengthsAlone(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", FormatCode(""))
	assert.Equal(t, "ABC23", FormatCode("ABC23"))
	assert.Equal(t, "ABC2345", FormatCode("ABC2345"))
}
