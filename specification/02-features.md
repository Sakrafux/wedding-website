# 02 — Feature Overview

Status: draft · Last updated: 2026-08-21

High-level map only. Detailed, isolated user stories live in `specification/features/<feature>.md` and are written per feature when we get to it. This file exists to answer "what is in this product and how important is it".

Priorities: **P0** = site is useless without it · **P1** = ship for invite send-out · **P2** = nice, ship if time · **P3** = post-wedding phase.

| ID | Feature | Actor | Priority | One-line description |
|----|---------|-------|----------|----------------------|
| F1 | Household login | Guest | P0 | Printed per-household code redeemed for a long-lived session. |
| F2 | Informational content | Guest | P0 | Schedule, venue, travel, dress code, gifts, FAQ, contact. |
| F3 | RSVP | Guest | P0 | One submission per household, per-member attendance + meal + allergies. |
| F4 | Plus-ones & children | Guest | P0 | Households add companions and kids themselves; we see the additions. |
| F5 | Admin: households & guests | Admin | P0 | CRUD households/members, generate and export codes. |
| F6 | Admin: RSVP dashboard | Admin | P1 | Totals, meal counts, allergy list, no-answer list, CSV export. |
| F7 | Seating chart | Admin + Guest | P1 | Admins assign guests to tables; guests see the plan and their own seat. |
| F8 | Admin: budget tracking | Admin | P1 | Cost items, planned vs. actual, payment status, totals. |
| F9 | Curated gallery | Guest | P2 | Photos we publish before the wedding. |
| F10 | Guest photo uploads | Guest | P3 | Post-wedding uploads with quotas and admin moderation. |
| F11 | Cross-cutting quality | — | P1 | Mobile-first, accessible, German, human error messages. |

---

## F1 — Household login (P0)

Card carries a **generic** QR code (identical on all cards, points at site root) plus an **individually printed** household code. Print shop has confirmed it does variable-data printing; extra cost accepted (~€55).

Key decisions:

- Code is the only secret. No username, no password, no email.
- Alphabet excludes ambiguous glyphs: `23456789ABCDEFGHJKLMNPQRSTUVWXYZ`. 6 characters, printed as `ABC-234`.
- Input normalized: uppercased, whitespace and dashes stripped.
- After redeem, a confirmation screen names the household ("Willkommen, Familie Müller — bist das du?") so a mistyped-but-valid code is caught immediately.
- Session cookie: `HttpOnly`, `Secure`, `SameSite=Lax`, 365 days, rolling refresh.
- Failed attempts rate-limited per IP; friendly error, never a lockout.

## F2 — Informational content (P0)

Read-only pages authored by us: Start (greeting, date, countdown), Ablauf, Location, Anreise & Übernachtung, Dresscode, Geschenke, FAQ, Kontakt.

**Decided:** content is hardcoded directly in the React components. No CMS, no Markdown pipeline, no DB rows, no admin text editor. The volume is small and changes are rare. Accepted cost: a content change means a rebuild and redeploy. Runtime-variable flags (RSVP deadline, seating published, uploads open) are the exception and live in `app_setting`.

## F3 — RSVP (P0)

One RSVP per household, editable until a deadline, then read-only.

- Per member: attendance scope (`no` / `church_only` / `party_only` / `both`), meal choice (`all` / `vegetarian` / `vegan`), portion (`none` / `kids` / `full`), midnight snack yes/no, seating need, allergies. Children also give an age.
- Scope is per member because the exceptions live inside households — a grandmother at the ceremony but not the party, children only at the church. To avoid extra clicking, the form offers a household-level "Wir kommen zu: Kirche / Feier / beidem" selector that sets everyone at once, with per-member overrides beneath.
- Catering questions are only asked of guests who are coming to the party. A church-only guest is never asked about meals, snacks or seating — and never counted for them.
- Per household: transport seats **needed** and seats **offered** for the church → reception trip. This gives us a capacity gap, which tells us whether to organise extra cars or a shuttle. The app does not match riders to drivers; we do that ourselves.
- Per household: a general free-text note. This is the intentional catch-all — no structured field set will anticipate everything, so guests always have somewhere to write "wir kommen erst nach der Zeremonie" or "Oma braucht einen Platz nah am Ausgang". Notes are shown prominently in the admin dashboard, and a household with an unread note is flagged.
- No "maybe" — it makes catering numbers unusable.
- Attendance and scope are one field, so "declined, but coming to the party" cannot be expressed.
- All changes audit-logged (who, when, what).
- Save shows an unmissable confirmation plus a summary of what was submitted.

## F4 — Plus-ones & children (P0)

We do not know every guest's current situation, so households add their own companions rather than us guessing. Scale stays manageable (~60 households, most already known to us).

- A household can add members itself: name, type (adult / child), meal choice, portion, seating need, allergies, and an age for children.
- Added members are marked as **guest-added** and surfaced to us distinctly in the admin dashboard, so the delta from our original list is always visible.
- Soft cap per household (configurable, default e.g. 2 additions) with a "mehr? ruf uns an" hint rather than a hard wall.
- Removing a guest-added member is allowed until the RSVP deadline; removing a household member we seeded is not (they set attending = no instead).
- **Decided:** additions count toward the headcount immediately — no approval step, no pending state, no guest left waiting. They appear highlighted in the admin dashboard as a delta against our original list, and we intervene only if something looks wrong.
- Children are recorded with an age (as of the wedding date), which feeds caterer pricing brackets and venue counts. Physical needs are captured separately in `seating_need`, so we never infer a high chair from an age.
- Every addition and removal is audit-logged.

## F5 — Admin: households & guests (P0)

CRUD households and their members. Generate/regenerate codes. Export a print-shop-ready CSV of household name + code for variable-data printing. Shows last login per household.

## F6 — Admin: RSVP dashboard (P1)

Headcounts split by scope (church / party / both — three numbers, three different vendors), meal and portion counts, midnight snack count, children by caterer age bracket, special seating needs (high chairs, strollers, wheelchairs), consolidated allergy list, transport seats needed vs. offered and the resulting gap, list of guest-added members, unread household notes, nudge list, CSV export for the caterer. No SQL needed to get the headcount.

## F7 — Seating chart (P1)

Admins define tables and assign attending guests to seats. Guests read the result.

**Decided — the SVG bridge.** No drag-and-drop editor. The room layout is a hand-drawn SVG, authored outside the app and checked into the frontend. Each table shape carries a stable `id` attribute. In the DB, `seating_table.svg_element_id` points at that shape. The app only ever colors and labels existing shapes — it never positions anything. This buys a real spatial view for a fraction of the effort of a canvas editor.

**Admin editing** is list-based: define tables (label, capacity, `svg_element_id`), then assign attending guests to them. Over-capacity is a warning, not a block.

**Guest view** renders the full floor plan so guests can orient themselves in the room — but only their **own** table is highlighted and labeled with names. Other tables are shapes without occupants. Guests learn where they sit, not where everyone else sits.

Gated by the `seating_published` flag; a draft plan is invisible to guests, since a visible draft invites lobbying.

Constraint: only guests attending the party (`party_only` or `both`) are assignable, including guest-added members; church-only guests are never seated. Flipping an RSVP to "no" or "church only", or deleting an added member, does **not** silently drop the assignment — it surfaces as a stale assignment for us to resolve.

Open for the detailed story: printable output for the day itself, and whether the SVG needs a second, zoomed variant for phones.

## F8 — Admin: budget tracking (P1)

Admin-only. Never exposed to a guest session — enforced server-side, not by hiding a nav link.

- Cost items with category, planned amount, actual amount, vendor, due date, payment status, note.
- Totals: planned vs. actual vs. paid, and delta against an overall budget cap.
- CSV export.

Open question: per-head cost derivation from the live headcount (e.g. catering per person) — useful, but couples budget to RSVP data.

## F9 — Curated gallery (P2)

Photos we upload. Grid, lazy loading, lightbox. Server-side thumbnails, originals kept, EXIF stripped on ingest (location especially).

## F10 — Guest photo uploads (P3, post-wedding)

Feature-flagged, opens after the wedding.

- Multi-file upload from a phone with progress and retry.
- Per-household quota to bound disk use (e.g. 100 files / 2 GB).
- JPEG, PNG, HEIC, MP4; validated by content sniffing, not file extension.
- Default auto-publish (trusted guests) with admin hide/delete; strict queue-first mode available as a config toggle.
- Uploader household recorded and displayed.
- Bulk ZIP download so the archive is not trapped in the app.

## F11 — Cross-cutting quality (P1)

- Mobile-first; desktop secondary.
- Accessibility: visible focus, labelled inputs, contrast ≥ 4.5:1, usable at 200% font size. Our guest demographic needs this more than most.
- Informational content cached so a bad venue signal does not blank the page.
- German throughout, informal "du" (to be confirmed).
- Error messages written for humans, with a phone number as last resort.

---

## Explicitly rejected

| Idea | Why not |
|---|---|
| Shared site password + self-identify | Older guests would pick the wrong household by mistake. Individual printed codes chosen instead. |
| Email magic links | No reliable email addresses for all guests; adds an SMTP dependency. |
| "Maybe" RSVP option | Makes catering numbers unusable. |
| Public site with a private area | Doubles design work; nothing here is for strangers. |
| Separate save-the-date mailing | Only worth it when the invitation is too far out to be actionable, or when guests must book long-haul travel. Neither holds at nine months with a mostly local guest list. Splitting would also mean two variable-data print runs, or a save-the-date pointing at a site nobody can log into. One mailing in Oct/Nov 2026 does both jobs. |
