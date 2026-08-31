# `F3-B03` — `PUT /api/rsvp` with validation

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F3-B02`

## Story

As a guest, I want one save that writes my household's whole answer at once, so that pressing "Speichern" either stores everything I filled in or nothing at all.

## Scope

**In:**

- `RSVP.Save(ctx, householdID, submission, options)` — one transaction over the household row and every member row.
- `PUT /api/rsvp`, behind `RequireHousehold`.
- Request validation with `validator/v10`, per-field German messages.
- The two new failure codes: `rsvp_closed` (`F3-B04` owns when it fires) and `member_set_mismatch`.
- Setting `rsvp_submitted_at` on the first save and `rsvp_updated_at` on every save.

**Out:**

- Deadline rule → `F3-B04`. Audit rows → `F3-B05`. Both land on this endpoint; both are their own story so that neither is "done" by being half-written here.
- Adding or removing members → `F4-B02`, `F4-B03`. This endpoint answers *for* the members that exist and never changes the set.

## Instructions

1. `PUT`, not `PATCH`, and the body carries the **complete** answer for the household. The form is one screen with one save button (`05-design`), so a partial body would be a shape no client produces; a full replace also makes the request idempotent, which matters on a phone that retries.
2. The body must list **exactly** the household's living members, by id. A member in the body that the household does not own, a member missing from the body, or a duplicate id → `member_set_mismatch`, 409, *"Die Liste der Personen hat sich geändert. Bitte lade die Seite neu."* This is the stale-tab case: a household that added a plus-one on a phone and still has the form open on a laptop would otherwise silently write the older set back. Refusing is the only answer that cannot lose an answer, and 409 rather than 400 because nothing in the body is malformed — the world moved.
3. Every member's `attending` is **required** on save. There is no way to store half an answer: the whole point of `rsvp_submitted_at` is that a household appears on the nudge list until they have actually told us who is coming, and a save that leaves two people at `null` while setting the submitted timestamp makes that list lie. The frontend keeps this from being a surprise by marking the unanswered cards before submit (`F3-F04`).
4. Run everything through `NormalizeGuestAnswer` and `NormalizeHouseholdAnswer` (`F3-B01`) **before** writing. Validation rejects what is malformed; normalization decides what is meaningless. A `meal_choice` sent for a `church_only` guest is therefore not an error — the form may legitimately still have it in state when somebody flips the scope last — it is simply not stored.
5. One transaction, on the write connection. The household row and n member rows are one answer; a partial write is a household that told us four people are coming and one meal.
6. `rsvp_submitted_at` is set on the first successful save only and never moved afterwards; `rsvp_updated_at` is set on every successful save. The pair is what `F6` reads for "answered" versus "changed since we last looked", and overwriting the first would erase the distinction.
7. A save that changes nothing still updates `rsvp_updated_at`? **No.** Compare against the loaded state and, if nothing changed, return `200` with the current body and write neither timestamp nor audit row — same rule as `Households.Update` in `F5-B01`, and for the same reason: `rsvp_note_seen_at` is compared against `rsvp_updated_at` to decide whether a note is unread, so a no-op save would silently re-flag a note we have already read.
8. Validation rules, keyed by JSON field name: `attending` required and one of the four values; `meal_choice` one of three when present; `portion` one of three; `seating_need` one of four; `dietary_note` max 500; `rsvp_note` max 2000; `transport_seats_needed` / `transport_seats_offered` 0–20; `age` 0–17 and only on `kind = 'child'`, reusing the `F5-B02` domain rule so the message is a field error and not a driver error.
9. Field errors are keyed per member: `members.31.attending`, using the member **id** and not the array index, because the frontend renders member cards keyed by id and an index would break the moment the list is filtered. Document the key shape in the DTO — it is a contract `F3-F04` renders from.
10. **When this ships, remove the `F3-B03` forward reference in `tests/integration/error_envelope_test.go`** — it names this story as where the first per-field rules arrive. `F5-B01` already delivered the validator mapping; check the comment in `httpio/respond.go` too and delete whichever pointer is now stale.

## Contract

```http
PUT /api/rsvp
```

Request:

```json
{
  "transport_seats_needed": 0,
  "transport_seats_offered": 2,
  "has_stroller": false,
  "rsvp_note": "Oma braucht einen Platz nah am Ausgang.",
  "members": [
    {
      "id": 31,
      "attending": "both",
      "meal_choice": "vegetarian",
      "portion": "full",
      "midnight_snack": true,
      "seating_need": "normal",
      "dietary_note": "Nussallergie",
      "age": null
    }
  ]
}
```

Response `200`: the `F3-B02` body, as stored — normalized, so the client sees what was actually kept rather than what it sent.

Errors: `validation_failed` → 400 with `fields` · `member_set_mismatch` → 409 · `rsvp_closed` → 409 (`F3-B04`) · `unauthenticated` → 401

## Test plan

- [ ] Integration: a full save round-trips through `GET /api/rsvp`.
- [ ] Integration: saving with a `meal_choice` on a `church_only` member stores no meal choice, and the response says so.
- [ ] Integration: first save sets both timestamps; the second save moves only `rsvp_updated_at`.
- [ ] Integration: a save that changes nothing moves neither timestamp.
- [ ] Integration: a member id from another household → 409 `member_set_mismatch`, and nothing is written.
- [ ] Integration: a body missing one of the household's members → 409, nothing written.
- [ ] Integration: a member with `attending: null` → 400 with `members.<id>.attending` in `fields`.
- [ ] Negative: a failure partway through leaves no partial write — force it (e.g. an invalid age on the last member) and assert the household row is untouched.
- [ ] Privacy: the response passes `assertNoLeak`.

## Done when

- [ ] A household's complete answer can be submitted and re-submitted through the API alone, and no combination of inputs stores a catering answer for somebody who is not at the party.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
