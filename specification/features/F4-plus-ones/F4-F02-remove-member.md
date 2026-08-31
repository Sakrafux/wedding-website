# `F4-F02` — Remove member, pre-deadline only

**Epic:** F4 — Plus-one · **Layer:** frontend · **Depends on:** `F4-B03`, `F4-F01`

## Story

As a guest, I want to take somebody I added back off the list, so that a mistyped plus-one is a correction rather than a phone call.

## Scope

**In:**

- The remove control on guest-added member cards, and its confirmation.
- The absence of that control on seeded members, and what is offered instead.
- Handling the deadline and the two refusals.

**Out:**

- Removing seeded members. There is no path (`F4-B01`), and this story's job is to make that legible rather than to work around it.

## Instructions

1. Consume the `F4-B03` contract exactly.
2. The control appears **only** on cards with `origin = 'guest_added'`. On a seeded member there is no disabled button and no explanation nobody asked for — the way to say somebody is not coming is the scope control that is already on the card.
3. Label it "entfernen", not "löschen": it is a soft delete and the German should not promise otherwise. Same wording as `F5-F02`, for the same reason.
4. Confirm before removing, naming the person. One tap from a list is how the wrong child gets removed on a phone.
5. Removal takes effect immediately against the server; the card disappears on success. No optimistic removal — a card that vanishes and reappears is worse than a brief pending state, and this is a rare action.
6. Removing the plus-one brings the add trigger back (`F4-F01`), from the refreshed `can_add_plus_one`. It is the only way to fix a mistyped companion, so the path has to work.
7. `cannot_remove_member` should be unreachable from this UI, since the control is not rendered for seeded members. Handle it anyway and show the API sentence — it is the case where a stale tab is looking at a member whose origin we changed in admin.
8. After the deadline, no remove control is rendered at all: the whole page is read-only (`F3-F05`), and this story adds nothing to it.

## Test plan

- [ ] Component: the control renders on guest-added cards and not on seeded ones.
- [ ] Component: removal asks for confirmation naming the person, and only then calls the API.
- [ ] Component: on success the card disappears and the add trigger returns in place of the explanation.
- [ ] Component: a failed removal leaves the card in place with the API message shown.
- [ ] Component: in the read-only view no remove control exists.
- [ ] Accessibility: the confirmation traps focus; the control's accessible name includes the member's name, not just "entfernen".

## Done when

- [ ] A household can undo an addition in two taps, and cannot find any control suggesting they might remove somebody we invited.
- [ ] Checkbox ticked in `README.md`.
