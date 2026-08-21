# TODO — Remaining Planning Work

Status: as of 2026-08-21 · Planning phase, no code written yet.

This tracks what is left **to plan**, not to build. Implementation tasks come after the specification is complete.

## Documents still to write

- [ ] `05-design.md` — visual direction, typography, colour, layout patterns, component inventory, accessibility specifics. Needs your taste as input; blocked on the questions under "Design" below.
- [ ] `06-privacy-security.md` — what personal data we hold and why, GDPR posture for a private family site, retention and deletion, secret handling (the SQLite file contains plaintext login codes), rate limiting, session handling, what the audit log is for.
- [ ] `07-roadmap.md` — milestones anchored to the wedding date, with hard gates (see "Sequencing gates" below).
- [ ] `specification/features/*.md` — detailed user stories per feature, one file each. Written just before each feature is built, not all up front. Priority order: F1 login, F3 RSVP, F4 plus-ones, F5/F6 admin, F7 seating, F8 budget, F9/F10 gallery.
- [ ] `README.md` — human-facing summary. Written **last**, once the specs are stable.

## Decisions needing your input

### Wedding facts (block content, seating, and the roadmap)

- [ ] Exact wedding date. Everything time-based depends on it: countdown, RSVP deadline, "age at the wedding", roadmap milestones.
- [ ] Church and reception venues — names, addresses, travel notes.
- [ ] Schedule of the day for the Ablauf page.
- [ ] Room/table layout for the reception, enough to draw the floor-plan SVG.
- [ ] Dress code wording; gift wishes and bank details.
- [ ] Caterer age brackets for children's pricing (e.g. 0–3 free / 4–12 reduced / 13+ full) — derived at read time, so this can land late, but the caterer decides the numbers.
- [ ] Confirm the print shop actually does variable-data printing for the household codes. The whole F1 login design assumes it. Fallback if not: printed code slips inserted per envelope.

### Product decisions still open

- [ ] Launch shape: one launch with the invitations, or a "save the date" phase first?
- [ ] RSVP deadline date, and what the site shows after it passes.
- [ ] Default soft cap on guest-added members (proposed: 2).
- [ ] Informal "du" vs formal "Sie" throughout the German copy. (Proposed: du.)
- [ ] Does an admin-only "we think they're probably coming" state add value over plain no-answer? Guests would never see it.
- [ ] Track invitation send-out date per household, to time reminder nudges?
- [ ] Photo retention: how long does the gallery stay online after the wedding, and what happens to the files afterwards?
- [ ] Guest upload quotas — file count and total size per household (proposed: 100 files / 2 GB).

### Seating detail (for the F7 story)

- [ ] Who draws the floor-plan SVG, and in what tool? It needs stable `id` attributes per table shape — that constraint has to survive re-exports from the drawing tool.
- [ ] Printable seating output for the day itself: what does it look like, and is it produced by the app or by hand from the data?
- [ ] Does the phone view need a separate, zoomed SVG variant, or does pan/zoom on one drawing suffice?

### Design (blocks `05-design.md`)

- [ ] Do you have wedding stationery, colours, or fonts the site should match?
- [ ] Overall feel: warm/traditional, minimal/modern, playful?
- [ ] Photography available for hero imagery, or is it type-and-colour only at launch?
- [ ] Dark mode: worth it, or a distraction for this audience?

### Technical loose ends

- [ ] Pick the Go thumbnailing library (deferred until F8/F9 is built).
- [ ] Confirm the reverse proxy in use on your server, so `TRUSTED_PROXY_CIDRS` and the deployment notes are concrete.
- [ ] Decide the Docker Compose shape and volume paths for `DB_PATH` and `PHOTO_DIR`.
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

- [ ] `06-privacy-security.md` is referenced from `03-data-model.md` and `04-architecture.md` but does not exist yet.
- [ ] German UI label mapping for every English enum value is not written down anywhere. It should live in one place in the frontend, and probably be listed in `05-design.md`.
- [ ] The API sketch in `04-architecture.md` predates the attendance-scope and transport fields; the RSVP payload shape needs a refresh once the F3 story is written.
- [ ] No decision recorded on what a guest sees between "RSVP deadline passed" and "seating published" — a period where the site has little to say.
