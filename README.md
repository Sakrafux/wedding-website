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

One container, behind a reverse proxy that terminates TLS. The Compose file lives on the server and is deliberately not in this repository: it describes a machine — host paths, networks, container names — not the app.

Configuration is environment variables only, no config file. Required: `DB_PATH`, `PHOTO_DIR`, `ADMIN_USER`, `ADMIN_PASSWORD`. Optional: `PORT`, `LOG_LEVEL`, `SESSION_COOKIE_SECURE`, `TRUSTED_PROXY_CIDRS`. A missing required variable is a hard failure at startup, not a silent default — the error names every problem at once, so one restart tells you everything that is wrong.

What the deployment has to provide:

- **Two writable directories**, mounted at `DB_PATH`'s parent and at `PHOTO_DIR`. 
- **A network the proxy can reach it on.** Nothing is published on the host: the proxy is not merely the intended way in but the only one. The container's name is what the proxy addresses, so pin it rather than letting Compose derive one from the directory.
- **`SESSION_COOKIE_SECURE=true`** — the proxy terminates TLS, so the process speaks plain HTTP while the cookie must still be marked `Secure`.

Both mounts hold durable state, so both belong in the backup.

### Where it is served

The site has its **own subdomain** and is served at the root, so nothing carries a path prefix: Vite builds asset URLs against `/`, the router matches at `/`, and Go serves the same paths the browser asks for. Earlier it lived under `/hochzeit` on a shared hostname; that is gone, along with `PUBLIC_BASE_PATH`, the Vite `base` and the prefix-stripping middleware.

The subdomain also makes `/robots.txt` ours, and the app serves it. Crawler exclusion is belt and braces: that file, the `X-Robots-Tag: noindex, nofollow` header from `E0-07` on every response, and the `<meta name="robots">` in `index.html`.

The image is a three-stage build — pnpm bundle, `CGO_ENABLED=0` Go binary, `distroless/static:nonroot` runtime — and lands at roughly 23 MB with no shell in it. It carries no `HEALTHCHECK` for that reason; `GET /api/health` is there for the proxy to call.

```bash
make docker-build         # build the image
make docker-push          # push to server-andreas.local:5000
```

The registry speaks plain HTTP, so the pushing daemon needs it under `insecure-registries` in `/etc/docker/daemon.json`. Pulling from the registry's own host over `localhost:5000` needs no such entry, since Docker trusts loopback by default.

### Reverse proxy

Caddy, terminating TLS for the app's own subdomain:

```caddyfile
<subdomain> {
    reverse_proxy wedding:8080
}
```

`wedding` is the container name, so that name is pinned in Compose. A site block for its own hostname needs no path handling at all — the app owns every path under it.

`reverse_proxy` sets `X-Forwarded-For` and `X-Forwarded-Proto` by default. Neither needs configuring, but both must survive — do not add a `header_up` that clears them.

### Deploy procedure

From the dev machine:

```bash
make docker-push                          # builds the frontend, the binary, the image; pushes to the registry
```

On the server, from a shell in the directory holding the Compose file:

```bash
docker compose pull
docker compose up -d
docker compose logs -n 20                 # migration lines, then "listening"
```

Then check from outside, over HTTPS:

```bash
curl -si https://<subdomain>/api/health   # 200 + JSON, security headers intact
curl -s  https://<subdomain>/rsvp | head  # the SPA shell, not a proxy 404
```

`ADMIN_PASSWORD` is stored in plaintext wherever the environment is defined — keep that file at mode `0600`, and generate the value rather than inventing one:

```bash
openssl rand -base64 24
```

`TRUSTED_PROXY_CIDRS` must cover the network the proxy reaches the container on:

```bash
docker network inspect <network> --format '{{(index .IPAM.Config 0).Subnet}}'
```

It cannot be fully verified yet: until `F1-B05` resolves `X-Forwarded-For`, the logged `remoteIP` is the direct peer — the proxy's own address — which is correct but proves nothing about the CIDR. Check it by observation when `F1-B05` lands: load the site from a phone on mobile data and confirm the logged client IP is the phone's. A wrong value makes every guest look like one client, so login rate limiting protects nothing while appearing to work.

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

### Test data

A fresh database has no households, so the login screen has nothing to let you in with. `make seed` inserts some:

```bash
make seed                                # one household, two members
make seed SEED_ARGS='-households 5 -guests 3'
```

It prints each household's id, name and login code — type one into the login screen. Codes are freshly generated, never fixed, so what you test with is shaped exactly like what gets printed on an invite.

**`cmd/seed` is a development tool only.** It writes obviously synthetic households to whatever `DB_PATH` points at and prints login codes — the app's only secret — in plain text. It is not in the image: the `Dockerfile` builds `./cmd/wedding` alone, and that absence is the whole guard. Real households come from the admin UI (`F5-B01`, `F5-B03`).

To start over, delete the file: `rm local/wedding.db*`. The command has no reset flag on purpose — one that read `DB_PATH` from the environment would be one shell mistake away from the deployed volume.

The frontend runs as a second server, not as part of the Go build:

```bash
corepack enable                          # once
cd web && pnpm install
pnpm dev                                 # http://localhost:5173/
```

Open **5173**, not 8080: Vite forwards `/api/…` to Go, so the browser talks to one origin exactly as it does in production. The single origin is not just convenience either: the session cookie is `HttpOnly; SameSite=Lax`, and a two-origin dev setup would need CORS and a weaker cookie policy than production ever uses. Port 8080 on its own serves the API and whatever was last built into `web/dist`, embedded at compile time — stale unless you just ran `make build`.

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
