# `F4-B03` — `DELETE /api/rsvp/members/{id}`

**Epic:** F4 — Plus-one · **Layer:** backend · **Depends on:** `F4-B02`

## Story

As a guest, I want to remove somebody I added by mistake or who can no longer come, so that a wrong plus-one does not have to be corrected by phone.

## Scope

**In:**

- `RSVP.RemoveMember(ctx, householdID, guestID, options)`.
- `DELETE /api/rsvp/members/{id}`, behind `RequireHousehold`.
- The two refusals: not ours, and not guest-added.
- Deadline enforcement; audit as `delete`.

**Out:**

- Removing a seeded member. There is no path for it and there must not be: they answer `no`.
- Hard deletion → nothing does this to a guest; see `F5-B02`.

## Instructions

1. Soft delete, exactly as `F5-B02` does it. The row and its `deleted_at` remain, and the seat assignment stays for `F7` to report as stale.
2. A guest id belonging to another household is `not_found`, not `forbidden`. A household must not be able to probe which ids exist by reading the difference — the threat model is drive-by strangers, and this costs nothing.
3. A seeded member reports its own code: `cannot_remove_member` → 409 → *"Diese Person haben wir eingetragen. Wenn sie nicht kommt, wähl bitte «Kommt nicht» aus."* The message names the actual remedy, because the guest's goal is to say somebody is not coming and the form can do that.
4. Enforce the deadline through the `F3-B04` option.
5. Removing somebody who has already answered is allowed (`F4-B01`).
6. Removing the plus-one returns the household to one member, so it may add again (`F4-B01`). That is deliberate: the alternative is a household stuck with a typo they cannot fix.
7. Audit `entity = 'guest'`, action `delete`, `actor_type = 'household'`, `before` carrying the name and origin. This is the row that explains a headcount that went down.
8. Respond `204`. The frontend already knows which row it removed and refetching a whole form to learn it is gone is a round trip for nothing.

## Contract

```http
DELETE /api/rsvp/members/{id}
```

Response `204`.

Errors: `not_found` → 404 · `cannot_remove_member` → 409 → *"Diese Person haben wir eingetragen. Wenn sie nicht kommt, wähl bitte «Kommt nicht» aus."* · `rsvp_closed` → 409 · `unauthenticated` → 401

## Test plan

- [ ] Integration: removing a guest-added member drops them from `GET /api/rsvp` and leaves the row with `deleted_at` set.
- [ ] Integration: removing a seeded member → 409 `cannot_remove_member`, nothing written.
- [ ] Integration: removing another household's member → 404, nothing written.
- [ ] Integration: after the deadline → 409 `rsvp_closed`.
- [ ] Integration: the removal makes `can_add_plus_one` true again — the one path back from a mistyped name.
- [ ] Integration: exactly one audit row, `delete`, `actor_type = 'household'`.

## Done when

- [ ] A household can undo its own additions and cannot touch anybody else's, and no removal erases a person from the record.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
