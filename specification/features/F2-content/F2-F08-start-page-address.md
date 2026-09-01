# `F2-F08` — Start page: one sentence about the answer, and the right number of people

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F02`, `F3-F09`

## Story

As a guest, I want the start page to address me the way I was invited and to say one thing about the RSVP, so that it does not thank me for answering and ask me to answer in the same breath.

## Scope

**In:**

- The RSVP block says one thing: either "please answer by …" or "thanks — changeable until …".
- Singular or plural address, chosen by how many people we seeded into the household.

**Out:**

- The RSVP form's own copy, which stays plural: it addresses a household, and a companion may have been added to it.
- Any new endpoint. `GET /api/me` already carries `members[].origin`.

## Instructions

1. Two sentences saying opposite things was the bug: the thanks and the request were rendered together because the deadline sentence was unconditional. Answered → the deadline becomes a "changeable until" date, which is the only thing left worth knowing. Unanswered → the request, unchanged.
2. The address follows the **seeded** member count, not the current one: a single guest who adds a companion is still the person we wrote to, and switching to "ihr" the moment they name somebody reads as the site correcting them. `origin === "seeded"` is already on `/api/me`.
3. Both forms of every affected sentence live in `labels.ts` as a `Record<"singular" | "plural", …>`, so the pair is visible in one place and a missing variant is a type error. Not built by string surgery: German does not conjugate by find-and-replace.
4. The greeting stays built *around* the display name rather than in front of it ("Luki & Paddi" is a valid display name).
5. Nothing else on the page moves — the hero, the countdown and the one primary action are `F2-F02`'s and are right.

## Test plan

- [ ] Component: a household seeded with one adult reads in the singular, and still does after a companion is added.
- [ ] Component: a household seeded with two or more reads in the plural.
- [ ] Component: an answered household sees the thanks and the "changeable until" date, and no request to answer.
- [ ] Component: an unanswered household sees the request and no thanks.

## Done when

- [ ] The start page says one true thing about the RSVP, in the number the guest was invited in.
- [ ] Checkbox ticked in `README.md`.
