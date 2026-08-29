// Package integration drives the JSON API end to end, through the real router and
// the real middleware chain — the same NewRouter that main calls, so a test cannot
// pass against a wiring the deployed binary does not have.
//
// Each test runs against its own migrated temp-file SQLite, built by newTestApp in
// harness_test.go; there is no mocking layer, because with a pure-Go driver a real
// database is both easier to set up and stronger evidence than a fake.
//
// This is also where the privacy guarantees are checked: guest-facing responses
// must never contain code, admin_note or any budget field.
package integration
