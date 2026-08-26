# `E0-03` — SQLite connection, pragmas, pools

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-02`

## Story

As a developer, I want the database opened correctly once, centrally, so that no later story has to remember the pragma list and no concurrent RSVP submission ever hits `SQLITE_BUSY`.

## Scope

**In:**

- Open `modernc.org/sqlite` via `jmoiron/sqlx`.
- Two handles: a **write** pool capped at one connection, and a read pool.
- The four pragmas applied to **every** connection, not once at open.
- A readiness endpoint, `GET /api/ready`, verifying both pools respond.

**Out:**

- Migrations → `E0-04`.
- Any store or query → later stories.

## Instructions

1. Pragmas, per [04-architecture](../../04-architecture.md): `journal_mode = WAL`, `foreign_keys = ON`, `busy_timeout = 5000`, `synchronous = NORMAL`.
2. Apply them via the DSN or a connection hook so **every pooled connection** gets them. Applying them once after `sql.Open` configures exactly one connection and silently leaves the rest with `foreign_keys = OFF` — a class of bug that shows up as orphaned rows months later.
3. Write handle: `SetMaxOpenConns(1)`. SQLite has one writer; a pool of writers produces contention, not throughput.
4. Read handle: separate `sqlx.DB`, open in read-only mode, a handful of connections.
5. Both handles closed on shutdown, before the server stops.
6. Create the parent directory of `DB_PATH` if absent, so a fresh volume mount does not require a manual `mkdir`.
7. The database check is a **separate** `GET /api/ready`, not an extension of `/api/health`. `/api/health` is what a restart policy polls, and a restart cannot fix an unmounted volume or a wrong `DB_PATH` — failing it there turns a legible error into a crash loop. `/api/ready` answers 503 when a pool is unreachable and keeps the reason in the log, since it is unauthenticated and a driver error carries the database path.

## Test plan

- [ ] Integration: open against a temp file; `PRAGMA foreign_keys` returns 1 on a connection taken from each pool.
- [ ] Integration: `PRAGMA journal_mode` returns `wal`.
- [ ] Integration: a foreign-key violation is actually rejected — the real proof that pragmas apply per connection.
- [ ] Integration: two concurrent writes both succeed, serialised, with no `SQLITE_BUSY`.

## Done when

- [ ] The app opens a database at `DB_PATH` and `GET /api/ready` reports it reachable.
- [ ] Checkbox ticked in `README.md`.
