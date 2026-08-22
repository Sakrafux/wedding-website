# `E0-04` — Migration runner

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-03`

## Story

As a developer, I want embedded migrations applied automatically at startup so that deploying is a container restart and the schema can never drift from the binary.

## Scope

**In:**

- `internal/infrastructure/persistence/migrations/` with `go:embed`.
- A runner that reads `NNNN-name.sql` files in lexical order, applies the unapplied ones, and records each in `schema_migration`.
- Called from `main` before the HTTP server starts listening.

**Out:**

- The schema itself → `E0-05`.
- Down-migrations — forward-only by decision.

## Instructions

1. `schema_migration(version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`. Create it if absent as step zero.
2. Zero-padded four-digit prefixes so lexical order is numeric order.
3. Each migration runs **inside a transaction** with its `schema_migration` insert. A partially applied migration must not be recordable as applied.
4. Any failure aborts startup with a non-zero exit. A container that will not start is a far better signal than one running against a half-migrated database.
5. Log each applied version. On a no-op run, log one line saying the schema is current — silence on startup is indistinguishable from a bug.
6. Refuse to start if the database contains a version **not present** in the embedded set: that means the binary is older than the database, which is a rollback nobody meant to perform.

## Test plan

- [ ] Integration: fresh database → all migrations applied, `schema_migration` populated.
- [ ] Integration: run twice → second run is a no-op, nothing re-applied.
- [ ] Integration: a deliberately broken SQL file → startup fails, and its version is **not** recorded.
- [ ] Integration: unknown version present in the database → startup refuses.

## Done when

- [ ] Deleting the database file and restarting produces a fully migrated schema.
- [ ] Checkbox ticked in `README.md`.
