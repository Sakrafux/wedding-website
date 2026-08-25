// Package integration drives the JSON API end to end, through the real router and
// the real middleware chain — the same NewRouter that main calls, so a test cannot
// pass against a wiring the deployed binary does not have.
//
// From E0-11 onward each test runs against its own temp-file SQLite with the
// migrations applied; there is no mocking layer, because with a pure-Go driver a
// real database is both easier to set up and stronger evidence than a fake.
//
// This is also where the privacy guarantees are checked: guest-facing responses
// must never contain code, admin_note or any budget field.
package integration
