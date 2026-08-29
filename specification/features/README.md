# Feature Backlog

Status tracker for the whole build. **This file is the single source of truth for progress** — story files describe work, never their own status. If a checkbox here and a story file disagree, this file wins, because the other one is not allowed to have an opinion.

## How this is organised

Each epic is a directory. Each story is one file inside it.

```text
specification/features/
├── README.md              # this file
├── _TEMPLATE.md           # story template
├── E0-setup/              # project setup — not a feature, still required
├── F1-login/
├── F5-admin-households/
├── …
└── E-ops/                 # non-code gates: printing, restore rehearsal, send-out
```

**Story IDs.** `<EPIC>-<B|F><NN>` — `B` for backend, `F` for frontend. `F3-B02` is the second backend story of the RSVP epic. Epics that are not feature work (`E0`, `E-ops`) use a plain sequence instead, since the split does not apply.

**IDs are permanent.** Never renumber. To insert work later, append a higher number or use a gap — the ID appears in commit messages and in this index, and renumbering invalidates both.

**Backend leads, per slice.** Within an epic, a capability is a backend story followed immediately by the frontend story that consumes it. The backend story defines the endpoint and DTO shape; **the frontend story consumes that contract and does not invent fields.** That is what makes "backend first" mean something more than "the frontend finds out later that the API is wrong".

**Story size.** One story = one sitting. Independently mergeable, ends with a passing test. If it will not fit in an evening or two, it is an epic wearing a disguise — split it.

**Detail arrives just in time.** Only the epic being built has written story files. Everything else lives here as a bare checkbox with a title, so the tree is complete while the detail is still allowed to change. Writing F7's test plan today would encode assumptions that F1 has not made yet.

Epics map to the milestones in [07-roadmap](../07-roadmap.md). The roadmap says *when and why*; this file says *what is left*.

---

## E0 — Project setup · M0

Nothing user-visible. Exit criterion: the container runs on the real server, behind the real reverse proxy, over HTTPS.

- [x] `E0-01` Go module, package layout, chi router, httplog
- [x] `E0-02` Config from environment, hard fail on missing required vars
- [x] `E0-03` SQLite connection: pragmas, single-writer pool, read pool
- [x] `E0-04` Migration runner and `schema_migration`
- [x] `E0-05` Migration `0001` — full schema
- [x] `E0-06` Error envelope, request ID, panic recovery
- [x] `E0-07` Security headers, CSP, robots noindex
- [x] `E0-08` Frontend scaffold: Vite, TS, Tailwind, shadcn, design tokens
- [ ] `E0-09` `go:embed` of `dist/` and SPA fallback
- [ ] `E0-10` Dockerfile, Compose, volumes
- [ ] `E0-11` Integration test harness against temp-file SQLite
- [ ] `E0-12` Deploy to the real server behind the real proxy

## F1 — Household login · M1 · P0

- [ ] `F1-B01` Code generation and normalisation (domain)
- [ ] `F1-B02` Session store: create, hash, validate, refresh, revoke, purge
- [ ] `F1-B03` Session middleware and request auth context
- [ ] `F1-B04` `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/me`
- [ ] `F1-B05` Rate limiting and trusted-proxy client IP resolution
- [ ] `F1-B06` Audit logging for `login` and `login_failed`
- [ ] `F1-F01` Login screen and `CodeInput`
- [ ] `F1-F02` Household confirmation screen
- [ ] `F1-F03` App shell, route guards, TanStack Query setup
- [ ] `F1-B07` Admin login and `/api/admin` gate
- [ ] `F1-F04` Admin login screen and admin shell

## F5 — Admin: households & guests · M1 · P0

- [ ] `F5-B01` Household store and CRUD endpoints
- [ ] `F5-B02` Guest store and CRUD endpoints
- [ ] `F5-B03` Code generation and regeneration per household
- [ ] `F5-B04` `codes.csv` and `guests.csv` exports
- [ ] `F5-F01` Household list with last-login column
- [ ] `F5-F02` Household detail and member editing
- [ ] `F5-F03` Export actions

## F3 — RSVP · M2 · P0

**Gate 1 closes here: the RSVP field set is frozen.** Write the stories before building; every later field addition is a migration against live guest data.

- [ ] `F3-B01` Domain: attendance scope gates catering, all invariants
- [ ] `F3-B02` `GET /api/rsvp`
- [ ] `F3-B03` `PUT /api/rsvp` with validation
- [ ] `F3-B04` Deadline enforcement, server-side
- [ ] `F3-B05` Audit logging for every RSVP mutation
- [ ] `F3-F01` Form shell and household scope selector
- [ ] `F3-F02` Member cards with scope-gated catering fields
- [ ] `F3-F03` Transport seats and household note
- [ ] `F3-F04` Save, confirmation, and summary
- [ ] `F3-F05` Read-only rendering after the deadline

## F4 — Plus-ones & children · M2 · P0

- [ ] `F4-B01` Domain: soft cap, deletion rules, `guest_added` origin
- [ ] `F4-B02` `POST /api/rsvp/members`
- [ ] `F4-B03` `DELETE /api/rsvp/members/{id}`
- [ ] `F4-F01` Add-member sheet with cap hint
- [ ] `F4-F02` Remove member, pre-deadline only

## F2 — Informational content · M3 · P0

Frontend only — content is hardcoded in components.

- [ ] `F2-F01` Layout, navigation, bottom bar
- [ ] `F2-F02` Start: hero, greeting, countdown
- [ ] `F2-F03` Ablauf
- [ ] `F2-F04` Location, Anreise & Übernachtung
- [ ] `F2-F05` Dresscode, Geschenke
- [ ] `F2-F06` FAQ, Kontakt
- [ ] `F2-F07` Datenschutz

## F11 — Cross-cutting quality · M3 · P1

- [ ] `F11-01` Accessibility pass: keyboard, focus, contrast, 200% zoom
- [ ] `F11-02` German error messages end to end, request ID surfaced
- [ ] `F11-03` Mobile device QA on real hardware

## F6 — Admin: RSVP dashboard · M4 or M5 · P1

**The release valve.** If send-out gets tight, ship without this and use `guests.csv`.

- [ ] `F6-B01` Headcount queries by scope
- [ ] `F6-B02` Meal, portion, snack, age bracket, seating-need queries
- [ ] `F6-B03` Dietary list, transport gap, delta list, nudge list
- [ ] `F6-B04` `GET /api/admin/dashboard`
- [ ] `F6-B05` Caterer CSV export
- [ ] `F6-F01` Dashboard page with stat tiles
- [ ] `F6-F02` Note inbox with seen/unseen
- [ ] `F6-F03` Export actions

## F8 — Admin: budget · M5 · P1

- [ ] `F8-B01` Domain: rollup math, per-head resolution, `external_cents`
- [ ] `F8-B02` Budget item CRUD
- [ ] `F8-B03` Rollup endpoint and CSV export
- [ ] `F8-F01` Budget table
- [ ] `F8-F02` Rollup and totals view

## F9 — Curated gallery · M5 · P2

- [ ] `F9-B01` Photo storage, content-addressed names, thumbnailing library decision
- [ ] `F9-B02` Admin upload and list endpoints
- [ ] `F9-B03` Authenticated photo serving
- [ ] `F9-F01` Grid, lazy loading, lightbox

## F7 — Seating · M6 · P1

Blocked until the RSVP deadline has passed (Gate 4) and the floor-plan SVG exists.

- [ ] `F7-B01` Domain: assignment validity, stale detection
- [ ] `F7-B02` Seating unit and seat CRUD with `svg_element_id`
- [ ] `F7-B03` Per-seat assignment endpoints, church and party
- [ ] `F7-B04` Guest seating endpoint behind `seating_published`
- [ ] `F7-F01` SVG integration plus a test asserting every `svg_element_id` matches a shape
- [ ] `F7-F02` Admin seating unit and seat editor
- [ ] `F7-F03` Assignment UI and stale-assignment alert
- [ ] `F7-F04` Guest seating view with plain-text fallback
- [ ] `F7-F05` Printable seating output

## F10 — Guest uploads · M8 · P3

- [ ] `F10-B01` Upload endpoint: content sniffing, size cap, quota
- [ ] `F10-B02` Moderation: hide, delete
- [ ] `F10-B03` Bulk ZIP export
- [ ] `F10-F01` Upload dropzone with progress and retry
- [ ] `F10-F02` Moderation grid

## E-ops — Gates and non-code work

These are the steps that get skipped when they live only in prose. They are checkboxes for the same reason everything else is.

- [ ] `E-OPS-01` Seed all households and members from the real guest list
- [ ] `E-OPS-02` Generate codes, export `codes.csv`
- [ ] `E-OPS-03` **Gate 2** — F1 verified end to end on real hardware *before* printing
- [ ] `E-OPS-04` Proof-read a physical card, including the code in the final typeface
- [ ] `E-OPS-05` **Gate 3** — backup restore rehearsed into a fresh container
- [ ] `E-OPS-06` Print run
- [ ] `E-OPS-07` Send-out, and delete the code CSV from both ends
- [ ] `E-OPS-08` Post-send-out: watch `last_login_at`, phone the households that never appear
- [ ] `E-OPS-09` **Gate 4** — RSVP deadline reached, form read-only, seating may start
- [ ] `E-OPS-10` Floor-plan SVG drawn with stable ids
- [ ] `E-OPS-11` Final counts to the caterer
- [ ] `E-OPS-12` Publish seating, print plan, place cards, allergy sheet
- [ ] `E-OPS-13` Code freeze and verified backup before the day
- [ ] `E-OPS-14` Open guest uploads
- [ ] `E-OPS-15` **M9** — wind-down: archive, snapshot, delete volumes and stale copies
