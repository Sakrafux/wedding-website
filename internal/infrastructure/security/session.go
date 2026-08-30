package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// sessionTokenBytes is 32, i.e. 256 bits of entropy. Guessing one is not a threat
// anybody needs to model; the number is the usual one because there is no reason
// to be clever with it.
const sessionTokenBytes = 32

// NewSessionToken returns a fresh session token — the value that goes into the
// cookie and is never stored anywhere.
//
// base64url without padding, so the token is cookie-safe without escaping and
// survives being logged, copied or pasted without a '=' turning into '%3D' halfway.
//
// No error return: since Go 1.24 crypto/rand.Read cannot fail, it panics if the
// system entropy source is broken. There is no useful fallback for that here — a
// session token from a degraded source would be worse than no service at all.
func NewSessionToken() string {
	token := make([]byte, sessionTokenBytes)
	_, _ = rand.Read(token)

	return base64.RawURLEncoding.EncodeToString(token)
}

// HashSessionToken returns the hex SHA-256 of token, which is what session.id
// holds. Look a session up by hashing the presented token, never by storing it.
//
// A leaked database therefore yields no usable sessions. That is defence in depth
// rather than a barrier — the same file also holds the login codes in plaintext —
// but it is the difference between "the attacker can log in as anyone" and "the
// attacker can log in as anyone once they have also typed a code".
//
// Plain SHA-256 and not bcrypt or argon2: the input is 256 random bits, not a
// human-chosen password, so there is no dictionary to slow down and a per-request
// key-derivation cost would buy nothing.
func HashSessionToken(token string) string {
	digest := sha256.Sum256([]byte(token))

	return hex.EncodeToString(digest[:])
}
