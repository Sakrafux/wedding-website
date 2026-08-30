package security

import "crypto/subtle"

// AdminCredentials is the single admin's login, as configured in the environment.
//
// There is no admin_user table, no password hash and no reset flow: one admin,
// credentials in ADMIN_USER and ADMIN_PASSWORD, changed by editing the environment
// and restarting. Hashing would protect the password from a database leak, and the
// password is not in the database.
type AdminCredentials struct {
	User     string
	Password string
}

// Matches reports whether the submitted pair is the configured one.
//
// Both halves are compared, and both comparisons run even when the first already
// failed. Returning as soon as the username mismatched would make a wrong username
// measurably faster to reject than a wrong password, which tells an attacker when
// they have found the username — and the cost of not returning early is nil.
//
// subtle.ConstantTimeCompare still returns early when the lengths differ, so the
// length of each half remains observable. That is accepted: the username is not a
// secret, and a password's length is worth far less than the password.
func (credentials AdminCredentials) Matches(user, password string) bool {
	userMatches := subtle.ConstantTimeCompare([]byte(credentials.User), []byte(user)) == 1
	passwordMatches := subtle.ConstantTimeCompare([]byte(credentials.Password), []byte(password)) == 1

	return userMatches && passwordMatches
}
