# `F4-F03` — Adding a companion must not submit the RSVP form

**Epic:** F4 — Plus-one · **Layer:** frontend · **Depends on:** `F4-F01`

## Story

As a guest who has already answered, I want adding my companion to add my companion, so that the page does not save an answer I did not press save on and then show me a list without them in it.

## Scope

**In:**

- The plus-one sheet's submit stops at the sheet.
- `RSVPForm` ignores a submit event that did not come from its own `<form>`.
- A regression test for the whole sequence: answered household → add a companion → the card appears and nothing was saved.

**Out:**

- The rule about who may add → `F4-B02`. The server was right throughout; this is a frontend bug.

## Instructions

1. The cause, for the record: `DialogContent` is a Radix **portal**, which moves the DOM node but not the React tree. The sheet's `<form>` is therefore still a React child of the RSVP `<form>`, and React's synthetic `submit` bubbles along the React tree — so adding a companion also ran `RSVPForm`'s submit handler. That `PUT /api/rsvp` carried the member list from *before* the addition, its response was written into the query cache by `useSaveRSVP`, and the freshly added member was overwritten out of existence. Symptoms, all one bug: the post-save summary appearing unbidden, the missing card, `can_add_plus_one` back to `true`, and a 409 on the second attempt.
2. `stopPropagation` on the sheet's submit event, with that reason named in a comment. Anything else — moving the sheet outside the form, replacing the inner form with a div — either breaks Enter-to-submit inside the sheet or moves the trap one refactor away.
3. `RSVPForm`'s handler additionally ignores an event whose `target` is not its own form. Belt and braces on purpose: any future portalled form under this component would otherwise re-open exactly this hole, and the check is one line.
4. Test the sequence, not the unit. A test that renders the sheet alone would have passed throughout.

## Test plan

- [ ] Component: an answered household adds a companion — the new card appears, the summary does **not**, and no `PUT /api/rsvp` was sent.
- [ ] Component: adding twice in a row is impossible, because the trigger is gone after the first (`can_add_plus_one` from the response).
- [ ] Component: pressing Enter in the name field adds the companion and does not save the form.
- [ ] Component: the save button still saves.

## Done when

- [ ] Adding a companion is a single action with a single request, on the first try.
- [ ] Checkbox ticked in `README.md`.
