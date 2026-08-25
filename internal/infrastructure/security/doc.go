// Package security holds session token generation and hashing, household login
// code normalization and generation, and the constant-time admin credential compare.
//
// No JWT and no password hashing: sessions are opaque random tokens stored in the
// database, because a household session must survive a year and stay revocable.
// The single admin's credentials come from the environment in plaintext.
package security
