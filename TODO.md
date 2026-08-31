# TODO — Remaining Planning Work

Status: as of 2026-08-31 · Building. E0 (setup), F1 (login) and F5 (admin households & guests) are done; see [specification/features/README.md](specification/features/README.md).

This tracks what is left **to plan**, not to build. Build work lives in [specification/features/README.md](specification/features/README.md).

## Documents still to write

- [x] [`specification/05-design.md`](specification/05-design.md) — written 2026-08-22. Warm/traditional, cream + sage + terracotta, Cormorant Garamond display / Source Serif 4 body, 18px base, no dark mode, informal "du". Includes the component inventory and the full German enum label map.
- [x] [`specification/06-privacy-security.md`](specification/06-privacy-security.md) — written 2026-08-22. Data inventory, GDPR posture (Art. 2(2)(c) household exemption, behave as if it applied anyway), retention and end-of-life procedure, "the database is a secret" and the leak-recovery path, code strength maths, rate limiting, sessions, response hygiene, headers/CSP, photo EXIF risk, logging rules, and a table of deliberate non-defences.
- [x] [`specification/07-roadmap.md`](specification/07-roadmap.md) — written 2026-08-22. M0–M9, **undated by design** — order is the plan. Three anchors only: wedding 2027-07-17, send-out Oct/Nov 2026, RSVP deadline ~2 months before the wedding. Splits features into must-ship-before-send-out and can-follow, names F6 as the release valve, and carries the facts/risk tables.
- [~] [`specification/features/`](specification/features/README.md) — epic directories with one file per story, tracked in `specification/features/README.md`. **Written: E0 setup (12 stories), F1 login (11 stories).** Every other epic exists in the index as bare checkboxes; detail gets written just before that epic is built. Order: F5 admin households, F3 RSVP, F4 plus-ones, F2 content, F11 quality, F6 dashboard, F8 budget, F9 gallery, F7 seating, F10 uploads.
- [x] Root [`README.md`](README.md) — written 2026-08-22. What it is, how it is built, the document index, conventions, and the intended run/operate commands. The "Running it" section is marked as unverified until `E0-12` lands, and needs a pass once it has.

## Decisions needing your input

### Wedding facts (block content, seating, and the roadmap)

- [x] Wedding date: **2027-07-17** assumed as the working date. Set by venue availability, so it becomes firm when the venue is booked — expected early September 2026. Alternatives were 07-16, 07-23, 07-24, all within nine days.
- [ ] Church and reception venues — names, addresses, travel notes. Being fixed within ~2 weeks of 2026-08-22; this also fixes the wedding date.
- [ ] Schedule of the day for the Ablauf page.
- [ ] Room/table layout for the reception, enough to draw the floor-plan SVG.
- [ ] Dress code wording; gift wishes and bank details.
- [ ] Caterer age brackets for children's pricing (e.g. 0–3 free / 4–12 reduced / 13+ full) — derived at read time, so this can land late, but the caterer decides the numbers.
- [x] Print shop does variable-data printing for the household codes — confirmed. F1 stands as specified; the code-slip fallback is not needed.

### Product decisions still open

- [x] Launch shape: **one launch with the invitations**, sent Oct/Nov 2026. No separate save-the-date — reasoning recorded in [02-features](specification/02-features.md) and [07-roadmap](specification/07-roadmap.md).
- [ ] RSVP deadline: **~2 months before the wedding**, so around mid-May 2027 — exact date still to pick, and it must be picked before send-out because it gets **printed on the card**. Also open: what the site shows after it passes.
- [ ] Default soft cap on guest-added members (proposed: 2).
- [x] Informal "du" vs formal "Sie" throughout the German copy. **Decided: du**, with neutral phrasing preferred where it reads naturally. See [05-design](specification/05-design.md).
- [ ] Does an admin-only "we think they're probably coming" state add value over plain no-answer? Guests would never see it.
- [ ] Track invitation send-out date per household, to time reminder nudges?
- [ ] Photo retention: how long does the gallery stay online after the wedding, and what happens to the files afterwards? Blocks the end-of-life procedure in [06-privacy-security](specification/06-privacy-security.md).
- [ ] Can the caterer export be name-free (counts + allergies keyed by table) instead of a full guest list? Would remove the largest planned data disclosure.
- [ ] **Contact phone number for the login fallback.** After two failed attempts the login screen offers "Klappt es nicht? Ruf uns an: …". `web/src/lib/labels.ts` carries a placeholder (`contactPhoneNumber`). Needed before send-out: it is the only escape hatch a guest has, and a placeholder reaching the print run strands exactly the people it exists for.
- [ ] Draft the German "Datenschutz" page text, alongside the F2 content pages.
- [ ] Guest upload quotas — file count and total size per household (proposed: 100 files / 2 GB).

### Answered against the F5 stories — 2026-08-31

- [x] **Admin-entered RSVP answers: yes, and as the same page.** Rather than a second form in admin, the admin gets the guest RSVP page addressed by household id. One form component and one use case, parameterised by *which* household instead of by *how you authenticated*. Recorded as `F3-B06` / `F3-F06`; the constraints it puts on the rest of F3 are under "Carried into F3" below.
- [x] **Regenerating a code revokes that household's sessions.**
- [x] **CSV encoding: UTF-8 BOM, semicolon-delimited, quoted.**
- [x] **`codes.csv` stays `haushalt;code`, with the code ungrouped (`ABC234`).** The dash was dropped entirely on 2026-08-31 — see `02-features` — so there is no separator for the print shop to get wrong. Still worth confirming the column names with them before `E-OPS-06`.
- [x] **Households hard-delete, guests soft-delete.** The asymmetry has a reason: only we delete a household, whereas a household removes its own plus-ones, and a person who was once counted has to stay explicable.
- [x] **`guests.csv` is a full dump** — every column of `guest` and of its household, soft-deleted rows included with `deleted_at` visible. It is therefore not a headcount, and both the story and the download label say so.
- [x] **German admin URLs.**
- [x] **F3 stories are written after F5 is built.** F5 is now built — F3's story files are the next planning work.
- [x] **No `formatted_code` field.** The reissue response returns `code` alone: the dash is gone, so the stored form is the printed form, and a second field would be the same six characters under another name. `F5-B03` updated on 2026-08-31.

### Raised in the 2026-08-31 code review

- [ ] **Admin settings endpoint.** The admin login response deliberately carries no flags, but the admin needs to *flip* `rsvp_open`, `seating_published`, `gallery_visible` and `uploads_open`. That is a read/write endpoint (`GET`/`PATCH /api/admin/settings`) and wants a story in F6 — not a read-only copy smuggled into the login body.
- [ ] **Split `internal/application` into `application/auth`, `application/rsvp`, …** once the second use case exists. It is what makes `auth.Bootstrap` possible instead of `application.Bootstrap`; doing it with one use case in the package buys nothing.
- [ ] **`web.Dependencies` growth.** Added 2026-08-31 with config, database and auth. Every new use case is a field there and a line in `main`, never a new `NewRouter` parameter.
- [x] **Dash dropped entirely, 2026-08-31.** Six characters need no grouping, and the separator was paying for a display format, a `FormatCode` function and an input field that had to decide what to do with a dash the guest typed. Codes are printed, exported and displayed as `ABC234`; input still strips dashes, because habit and word processors produce them.

### Dynatrace / F12 — open questions

- [ ] Which tenant, and a token scoped to ingest only (`openTelemetryTrace.ingest`, `metrics.ingest`, `logs.ingest`) — nothing that can read back.
- [ ] Traces must carry **no household id, no guest name and no login code** as attributes. Route patterns only, never resolved paths that embed an id.
- [ ] OneAgent in the image versus the OTel SDK in the binary: the SDK is more code and more learning, OneAgent is less of both and fights the single-static-binary design. Leaning SDK.
- [ ] Frontend RUM means loading a third-party script into a `script-src 'self'` CSP. Decide whether that trade is worth it *before* writing `F12-02`.

### Seating detail (for the F7 story)

- [ ] Who draws the floor-plan SVG, and in what tool? It needs stable `id` attributes per table shape — that constraint has to survive re-exports from the drawing tool.
- [ ] Printable seating output for the day itself: what does it look like, and is it produced by the app or by hand from the data?
- [ ] Does the phone view need a separate, zoomed SVG variant, or does pan/zoom on one drawing suffice?

### Design — resolved 2026-08-22, see [05-design](specification/05-design.md)

- [x] Stationery/colours/fonts to match: none exist. The site defines the palette.
- [x] Overall feel: warm/traditional.
- [x] Hero photography: engagement photos exist, hero ships from day one. Final selection and crop still open.
- [x] Dark mode: rejected.

### Technical loose ends

- [ ] Pick the Go thumbnailing library (deferred until F8/F9 is built).
- [x] Reverse proxy: Caddy, own subdomain on a shared Docker network. Snippet and `TRUSTED_PROXY_CIDRS` in the README deploy section.
- [x] Compose shape and volumes decided in `E0-10`/`E0-12`: one `/data` mount plus a photo mount, bind-mounted, documented in the README.
- [ ] CI: any, or local build and push only?

## Sequencing gates

Non-negotiable ordering, mostly to avoid painful migrations or reprints:

1. **Freeze the RSVP field set before invitations go out.** The field set grew three times during this planning session. Once real answers are in the database, every addition is a migration against live guest data. The RSVP form's shape is the last thing to change, not the first.
2. **Print the codes only after F1 is built and tested end to end.** A code format bug discovered after 60 cards are printed is unrecoverable.
3. **Rehearse a backup restore before invitations go out.** An untested backup is not a backup, and after send-out the guest list is irreplaceable.
4. **Seating work starts only after the RSVP deadline.** Assigning seats against a moving headcount wastes the effort twice.
5. **Guest uploads (F10) ship after the wedding**, not before. It is the one feature with no deadline pressure at all.

## Consistency debt in the current specs

Small things I know are loose, worth a pass before implementation:

- [x] `06-privacy-security.md` is referenced from `03-data-model.md` and `04-architecture.md` — now exists.
- [x] German UI label mapping for every English enum value — now in [05-design](specification/05-design.md), to be implemented as `web/src/lib/labels.ts`.
- [ ] The API sketch in [04-architecture](specification/04-architecture.md) predates the attendance-scope and transport fields; the RSVP payload shape needs a refresh once the F3 stories are written. `F1-B04` already defines the real `/api/me` shape, so the sketch is drifting.
- [ ] **Carried into F3, decided 2026-08-31 when `F3-B06` was agreed.** Write these into the F3 stories rather than rediscovering them:
  - The RSVP form component takes its data and its save mutation as **props** and never fetches for itself, so the guest route and the admin route can hand it different sources. This constrains `F3-F01`–`F3-F04`.
  - One `application` use case takes a household id. The guest handler passes the session's household; the admin handler passes the path parameter. Ownership is checked in exactly one place either way.
  - **Deadline enforcement is an argument to that use case, not a rule inside it** (`F3-B04`). The admin path exists for the call that comes in late, so it must be able to write after `rsvp_deadline`.
  - **An admin edit is audited as `actor_type = 'admin'`** (`F3-B05`). The audit log settles "but I said we were coming", and recording our own edit as the household's answer would mislead at the one moment it matters.
  - An admin edit sets `rsvp_submitted_at`, so the household drops off the nudge list. If we took it down on the phone, they have answered.
- [ ] No decision recorded on what a guest sees between "RSVP deadline passed" and "seating published" — a period where the site has little to say. Roughly five weeks in late spring 2027; [07-roadmap](specification/07-roadmap.md) wants this decided during M5.
