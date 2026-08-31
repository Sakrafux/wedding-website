# Journal

Work log for the wedding web app. Newest entry first. One `##` heading per day: a `Done` list, a `Decisions` list for anything newly decided, then a `Time:` line and a `Cost:` line per model used.

Entries stay short. The reasoning behind a decision belongs in the spec, the story file or a code comment — this file records *that* it was decided and *when*, and points at where it lives.

## 2026-08-31

Done:

- `F1-B08` — new story: `cmd/seed`, the local dev tool that inserts households with real generated codes. Spec in `specification/features/F1-login/F1-B08-dev-seed.md`.
- `cmd/seed/main.go`: `-households` (default 1) and `-guests` (default 2), loads the app's own environment, migrates a fresh file, one transaction per household, retries a colliding code against the UNIQUE index, prints id/name/`FormatCode` code to stdout behind a `DEVELOPMENT ONLY` banner.
- `cmd/seed/main_test.go`: counts, append-on-second-run, non-positive counts rejected before the file is created, members are seeded adults with NULL `attending`, fresh-database migration, and printed codes resolving through `HouseholdStore.FindByCode`.
- `make seed` (with `SEED_ARGS`) and a "Test data" subsection in `README.md`'s local development chapter.
- Verified end to end: `make seed` then a login with the printed `2BC-96G` against the running binary returns the bootstrap body.
- Fixed missing pointer cursor on every button: Tailwind v4's preflight resets `button` to `cursor: default` and the shadcn v4 button recipe no longer adds it back. Restored globally in `index.css`'s base layer, not per component.

Decisions:

- Codes are generated, never fixed — reasons in the story's Scope.
- No reset flag and no runtime production guard; the guard is that the `Dockerfile` builds `./cmd/wedding` alone. See `cmd/seed`'s package comment.
- The insert SQL stays in `cmd/seed` until `F5-B01` gives `HouseholdStore` a create method; forward reference noted at `insertHousehold`.

Time: <h>
Cost: Opus 5 — $<x>

## 2026-08-30

Done:

- `F1-B01` — `internal/domain/code.go`: `GenerateCode` (crypto/rand, unbiased because 32 divides 256), `NormalizeCode` (case, whitespace incl. NBSP, dashes incl. en/em), `ValidateCode` (shape only, `ErrMalformedCode`), `FormatCode` (`ABC234` → `ABC-234`). Unit tests cover the normalisation table, the rejected shapes and the generator's distribution.
- `F1-B02` — `domain/session.go` holds the lifetime and refresh policy (household 365d, admin 8h, roll at most daily); `security/session.go` the 32-byte base64url token and its SHA-256 id; `persistence/session.go` the store, plus `PurgeExpiredPeriodically` wired into `main` on the signal context.
- `persistence/timestamp.go` and `persistence/errors.go`: the one RFC3339-UTC storage format, and `ErrNotFound` so callers need not import `database/sql`.
- `F1-B03` — `middleware/session.go`: `SessionGate.Resolve` on the whole `/api` tree, `RequireHousehold`, `RequireAdmin`, typed context accessors, and the session cookie's issue/clear helpers in one place.
- `F1-B04` — `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/me` over `application/auth.go`; `dto.BootstrapResponse` shared by login and `/api/me`. `httpio.DecodeJSON` added: JSON content type required, unknown fields rejected, 1 MiB cap.
- `persistence/household.go` and `persistence/setting.go`: the read side F1 needs; `F5-B01`/`F5-B02` widen the first into the full CRUD.
- Integration tests: session store round trip, hashing, expiry, purge; the gates against real routes; login normalisation variants, cookie attributes both ways on `SESSION_COOKIE_SECURE`, `last_login_at`, re-login replacing a session, logout idempotence, and `assertNoLeak` on both bootstrap bodies.
- Harness: `newTestApp` takes options (`withExtraRoutes`, `withSecureCookies`), plus `logIn`, `putSessionCookie`, `countSessions`, `setCookie`. Fixed `withGuests` numbering, and the default display name no longer embeds the household's own code.
- Verified end to end against the built binary: `abc-234` logs in, `/api/me` matches, `/api/admin/*` is 401 for a household session, a form-encoded login is 400, logout is 204 and revokes.
- `F1-B05` — `middleware/clientip.go` resolves the caller against `TRUSTED_PROXY_CIDRS`, rightmost untrusted hop, IPv4-mapped proxies unmapped first; `middleware/ratelimit.go` is an in-memory sliding window, 10 guest / 5 admin failures per hour per address, one limiter per endpoint. Failure is read off the response status, so an endpoint cannot forget to report one.
- `main` warns at startup when `TRUSTED_PROXY_CIDRS` is empty — the misconfiguration that silently makes the limiter key on the proxy.
- `F1-B06` — `domain/audit.go` (entry constructors) and `persistence/audit.go`. Login, admin login and failed attempts recorded; audit failures are logged and never fail the request.
- `F1-B07` — `security/credentials.go` (`subtle.ConstantTimeCompare`, both halves, no early return), `POST /api/auth/admin/login`, and `GET /api/admin/me`.
- `F1-F01`–`F1-F04` — `lib/api.ts` (one fetch wrapper, `ApiError` vs `NetworkError`), `lib/session.ts` (`me` and the admin session as the only session state), `lib/code.ts`, `CodeInput`, the login screen, the confirmation screen, the guest and admin layouts with their guards, and the admin shell with disabled placeholders.
- German screen copy moved into `labels.ts` alongside the enum maps; the file is now the whole of what a guest can read.
- Frontend test setup: vitest, jsdom, Testing Library. 34 component tests driving the real router and real guards against a stubbed `fetch`. `make check` runs both halves.
- Verified against the built binary: SPA deep links, guest login, admin login, the admin gate refusing a household session, 10 failures then 429, and an audit log with no code or password in it.
- Wrote the seven `F5` story files in `specification/features/F5-admin-households/`: household CRUD, guest CRUD, code assignment and regeneration, the two CSV exports, and the three frontend stories.
- Collected eight open questions against them in `TODO.md`, each with the default the story already assumes, so F5 is buildable unanswered.

Decisions:

- The login-code alphabet excludes exactly `0`, `O`, `1`, `I` — not `L`/`U`/`V`, which `F1-B01` and `requestid.go` both claimed. Both corrected; reasoning in `domain/code.go`.
- Malformed codes are a plain sentinel (`domain.ErrMalformedCode`), not an `ErrorCode`: no response may ever distinguish them from unknown ones. Login collapses both to `unknown_login_code`, which replaces the story's `invalid_code`.
- `/api/admin` is mounted with `RequireAdmin` and a catch-all before it has any routes. chi builds a sub-router's middleware chain only once a route exists, so an empty guarded subtree silently answers 404 to everyone — the catch-all is what makes the gate real. `F1-B07` adds the login behind it.
- `Content-Type: application/json` is enforced in `httpio.DecodeJSON` rather than in a router middleware — the CSRF control from `06-privacy-security` lands where handlers already parse, and leaves bodyless `POST`s alone.
- Client IP for the audit trail is the direct peer until `F1-B05`; `X-Forwarded-For` is deliberately not read yet.
- Session `last_seen_at` means "when the session was last extended", accurate to a day. The cost of not writing on every read.
- `entity`/`entity_id` for auth events are the household (or `admin`, id 0), not `session` as `F1-B06` said: session ids are a token hash and cannot go in an INTEGER column. Story corrected.
- Rate-limit visibility is one log line per refusal, not a counter, and refusals are not audited — nothing was attempted, and recording them would let anyone grow `audit_log` by hammering. `F1-B05` updated.
- `GET /api/admin/me` added to `F1-B07`'s scope: the admin route guard has nothing else to ask, since `/api/me` answers 401 for an admin.
- `CodeInput` carries no `maxLength` — the attribute counts raw characters and truncated a pasted `ABC-234` to `ABC-23`. The cap lives in `normalizeCode`, after the dash is stripped. Caught by a component test.
- No auto-formatted dash in the code field: it would move the caret on every keystroke, and backspace jumping to the end is unrecoverable for this audience. The hint text carries the printed form.
- Sessions expiring mid-visit are handled in the layout components as well as in `beforeLoad`: a guard only re-runs on navigation, and a revoked session arrives as a 401 on whatever query was running.
- Redirect targets from the URL are validated as single-slash internal paths, so `?redirect=https://…` cannot bounce a guest off the site straight after login.
- The line F5 draws: an admin edits anything **we** record, nothing the household **answered**. Transport seats, `has_stroller`, `seating_need` and `dietary_note` are on our side — they get told to us on the phone. `attending`, `meal_choice`, `portion`, `midnight_snack` and `rsvp_note` stay with F3, so the field set has one definition.
- Guests are soft-deleted, households are not. A guest was counted and may hold a seat; a household that never existed leaves nothing behind but its audit trail.
- Code collisions are handled by retrying a failed insert against the `UNIQUE` index, capped at 5. Checking first is redundant when the database already enforces it, and an unbounded retry hides a broken generator behind a hung request.
- CSV exports are UTF-8-with-BOM, semicolon-delimited, quoted throughout — a decision about German Excel rather than about correctness, recorded as such in `F5-B04`.
- Exports are logged, not audited: `audit_log.action` has no `read` value, and adding one would turn that table from a record of changes into a general event log.
- `validator/v10` arrives with `F5-B01`, not `F3-B03`. The comment in `httpio/respond.go` names the wrong story and must be updated when F5-B01 ships.
- Only F5 written, per the just-in-time rule in `CLAUDE.md`. F3 is drafted after F5 is built.
- **Admin RSVP is the guest page addressed by household id**, not a second admin form — `F3-B06` / `F3-F06`, appended to the index. One form component, one use case, parameterised by *which* household rather than by *how you authenticated*. Four consequences for the unwritten F3 stories are recorded in `TODO.md` under "Carried into F3": the form takes data and save as props, the deadline becomes an argument rather than an internal rule, an admin edit audits as `admin`, and it sets `rsvp_submitted_at`.
- `guests.csv` is a full table dump — every guest and household column, soft-deleted rows included with `deleted_at` third. It is therefore not a headcount, and the story and the download label both say so; `F6-B05` is the counted one.
- Remaining F5 assumptions confirmed unchanged: session revocation on regeneration, CSV encoding, `codes.csv` layout, hard-delete for households, German admin URLs.

Time: 4h
Cost: 37.72

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
- `E0-10` — three-stage `Dockerfile`: node 24 + corepack bundle, `golang:1.26-alpine` with `CGO_ENABLED=0`, `distroless/static-debian12:nonroot` runtime. 23.1 MB, no shell, UID 65532.
- `/data` and `/data/photos` staged in the build stage and copied with `--chown=nonroot`, so a fresh named volume inherits writable directories — the runtime image has no shell to `mkdir` with.
- Compose stays out of the repo entirely: one `/data` volume, `restart: unless-stopped`, no published ports. What the file must contain is documented in the README instead.
- `.dockerignore` covering `node_modules`, `web/dist`, `.git`, `local/`, `*.db*`, `.env`.
- `Makefile`: `docker-build` / `docker-push` targets against `server-andreas.local:5000`.
- Verified end to end: build from a clean context, `compose up` migrates and answers `/api/health`, restart preserves the database, a missing `ADMIN_PASSWORD` fails at startup with the `E0-02` message.
- Root README: "Running it" rewritten for the real container path, registry push and the `insecure-registries` requirement.
- `E0-11` — `newTestApp(t)` is now the single entry point: temp-file SQLite, migrations, the production router and a cookie-jar client, all cleaned up via `t.Cleanup`. Every existing integration test migrated onto it; `newTestServer`/`newTestDatabase` are gone.
- Request helpers on the app (`get`, `postJSON`) return a `testResponse` with the body already read — `Status`, `ContentType`, `Body`, `decodeJSON`, `errorEnvelope`.
- `fixtures_test.go`: `seedHousehold(t, pool, withCode/withDisplayName/withAdminNote/withGuests/withAdult/withChild)`, plus the low-level `insert*` helpers moved over from `schema_test.go`. Codes come from a counter over the F1-B01 alphabet.
- `assertNoLeak` walks the decoded JSON and rejects `code`, `admin_note` and the budget columns by field name, with `$.error.code` the one documented exception; extra secret *values* can be passed in.
- `harness_smoke_test.go` tests the harness itself: smoke, real file not `:memory:`, temp DB gone after the test, two parallel subtests isolated, fixture defaults and options, and `findLeak` proven to fire on each private field.
- `E0-12` — the app is deployed and reachable over HTTPS through the real proxy, under the path prefix `/hochzeit`.
- Path prefix wired through three places, each from one literal: Vite `base` (bundle asset URLs), the proxy's `handle_path` (strips it), `PUBLIC_BASE_PATH` (scopes the session cookie). Router `basepath` and the new `web/src/lib/api.ts` both read `import.meta.env.BASE_URL`.
- `middleware.StripPublicBasePath` lets the binary answer prefixed URLs itself, so `make preview` and a direct curl behave like the proxied site. `NewRouter` now takes the `Config`.
- `PUBLIC_BASE_PATH` added to config with normalisation (leading slash, no trailing) and tests; a non-path value fails startup rather than silently widening the cookie's scope.
- Integration tests: the embedded `index.html` must reference assets under the prefix, and the prefixed paths must answer like the stripped ones.
- README: deploy procedure, the Caddy block, what the deployment must provide, and how to derive `TRUSTED_PROXY_CIDRS`. `compose.example.yaml` deleted — the Compose file describes a machine, not the app.
- Image pushed to `server-andreas.local:5000/wedding:latest` (digest `sha256:3e625cdf…`). Docker here is Rancher Desktop, so `insecure-registries` and the `10.0.0.45 server-andreas.local` hosts entry had to go inside the lima VM via `rdctl shell`, not on the host.
- Path prefix removed again: the app got its own subdomain. Gone are `PUBLIC_BASE_PATH` and its config normalisation, `middleware.StripPublicBasePath`, Vite's `base`, the router `basepath`, `web/src/lib/api.ts` and the two prefix integration tests. `NewRouter` takes `(logger, database)` again; the dev proxy forwards `/api` unchanged.
- README, `04-architecture`, `06-privacy-security`, `.env.example`, `TODO.md` and the `E0-12` story updated to the subdomain deployment; Caddy block is now a plain site block.

Decisions:

- Latin-Extended is bundled alongside Latin: guest surnames carry Czech, Polish and Hungarian diacritics. Rationale in `src/styles/fonts.css`.
- Root font size is `1.125rem`, so Tailwind's spacing unit is 4.5px rather than 4px; type and spacing scale together and the px values in `05-design.md` are read as ratios. Rationale in `src/index.css`.
- `src/routeTree.gen.ts` is committed — `pnpm build` type checks before Vite generates it. Noted in `web/README.md`.
- **oxlint + prettier, no eslint.** Rationale in `web/README.md` and `CLAUDE.md`; revisit only if a type-aware rule would have caught a real bug.
- Versions stay as `^` ranges pinned by `pnpm-lock.yaml`; `E0-10` must install with `--frozen-lockfile`, and `pnpm audit` joins `govulncheck` in the pre-deploy check.
- Embed lives in `web/embed.go`, not next to the handler: an embed directive cannot reach outside its package directory. Package named `frontend` to avoid colliding with `infrastructure/web`.
- SPA fallback answers 200, and nothing outside `/api` ever 404s — the handler cannot tell a typo from a client-side route. Rationale in `static.go`.
- Content types come from an explicit table, not `mime.TypeByExtension`, whose answers depend on the image's `/etc/mime.types`.
- The harness lives in `harness_test.go`, not the `harness.go` the story named: it is only ever used by tests, and a non-test file would compile it into the package proper.
- `assertNoLeak` splits into a pure `findLeak(body) string` plus a thin assertion, so the privacy check is itself testable without a fake `*testing.T`.
- `google/go-cmp` deliberately not added yet: nothing so far needs a struct diff `testify` cannot express, and an unused dependency is a decision nobody made.
- The login/cookie helper the story asks for is deferred to `F1-B04`, which is where the endpoint appears; the cookie jar is already in the client.
- Site lives at `/hochzeit`, not `/apps/wedding`: the URL is printed on the invitation card, so it is user-facing text and follows the German rule.
- A path prefix rather than its own hostname — accepted cost: the frontend carries a base path, and the cookie must be scoped so it is not sent to the neighbouring apps on that hostname.
- `TRUSTED_PROXY_CIDRS` cannot be verified until `F1-B05` resolves `X-Forwarded-For`; the test-plan item stays open in the story file with that note.
- The image is built from a named registry path (`server-andreas.local:5000/wedding`) rather than a bare name: the server's registry is plain HTTP, so both daemons need it in `insecure-registries`. Noted in the README.
- Compose takes `TRUSTED_PROXY_CIDRS` with `${VAR?}` and not `${VAR:?}` — empty is a meaningful value ("trust no proxy"), only an absent variable is an error.
- No `HEALTHCHECK` in the image: distroless has no shell or curl to run one. Liveness is the reverse proxy's job against `/api/health`.
- Named volume rather than a bind mount, since all state lives under one `/data` root either way and the backup story is unchanged.
- TanStack router and query devtools mounted from `src/components/Devtools.tsx` behind a dynamic `import()`, not a static import guarded by `import.meta.env.DEV` — a static import survives dead-code elimination and would ship the panels to guests. Verified absent from `dist/`.
- Own subdomain instead of a path prefix, reversing the E0-12 decision above. Session cookie goes back to `Path=/` (nothing else lives on that hostname); rationale in `06-privacy-security.md` and the README's "Where it is served".
- `web/src/lib/api.ts` deleted rather than kept returning `/api${path}`: with no prefix it wrapped nothing and had no callers.

Time: 6h
Cost: $38.17

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
Cost: $9.45

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
Cost: $4.11

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
Cost: $11.31

## 2026-08-21

Done:

- Project set up from scratch. `CLAUDE.md` plus the first spec batch: `01-vision-scope.md`, `02-features.md`, `03-data-model.md`, `04-architecture.md`, and the initial TODO list.

Decisions:

- Go single binary with an embedded React/Vite frontend, SQLite, trimmed hexagonal layout.

Time: 3h
Cost: $6.57
