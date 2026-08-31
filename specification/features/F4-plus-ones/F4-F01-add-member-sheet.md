# `F4-F01` — Add-plus-one sheet, and the hint for everyone else

**Epic:** F4 — Plus-one · **Layer:** frontend · **Depends on:** `F4-B02`, `F3-F02`

## Story

As a guest invited on my own, I want to add the person I am bringing straight from the answer form, and as anybody else I want to know how to get another person added, so that neither of us has to guess whether the list is fixed.

## Scope

**In:**

- `AddPlusOneSheet`: one field, the companion's name.
- The trigger, shown only when `can_add_plus_one` is true.
- The explanation shown in its place when it is false, with our phone number.

**Out:**

- Removal → `F4-F02`. Answering for the companion — they appear as an ordinary member card (`F3-F02`).
- Any admin view of the additions → `F6-F01`.

## Instructions

1. Consume the `F4-B02` contract exactly, and branch on `can_add_plus_one` alone. Do not re-derive the rule from the member list: two implementations of "who may add" is how the screen and the server end up disagreeing in front of a guest.
2. A `Sheet` on mobile, a `Dialog` on desktop. One field does not deserve a page, and navigating away from the form risks the answers already typed.
3. **One field: the name.** No kind, no age, no meal — the server takes nothing else (`F4-B02`), and a form asking for more would promise something the API refuses.
4. The trigger sits after the member cards, secondary, phrased as the question it answers: "Kommst du zu zweit? Begleitung hinzufügen". It is the single most consequential thing a solo guest can do on this page, and it must not read as an administrative option.
5. When `can_add_plus_one` is false, render an explanation in the same place — not a disabled button (fails contrast, reads as broken) and not nothing (a guest wondering whether they missed it will phone anyway, which is the call we were trying to make unnecessary). Two sentences: further people we enter ourselves, and the number from `contactPhoneNumber`. It must read as an offer, not a refusal — the answer is yes.
6. Households of two or more see that explanation from the start. That is most households, so the block is quiet: `ink-muted`, no icon, no `warning` colour. Nothing has gone wrong.
7. A 409 `plus_one_not_allowed` — two tabs, one form each — shows the API sentence in the sheet and re-renders the page from a refetch.
8. On success, append the returned member, close the sheet, and scroll to the new card so the next question — is this person coming — is on screen.
9. The new card is visibly unanswered, with the same neutral marking as `F3-F02` uses, and only becomes an error after a submit attempt.
10. Adding does **not** save the rest of the form. The asymmetry is real and confusing, so the confirmation says the person was added and still has to be answered for.
11. `rsvp_closed` means the deadline passed while the form was open: show the API sentence and switch the page to `F3-F05`.
12. The name field carries a help popover like every other guest field (`05-design`): it says who counts as a plus-one and that anybody else goes through us, which is the same question the guest is about to phone about.
13. All copy in `labels.ts`, phone number included — no number written inline.

## Test plan

- [ ] Component: with `can_add_plus_one` true, the trigger renders and the sheet takes exactly one field.
- [ ] Component: with it false, the explanation and phone number render and no disabled button exists.
- [ ] Component: adding appends the card, closes the sheet, and flips the page to the explanation state.
- [ ] Component: a 409 `plus_one_not_allowed` shows the API sentence.
- [ ] Component: the new card is marked unanswered and blocks submit until answered.
- [ ] Component: `rsvp_closed` switches the page to the read-only view.
- [ ] Accessibility: the sheet traps focus, closes on Escape, and returns focus to the trigger.

## Done when

- [ ] A solo guest adds their companion in two taps, and every other household reads one calm sentence telling them how to get somebody added.
- [ ] Checkbox ticked in `README.md`.
