# Wedding Website

A private, invitation-only web app for our wedding. Guests log in with a code printed on their invitation card, find out what is happening, tell us whether they are coming, and later see where they sit and look at the photos. For us it doubles as the planning tool — guest list, RSVP overview, seating and budget in one place instead of three spreadsheets.

Roughly 80 guests across 60 households. German only. Mobile-first. Not public, not indexed, no accounts, no email.

**Status: early build.** The specs are complete enough to build from; progress lives in `specification/features/README.md`. Currently in `E0` — project setup, nothing user-visible yet.

## What it does

| | |
|---|---|
| **Login** | One code per household, printed on the card. No username, no password, no email. Session lasts a year |
| **Content** | Schedule, venue, travel, dress code, gifts, FAQ — hardcoded in components, no CMS |
| **RSVP** | Per household, per member: attendance scope, meal, portion, allergies, seating needs, transport |
| **Plus-ones** | Households add their own companions and children; we see the delta |
| **Admin** | Guest list, RSVP dashboard, CSV exports, seating, budget, photo moderation |
| **Seating** | Hand-drawn floor-plan SVG; admins assign, guests see their own table |
| **Gallery** | Curated before the wedding, guest uploads after it |

## How it is built

One Go binary that serves both the JSON API and the embedded React frontend, in one Docker container, behind an existing reverse proxy that terminates TLS. One SQLite file and one photo directory hold all durable state.

**Backend** — Go, `chi`, `sqlx`, `modernc.org/sqlite` (pure Go, no CGO), `validator/v10`. No ORM, no JWT, no external services.

**Frontend** — React, TypeScript, Vite, TanStack Router + Query, Tailwind, shadcn/ui. Built with **pnpm**, embedded into the binary via `go:embed`, so a frontend/backend version mismatch is structurally impossible.

**Layout** — trimmed hexagonal: `domain` imports nothing internal, `application` holds no SQL and never imports `web`, handlers parse and delegate but contain neither. Responses are always explicit DTOs, never serialised domain structs — that is a privacy control, since `household.code` and `admin_note` must never reach a guest.

```text
cmd/wedding/              entrypoint: config, wiring, migrate, serve
internal/domain/          entities, enums, invariants — pure, no dependencies
internal/application/     use cases
internal/infrastructure/  web · persistence · configuration · security · photo
tests/integration/        API tests against a real temp-file SQLite
web/                      React app; dist/ is the go:embed target
specification/            the documents below
```

## Documentation

Read in order. Each one records the decisions *and* the rejected alternatives, so the reasoning survives longer than anyone's memory of it.

| Document | What is in it |
|---|---|
| [01 — Vision & Scope](specification/01-vision-scope.md) | Purpose, audience, principles, what is deliberately out of scope |
| [02 — Features](specification/02-features.md) | F1–F11 with priorities. The map of the product |
| [03 — Data Model](specification/03-data-model.md) | Schema, enums, invariants, derived views, rejected fields |
| [04 — Architecture](specification/04-architecture.md) | Stack, layering, API surface, auth, SQLite config, migrations |
| [05 — Design](specification/05-design.md) | Colour, typography, spacing, components, accessibility, German label map |
| [06 — Privacy & Security](specification/06-privacy-security.md) | What data we hold and why, retention, sessions, headers, deliberate non-defences |
| [07 — Roadmap](specification/07-roadmap.md) | Milestones, sequencing gates, what must ship before invitations go out |
| [Feature backlog](specification/features/README.md) | Epic and story checklist — the build tracker |
| [TODO](TODO.md) | Planning-level: open decisions, facts we are waiting on, spec debt |
| [CLAUDE.md](CLAUDE.md) | Conventions and decided facts, for humans and for Claude Code |

## Working on it

Build order lives in [specification/features/README.md](specification/features/README.md). Epics are directories, stories are files, and each story carries its own instructions and test plan. Backend story first, frontend story second — the backend defines the contract, the frontend consumes it.

Conventions that matter:

- Spec and code in English. **All user-facing text in German**, informal *du*. German strings live in `web/src/lib/labels.ts`, never inline.
- **Enum values are English everywhere** — database, API, Go, TypeScript. German exists only as display labels.
- Markdown is not manually wrapped. One paragraph, one line.
- Record rejected options and the reason, not just the chosen one.
- Story IDs are permanent. Never renumber.

## Running it

```bash
cp .env.example .env                      # then fill it in
cp compose.example.yaml compose.yaml      # then adapt it to the server
docker compose up -d                      # migrates on startup, serves on 127.0.0.1:8080
```

Both copies are gitignored. `.env` holds `ADMIN_PASSWORD` in plaintext, and the real `compose.yaml` describes the server — networks, proxy wiring, paths — which is deployment detail, not app detail. `compose.example.yaml` is the committed shape.

Configuration is environment variables only, no config file. Required: `DB_PATH`, `PHOTO_DIR`, `ADMIN_USER`, `ADMIN_PASSWORD`. Optional: `PORT`, `LOG_LEVEL`, `SESSION_COOKIE_SECURE`, `TRUSTED_PROXY_CIDRS`. A missing required variable is a hard failure at startup, not a silent default — the error names every problem at once, so one restart tells you everything that is wrong.

The image sets `DB_PATH` and `PHOTO_DIR` itself, both under the single `/data` volume; Compose supplies the rest and refuses to start if `ADMIN_USER`, `ADMIN_PASSWORD` or `TRUSTED_PROXY_CIDRS` are absent from the `.env` next to it. `TRUSTED_PROXY_CIDRS` may be empty — that means "trust no proxy" — but it has to be written down.

The port is published on loopback only. The reverse proxy is the sole way in; binding `0.0.0.0` would expose the app on the LAN without TLS. Override with `BIND_ADDR` / `HOST_PORT` if the proxy lives on another interface or in another Docker network.

The image is a three-stage build — pnpm bundle, `CGO_ENABLED=0` Go binary, `distroless/static:nonroot` runtime — and lands at roughly 23 MB with no shell in it. It carries no `HEALTHCHECK` for that reason; `GET /api/health` is there for the proxy to call.

```bash
make docker-build         # build the image
make docker-push          # push to server-andreas.local:5000
```

The registry speaks plain HTTP, so both the pushing and the pulling Docker daemon need it under `insecure-registries` in `/etc/docker/daemon.json`.

## Local development

Requires Go 1.26 and Node 24 with corepack. The frontend's package manager is pinned in `web/package.json`; `corepack enable` is what makes your shell honour that pin instead of whatever pnpm happens to be on your `PATH`.

```bash
cp .env.example .env                     # gitignored: it holds ADMIN_PASSWORD
mkdir -p local/photos                    # matches the DB_PATH/PHOTO_DIR below

make run                                 # exports .env, then go run ./cmd/wedding
```

The `Makefile` is the command list: `make build` builds the frontend and *then* the binary that embeds it (that order is the point — a stale `web/dist` baked into a fresh binary is the skew the embed exists to rule out), `make test` runs tests, `gofmt` and `go vet`, `make lint` and `make fmt` cover both languages.

For local runs set `DB_PATH=./local/wedding.db`, `PHOTO_DIR=./local/photos` and `SESSION_COOKIE_SECURE=false`. `/local/` is gitignored. The `Secure` flag has to come off because the browser refuses to send a `Secure` cookie over `http://localhost`, which makes login fail with no visible error.

Nothing loads `.env` automatically — there is no dotenv dependency, on purpose, since the deployed container gets its environment from Compose. Export it yourself as above, or pass the variables inline.

The frontend runs as a second server, not as part of the Go build:

```bash
corepack enable                          # once
cd web && pnpm install
pnpm dev                                 # http://localhost:5173
```

Open **5173**, not 8080. Vite serves the app and proxies `/api` through to the Go process on 8080, so the browser sees a single origin — which is not just convenience: the session cookie is `HttpOnly; SameSite=Lax`, and a two-origin dev setup would need CORS and a weaker cookie policy than production ever uses. Port 8080 on its own serves the API and whatever was last built into `web/dist`, embedded at compile time — stale unless you just ran `make build`.

Frontend checks, run from `web/`:

```bash
pnpm build                               # type check, then bundle into dist/
pnpm lint                                # oxlint
pnpm format                              # prettier
```

More detail, and why each dependency is there: [web/README.md](web/README.md).

Check it is up:

```bash
curl -s localhost:8080/api/health        # {"status":"ok"} — liveness, no dependencies
curl -s localhost:8080/api/ready         # {"status":"ok","database":"ok"}, 503 if not
curl -si localhost:8080/api/nope         # JSON 404 envelope, not chi's HTML
```

`/api/ready` is the one that proves the process opened the database it was configured with; `/api/health` deliberately never fails on a subsystem, so a restart policy cannot loop on it.

Startup logs the whole config as JSON with `ADMIN_PASSWORD` redacted, so the first log line tells you which values the process actually got. `LOG_LEVEL=debug` is the useful setting locally. Ctrl-C exercises the real graceful-shutdown path rather than killing the process.

Tests:

```bash
make test                                # go test ./..., gofmt, go vet
```

The integration tests that assert the SPA fallback skip when the binary was built without a frontend bundle, so run `make build-web` (or `make build`) at least once to exercise them.


## Operating it

All durable state lives under `DB_PATH` and `PHOTO_DIR`. Back up that one directory and you have backed up everything.

**Never `cp` a live SQLite file.** WAL mode means a plain copy can capture a torn state that looks like a valid backup. Use `VACUUM INTO '/backup/wedding-YYYYMMDD.db'`.

Rehearse a restore before the invitations go out. After send-out the guest list is irreplaceable, and an untested backup is not a backup.

The database contains household login codes in plaintext — deliberately, because a code must be readable back for a guest who loses their card. Keep dumps and backups off shared drives, and don't commit one.

## Licence

None. Private project, not intended for reuse.
