# TODO — Remaining Planning Work

Status: as of 2026-09-01 · Building. E0 (setup), F1 (login), F5 (admin households & guests), F3 (RSVP) and F4 (plus-one) are done, and F2 (content) is built except for the facts it needs; see [specification/features/README.md](specification/features/README.md). A round of UI feedback on 2026-09-01 added thirteen stories across F3, F4, F5, F2 and F11 — all built the same day.

This tracks what is left **to plan**, not to build. Build work lives in [specification/features/README.md](specification/features/README.md).

## Documents still to write

- [x] [`specification/05-design.md`](specification/05-design.md) — written 2026-08-22. Warm/traditional, cream + sage + terracotta, Cormorant Garamond display / Source Serif 4 body, 18px base, no dark mode, informal "du". Includes the component inventory and the full German enum label map.
- [x] [`specification/06-privacy-security.md`](specification/06-privacy-security.md) — written 2026-08-22. Data inventory, GDPR posture (Art. 2(2)(c) household exemption, behave as if it applied anyway), retention and end-of-life procedure, "the database is a secret" and the leak-recovery path, code strength maths, rate limiting, sessions, response hygiene, headers/CSP, photo EXIF risk, logging rules, and a table of deliberate non-defences.
- [x] [`specification/07-roadmap.md`](specification/07-roadmap.md) — written 2026-08-22. M0–M9, **undated by design** — order is the plan. Three anchors only: wedding 2027-07-17, send-out Oct/Nov 2026, RSVP deadline ~2 months before the wedding. Splits features into must-ship-before-send-out and can-follow, names F6 as the release valve, and carries the facts/risk tables.
- [~] [`specification/features/`](specification/features/README.md) — epic directories with one file per story, tracked in `specification/features/README.md`. **Written: E0 setup (12), F1 login (11), F5 admin households (7), F3 RSVP (12), F4 plus-ones (5), F2 content (7), F11 quality (3).** The remaining epics exist in the index as bare checkboxes; detail gets written just before that epic is built. Order left: F6 dashboard, F8 budget, F9 gallery, F7 seating, F10 uploads.
- [x] Root [`README.md`](README.md) — written 2026-08-22. What it is, how it is built, the document index, conventions, and the intended run/operate commands. The "Running it" section is marked as unverified until `E0-12` lands, and needs a pass once it has.

## Decisions needing your input

### Wedding facts (block content, seating, and the roadmap)

- [x] Wedding date: **2027-07-17** assumed as the working date. Set by venue availability, so it becomes firm when the venue is booked — expected early September 2026. Alternatives were 07-16, 07-23, 07-24, all within nine days.
- [ ] Church and reception venues — names, addresses, travel notes. Being fixed within ~2 weeks of 2026-08-22; this also fixes the wedding date. **Now blocking three built pages:** `/location` teases both venues and carries the transfer text, and `/location/kirche` and `/location/feier` carry the address, the map link and the arrival section each — all saying "steht noch nicht fest" where the facts go (`locationLabels` in `web/src/lib/labels.ts`). The split landed 2026-09-01 (`F2-F09`); the facts are still the one content gap that blocks send-out, and `F2-F04` refuses to ship in that state.
- [ ] Schedule of the day for the Ablauf page. `/ablauf` ships thin, as 07-roadmap allows: four entries (Trauung, Empfang, Abendessen, Feier) with no times, each rendering "Uhrzeit steht noch nicht fest" rather than a guessed one. Entries live in `scheduleLabels`.
- [ ] Room/table layout for the reception, enough to draw the floor-plan SVG.
- [ ] Hotel and pension list for `/location/feier`. Part of the venue facts above, but it hangs off the party venue rather than the church — that is where the section now lives.
- [ ] Dress code wording; gift wishes and bank details. `/dresscode` and `/geschenke` are built with placeholder copy in `dresscodeLabels` / `giftLabels`; the gift page hides the IBAN block until the account is real, which it detects from the placeholder value.
- [ ] Caterer age brackets for children's pricing (e.g. 0–3 free / 4–12 reduced / 13+ full) — derived at read time, so this can land late, but the caterer decides the numbers.
- [x] Print shop does variable-data printing for the household codes — confirmed. F1 stands as specified; the code-slip fallback is not needed.

### Product decisions still open

- [x] Launch shape: **one launch with the invitations**, sent Oct/Nov 2026. No separate save-the-date — reasoning recorded in [02-features](specification/02-features.md) and [07-roadmap](specification/07-roadmap.md).
- [ ] RSVP deadline: **~2 months before the wedding**, so around mid-May 2027 — exact date still to pick, and it must be picked before send-out because it gets **printed on the card**. Also open: what the site shows after it passes.
- [x] **One plus-one, single-person households only.** Decided 2026-08-31, in two steps: first the soft cap became a hard cap of 2, then the number went away entirely. A household we seeded as one person may add one adult; everybody else adds nobody, and no guest may ever add a child. Refused server-side (`plus_one_not_allowed`, 409) with our number in the sentence; the admin path is uncapped and is how every other addition happens. Consequence accepted: guests have almost no control over their member list. `default_addition_limit` becomes dead configuration and is dropped in migration `0003`. Touched `CLAUDE.md`, `02-features`, `03-data-model`, `04-architecture`, `05-design`, `07-roadmap`, `features/README.md`, all five F4 stories, `F3-B02`, `F5-B02` and `domain/setting.go`.
- [x] Informal "du" vs formal "Sie" throughout the German copy. **Decided: du**, with neutral phrasing preferred where it reads naturally. See [05-design](specification/05-design.md).
- [ ] Does an admin-only "we think they're probably coming" state add value over plain no-answer? Guests would never see it.
- [ ] Track invitation send-out date per household, to time reminder nudges?
- [ ] Photo retention: how long does the gallery stay online after the wedding, and what happens to the files afterwards? Blocks the end-of-life procedure in [06-privacy-security](specification/06-privacy-security.md).
- [ ] Can the caterer export be name-free (counts + allergies keyed by table) instead of a full guest list? Would remove the largest planned data disclosure.
- [x] **Contact phone numbers, 2026-08-31.** Login fallback uses **+43 650 9408100** (`contactPhoneNumber` in `web/src/lib/labels.ts`, set 2026-08-31; `contacts` there holds both names and numbers for `/kontakt`). `/kontakt` lists both that number and **+43 677 63668655**; One number in the fallback sentence, two on the contact page — see `F11-02` and `F2-F06`.
- [~] **Datenschutz drafted, 2026-08-31.** `/datenschutz` is written in guest German from `06-privacy-security` (`privacyLabels`) and sits behind the login with the other content pages — the "readable without logging in" requirement was reversed the same day, reasoning in `06-privacy-security` and `F2-F07`. Wants a proof-read, and it deliberately states no gallery lifetime — that number is the open photo-retention question below.
- [x] **Bank details are published on `/geschenke`, 2026-08-31.** Content pages compile into the bundle and the SPA is served unauthenticated, so an IBAN there is semi-public; accepted, since an IBAN lets somebody send money rather than take it. The IBAN and account holder themselves are still to supply. See `F2-F05`.
- [x] **Accessibility scope, 2026-08-31.** No automated checker (`axe-core` rejected — it cannot see targets, zoom, focus visibility or copy register) and no screen-reader pass; nobody in the guest list uses one. Manual checklist, kept **in `F11-01` itself** — a separate `specification/08-accessibility.md` was rejected the same day: the rules live in `05-design` and would be restated, and a table of tick marks is progress, which belongs in `features/README.md`.
- [x] **One countdown, to the wedding**, 2026-08-31. The RSVP deadline is a written-out date beside the answer link, not a second counter. `F2-F02`.
- [x] **Hero carries names and date only**; the venues get their own page. `F2-F02`, `F2-F04`.
- [~] **Proof-read the F2 copy.** Added 2026-08-31 with the content pages: `startLabels`, `scheduleLabels`, `locationLabels`, `dresscodeLabels`, `giftLabels`, `faqLabels`, `contactLabels` and `privacyLabels` were written with the pages and have not been read back. The FAQ list is a guess until the first questions arrive after send-out.
- [~] **Hero photo for `/start`.** A placeholder image is checked in at `web/src/assets/hero.{jpg,webp}` at the aspect ratios the layout uses (4:5 mobile, 16:9 desktop). Swap the files; nothing else changes.
- [~] **Help texts on every guest form field, behind a `?` popover** beside the label — inline help on every field would bloat the form and put an invisible length limit on the copy. Rule added to `05-design`'s form behaviour section on 2026-08-31, including the accessibility requirements. The F3/F4 stories were retrofitted the same day (`F3-F01`, `F3-F02`, `F3-F03`, `F4-F01`). The F3 sentences were written with the form on 2026-08-31 (`rsvpLabels` in `web/src/lib/labels.ts`) and want a proof-read; F4's and F2's are still to write.
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

### Answered in the 2026-09-01 UI feedback round

Decisions taken with the thirteen stories built that day. Each one is recorded where it belongs — this list is the index, not the reasoning.

- [x] **Transport is one question with a direction.** Needing seats and offering them are mutually exclusive: the form asks nothing / we need / we can offer, and the server refuses a body claiming both. `F3-B07`, `F3-F07`; also `02-features`, `03-data-model`.
- [x] **`with_parent` and `high_chair` are children only**, refused server-side on every write, and hidden from an adult's card. One `seating_need` field for both venues was kept — a high chair being a party matter is copy, not a second column. `F3-B08`, `F3-F08`.
- [x] **`meal_choice` defaults to `all`** once a scope covers the party. Accepted cost: "said: eats everything" and "did not look at the field" are no longer distinguishable. `F3-F08`.
- [x] **`/zusagen` shows the stored answer when a household has answered**, with the form one tap away; the admin page stays form-first. `F3-F09`.
- [x] **The plus-one sheet was submitting the RSVP form** — a Radix portal keeps the React tree, so the synthetic submit bubbled. That one bug produced all four reported symptoms. `F4-F03`.
- [x] **The RSVP-answered household fields have one writer.** Transport counts and `has_stroller` left the admin household create and patch bodies; the detail page shows them read-only. `F5-B05`, `F5-F05`.
- [x] **"Zuletzt gespeichert um HH:MM", persistent**, instead of a transient "Gespeichert." — and therefore **no artificial minimum duration** on the saving state, which was the alternative. `F5-F05`.
- [x] **No pagination on the household list.** Sixty rows with a search box and two filters is a screen you scan; the table moved to the bottom of the page instead, so no control drifts as households are added, and a filter for "ohne Antwort" was added beside the never-logged-in one. `F5-F04`.
- [x] **The start page says one thing about the RSVP**, and addresses a household in the number we **seeded** it with — a plus-one does not turn "du" into "ihr". `F2-F08`.
- [x] **Location is an overview plus one page per venue**, with Übernachtung on the party page and the transfer on the overview. `F2-F09`.
- [x] **The navigation skeleton was a guard awaiting a cached session**, not code splitting. Fixed at the cause; the full-screen skeleton is now the cold load's only. No spinner. `F11-04`.
- [x] **Editable fields are filled white.** The shadcn defaults were transparent, which on `paper` reads as disabled; radius and type size unified across `Input`, `Textarea` and `Select`. `F11-05` — the code is done, the side-by-side look is still worth one manual pass.

### Raised in the 2026-08-31 code review

- [ ] **Admin settings endpoint.** The admin login response deliberately carries no flags, but the admin needs to *flip* `rsvp_open`, `seating_published`, `gallery_visible` and `uploads_open`. That is a read/write endpoint (`GET`/`PATCH /api/admin/settings`) and wants a story in F6 — not a read-only copy smuggled into the login body.
- [x] **Split `internal/application` into `application/auth`, `application/households`, `application/exports`, `application/rsvp`** — done 2026-08-31 with F3-B02, the second use case. Each subpackage exposes one `UseCase` type; the parent package keeps only `ErrNotFound` and `TranslateNotFound`, which every use case needs and none may import from a sibling.
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
- [x] The API sketch in [04-architecture](specification/04-architecture.md) was refreshed on 2026-08-31, once the F3/F4 stories fixed the real shapes: `PUT /api/rsvp` as a full replace, `POST /api/rsvp/members` taking a name, and the admin RSVP routes.
- [x] **Carried into F3, decided 2026-08-31 when `F3-B06` was agreed** — all five are in the built code as of 2026-08-31; kept here as the record of what was decided:
  - The RSVP form component takes its data and its save mutation as **props** and never fetches for itself, so the guest route and the admin route can hand it different sources. This constrains `F3-F01`–`F3-F04`.
  - One `application` use case takes a household id. The guest handler passes the session's household; the admin handler passes the path parameter. Ownership is checked in exactly one place either way.
  - **Deadline enforcement is an argument to that use case, not a rule inside it** (`F3-B04`). The admin path exists for the call that comes in late, so it must be able to write after `rsvp_deadline`.
  - **An admin edit is audited as `actor_type = 'admin'`** (`F3-B05`). The audit log settles "but I said we were coming", and recording our own edit as the household's answer would mislead at the one moment it matters.
  - An admin edit sets `rsvp_submitted_at`, so the household drops off the nudge list. If we took it down on the phone, they have answered.
- [ ] No decision recorded on what a guest sees between "RSVP deadline passed" and "seating published" — a period where the site has little to say. Roughly five weeks in late spring 2027; [07-roadmap](specification/07-roadmap.md) wants this decided during M5.
