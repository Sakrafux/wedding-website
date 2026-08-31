# `F5-B01` — Household store and CRUD endpoints

**Epic:** F5 — Admin: households & guests · **Layer:** backend · **Depends on:** `F1-B07`

## Story

As an admin, I want to create, read, update and delete households, so that the real guest list can be entered before the invitations are printed.

## Scope

**In:**

- `HouseholdStore`: list, create, update, delete. The read side already exists from `F1-B04`.
- `GET|POST /api/admin/households`, `GET|PATCH|DELETE /api/admin/households/{id}`.
- Admin DTOs, which **do** carry `code` and `admin_note`.
- Request validation with `validator/v10`.
- Audit rows for every mutation.

**Out:**

- Members → `F5-B02`. Code generation and regeneration → `F5-B03`. Exports → `F5-B04`.
- Editing the household's **answers** — `attending`, `meal_choice`, `portion`, `midnight_snack`, `rsvp_note`, `rsvp_submitted_at`. Those are `F3-B06`, which reaches them through the *same* use case the guest form uses, addressed by household id instead of by session. Duplicating the form here would mean two places where the field set has to stay in step, which is the one thing Gate 1 exists to prevent.

The line this epic draws: an admin may edit anything **we** record, and nothing the household **answered**. Transport seat counts, `has_stroller`, `seating_need` and `dietary_note` fall on our side despite looking like RSVP fields — they are logistics somebody tells us on the phone in March and we have to be able to write down.

## Instructions

1. Extend the existing `HouseholdStore`. `FindByID`, `FindByCode`, `ListMembers` and `TouchLastLogin` are already there and are not to be duplicated.
2. `List` returns every household with the columns the list screen needs: id, display name, code, `last_login_at`, `rsvp_submitted_at`, and a member count. One query with a `LEFT JOIN` and a `COUNT`, not a query per row — sixty households is small, but a loop that issues sixty-one queries is a habit, not a size.
3. Order by `display_name COLLATE NOCASE`. The admin scans this list looking for a name; insertion order is meaningless to that task.
4. Creating a household assigns a code in the same transaction, via `F5-B03`. A household without a code is a household nobody can log in as, and there is no screen that would show you that.
5. `PATCH` updates `display_name`, `admin_note`, `transport_seats_needed`, `transport_seats_offered` and `has_stroller`. It does **not** update `code` — regeneration is its own endpoint with its own consequences, and a code changed by a stray field in a form body is a code nobody knows changed.
6. Absent fields in a `PATCH` body leave the column alone; present-but-empty clears it. Use pointer fields in the request DTO so "not sent" and "sent as empty" are distinguishable — with a plain string they are the same value, and clearing an `admin_note` becomes impossible.
7. `DELETE` removes the row. `guest` cascades by foreign key; `seat_assignment` cascades from `guest`. Sessions are not deleted here and do not need to be: `Auth.ResolveSession` already treats a session whose household is gone as anonymous and deletes it on sight.
8. Deleting a household with answered RSVPs is allowed but must be deliberate — the frontend confirms (`F5-F02`), the API does not second-guess. `audit_log` keeps the record, which is the whole reason it outlives the row.
9. Validation with `validator/v10`, mapped into `httpio.ValidationError` keyed by the JSON field name. This is the first endpoint in the app with per-field rules; the mapping helper lands here and everything later reuses it.
   **When this ships, update the comment in `httpio/respond.go`** — it currently names `F3-B03` as where the validator mapping arrives.
10. Rules: `display_name` required, 1–120 characters. `admin_note` optional, max 2000. Transport seat counts 0–20 — a household is not a coach.
11. Every mutation writes an audit row: `actor_type = 'admin'`, `entity = 'household'`, `entity_id`, action `create` / `update` / `delete`. `before` and `after` carry **only the changed fields**, per `F1-B06`. Never put `code` in an audit payload — the audit log must not become a second copy of the key list.

## Contract

```http
GET /api/admin/households
```

Response `200`:

```json
{
  "households": [
    {
      "id": 12,
      "display_name": "Familie Müller",
      "code": "ABC234",
      "member_count": 4,
      "last_login_at": "2026-11-03T18:22:00Z",
      "rsvp_submitted_at": null
    }
  ]
}
```

```http
POST /api/admin/households
PATCH /api/admin/households/{id}
```

Request (`POST` requires `display_name`; `PATCH` takes any subset):

```json
{
  "display_name": "Familie Müller",
  "admin_note": "Kommen mit dem Zug",
  "transport_seats_needed": 0,
  "transport_seats_offered": 4,
  "has_stroller": false
}
```

Response `200` / `201`: the single household object above, plus `admin_note`, the transport counts and `has_stroller`.

```http
GET /api/admin/households/{id}
```

Response `200`: the household object, with its members embedded (`F5-B02` defines the member shape).

```http
DELETE /api/admin/households/{id}
```

Response `204`.

Errors: `not_found` → 404 · `validation_failed` → 400 with `fields` · `unauthenticated` → 401

> `code` and `admin_note` appear here **on purpose**. This is the one place in the API that may return them, and it is reachable only behind `RequireAdmin`.

## Test plan

- [ ] Integration: create → list → read → update → delete, end to end.
- [ ] Integration: a created household comes back with a code that satisfies `domain.ValidateCode`.
- [ ] Integration: `PATCH` with one field leaves the others untouched; `PATCH` with `admin_note: ""` clears it.
- [ ] Integration: `PATCH` cannot change `code` — send one and assert the stored code is unchanged.
- [ ] Integration: deleting a household removes its guests and its seat assignments.
- [ ] Integration: a session belonging to a deleted household resolves as anonymous, and its row is gone.
- [ ] Integration: validation failures return 400 with `fields` keyed by the JSON name, in German.
- [ ] Integration: every mutation writes exactly one audit row with the right action; no audit payload contains the code.
- [ ] **Privacy: the household-session table-drive from `F1-B07` still returns 401 for every one of these routes.** Adding a route is the moment the guard is forgotten.
- [ ] **Privacy: `assertNoLeak` is not weakened.** Admin responses legitimately carry `code` and `admin_note`, so admin tests simply do not call it. Do **not** remove entries from `forbiddenFields` — add a test that walks the non-admin routes and asserts a clean body for each, so the guard gets stronger here rather than weaker.

## Done when

- [ ] The real guest list can be entered through the API alone.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
