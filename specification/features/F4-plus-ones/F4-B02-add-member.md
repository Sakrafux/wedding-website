# `F4-B02` — `POST /api/rsvp/members`

**Epic:** F4 — Plus-one · **Layer:** backend · **Depends on:** `F4-B01`, `F3-B03`

## Story

As a guest invited on my own, I want to add the person I am bringing, so that we are counted as two without a phone call about a name we already know.

## Scope

**In:**

- `RSVP.AddPlusOne(ctx, householdID, name, options)`.
- `POST /api/rsvp/members`, behind `RequireHousehold`.
- `can_add_plus_one` on the `F3-B02` response, so the form knows which of the two states to render.
- Refusal with `plus_one_not_allowed` → 409; deadline through the `F3-B04` option; audit as `create`.

**Out:**

- Removal → `F4-B03`. Answering for the plus-one → the ordinary `PUT /api/rsvp`, which now includes them.
- Adding children, or a second person, or anybody to a household of two. Not a gap: `F5-B02` is the path, and `F4-B01` says why.
- An approval or pending state. Rejected in `02-features`: the addition counts immediately and shows up as a delta.

## Instructions

1. The request body is **one field, the name**. No `kind`, no `age`, no meal — a plus-one is an adult by construction (`F4-B01`), and the answers come through the RSVP form like everybody else's. A body that carried `kind` would be a body somebody eventually sets to `child`.
2. `origin = 'guest_added'`, always, ignoring anything the body says. The delta view rests on that column and it is not an input.
3. Check `CanHouseholdAddPlusOne` against the household's living members **inside the write transaction**, not before it. Two phones submitting at once is unlikely and cheap to exclude.
4. Refusal is `plus_one_not_allowed` → 409 → *"Weitere Personen tragen wir gern für euch ein — ruf uns bitte kurz an: +43 650 9408100."* One sentence for every reason (`F4-B01`): it does not say which rule was hit, because the answer is the same and the guest's next step is the same.
5. Enforce the deadline through the same option as `F3-B04`. After it, `rsvp_closed` — checked first, since a closed form is a closed form regardless of who is being added.
6. Validation: `name` required, 1–160 characters, matching `F5-B02`. One rule set, two callers.
7. Extend the `F3-B02` response with **`can_add_plus_one`** — a boolean, computed from the same domain function the endpoint enforces with. A boolean rather than the counts, because the frontend has one decision to make and two integers would let it re-derive the rule and get it wrong.
8. Audit `entity = 'guest'`, action `create`, `actor_type = 'household'`, `after` carrying name, kind and origin. This is the row that later answers "where did this person come from".
9. Return the created member in the `F3-B02` member shape, plus the recomputed `can_add_plus_one` (now false), so the frontend appends and re-renders without a refetch.

## Contract

```http
POST /api/rsvp/members
```

Request:

```json
{ "name": "Isabella Michelbacher" }
```

Response `201`:

```json
{
  "member": {
    "id": 44,
    "name": "Isabella Michelbacher",
    "kind": "adult",
    "age": null,
    "origin": "guest_added",
    "attending": null,
    "meal_choice": null,
    "portion": "full",
    "midnight_snack": false,
    "seating_need": "normal",
    "dietary_note": ""
  },
  "can_add_plus_one": false
}
```

Errors: `validation_failed` → 400 with `fields` · `plus_one_not_allowed` → 409 · `rsvp_closed` → 409 · `unauthenticated` → 401

## Test plan

- [ ] Integration: a single-person household adds one, and the member appears in `GET /api/rsvp` as an unanswered adult with `origin = 'guest_added'`.
- [ ] Integration: the second attempt → 409 `plus_one_not_allowed`, nothing written.
- [ ] Integration: a household seeded with two members → 409 on the first attempt.
- [ ] Integration: removing the plus-one (`F4-B03`) makes `can_add_plus_one` true again, and the next add succeeds.
- [ ] Integration: a body carrying `kind: "child"` or an `age` is rejected or ignored — decide which and assert it; either way no child is ever created here.
- [ ] Integration: after the deadline → 409 `rsvp_closed`, and that code wins over `plus_one_not_allowed` when both apply.
- [ ] Integration: the admin path adds a member to a household that has already used its plus-one.
- [ ] Integration: exactly one audit row, `create`, `actor_type = 'household'`.
- [ ] Privacy: the response passes `assertNoLeak`.

## Done when

- [ ] A guest invited alone can add their companion and answer for them, no guest path can produce a child or a third member, and we can still add anybody by hand.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
