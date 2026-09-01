# `F5-F05` — Household detail: answered fields are read-only, and saving says when

**Epic:** F5 — Admin: households & guests · **Layer:** frontend · **Depends on:** `F5-B05`, `F5-F02`

## Story

As an admin, I want the household page to edit only what is ours to edit and to tell me when it last saved, so that I never wonder whether the change I just made went in.

## Scope

**In:**

- Transport counts and `has_stroller` shown as text in the RSVP section, not as inputs.
- "Zuletzt gespeichert um HH:MM" in place of "Gespeichert.", on every save block on the page.

**Out:**

- Editing those fields, which happens on the RSVP page the section already links to (`F3-B06`).
- An artificial minimum duration for the saving state. Considered on 2026-09-01 and dropped: the flash was only a problem because the confirmation disappeared, and a delay would slow down the one screen used sixty times in a row.

## Instructions

1. The three fields are answers, not settings. Editing them here bypassed the RSVP rules (`F5-B05`), and showing them beside `display_name` invited exactly that. They move into the RSVP section, next to the link that edits them properly, and read as "—" when nothing was answered.
2. The timestamp is the browser's own clock at the moment the response arrived, and the copy says what that is: "zuletzt gespeichert um", not "gespeichert am". The alternative — an `updated_at` on the response — would be a column, a migration and a DTO field to carry a sentence that only has to distinguish "just now" from "a while ago".
3. It persists until the next save or a reload. That is the whole fix for the flash: a confirmation that stays does not need to be shown for a minimum time to be seen.
4. Same treatment on the member blocks, which have the same chip and the same problem.
5. Still `role="status"`, so it is announced once rather than on every keystroke.

## Test plan

- [ ] Component: the household form has no transport or stroller inputs.
- [ ] Component: the RSVP section shows the stored transport answer and the pram as text.
- [ ] Component: saving the household shows "Zuletzt gespeichert um …" and it is still there afterwards.
- [ ] Component: saving a member shows the same, scoped to that member's block.

## Done when

- [ ] Every editable field on the page is ours to edit, and no save leaves any doubt.
- [ ] Checkbox ticked in `README.md`.
