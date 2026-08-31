# `F3-F04` — Save, confirmation, and summary

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-B03`, `F3-F03`

## Story

As a guest, I want an unmistakable confirmation of what I just submitted, so that I know I am finished and do not answer a second time next week to be sure.

## Scope

**In:**

- The submit button, its pending state, and the mutation passed in as a prop.
- `RsvpSummary`: a recap of exactly what was saved, per member.
- The error paths: field errors, `member_set_mismatch`, `rsvp_closed`, network failure.

**Out:**

- The read-only rendering after the deadline → `F3-F05`.
- Any email or printed confirmation. There is no email address in this product.

## Instructions

1. The save mutation arrives as a prop, like the data (`F3-F01`). The route wires it to `PUT /api/rsvp`; `F3-F06` wires the same component to the admin endpoint.
2. Before submitting — and **only** then; the form never opens with errors on it (`F3-F02`) — check that every member has a scope. Unanswered members are marked in place, the page scrolls to the first one, and the error summary at the top links to each — the server rejects the same case (`F3-B03`), but a guest who is missing one answer should not have to wait for a round trip to find out which.
3. One button, "Speichern", full width, primary. Disabled only while the request is in flight, and it says so ("Wird gespeichert …") rather than just spinning.
4. On success, replace the form with `RsvpSummary` rather than showing a toast. A toast is the standard move and it is wrong here: it disappears, and this audience needs the answer to still be on screen when they look up from the phone. Rejected deliberately.
5. The summary lists every member with their scope and, for party guests, their meal, portion and snack — in German labels, not enum values — plus the transport answers and the note. It is a recap of the **response body**, not of the form state, so a value the server normalized away is visibly absent.
6. Above it, one sentence confirming that we have it, and the date they answered. Below it, "Antwort ändern", which returns to the form, and a line saying it can be changed until the deadline, with the date written out.
7. Field errors from a 400 are distributed to their controls by the `members.<id>.<field>` key, plus a summary at the top of the form linking to each (`05-design`).
8. `member_set_mismatch` → the German sentence from the API, verbatim, plus a button that refetches and re-renders the form. Do not attempt to merge state: the household list changed, and the honest move is to show the new one.
9. `rsvp_closed` → the API sentence, and the form switches to the `F3-F05` read-only view. This is the race where somebody had the form open as the deadline passed.
10. A network failure keeps every entered value on screen. Losing a filled-in form to a dropped connection in a village is the failure this audience will actually hit.
11. After a successful save, invalidate the `/api/me` query too: `rsvp_submitted_at` drives the navigation label (`F2-F01`), and a bar still saying "Zusagen" after a successful answer is a bar that gets pressed again.

## Test plan

- [ ] Component: submitting posts exactly the form state, in the `F3-B03` request shape.
- [ ] Component: a member with no scope blocks submit, is marked, and is linked from the summary at the top.
- [ ] Component: on success the summary replaces the form and lists the saved values in German.
- [ ] Component: the summary shows what the server returned, not what was typed — feed it a response with a cleared meal choice and assert it is absent.
- [ ] Component: a 400 with field errors renders them on the right controls.
- [ ] Component: `member_set_mismatch` shows the message and offers a reload that refetches.
- [ ] Component: `rsvp_closed` switches to the read-only view.
- [ ] Component: a network error leaves the entered values intact.
- [ ] Accessibility: the success state moves focus to the confirmation heading and is announced.

## Done when

- [ ] A household can answer, see exactly what was recorded, and come back to change it — and no failure mode discards what they typed.
- [ ] Checkbox ticked in `README.md`.
