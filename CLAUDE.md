# CLAUDE.md

Private wedding web app. Currently in **specification phase** — no code yet. Specs live in `specification/`. Read those before proposing implementation.

## Decided facts

**Stack**

- Backend: Go. Single binary, serves both the JSON API and the frontend. Deps: `go-chi/chi/v5`, `go-chi/httplog/v2`, `jmoiron/sqlx`, `go-playground/validator/v10`, `golang.org/x/crypto`, `modernc.org/sqlite` (pure Go, no CGO). **No** `go-chi/cors` (same origin), no ORM, no pgx.
- Frontend: React + TypeScript + Vite, TanStack Router, TanStack Query, Tailwind, shadcn/ui. Package manager is **pnpm** (via corepack, version pinned in `package.json`) — never npm or yarn. Built `dist/` embedded via `go:embed` — one artifact, FE/BE skew impossible. SPA fallback to `index.html` for non-`/api` paths.
- DB: SQLite at `DB_PATH`, mounted volume. WAL, `foreign_keys=ON`, `busy_timeout=5000`. **Single writer connection** — do not pool writes, SQLite has one writer.
- Migrations: numbered `.sql` files via `go:embed`, applied in order at startup in a transaction, version tracked in `schema_migration`. Forward-only; no down-migrations.
- Deployment: one Docker container on the user's own personal server, behind an existing reverse proxy that terminates TLS. Go speaks plain HTTP; trusts `X-Forwarded-For` only from `TRUSTED_PROXY_CIDRS`.
- Admin: **single** admin, credentials from `ADMIN_USER` / `ADMIN_PASSWORD` env vars in plaintext. No `admin_user` table, no reset flow. Compare with `subtle.ConstantTimeCompare`.
- Config is env-vars only, no config file. Missing required vars = hard fail at startup.
- All durable state under `DB_PATH` and `PHOTO_DIR`. Backups are the user's server-side concern, not app code — but never `cp` a live WAL-mode SQLite file; use `VACUUM INTO`.
- No external SaaS in the critical path.

**Product**

- Audience: ~80 guests / ~60 households. German only. Mobile-first.
- Wide age range, low tech confidence → accessibility and "log in once" matter more than usual.
- Auth: per-household code, printed individually on the invite card (variable-data printing). Generic QR on the card points at the site root. Code is the only secret; no username, no password, no email.
- Code format: 6 chars from `23456789ABCDEFGHJKLMNPQRSTUVWXYZ` (no ambiguous glyphs), printed as `ABC-234`. Input normalized: uppercase, strip spaces/dashes.
- Session: `HttpOnly`, `Secure`, `SameSite=Lax` cookie, 365-day lifetime.
- RSVP unit is the **household**, not the individual.
- Households **can** add plus-ones and children themselves. Guest-added members are flagged as such and shown separately to admins. Soft cap, not a hard wall.
- Admin sessions are short-lived (hours); household sessions last 365 days with rolling refresh. Different risk profiles.
- Budget data is admin-only and must be enforced server-side.
- Page content (schedule, travel, dress code, FAQ, …) is **hardcoded in React components**. No CMS, no Markdown pipeline, no DB-backed text. Content change = rebuild + redeploy; accepted.
- Seating: layout is a **hand-drawn SVG** checked into the frontend, with stable `id` attributes per table shape. `seating_table.svg_element_id` links DB → shape. App only colors/labels existing shapes, never positions them. No drag-and-drop editor. Guests see the full plan but only their own table is labeled with names.
- Single-day wedding. No multi-day / arrival-day tracking.
- **All enum values are English** — DB, API, Go, TypeScript. German only as display labels mapped in the frontend.
- `guest.attending` is a **single** field carrying attendance *and* scope: `no` | `church_only` | `party_only` | `both`, NULL = unanswered. Contradictory states are unrepresentable. Scope is per guest (exceptions live inside households); the form offers a household-level bulk selector as a UI convenience only.
- **Scope gates catering.** `meal_choice`, `portion`, `midnight_snack`, and seating apply only to `party_only` / `both`. Every derived count keys off scope, never off "is attending" — otherwise we pay for meals nobody eats.
- Meal choice: `all` | `vegetarian` | `vegan`. Portion (orthogonal): `none` | `kids` | `full`, default `full` — `none` covers infants and adults not eating. `midnight_snack` is an independent boolean.
- Transport: `household.transport_seats_needed` / `transport_seats_offered` (church → reception). Used to compute a capacity gap for shuttle planning. The app does **not** match riders to drivers, and stores no phone numbers.
- `guest.age` (children only) means **age at the wedding date**, asked that way in the UI so it doesn't drift. Caterer age brackets derived at read time, not stored.
- `guest.seating_need` enum drives physical arrangements: `normal` | `with_parent` | `high_chair` | `stroller` | `wheelchair`. Applies to adults too.
- Money stored as `INTEGER` cents. Timestamps: UTC RFC3339 `TEXT` (readability when inspecting DB by hand beats epoch ints at this scale). Flags declared `BOOLEAN`.
- Photo originals kept byte-for-byte, **EXIF intact** — private gallery, and stripping breaks orientation + degrades the archive.
- Household login codes stored in **plaintext** (must be reprintable; accepted risk) → the SQLite file is a secret.

**Timeline**

- Wedding: **2027-07-17** (working assumption; alternatives are within nine days, so a change is a copy edit).
- **Invitations go out Oct/Nov 2026** — that, not the wedding, is the real deadline. F1/F3/F4/F5/F2/F11 must ship before it; F6/F7/F8/F9/F10 may follow. Dev target: done by end of 2026, ideally Sep/Oct.
- RSVP deadline ~2 months before the wedding (mid-May 2027). Milestones in `07-roadmap.md` are deliberately undated — order is the plan.

## Architecture rules

Layout is **trimmed hexagonal** (`cmd/`, `internal/domain`, `internal/application`, `internal/infrastructure/{web,persistence,configuration,security,photo}`, `tests/integration`, `web/`). Enforce:

1. `domain` imports no other internal package. Business rules are pure functions over plain structs.
2. `application` never imports `web` and contains no SQL.
3. Handlers contain no business rules and no SQL — parse, delegate, format.
4. **Never serialize domain structs to JSON.** Always explicit DTOs in `web/dto`. Reason is privacy, not purity: `household.code` and `admin_note` must never reach a guest response.

Deliberately absent: port interfaces, mocks, JWT, bcrypt, ORM. Add an interface only when something real needs substituting.

Tests: domain unit tests for RSVP/seating/budget invariants + integration tests over the HTTP API against a real temp-file SQLite. Deps: `stretchr/testify`, `google/go-cmp`. No mocking library.

## Conventions

- Spec and code comments in English. All user-facing UI text in German, informal **"du"**.
- **Documentation: names first, comments for decisions.** Variable and function names must be sensible and legible — a good name removes the need for a comment. Doc comments on non-trivial functions (exported or not): what it does, why it exists, non-obvious constraints. Inline comments where *appropriate*: non-trivial logic and, above all, **decisions** — why this threshold, why this order, why we deviate from the obvious approach. Also document unusual terms, settings, and magic values.
- **Do not** comment the self-evident (`== nil` checks, obvious loops, getters). Length is not a trigger: a long function that parses and destructures a request body is trivial and needs no commentary; a short function encoding a business decision does.
- **Deliberate omissions in DTOs must be commented.** Where a DTO leaves out a field on purpose (`household.code`, `admin_note`), say so and why — otherwise a later reader "fixes" the gap and leaks it. Same for any other privacy- or security-motivated absence.
- **A stale comment is a bug.** Change the code, update or delete the comment. A comment that lies is worse than no comment.
- **Forward references carry a story ID.** A comment describing work a later story will do names that story (`E0-06 replaces this`). Before ticking a story in `features/README.md`, run `grep -rn "<ID>" --include='*.go' .` and resolve every hit. A pointer to finished work is a stale comment, and nothing else prompts you to look — the code it describes still reads as correct.
- Go doc comments start with the identifier and are full sentences: `// ResolveHousehold returns the household for a login code, or ErrUnknownCode.` Keeps `go doc` output readable.
- Every migration `.sql` file opens with a header comment stating its intent. Migrations are forward-only with no down-migrations, so that comment is the only record of why the change happened.
- Design system (colours, type, components, German enum label map) lives in `specification/05-design.md`. German strings belong in `web/src/lib/labels.ts`, never inline in a component.
- **Markdown: no manual line wrapping.** One paragraph = one line. Let the editor soft-wrap.
- Feature specs: high-level overview in `specification/02-features.md`. Detail lives in `specification/features/<EPIC>/`, one directory per epic, one file per story, from `_TEMPLATE.md`.
- **Story IDs are permanent**: `<EPIC>-<B|F><NN>`, e.g. `F3-B02`. Never renumber — append or use a gap. `E0` (setup) and `E-ops` (gates) use a plain sequence.
- **`specification/features/README.md` is the only place build progress lives.** Checkboxes there; story files never carry a status line. Root `TODO.md` is the separate, planning-level list: open decisions, missing facts, spec debt — not implementation tasks.
- Backend story leads, frontend story follows, per slice. The backend story defines the endpoint and DTO shape; the frontend story consumes that contract and invents no fields.
- Story detail is written just before its epic is built. Unwritten stories exist as bare checkbox lines in the index — a tree written months early encodes assumptions the previous epic invalidates.
- Record rejected options and the reason, not just the chosen one.
- At the end of a session, add or extend today's entry in `JOURNAL.md` — newest first, `##` date heading, a `Done:` list, a `Decisions:` list (omit if nothing was decided), then a `Time:` line and a `Cost:` line per model. Entries are terse bullets, one line each. **Do not restate reasoning that lives in a spec, story file or code comment** — record that a decision was made and where it lives. Only describe work actually done in the session; leave time and cost as `<h>` / `$<x>` for the user to fill in.

## Threat model

Trusted guests (friends and family). Defend against guest **mistakes** and drive-by strangers — not a determined insider. Exception: admin-only data (budget) must be genuinely inaccessible to guest sessions.
