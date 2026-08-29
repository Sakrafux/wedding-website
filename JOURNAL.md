# Journal

Work log for the wedding web app. Newest entry first. One `##` heading per day: a `Done` list, a `Decisions` list for anything newly decided, then a `Time:` line and a `Cost:` line per model used.

Entries stay short. The reasoning behind a decision belongs in the spec, the story file or a code comment — this file records *that* it was decided and *when*, and points at where it lives.

## 2026-08-29

Done:

- `E0-07` — `middleware.SecurityHeaders`: CSP, `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`, `Permissions-Policy`, `X-Robots-Tag`, exactly the table in `06-privacy-security.md`. Registered outside the recoverer so an unrecovered panic still carries them.
- `System.Robots` serves `/robots.txt` with a blanket disallow, from Go rather than the frontend's `public/`.
- `tests/integration/security_headers_test.go`: header set asserted on an API response, an error response and a non-`/api` path, each also asserted to be set exactly once; `robots.txt` body and content type.
- CSP verification against the real bundle is deferred to `E0-08`/`E0-09`; header verification through the real proxy to `E0-12`.
- `E0-08` — `web/` scaffolded: Vite 8, React 19, TS 6 (`strict`, `noUncheckedIndexedAccess`), Tailwind 4, shadcn/ui with `Button`/`Input`/`Label`/`Card`, TanStack Router (file-based) and Query, pnpm pinned via `packageManager`.
- All design tokens from `05-design.md` live in `src/index.css` as Tailwind 4 `@theme` variables, with shadcn's own variable names mapped onto them; global focus-visible ring and `prefers-reduced-motion` in the base layer.
- Cormorant Garamond 600 and Source Serif 4 self-hosted as Latin + Latin-Extended `.woff2` in `src/assets/fonts/`, OFL note beside them.
- `src/lib/enums.ts` (English unions) and `src/lib/labels.ts` (the only German enum labels), each map typed `Record<Union, …>` — verified that removing a key fails `tsc`.
- Built `dist/index.html` contains no inline script or style, so the `E0-07` CSP holds against the real bundle.
- `web/README.md` rewritten: a justification per dependency (runtime and tooling), the lockfile/pinning policy, the linting and formatting rules, and the `@/` double-declaration warning. Root README covers setup, so `web/` links to it instead of repeating it.
- Prettier added with `prettier-plugin-tailwindcss`; `format` / `format:check` scripts. Double quotes, semicolons, 2-space indent, 120 columns — reasoning in `web/README.md`; 120 matches `lll`'s default and this repo's Go p99. Markdown is in `.prettierignore` so the repo keeps one Markdown convention.
- Root README "Local development" now covers the frontend: `corepack enable`, `pnpm dev` on 5173, why to open 5173 rather than 8080, and the frontend check commands.
- `E0-09` — `web/embed.go` (package `frontend`) embeds `dist` with `all:`, `internal/infrastructure/web/static.go` serves it: real file if present, `index.html` with 200 otherwise, explicit MIME table, `immutable` on `/assets/`, `no-cache` on `index.html`. Registered as the router's `NotFound`, so `/api` keeps its JSON 404.
- `web/dist/.gitkeep` committed with a `.gitignore` exception, so `go build` works on a clean checkout; a bundle-less binary serves the API and reports the missing shell per request.
- Vite empties `dist/` and takes `.gitkeep` with it, so `pnpm build` ends with `scripts/write-dist-gitkeep.mjs`. Verified: `go build` succeeds against a `dist/` holding only that file.
- `tests/integration/static_test.go`: root and deep link return the shell, unknown asset falls through, hashed asset vs. `index.html` cache headers, `/api` miss stays JSON. Skips when no bundle is embedded.
- `Makefile`: `build` (frontend then Go), `build-web`, `run` (exports `.env`), `preview` (exports `.env`), `test` (`go test`, `gofmt`, `go vet`), `fmt`, `lint`, `clean`.
- Smoke-tested the built binary: `/`, `/rsvp`, a hashed asset and `/api/nope` all correct, security headers intact on the SPA response.

Decisions:

- Latin-Extended is bundled alongside Latin: guest surnames carry Czech, Polish and Hungarian diacritics. Rationale in `src/styles/fonts.css`.
- Root font size is `1.125rem`, so Tailwind's spacing unit is 4.5px rather than 4px; type and spacing scale together and the px values in `05-design.md` are read as ratios. Rationale in `src/index.css`.
- `src/routeTree.gen.ts` is committed — `pnpm build` type checks before Vite generates it. Noted in `web/README.md`.
- **oxlint + prettier, no eslint.** Rationale in `web/README.md` and `CLAUDE.md`; revisit only if a type-aware rule would have caught a real bug.
- Versions stay as `^` ranges pinned by `pnpm-lock.yaml`; `E0-10` must install with `--frozen-lockfile`, and `pnpm audit` joins `govulncheck` in the pre-deploy check.
- Embed lives in `web/embed.go`, not next to the handler: an embed directive cannot reach outside its package directory. Package named `frontend` to avoid colliding with `infrastructure/web`.
- SPA fallback answers 200, and nothing outside `/api` ever 404s — the handler cannot tell a typo from a client-side route. Rationale in `static.go`.
- Content types come from an explicit table, not `mime.TypeByExtension`, whose answers depend on the image's `/etc/mime.types`.
- TanStack router and query devtools mounted from `src/components/Devtools.tsx` behind a dynamic `import()`, not a static import guarded by `import.meta.env.DEV` — a static import survives dead-code elimination and would ship the panels to guests. Verified absent from `dist/`.

Time: 1.5h (tentative)
Cost: $10.18 (tentative)

## 2026-08-27

Done:

- `E0-05` — `0001-initial-schema.sql`: all nine tables, `CHECK` on every enum, FK behaviours, defaults, indexes and the five `app_setting` seed rows. `rsvp_deadline` seeded to `2027-05-17T21:59:59Z`.
- `tests/integration/schema_test.go`: one rejected insert per enum column, duplicate code, household cascade, both seating RESTRICT paths, one-seat-per-guest-per-venue, both venue-mismatch paths, guest and household defaults, seeded settings.
- `database_test.go` now uses `household`/`guest` instead of the inline `parent`/`child` stand-in; the `E0-05` forward references are gone.
- Data model, `02-features`, `06-privacy-security`, `07-roadmap`, `CLAUDE.md` and the F7 story titles updated for the seating rework.
- `E0-06` — `middleware.RequestID` (7 chars, unambiguous base32, `X-Request-Id` on every response) and `middleware.Recoverer` (stack into the structured log, envelope out); both replace chi's versions, so the router now uses `httplog.Handler` instead of `httplog.RequestLogger`.
- Error envelope gained `request_id` and an optional `fields` map.
- Tests: `httpio/write_test.go` (one ID in header, body and log line; client-supplied header ignored; `fields` omitted when empty) and `tests/integration/error_envelope_test.go` (unknown route, success header, panic leaks nothing, field-keyed validation). Harness gained `newTestServerWithRoutes` and now discards log output.
- `domain/errors.go` — `Error{Code, cause}` plus `NewError`/`WrapError`, and the central `ErrorCode` list (first entry: `unknown_login_code`). `httpio.RespondError` maps code → status + German message via the `errorResponses` table; tests cover wrapped codes, unmapped codes and non-domain errors failing closed.
- `httpio` narrowed to `WriteJSON` + `RespondError`: `WriteError`, `WriteValidationError` and `WriteInternalError` are gone, the transport codes (`not_found`, `method_not_allowed`, `not_ready`, `validation_failed`, `internal_error`) moved into the same table as sentinels, and validation became `httpio.ValidationError`. Handlers and the recoverer now pass errors, not statuses.

Decisions:

- **Both venues are seated.** `seating_table` → `seating_unit` with a `venue` enum (`church` | `party`); a church pew is the same object as a party table. Was an oversight in the data model.
- **Seats are rows.** New `seat` table transcribed by hand from the SVG, so an assignment names a specific place. Rejected: an `svg_element_id` on `seat_assignment` with no seat rows — cheaper to set up, but free-seat queries and place cards both want the row. See `03-data-model.md`.
- `seating_unit.capacity` dropped — capacity is `COUNT(seat)`, so over-assignment is unrepresentable instead of a warning.
- `venue` is denormalised down `seating_unit → seat → seat_assignment` behind composite FKs, which buys `UNIQUE (guest_id, venue)` without a trigger. Rationale in `03-data-model.md`.
- `guest.last_name` is required: households of two surnames are common.
- `stroller` removed from `guest.seating_need` → `household.has_stroller` BOOLEAN. A pram is parked, not sat on, belongs to the household, and nobody brings two — so a flag, not a count.
- No "log out other devices" — one household, one shared code. `session.user_agent`/`ip` stay, for the audit trail only.
- Request ID is 7 speakable base32 chars, generated with `math/rand/v2`, stored under chi's context key so httplog logs the same value; an incoming `X-Request-Id` is ignored. Rationale in `middleware/requestid.go`.
- Validation failures answer **400**, not 422 — the frontend branches on the presence of `fields`, so a second status buys nothing. In `httpio`'s `errorResponses` table.
- **Domain errors carry a code, not a message or a status.** One `errorResponses` table in `httpio` maps code → status + German text, so it is the full list of what a guest can be shown; unmapped codes and non-domain errors fail closed to a generic 500. Rejected `AppError{Type, Message}` with the message echoed to the client; reasoning in `04-architecture.md`.
- **No exported writer takes a status, code or message.** `WriteError(status, code, message)` was a second write path next to the table and therefore a second chance to leak; every failure now goes through `RespondError`.
- `domain.ErrorCode` constants live in one block in `domain/errors.go`, not next to each rule — deliberate break with "keep it near the code", so the list can be read against the message table.

Time: 2h
Cost: $8.36

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

Time: 1.5h

Cost: Opus 5 — $9.45

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
