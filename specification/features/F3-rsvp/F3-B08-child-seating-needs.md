# `F3-B08` — `with_parent` and `high_chair` are children only

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F3-B01`, `F5-B02`

## Story

As an admin drawing the seating plan, I want a lap seat and a high chair to be recordable for children only, so that a seat count is never lost to an adult who was marked as sitting on somebody's lap.

## Scope

**In:**

- `domain.ResolveSeatingNeed`, refusing `with_parent` and `high_chair` for `kind = adult`.
- Enforced on every write that sets `seating_need`: the RSVP save, the admin guest create, the admin guest patch.
- The `kind`-change case: turning a child with a high chair into an adult is refused, not silently reset.

**Out:**

- Hiding the two options in the form → `F3-F08`.
- Anything about *where* a high chair goes; church seating has no high chairs, which is copy (`F3-F08`) rather than a second field.

## Instructions

1. `ResolveSeatingNeed(kind GuestKind, need SeatingNeed) (SeatingNeed, error)` mirrors `ResolveAge`: it returns the value to store, or `ErrSeatingNeedOnAdult`. Same shape, same reason — the rule is a business fact and the field name and German sentence are the web layer's.
2. `with_parent` means the person consumes no seat of their own (`F7` must not assign them one) and `high_chair` means a chair we hire per child. Both are refused for an adult because both would corrupt a count we buy things with, and an adult who needs a special seat is what `wheelchair` and the household note are for.
3. Called from `ApplyGuestAnswer`, from `ApplyGuestPatch` — after `kind` is settled, so the refusal describes the state the request asked for — and from the admin guest create path in `application/households`.
4. Refused rather than reset on a `kind` change: resetting would answer a question the request did not ask, and the admin who just turned a child into an adult is the one person able to say what that person now needs.
5. `httpio`'s age translator grows into one guest-field translator, since both rules now land as field errors on the same bodies. Keep the member-id prefix (`members.<id>.seating_need`) so the message reaches the right card.

## Contract

```http
PUT /api/rsvp
PUT /api/admin/households/{id}/rsvp
POST /api/admin/households/{id}/guests
PATCH /api/admin/guests/{id}
```

Errors: `validation_failed` → `400` → `Hochstuhl und „sitzt bei den Eltern“ tragen wir nur für Kinder ein.` on `seating_need`, or `members.<id>.seating_need` on the RSVP bodies.

## Test plan

- [ ] Unit: both child-only values are refused for an adult and accepted for a child; `normal` and `wheelchair` are accepted for both.
- [ ] Unit: patching a child with `high_chair` to `kind = adult` is refused, and the guest is left unchanged.
- [ ] Integration: the RSVP PUT keys the error to the member.
- [ ] Integration: the admin guest create and patch both refuse it.

## Done when

- [ ] No adult in the database carries a child-only seating need, whichever endpoint wrote it.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
