# `F3-B01` — Domain: attendance scope gates catering

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F5-B02`

## Story

As a developer, I want the RSVP rules as pure functions over plain structs, so that "a church-only guest is never counted for a meal" is one testable fact rather than a condition repeated in a handler, a form and three dashboard queries.

## Scope

**In:**

- The `Attending`, `MealChoice` and `Portion` enums on `domain`, alongside the `SeatingNeed` and `GuestKind` that already exist.
- The RSVP answer fields on `domain.Guest`: `Attending *Attending`, `MealChoice *MealChoice`, `Portion`, `MidnightSnack`. `SeatingNeed` and `DietaryNote` are already there.
- `NormalizeGuestAnswer` — the one function that applies the scope gate, and the invariants it enforces.
- Household-level answer fields and their rules: transport seat counts, `rsvp_note`.
- Scope predicates the rest of the app reads instead of comparing strings: `AttendsChurch`, `AttendsParty`, `Attends`.

**Out:**

- Persistence, endpoints, DTOs → `F3-B02`, `F3-B03`.
- Deadline → `F3-B04`. The domain has no clock and no opinion about *when* an answer may be written.
- Adding and removing members → `F4-B01`.
- Any aggregate count → `F6-B01`. This story owns one guest's answer; counting is a query, not a rule.

## Instructions

1. `Attending` is a pointer on `Guest` because NULL is a real state — "has not answered" is what the nudge list is built from, and a zero value would make an unanswered guest indistinguishable from one who declined. Do **not** introduce an `unanswered` enum member: the database column is nullable, and a fifth value would be a second way to say the same thing.
2. `NormalizeGuestAnswer(Guest) Guest` is the whole scope gate. Given an answer, it returns the answer as it will be stored: for `no` and `church_only` it resets the catering fields to their schema defaults — `MealChoice` to nil, `Portion` to `full`, `MidnightSnack` to false.
3. **Reset rather than preserve, and reset to the defaults rather than to `none`.** Preserving means a guest who switches from `both` to `church_only` leaves a meal choice behind that every later reader has to remember to ignore, and one that forgets pays for a plate. Resetting to `portion = 'none'` was rejected because it reads as an answer the household never gave — the neutral default is the honest record of "not asked". The scope is the single source of truth, and the stored row agrees with it.
4. `SeatingNeed` and `DietaryNote` are **not** scope-gated: a wheelchair space is needed in the church pew as much as at the table, and an allergy is worth knowing wherever somebody eats. Say so in the doc comment, because their absence from the reset list otherwise reads as an oversight.
5. `Age` is editable through the RSVP for `kind = 'child'` and obeys the rules `F5-B02` already wrote (`ErrAgeOnAdult`, `ErrAgeOutOfRange`). Reuse those; do not restate the range.
6. `Kind` is **not** editable through the RSVP. A household turning an adult into a child changes a caterer bracket by pressing a radio button, and the case it would serve — we typed the wrong thing — is ours to fix in `F5-F02`. `F4` sets `kind` when a member is added, which is the only moment it is a real question.
7. Transport: `NormalizeHouseholdAnswer` zeroes `TransportSeatsNeeded` and `TransportSeatsOffered` when no member is `both`. Per `03-data-model`, the church → reception trip only exists for a guest attending both; a household whose members all attend one half and which still carries a seat count inflates the shuttle gap, which is the one number those fields exist to produce. The form hides the fields in that case (`F3-F03`); the server does not rely on the form.
8. Seat counts are capped at 20 apiece, the same bound `F5-B01` uses. One bound, one reason: a household is not a coach.
9. `Attends`, `AttendsChurch` and `AttendsParty` are the readable form of the four-value comparison, and every later epic uses them — `F6`'s counts, `F7`'s assignment validity, `F8`'s per-head resolution. A `switch` on the raw string in a query file is the thing they exist to prevent.
10. `rsvp_note` gets a sane cap of 2000 characters, matching `admin_note`. Not a product rule, a protection against a paste accident: the note is displayed in full in the admin inbox.

## Test plan

- [ ] Unit: `NormalizeGuestAnswer` clears meal, portion and snack for `no` and for `church_only`, and leaves them untouched for `party_only` and `both`.
- [ ] Unit: normalization is idempotent — normalizing twice equals normalizing once.
- [ ] Unit: `seating_need` and `dietary_note` survive normalization at every scope.
- [ ] Unit: `NormalizeHouseholdAnswer` zeroes both seat counts when no member is `both`, and keeps them when one is.
- [ ] Unit: the scope predicates agree with the four enum values, exhaustively — a table test over all four, so a fifth value added later fails here first.
- [ ] Unit: an age on an adult and an out-of-range child age still report the `F5-B02` sentinels.

## Done when

- [ ] The scope gate exists in exactly one function, and nothing outside `domain` compares `attending` to a string literal.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
