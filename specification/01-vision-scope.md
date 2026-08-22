# 01 — Vision & Scope

Status: draft · Last updated: 2026-08-21

## Purpose

A private web application for our wedding. It is the single place guests go to find out what is happening, tell us whether they are coming, see where they sit, and — after the event — look at and contribute photos.

For us, it doubles as the planning tool: guest list, RSVP overview, seating, and budget in one place instead of three spreadsheets.

It replaces: a paper insert with schedule/travel info, RSVP cards mailed back, a printed seating plan taped to a wall, and a WhatsApp group full of photos nobody can find later.

## Audience

- **Guests** — ~80 people across ~60 households. Wide age range; a meaningful share are not confident with technology. Mostly mobile phones.
- **Us (admins)** — two people. Need to manage the guest list, see who is coming, plan seating, track budget, publish content, moderate photos.

## Guiding principles

1. **Guest UX beats everything.** If a 75-year-old cannot RSVP alone on a phone, the feature has failed. One login, ever. Big targets, plain German, no jargon.
2. **Private by default.** Nothing is publicly readable, nothing is indexed.
3. **Small and self-contained.** One container, one binary, one SQLite file. No external SaaS in the critical path.
4. **Trusted-guest threat model.** Guests are friends and family. We defend against mistakes and drive-by strangers, not against a determined insider. The exception is admin-only data (budget), which guests must never reach.
5. **Boring tech.** This must still work unattended for months.

## In scope

- Static informational content (schedule, venue, travel, accommodation, dress code, gifts, FAQ, contact).
- Per-household login via a printed code.
- Household RSVP: per-member attendance, meal choice, allergies, free-text note.
- Guest-added plus-ones and children, visible to us as additions.
- Seating chart: planned and edited by admins, readable by guests.
- Budget tracking: admin-only. Planned vs. actual per cost item.
- Admin area: household/guest CRUD, RSVP dashboard + export, content control, seating editor, budget, photo moderation.
- Photo gallery — curated by us before the wedding, guest uploads after it.

## Out of scope

- Public marketing site / SEO / social sharing cards.
- Multi-language. **German only.**
- Vendor management, contract storage, task/todo planning.
- Gift registry with purchase tracking. A text section with bank details and wishes is enough.
- Email sending as a hard dependency. Codes travel on paper.
- Native mobile apps.
- Live streaming the ceremony.

## Success criteria

- ≥ 90% of households RSVP through the site without us helping them by phone.
- Zero households locked out or logged in as the wrong household.
- RSVP data exportable to CSV for the caterer without manual cleanup.
- Seating chart finalized in the app, printable, and guests can find their table themselves on the day.
- Budget figures never visible to a non-admin session.
- Site available from invite send-out through 3 months post-wedding.

## Open questions

- Venue and table layout — still needed for content and seating. The wedding date is assumed to be 2027-07-17, and invitations go out in Oct/Nov 2026 as a single launch (no save-the-date phase).
- Photo retention: how long do we keep the gallery online after the wedding?
