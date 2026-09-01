# `F3-B07` — A household needs seats or offers them, never both

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F3-B01`, `F3-B03`

## Story

As an admin planning the shuttle, I want a household's transport answer to point in one direction only, so that the capacity gap is a number and not a contradiction.

## Scope

**In:**

- A domain rule refusing `transport_seats_needed > 0` together with `transport_seats_offered > 0`.
- The rule enforced on `PUT /api/rsvp` and on the admin PUT, as field errors.

**Out:**

- The control that makes the conflict unreachable in the browser → `F3-F07`.
- The capacity gap itself → `F6`. This story only keeps its inputs meaningful.
- The admin household `PATCH`, which stops carrying these fields entirely → `F5-B05`.

## Instructions

1. `domain.ValidateTransportSeats(needed, offered int) error` returns `ErrTransportSeatsConflict` when both are above zero, and nil otherwise. A sentinel, not a `domain.Error`, for the same reason `ErrAgeOnAdult` is one: the answer has to land next to a field, and field names are a shape the domain does not know.
2. A household that both needs and offers seats is not a case we cannot serve — it is a case we cannot *plan*: the pair feeds one subtraction, and a household on both sides of it inflates the gap and the supply at once. Refused rather than normalized, because the body said two things that cannot both be true and silently picking one would answer a question nobody asked.
3. Called from `rsvp.Save` after the member set matches and before the answer is applied. Not inside `ApplyHouseholdAnswer`: normalization zeroes both counts when nobody attends `both`, so a conflicting pair would sometimes be stored as a legal one and sometimes be refused, depending on scopes.
4. Mapped in `httpio` onto **both** field keys with the same sentence. The form shows one stepper at a time (`F3-F07`), so keying only one field is how the message ends up attached to a control that is not on screen.
5. Zero and zero stays legal, and so does either one alone. There is no "direction" column: the pair of counts *is* the direction, and a third field would be a second way to say the same thing.

## Contract

```http
PUT /api/rsvp
PUT /api/admin/households/{id}/rsvp
```

Errors: `validation_failed` → `400` → `Sagt uns bitte entweder, wie viele Plätze ihr braucht, oder wie viele ihr anbieten könnt — beides zusammen können wir nicht planen.` on `transport_seats_needed` and `transport_seats_offered`.

## Test plan

- [ ] Unit: both above zero is refused; either alone and both zero are accepted.
- [ ] Integration: the guest PUT answers 400 with both field keys set.
- [ ] Integration: the admin PUT is refused identically — the rule is not a guest-only courtesy.
- [ ] Negative: a conflicting pair sent by a household where nobody attends `both` is still refused, and not quietly zeroed by normalization.

## Done when

- [ ] No stored household both needs and offers seats, whichever endpoint wrote it.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
