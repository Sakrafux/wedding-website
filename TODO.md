# TODO — Remaining Planning Work

Status: as of 2026-08-22 · Planning phase, no code written yet.

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
- [ ] Draft the German "Datenschutz" page text, alongside the F2 content pages.
- [ ] Guest upload quotas — file count and total size per household (proposed: 100 files / 2 GB).

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
- [x] Reverse proxy: Caddy, path-routed as `/hochzeit` on a shared Docker network. Snippet and `TRUSTED_PROXY_CIDRS` in the README deploy section.
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
- [ ] No decision recorded on what a guest sees between "RSVP deadline passed" and "seating published" — a period where the site has little to say. Roughly five weeks in late spring 2027; [07-roadmap](specification/07-roadmap.md) wants this decided during M5.
