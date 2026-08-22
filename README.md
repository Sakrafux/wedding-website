# Wedding Website

A private, invitation-only web app for our wedding. Guests log in with a code printed on their invitation card, find out what is happening, tell us whether they are coming, and later see where they sit and look at the photos. For us it doubles as the planning tool — guest list, RSVP overview, seating and budget in one place instead of three spreadsheets.

Roughly 80 guests across 60 households. German only. Mobile-first. Not public, not indexed, no accounts, no email.

**Status: specification phase. No application code yet.** The specs are complete enough to build from; `specification/features/` is where that starts.

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

Not runnable yet — this section is the intended shape, and gets verified as `E0-01` … `E0-12` land.

```bash
cp .env.example .env      # then fill it in
docker compose up -d      # migrates on startup, serves on $PORT
```

Configuration is environment variables only, no config file. Required: `DB_PATH`, `PHOTO_DIR`, `ADMIN_USER`, `ADMIN_PASSWORD`. Optional: `PORT`, `LOG_LEVEL`, `SESSION_COOKIE_SECURE`, `TRUSTED_PROXY_CIDRS`. A missing required variable is a hard failure at startup, not a silent default.

Local development runs the Go API and the Vite dev server separately; the production path embeds `web/dist` into the binary.

## Operating it

All durable state lives under `DB_PATH` and `PHOTO_DIR`. Back up that one directory and you have backed up everything.

**Never `cp` a live SQLite file.** WAL mode means a plain copy can capture a torn state that looks like a valid backup. Use `VACUUM INTO '/backup/wedding-YYYYMMDD.db'`.

Rehearse a restore before the invitations go out. After send-out the guest list is irreplaceable, and an untested backup is not a backup.

The database contains household login codes in plaintext — deliberately, because a code must be readable back for a guest who loses their card. Keep dumps and backups off shared drives, and don't commit one.

## Licence

None. Private project, not intended for reuse.
