package security_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/security"
)

func TestNewSessionTokenIsUnpredictableAndCookieSafe(t *testing.T) {
	t.Parallel()

	const sampleSize = 1000

	seen := make(map[string]bool, sampleSize)
	for range sampleSize {
		token := security.NewSessionToken()

		require.False(t, seen[token], "generated the same token twice")
		seen[token] = true

		// 32 bytes in unpadded base64url. Decoding is the real assertion: it
		// proves the encoding, and the length check proves the entropy.
		decoded, err := base64.RawURLEncoding.DecodeString(token)
		require.NoError(t, err, "token %q is not unpadded base64url", token)
		assert.Len(t, decoded, 32)

		// The characters that would have to be escaped in a Set-Cookie header, or
		// mangled by anything that URL-encodes the value on the way through.
		assert.False(t, strings.ContainsAny(token, "=+/"), "token %q needs escaping", token)
	}
}

func TestHashSessionTokenIsStableAndHidesTheToken(t *testing.T) {
	t.Parallel()

	token := security.NewSessionToken()
	hash := security.HashSessionToken(token)

	assert.Equal(t, hash, security.HashSessionToken(token), "the same token must hash to the same id")
	assert.NotEqual(t, hash, security.HashSessionToken(security.NewSessionToken()))

	// Hex SHA-256. The length is asserted rather than the value so the test does
	// not become a second implementation of the hash.
	assert.Len(t, hash, 64)
	assert.NotContains(t, hash, token)
}
