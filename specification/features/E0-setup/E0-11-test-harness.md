# `E0-11` — Integration test harness

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-05`

## Story

As a developer, I want a one-line way to spin up a fully migrated temporary database and an HTTP test server, so that writing an integration test is cheap enough that I actually write them.

## Scope

**In:**

- `tests/integration/harness.go`: temp-file SQLite, migrations applied, real router, `httptest.Server`, cleanup.
- Helpers: seed a household, log in and hold the cookie, issue JSON requests, decode the error envelope.
- One end-to-end smoke test proving the harness works.

**Out:**

- Feature-specific tests → the feature stories.

## Instructions

1. **Temp file, not `:memory:`.** In-memory SQLite behaves differently across connections, and this app deliberately uses two pools — the test must exercise the same configuration as production.
2. `newTestApp(t)` returns the server URL plus handles, and registers `t.Cleanup` to close and delete. No manual teardown in any test.
3. Every test gets its own database file, so tests can run in parallel and none depends on another's leftovers.
4. Fixture builders with sensible defaults and overrides — `seedHousehold(t, db, withGuests(2))`. Fixtures that require every field make tests unreadable and, worse, hide which field the test is actually about.
5. An `assertNoLeak(t, body)` helper that fails if a response contains `"code"`, `"admin_note"`, or any budget field name. Every guest-facing test calls it. This is the mechanical enforcement of the DTO privacy rule from [04-architecture](../../04-architecture.md) — a rule enforced only by discipline is a rule that gets broken during a late-night change.
6. Dependencies: `stretchr/testify` and `google/go-cmp`. Nothing else.

## Test plan

- [ ] The smoke test starts the harness, calls `/api/health`, and passes.
- [ ] Two tests running in parallel do not interfere.
- [ ] The temp database is removed after the run.
- [ ] `assertNoLeak` fails when handed a body containing `admin_note` — test the test.

## Done when

- [ ] A new integration test needs three lines of setup.
- [ ] Checkbox ticked in `README.md`.
