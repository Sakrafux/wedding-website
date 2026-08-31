# `F3-B04` — Deadline enforcement, server-side

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F3-B03`

## Story

As an admin, I want the RSVP to stop accepting guest changes at the deadline while I can still record a late answer myself, so that the caterer's numbers are final for everyone except the phone call we take on the last evening.

## Scope

**In:**

- `rsvp_closed` as a `domain.ErrorCode`, its status and its German sentence in the `httpio` table.
- The deadline check as an **argument** to `RSVP.Save`, not a rule inside it.
- The guest handler passing "enforce"; `F3-B06`'s admin handler passing "do not".

**Out:**

- What the site shows guests after the deadline → `F3-F05` for the form; the wider question of what the whole site says in the five weeks between the deadline and published seating is still open in `TODO.md`.
- Changing the deadline → an admin settings endpoint, which is `F6`'s (`TODO.md`). Until then it is a row in `app_setting`, edited by hand.

## Instructions

1. `Save` takes an explicit option — `EnforceDeadline bool` on a small options struct, not a boolean parameter at a call site where nobody can see what `true` means. The admin path exists precisely for the answer that arrives late, so the rule cannot live inside the use case as an unconditional check.
2. The clock is read once per request and passed in, so a test drives the deadline by moving the setting rather than by waiting or by monkey-patching time.
3. Use `Settings.RSVPOpen(now)` — the derived predicate that already exists — rather than comparing timestamps at the call site. There is one definition of "open", and `/api/me`, `GET /api/rsvp` and this check all read it.
4. The check happens **before** any write and before validation is worth reporting: a closed form should say it is closed, not list three field errors first.
5. `rsvp_closed` → 409, message: *"Die Rückmeldefrist ist vorbei. Wenn sich etwas geändert hat, ruf uns bitte kurz an."* Named as a state, with the way out stated — the person reading it is a guest whose plans changed, and the phone call is the actual remedy. 409 rather than 403: nothing is wrong with who they are.
6. Reading stays open after the deadline. `GET /api/rsvp` never enforces anything; a household must be able to see what they answered, which is also what `F3-F05` renders.
7. `F4`'s add and remove endpoints enforce the same deadline the same way, through the same option. Note it here so the rule is written once and `F4-B02` cites it.

## Contract

No new endpoint. `PUT /api/rsvp` gains:

Errors: `rsvp_closed` → 409 → *"Die Rückmeldefrist ist vorbei. Wenn sich etwas geändert hat, ruf uns bitte kurz an."*

## Test plan

- [ ] Integration: with the deadline in the past, `PUT /api/rsvp` → 409 `rsvp_closed` and nothing is written.
- [ ] Integration: with the deadline in the past, `GET /api/rsvp` still returns 200 with `editable: false`.
- [ ] Integration: the deadline exactly now counts as closed — assert the boundary, since `RSVPOpen` is a strict `Before`.
- [ ] Integration: the admin path (`F3-B06`) writes successfully with the deadline in the past. Written here, run again there.
- [ ] Integration: a closed form reports `rsvp_closed`, not `validation_failed`, when the body is also invalid.

## Done when

- [ ] The deadline holds against a request that bypasses the frontend entirely, and does not hold against us.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
