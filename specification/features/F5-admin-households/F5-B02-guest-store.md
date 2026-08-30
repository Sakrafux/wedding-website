# `F5-B02` — Guest store and CRUD endpoints

**Epic:** F5 — Admin: households & guests · **Layer:** backend · **Depends on:** `F5-B01`

## Story

As an admin, I want to add, edit and remove the people in a household, so that the guest list holds actual names before the cards are printed and the caterer is briefed.

## Scope

**In:**

- `GuestStore`: create, update, soft delete. `ListMembers` already exists from `F1-B04`.
- `POST /api/admin/households/{id}/guests`, `PATCH /api/admin/guests/{id}`, `DELETE /api/admin/guests/{id}`.
- The `kind` / `age` invariant, enforced in the domain rather than only by the `CHECK`.
- Audit rows for every mutation.

**Out:**

- Households themselves → `F5-B01`.
- Households adding their own members → `F4-B02`; that path sets `origin = 'guest_added'` and obeys the soft cap, neither of which applies here.
- Editing the household's answers (`attending`, `meal_choice`, `portion`, `midnight_snack`) → `F3`. `seating_need` and `dietary_note` are editable here: see the line drawn in `F5-B01`.

## Instructions

1. Guests created here are always `origin = 'seeded'`. That column is what the admin delta view reads to answer "what did the households add themselves", and an admin-created guest is not that.
2. `age` is **age at the wedding date** and is only meaningful for `kind = 'child'`. Enforce the pairing in a domain function, not only in the SQL `CHECK`: the constraint gives a driver error, and a driver error is not a message a form can put next to a field.
3. Changing `kind` from `child` to `adult` clears `age` in the same statement. Leaving a stale age behind would violate the column `CHECK` and, worse, would quietly feed a caterer bracket.
4. `PATCH` takes any subset, pointer fields, same rule as `F5-B01`: absent leaves alone, present-and-empty clears.
5. `DELETE` is a **soft** delete — set `deleted_at`. The guest was counted, may hold a seat assignment, and appears in the audit trail; erasing the row would leave those dangling and the history unexplainable. Every read filters `deleted_at IS NULL`, which `ListMembers` already does.
6. A soft-deleted guest keeps their `seat_assignment` row rather than losing it silently. `F7-B01` reports it as a stale assignment for a human to resolve — seats must not quietly vanish from a finished plan.
7. Deleting the last member of a household is allowed. An empty household is a real state: we know a name, we have not yet asked who is coming.
8. Validation: `first_name` and `last_name` required, 1–80 characters each. `kind` must be one of the enum values. `age` 0–17 when present, and rejected outright when `kind = 'adult'` — a child of 18 is an adult, and the caterer brackets are drawn below that.
9. `dietary_note` and `seating_need` are editable here too: both are things a household tells us by phone as often as through the form.
10. Audit every mutation with `entity = 'guest'` and the guest id. `before`/`after` carry only the changed fields.

## Contract

```http
POST /api/admin/households/{id}/guests
PATCH /api/admin/guests/{id}
```

Request (`POST` requires the names and `kind`; `PATCH` takes any subset):

```json
{
  "first_name": "Emil",
  "last_name": "Müller",
  "kind": "child",
  "age": 4,
  "seating_need": "high_chair",
  "dietary_note": "Nussallergie"
}
```

Response `200` / `201`:

```json
{
  "id": 31,
  "household_id": 12,
  "first_name": "Emil",
  "last_name": "Müller",
  "kind": "child",
  "age": 4,
  "origin": "seeded",
  "seating_need": "high_chair",
  "dietary_note": "Nussallergie"
}
```

```http
DELETE /api/admin/guests/{id}
```

Response `204`.

Errors: `not_found` → 404 · `validation_failed` → 400 with `fields` · `unauthenticated` → 401

## Test plan

- [ ] Integration: create → read back through the household detail → update → delete.
- [ ] Integration: a guest created here has `origin = 'seeded'`.
- [ ] Unit: `age` on an adult is rejected by the domain rule, with a field error rather than a driver error.
- [ ] Unit: changing `kind` from `child` to `adult` clears `age`.
- [ ] Integration: a soft-deleted guest disappears from `ListMembers` but the row and its `deleted_at` remain.
- [ ] Integration: a soft-deleted guest's `seat_assignment` still exists — it is `F7`'s job to report it, not this story's to delete it.
- [ ] Integration: deleting the last member leaves the household intact.
- [ ] Integration: `PATCH` on a guest of another household still works — guests are addressed by their own id, and the admin owns all of them. This is asserted so the route is not later "fixed" into a household-scoped one that the frontend does not call.
- [ ] Integration: household-session and anonymous requests to every route → 401.
- [ ] Integration: each mutation writes exactly one audit row.

## Done when

- [ ] A household's real members can be entered, corrected and removed through the API alone.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
