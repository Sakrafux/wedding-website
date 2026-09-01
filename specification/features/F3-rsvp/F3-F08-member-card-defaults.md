# `F3-F08` — Member card: a meal default, and seats that exist for this person

**Epic:** F3 — RSVP · **Layer:** frontend · **Depends on:** `F3-B08`, `F3-F02`

## Story

As a guest, I want the form to pre-answer the ordinary case and to offer only the options that apply to the person in front of me, so that answering for four people is a handful of taps.

## Scope

**In:**

- `meal_choice` defaults to `all` as soon as a scope covers the party.
- `with_parent` and `high_chair` shown for `kind = child` only.
- A sentence saying a high chair is a party matter and that a child sits in the pew normally.
- The deadline sentence above the form set in the emphasised style.

**Out:**

- Splitting `seating_need` into a church field and a party field — considered and rejected on 2026-09-01: the need is a property of the person, `F7` renders per venue, and a second enum is a migration plus a second question on every attending guest's card.
- The server-side rule → `F3-B08`.

## Instructions

1. The meal default is applied in `state.ts`, in one pure function, at both moments it is needed: when the draft is seeded from an answer that already covers the party, and when a scope changes to one that does. Two call sites, one rule — a default written into the radio group's `value` would be a default that never reaches the request body.
2. `all` and not "nothing selected": three quarters of the guest list eats everything, and the accepted cost is that we can no longer tell "said all" from "did not look at the field". Recorded here because it is a real loss — the caterer numbers are unaffected, since a plate is ordered either way.
3. Filter the seating options by `member.kind` from the response. Filter, never disable: a disabled radio fails contrast and reads as broken (`05-design`).
4. The high-chair help sentence covers the question the option raises rather than restating the label: in the church a child sits in the pew, the high chair is for the table. It goes in the field's popover, where the rest of the *why* lives.
5. The deadline sentence gets weight, not colour: it is the one date on the page and it is currently the quietest line on it.

## Test plan

- [ ] Component: choosing a party-covering scope selects "Isst alles" and submits `meal_choice = "all"`.
- [ ] Component: an answer that already covers the party and has no meal choice renders with "Isst alles" selected.
- [ ] Component: a member card for an adult offers exactly `normal` and `wheelchair`; a child's offers all four.
- [ ] Component: the high-chair help mentions the church.

## Done when

- [ ] No household is offered a high chair for an adult, and the common meal answer needs no tap.
- [ ] Checkbox ticked in `README.md`.
