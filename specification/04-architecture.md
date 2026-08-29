# 04 — Architecture

Status: draft · Last updated: 2026-08-21

## Topology

One Docker container on a personal server, behind an already-existing reverse proxy that terminates TLS.

```
Internet → Caddy (TLS, strips /hochzeit) → :8080 wedding container
                                   ├── Go binary (chi)
                                   │    ├── /api/*  JSON API
                                   │    └── /*      embedded SPA (index.html fallback)
                                   ├── /data/wedding.db      (mounted volume)
                                   └── /data/photos/         (mounted volume)
```

The Go process speaks plain HTTP internally. It does **not** manage certificates. It trusts `X-Forwarded-For` for client IPs (needed for login rate limiting) — but only when the request arrives from a configured trusted proxy CIDR, otherwise the header is spoofable and rate limiting becomes bypassable.

Single artifact: the React build is embedded into the binary via `go:embed`. There is no way to deploy a frontend that disagrees with its backend.

## Backend stack

Deliberately small and close to stdlib. This has to still build and run after a year of nobody touching it.

| Concern | Choice |
|---|---|
| Router | `go-chi/chi/v5` |
| Request logging | `go-chi/httplog/v2` |
| DB access | `jmoiron/sqlx` over `database/sql` |
| SQLite driver | `modernc.org/sqlite` (pure Go, no CGO) |
| Validation | `go-playground/validator/v10` |
| Crypto | `golang.org/x/crypto` |
| Images | one thumbnailing library, chosen when F9 is built |

Not used: `go-chi/cors` (same origin — the Go binary serves the SPA, so there are no cross-origin requests), any ORM, `pgx`.

Pure-Go SQLite means a static binary and a distroless/scratch runtime image. The performance gap to the C driver is irrelevant at 80 guests.

## Project layout

Trimmed hexagonal. The layering discipline is kept; the ceremony is not.

```text
.
├── cmd/wedding/                    # entrypoint: config, wiring, migrate, serve
├── internal/
│   ├── domain/                     # entities, enums, invariants. Imports no internal package.
│   │   ├── guest.go                #   RSVP rules, portion/seating/meal enums
│   │   ├── seating.go              #   assignment validity, stale detection
│   │   └── budget.go               #   rollup math (per-head, external_cents)
│   ├── application/                # use cases; orchestrates domain + stores
│   │   ├── rsvp.go
│   │   ├── seating.go
│   │   ├── budget.go
│   │   └── gallery.go
│   └── infrastructure/
│       ├── web/                    # router wiring: middleware chain + routes
│       │   ├── handler/            #   request handlers, one file per resource group
│       │   ├── httpio/             #   JSON + error-envelope writers
│       │   ├── dto/                #   explicit request/response types
│       │   ├── middleware/         #   session, admin gate, ratelimit, requestID
│       │   └── static.go           #   go:embed SPA + index.html fallback
│       ├── persistence/            # sqlx stores, concrete types, embedded migrations
│       ├── configuration/          # env parsing, sqlite pragmas/pools (a leaf package)
│       ├── security/               # session tokens, code normalization, admin compare
│       └── photo/                  # file storage, thumbnailing
├── tests/integration/              # API-level tests against a real temp SQLite
├── web/                            # React app (Vite); dist/ is the go:embed target
└── specification/                  # these documents
```

### Layering rules (the point of the pattern)

1. `domain` imports no other internal package. Business rules are pure functions over plain structs, testable with no HTTP and no database.
2. `application` never imports `web` and never contains SQL.
3. Handlers contain no business rules and no SQL — they parse, delegate, and format.

### Deliberate deviations from textbook hexagonal

- **No `application/port/` interfaces.** Ports exist to allow substitution. There will be exactly one database, forever, and one web adapter. An interface whose only implementations are the real store and a mock buys nothing. Extracting an interface later is a mechanical refactor in Go, not a rewrite — so this is deferred, not foreclosed.
- **No hand-maintained mocks.** Because the SQLite driver is pure Go, an integration test against a real temp-file database is both easier to write and stronger evidence than a mocked unit test. Mocks are also the artifact most likely to rot in a project touched once a month.
- **DTOs are kept — but for privacy, not purity.** `household.code` and `household.admin_note` must never appear in a guest's JSON. If handlers serialized domain structs directly, one added field would silently leak a login code. Explicit response types make that class of leak impossible rather than merely unlikely.
- **Handlers are resource-scoped structs, one file per group.** `System` (health, ready, `/api` fallbacks), later `Auth`, `RSVP`, `Admin` — each file holds the struct, its constructor and its endpoints as methods, dependencies passed in. Rejected: a single handler struct for the whole API, which would hold every store and service and let any endpoint reach any of them, so "which handlers can touch the budget" would stop being answerable from the type. Endpoints with no dependencies are methods too, so construction and route registration have one shape.
- **`web` is split into `handler/`, `httpio/`, `dto/`, `middleware/`.** A flat `web` package would end up around fifteen handler files next to the router, and — more decisive — `middleware` must reject requests in the same error envelope handlers use. Since `web` imports `middleware` to build the router, envelope writers living in `web` would be an import cycle. `httpio` holds them instead, so a `401` from the session gate and a `404` from a handler are the same shape by construction. `web` itself keeps only the wiring. The name is `httpio` and not `helper/` or `util/`, which exclude nothing and therefore collect everything; its success writer keeps the explicit `Write` prefix (`WriteJSON`) because a Go function called `Error` reads as the `error` interface method, and the error path is `RespondError`.
- **No JWT, no bcrypt in `security/`.** Sessions are opaque random tokens in a DB table: a household session must last a year and stay revocable, which stateless JWTs handle badly. Password hashing is moot with a single plaintext-env admin.

## Testing

Two layers, no mocking layer.

**Domain unit tests** cover the rules where a silent bug costs a real headcount or a wrong bill:

- Only `guest_added` members are deletable, and only before the RSVP deadline.
- `seating_need = 'with_parent'` consumes no seat and cannot hold an assignment.
- Attendance scope gates catering: church-only guests never appear in meal, portion, snack or seating counts.
- Flipping `attending` to `no` or `church_only` produces a *stale assignment*, never a silent unassignment.
- Budget rollup: per-head items track live headcount; `external_cents` is excluded from our own cost.
- Code normalization: case, whitespace, dashes; ambiguous glyphs never generated.

**Integration tests** drive the HTTP API against a real SQLite temp file with migrations applied, covering login, full RSVP submission and edit, plus-one addition past the soft cap, deadline lockout, admin authorization boundaries, and that guest-facing responses never contain `code`, `admin_note`, or any budget field.

Test dependencies: `stretchr/testify` (`require` for setup, `assert` where multiple failures are useful) and `google/go-cmp` for readable diffs on large structs like headcount rollups. Nothing else — stdlib `httptest` covers the HTTP side, and there is no interface to mock.

## API design

REST-ish JSON under `/api`. No GraphQL, no RPC framework.

- Session cookie auth; no bearer tokens, no `Authorization` header.
- Request bodies validated with `validator/v10`; validation failures return a field-keyed error object the frontend renders next to inputs.
- Errors: consistent JSON shape `{ "error": { "code": "...", "message": "...", "request_id": "...", "fields": { … } } }`. `message` is German and safe to show a guest verbatim; `request_id` is the short speakable ID also sent as `X-Request-Id` and logged with the request; `fields` appears only on validation failures.
- **Failures are `domain.Error` values carrying an `ErrorCode` and nothing else.** No message, no status: `httpio` owns the single table mapping code → HTTP status + German sentence, so that table is the complete list of what a guest can ever be shown. Transport-level failures (`not_found`, `method_not_allowed`, `not_ready`, `validation_failed`, `internal_error`) are codes in the same table, exposed as sentinel errors rather than as business codes, since routing and readiness are not business facts. A code missing from the table answers a generic 500 instead of improvising, and an error that is not a `domain.Error` is logged and answered generically — its text may hold SQL, a file path or a login code. Rejected: the usual `AppError{Type, Message}` whose `Message` is echoed to the client (as in the `realworld-tech-comparison` reference implementation) — it puts German UI copy behind the business rules and makes every leak a one-line mistake, and its `err.Error()` fallback leaks by default rather than failing closed.
- **`httpio` exports exactly two response functions:** `WriteJSON` for a success body and `RespondError(w, r, err)` for every failure. No exported writer takes a status, a code or a message, so a handler cannot phrase an error itself — the earlier `WriteError(status, code, message)` was a second write path and therefore a second chance to leak.
- Mutating endpoints require the session's household to own the target row — checked server-side on every request, never inferred from what the frontend sent.
- Health is split in two. `/api/health` touches no dependency: it is what the restart policy polls, and restarting cannot fix an unmounted volume or a wrong `DB_PATH` — it only replaces a legible error with a crash loop. `/api/ready` pings both database pools and answers 503 when either is unreachable; it is the post-deploy smoke check, and the reason for a failure stays in the log because the endpoint is unauthenticated.

Rough surface:

```
GET    /api/health                                  → liveness, no dependencies
GET    /api/ready                                   → dependencies usable, 503 if not

POST   /api/auth/login              { code }        → sets session cookie
POST   /api/auth/admin/login        { user, pass }  → sets admin session cookie
POST   /api/auth/logout
GET    /api/me                                      → household + members + flags

GET    /api/rsvp                                    → own household RSVP state
PUT    /api/rsvp                                    → attendance/meal/portion/needs/note
POST   /api/rsvp/members                            → add plus-one or child
DELETE /api/rsvp/members/{id}                       → only guest_added, pre-deadline

GET    /api/seating                                 → floor plan + own table, if published
GET    /api/gallery
POST   /api/gallery                                 → upload, if uploads_open

GET    /api/admin/households                        CRUD
GET    /api/admin/dashboard                         → headcounts, deltas, notes, stale seats
GET    /api/admin/export/guests.csv
GET    /api/admin/export/codes.csv                  → for variable-data printing
GET    /api/admin/seating                           CRUD tables + assignments
GET    /api/admin/budget                            CRUD budget items
GET    /api/admin/photos                            hide/delete, ZIP export
```

## Authentication and authorization

**Guests.** A household code is redeemed at `POST /api/auth/login`. Codes are normalized (uppercase, strip whitespace and dashes) before lookup. On success a random 256-bit token is generated; its hash is stored in `session`, the raw token goes into an `HttpOnly; Secure; SameSite=Lax` cookie with a 365-day lifetime and rolling refresh.

**Admin.** A single admin, credentials from environment variables, compared with `subtle.ConstantTimeCompare`. No `admin_user` table, no password reset flow, no admin CRUD. Changing the password means editing the env and restarting. Admin sessions are short-lived (hours) — a different risk profile from a guest session that must survive months.

**Authorization** is enforced in middleware by session subject type. Admin-only routes live under `/api/admin` and are rejected wholesale for household sessions. Budget endpoints exist only there. The frontend hiding a nav link is not a security control and is never treated as one.

**Rate limiting.** Per-IP limiting on both login endpoints (order of 10 failures/hour) with a friendly German error, never a lockout — locking out a confused guest is worse than the attack we are preventing. Failed attempts are written to `audit_log` as `login_failed`.

## SQLite configuration

Applied on every connection:

```
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;
```

`database/sql` pool capped at **one writer** connection (`SetMaxOpenConns(1)` for the write pool, or a single shared handle with serialized writes). SQLite has one writer; pretending otherwise produces `SQLITE_BUSY` under concurrent RSVP submissions. Reads can use a separate read-only pool.

## Migrations

Numbered `.sql` files embedded with `go:embed`, applied in order at startup inside a transaction, with the applied version tracked in a `schema_migration` table. No external tool, no manual deploy step — the container brings itself up to date or fails to start.

Forward-only. Down-migrations are not written; at this scale, restoring a backup is the honest rollback.

## Frontend stack

| Concern | Choice |
|---|---|
| Language | TypeScript |
| Build | Vite |
| Routing | TanStack Router — type-safe params, pairs with TanStack Query |
| Server state | TanStack Query |
| Styling | Tailwind CSS |
| Components | shadcn/ui (Radix + Tailwind, copied into the repo) |

shadcn/ui over Mantine specifically because it builds on Tailwind rather than shipping a competing style system, and because its Radix primitives are accessible by default — which matters given the guest demographic.

Page content (schedule, travel, dress code, FAQ, …) is hardcoded in components. No CMS, no Markdown pipeline.

**SPA fallback:** the Go static handler serves `index.html` for any unmatched path that is not under `/api` and does not look like a static asset. Deep links and refreshes work.

## Configuration

Environment variables only. No config file.

| Variable | Purpose |
|---|---|
| `PORT` | Listen port (default 8080) |
| `DB_PATH` | SQLite file, e.g. `/data/wedding.db` |
| `PHOTO_DIR` | Photo storage root, e.g. `/data/photos` |
| `ADMIN_USER` | Admin username |
| `ADMIN_PASSWORD` | Admin password, plaintext. Only readable with direct server access; accepted. |
| `SESSION_COOKIE_SECURE` | Off for local development only |
| `PUBLIC_BASE_PATH` | Public path prefix; scopes the session cookie |
| `TRUSTED_PROXY_CIDRS` | Whose `X-Forwarded-For` to believe |
| `LOG_LEVEL` | |

Absent required variables cause a hard fail at startup, not a silent default.

## Photo storage

Files on the mounted volume, never in SQLite. Stored under a content-addressed name; the user-supplied filename is metadata only, never a path component. Thumbnails are generated on ingest into a sibling directory and are regenerable — losing them is not data loss. Originals are kept byte-for-byte with EXIF intact.

## Logging and observability

`httplog` structured request logs to stdout, for the host's log collector. No APM, no metrics stack — for one container and 80 users, logs plus the audit table are enough. Errors carry a request ID that also appears in the user-facing error message, so a guest can read it out over the phone.

## Operational notes (outside the app)

Backups are handled on the server, not by the application. Two constraints the app guarantees to make that possible:

1. All state lives under the two mounted paths (`DB_PATH`, `PHOTO_DIR`). Nothing durable is written anywhere else.
2. **Never copy a live SQLite file.** With WAL enabled, `cp` can capture a torn state. Use `VACUUM INTO '/backup/wedding-YYYYMMDD.db'` or `sqlite3 ... ".backup"`, both of which are safe against a running writer.

A restore should be rehearsed once before invitations go out. An untested backup is not a backup.

## Rejected

| Option | Why not |
|---|---|
| nginx container serving static, Go serving `/api` | Contradicts the single-container goal for no benefit at this scale. |
| Frontend served from a mounted directory | Allows FE/BE version skew. Embedding makes it impossible. |
| `backend/` + `frontend/` top-level split | `go:embed` cannot cross the module root, so the SPA build would have to be copied into the Go tree before every build — an extra step and a stale copy waiting to happen in development. |
| `mattn/go-sqlite3` (CGO) | Needs a C toolchain in the build and a fatter runtime base; the speed is not needed. |
| ORM (GORM, ent) | Heavy abstraction over a ~10-table schema we fully understand. sqlx is the right altitude. |
| Full hexagonal with port interfaces + mocks | One DB and one web adapter forever; substitution is never exercised. Layering kept, ceremony dropped. |
| Mocking library / generated mocks | No interfaces to mock, and a real pure-Go SQLite in tests is better evidence. |
| JWT sessions | Household sessions must live a year and be revocable — a DB-backed opaque token does that; stateless JWTs do not. |
| `admin_user` table with CLI creation | Over-built for one admin at this stake level. Env vars chosen. |
| Down-migrations | Restoring a backup is the honest rollback path here. |
| Go-managed TLS (autocert) | A reverse proxy already exists and does this better. |
