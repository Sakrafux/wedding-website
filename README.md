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

The container path is not built yet — `E0-10` adds the Dockerfile and Compose file, and this is the intended shape:

```bash
cp .env.example .env      # then fill it in
docker compose up -d      # migrates on startup, serves on $PORT
```

Configuration is environment variables only, no config file. Required: `DB_PATH`, `PHOTO_DIR`, `ADMIN_USER`, `ADMIN_PASSWORD`. Optional: `PORT`, `LOG_LEVEL`, `SESSION_COOKIE_SECURE`, `TRUSTED_PROXY_CIDRS`. A missing required variable is a hard failure at startup, not a silent default — the error names every problem at once, so one restart tells you everything that is wrong.

## Local development

Requires Go 1.26. Nothing else yet; pnpm and Node join once `E0-08` scaffolds the frontend.

```bash
cp .env.example .env                     # gitignored: it holds ADMIN_PASSWORD
mkdir -p local/photos                    # matches the DB_PATH/PHOTO_DIR below

set -a && . ./.env && set +a             # nothing in the app reads .env itself
go run ./cmd/wedding
```

For local runs set `DB_PATH=./local/wedding.db`, `PHOTO_DIR=./local/photos` and `SESSION_COOKIE_SECURE=false`. `/local/` is gitignored. The `Secure` flag has to come off because the browser refuses to send a `Secure` cookie over `http://localhost`, which makes login fail with no visible error.

Nothing loads `.env` automatically — there is no dotenv dependency, on purpose, since the deployed container gets its environment from Compose. Export it yourself as above, or pass the variables inline.

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
go test ./...                            # unit + integration, no external services
gofmt -l . && go vet ./...
```

Once the frontend exists, local development runs the Go API and the Vite dev server side by side, with Vite proxying `/api`; the production path embeds `web/dist` into the binary instead.

## Operating it

All durable state lives under `DB_PATH` and `PHOTO_DIR`. Back up that one directory and you have backed up everything.

**Never `cp` a live SQLite file.** WAL mode means a plain copy can capture a torn state that looks like a valid backup. Use `VACUUM INTO '/backup/wedding-YYYYMMDD.db'`.

Rehearse a restore before the invitations go out. After send-out the guest list is irreplaceable, and an untested backup is not a backup.

The database contains household login codes in plaintext — deliberately, because a code must be readable back for a guest who loses their card. Keep dumps and backups off shared drives, and don't commit one.

## Licence

None. Private project, not intended for reuse.
