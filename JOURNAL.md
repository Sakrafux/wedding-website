# Journal

Work log for the wedding web app. Newest entry first. One `##` heading per day: a `Done` list, a `Decisions` list for anything newly decided, then a `Time:` line and a `Cost:` line per model used.

Entries stay short. The reasoning behind a decision belongs in the spec, the story file or a code comment — this file records *that* it was decided and *when*, and points at where it lives.

## 2026-08-26

Done:

- `E0-04` — `persistence.Migrate`: `go:embed` of `migrations/*.sql`, `schema_migration` created as step zero, each file applied in its own transaction with its version insert, unknown applied version refuses startup, one log line either way. Called from `main` before the listener starts, and from the integration harness in the same order.
- Unit tests over an in-memory `fs.FS` (order, idempotency, rollback of a broken file, unknown version, malformed names) plus one over the embedded set; integration test asserts a fresh database comes up migrated.
- `migrations/0001-initial-schema.sql` exists as a comment-only placeholder — `E0-05` fills it.
- `E0-03` — `configuration.Database`: write pool capped at one connection, separate read-only pool, pragmas carried in the DSN so every connection gets them, parent directory created, both handles closed on shutdown.
- `GET /api/ready` (`handler.System.Ready`) pinging both pools, 503 with the reason kept in the log.
- Integration tests: pragmas across concurrent connections per pool, FK violation rejected, concurrent writes serialised, read pool rejects writes and sees committed rows, both `/api/ready` paths. Test helpers moved to `tests/integration/harness_test.go`.
- `E0-02` — `configuration.Config` + `Load()`, all eight env vars, aggregated validation error, `ADMIN_PASSWORD` redacted in `LogValue`/`String`, unit tests.
- `main` loads config before the logger and takes the log level from it; failure goes to a plain stderr `slog` handler, exit 1.
- `.env.example` committed, `.env` and `/local/` gitignored, README `Local development` section.

Decisions:

- `schema_migration.version` is the four-digit prefix only, not the file name, so a file can be renamed for clarity without the database treating it as unapplied.
- `migrate` takes an `fs.FS` so the failure cases can be built in memory; a broken migration must not be a shippable file.
- Malformed names, duplicate versions and an empty migration set all fail startup rather than being skipped.
- The `0001` placeholder is comment-only rather than a stub table: an applied version whose contents later change is the migration failure nothing detects. Development databases created before `E0-05` must be deleted, not migrated.
- Health split into `/api/health` (liveness, no dependencies) and `/api/ready` (pools pinged, 503). Recorded in `04-architecture.md` and `E0-03` instruction 7; the story's "extend the health check" line was rewritten.
- Read pool DSN omits `journal_mode` — setting it writes the file header, which `mode=ro` cannot. Write pool opens first so the file and its `-wal`/`-shm` exist.
- `_txlock=immediate` on the write pool: a deferred transaction upgrading mid-flight can hit a `SQLITE_BUSY` that `busy_timeout` cannot resolve.
- Handlers take `*configuration.Database` directly. `configuration` is a leaf and stays one — wiring lives in `cmd/wedding`, so there is no cycle to defend against and no interface worth writing. Its `doc.go` and the layout in `04-architecture.md` both claimed it owns the DI wiring; corrected.
- Handlers are resource-scoped structs, one file per group: `handler/system.go` holds `System` plus health, ready and the `/api` fallbacks, all as methods even where there are no dependencies. Rejected a single handler struct for the whole API. In `04-architecture.md` under the deliberate deviations.
- Env var set but blank counts as absent — failure if required, default if optional.
- `Load()` reads `os.Getenv` directly; no injected lookup, `t.Setenv` covers the tests.
- `testify` pulled in early (it was pencilled in for `E0-11`).
- No dotenv dependency: Compose supplies the environment in production, so it would serve the dev shell alone.
- Makefile (`run`, `test`, frontend-then-Go `build`) deferred to `E0-09` — one command list, not two. Recorded as instruction 6 there.

Time: <h>

Cost: Opus 5 — $<x>

## 2026-08-25

Done:

- `E0-01` — Go module on 1.26, package tree with a `doc.go` per package, chi + `httplog`, `GET /api/health`, graceful shutdown, all four server timeouts explicit.
- Deps held to `chi/v5` + `httplog/v2`; integration test on stdlib `httptest`.

Decisions:

- `web` split into `handler/`, `httpio/`, `dto/`, `middleware/` beneath the router. Forced by the error envelope: envelope writers inside `web` would be an import cycle. Spec tree and deviations updated. Rejected `helper/`, and `respond.JSON`/`respond.Error` in favour of `WriteJSON`/`WriteError`.
- `middleware.RealIP` deliberately not installed — `F1-B05` resolves the client IP against `TRUSTED_PROXY_CIDRS` instead.
- Minimal error envelope lives in `web/dto` now; `E0-06` grows it rather than inventing a second shape.
- New `CLAUDE.md` convention: forward-reference comments name their story, and a story is not ticked until `grep -rn "<ID>" --include='*.go' .` is empty.
- Kept the root layout over a `backend/` + `frontend/` split — `go:embed` cannot cross the module root.

Time: 1h

Cost: Opus 5 — $4.11

## 2026-08-22

Done:

- `05-design.md`, `06-privacy-security.md`, `07-roadmap.md` written; privacy doc toned down to match the actual threat model.
- `specification/features/` split started: index README, `_TEMPLATE.md`, 23 story files in full (`E0-setup` 12, `F1-login` 11), every other epic as bare checkboxes, plus an `E-ops` epic for the non-code gates.
- Root `README.md` written, `TODO.md` moved to the root and scoped away from the build tracker.

Decisions:

- Wedding date 2027-07-17 as the working assumption; invitations Oct/Nov 2026 are the real deadline.
- Print shop confirmed for variable-data printing.
- No separate save-the-date — one mailing does both jobs; reasoning in `02-features.md`.
- pnpm as the frontend package manager.
- This journal is where effort and AI spend get tracked.

Time: 4.5h

Cost: Opus 5 (1M context) — $11.31

## 2026-08-21

Done:

- Project set up from scratch. `CLAUDE.md` plus the first spec batch: `01-vision-scope.md`, `02-features.md`, `03-data-model.md`, `04-architecture.md`, and the initial TODO list.

Decisions:

- Go single binary with an embedded React/Vite frontend, SQLite, trimmed hexagonal layout.

Time: 3h

Cost: Opus 5 (1M context) — $6.57
