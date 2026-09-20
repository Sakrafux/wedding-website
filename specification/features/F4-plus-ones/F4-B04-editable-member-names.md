# `F4-B04` — The household writes its own names

**Epic:** F4 — Plus-one · **Layer:** backend · **Depends on:** `F3-B03`, `F5-B02`

## Story

As a guest, I want to correct how the people in my household are called, so that the place card, the list and the greeting say the name the person actually goes by rather than the one we copied off an address book.

## Scope

**In:**

- `name` on the member objects of `PUT /api/rsvp` — required, 1–160 characters, trimmed, refused when blank.
- `domain.ResolveGuestName`, shared by the RSVP save and the admin patch, so one rule decides what a stored name looks like.
- `guest.seeded_name` (migration `0006`): the name we posted the invitation to, kept when the household rewrites `name`.
- `seeded_name` on `dto.AdminGuest`, and the `seeded_name` column in `guests.csv`.
- The rename is diffed and audited like every other RSVP field.

**Out:**

- `household.addressee` and `household.name`. The greeting and the file-by name stay ours (`F5-B02`) — a household editing how we file it is a change to our list, not to their answer.
- `kind`, `origin` and the member list itself. Unchanged from `F3-B03` and `F4-B01`, and for the same reasons.
- Any admin view of the rename beyond the one line on the detail page (`F4-F04`). No history screen, no revert button — the audit log already holds the before and after.

## Instructions

1. Add `name` to `dto.RSVPMemberRequest`, `validate:"required,max=160"` — the same bound as `AdminGuestCreateRequest.Name`, one rule set for all three callers. It goes on the **existing** save: one screen, one save button, and a separate rename endpoint would be a second write path with its own deadline and audit rules to keep in step.
2. `domain.ResolveGuestName` trims and refuses a blank result with `ErrEmptyName`. Trimmed rather than refused for stray whitespace, because a name pasted off a phone keyboard routinely carries a trailing space; blank is refused because nobody can seat, cater for or address an unnamed guest.
3. `httpio.GuestFieldValidationErrorUnder` maps `ErrEmptyName` to the `name` field, so the message lands on the right card.
4. **`seeded_name` has exactly one writer, and it is not this path.** `RSVPStore.SaveAnswer` writes `name`; `GuestStore.Update` writes both. `domain.ApplyGuestPatch` moves the seeded name along with the name for a `seeded` guest — an admin rename is us fixing our own typo, and a detail page reporting "Eingeladen als …" for a typo we corrected would be noise.
5. `seeded_name` is empty for a `guest_added` member, and stays empty. A copy of `name` there would tell a later reader we invited somebody we invented; `origin` and an empty column say the truth twice.
6. `seeded_name` never reaches a guest-facing DTO. Not a privacy leak in the `admin_note` sense — it is the household's own former name — but handing back a name they deliberately replaced is the app arguing with them.
7. The rename is subject to the deadline like the rest of the answer (`F3-B04`), and audited through the ordinary `Changes` diff.

## Contract

`PUT /api/rsvp` and `PUT /api/admin/households/{id}/rsvp`, request members gain one field:

```json
{
  "id": 30,
  "name": "Oma Erika",
  "attending": "both",
  "meal_choice": "all",
  "portion": "full",
  "midnight_snack": false,
  "seating_need": "normal",
  "dietary_note": "",
  "age": null
}
```

`GET /api/admin/households/{id}` members gain one field:

```json
{ "id": 30, "name": "Oma Erika", "seeded_name": "Erika Huber", "kind": "adult", "origin": "seeded" }
```

Errors: `validation_failed` → 400 with `fields["members.<id>.name"]` → *"Bitte gib einen Namen an."*

## Test plan

- [x] Unit: `ResolveGuestName` trims, and refuses a name that is blank once trimmed.
- [x] Unit: `ApplyGuestAnswer` stores the trimmed name, reports it in `Changes`, and leaves `SeededName` alone.
- [x] Unit: `ApplyGuestPatch` moves the seeded name for a `seeded` guest and leaves it empty for a `guest_added` one.
- [x] Integration: a household renames a member; the admin detail shows the new name and `seeded_name` still holds the invited one.
- [x] Integration: an admin rename moves both, so the detail page reports no rename.
- [x] Integration: a blank name → 400 keyed `members.<id>.name`, and nothing is written.
- [x] Integration: a 161-character name → 400.
- [x] Privacy: no `seeded_name` anywhere in the guest RSVP body.
- [x] Integration: `guests.csv` carries `seeded_name` — covered by the existing column-parity test against the table.

## Done when

- [x] A household can fix a name itself, and we can still tell whose invitation card that person is.
- [x] Tests above pass; `go test ./...` is green.
- [x] Checkbox ticked in `README.md`.
