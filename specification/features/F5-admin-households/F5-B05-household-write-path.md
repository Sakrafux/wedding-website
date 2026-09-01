# `F5-B05` — Transport counts and the pram leave the admin household write path

**Epic:** F5 — Admin: households & guests · **Layer:** backend · **Depends on:** `F5-B01`, `F3-B06`

## Story

As an admin, I want the fields a household answers to be writable in exactly one place, so that the RSVP page and the household page cannot disagree about what a household said.

## Scope

**In:**

- `transport_seats_needed`, `transport_seats_offered` and `has_stroller` removed from `AdminHouseholdCreateRequest` and `AdminHouseholdPatchRequest`.
- They stay in `AdminHousehold`, which is a read shape: the detail page shows them, read-only (`F5-F05`).

**Out:**

- Their editing path, which already exists and is not new work: `PUT /api/admin/households/{id}/rsvp` (`F3-B06`), the same form the household uses.

## Instructions

1. Three RSVP-answerable fields had two writers, and only one of them ran the RSVP rules — `F3-B07`'s direction rule among them. A second write path that skips the rules is not a shortcut, it is the bug report.
2. Unknown fields are already refused for the whole API (`DecodeJSON`), so a request that still sends them gets `validation_failed` rather than silence. That is the desired answer: the caller is using an endpoint that no longer owns those fields.
3. The create path takes the display name and the admin note, which is what `F5-F04`'s one-row form sends anyway.
4. `domain.HouseholdPatch` keeps its transport fields — `PUT /rsvp` patches through them. Only the two request DTOs shrink.

## Contract

```http
POST /api/admin/households
PATCH /api/admin/households/{id}
```

Request (both, any subset for `PATCH`):

```json
{ "display_name": "Familie Müller", "admin_note": "" }
```

Response: `AdminHousehold`, unchanged — transport counts and `has_stroller` still present, now read-only.

## Test plan

- [ ] Integration: a `PATCH` carrying `transport_seats_needed` is refused as `validation_failed`.
- [ ] Integration: a `PATCH` of `display_name` still works and leaves the transport counts alone.
- [ ] Integration: the RSVP `PUT` still writes them, which is the point of removing the other path.

## Done when

- [ ] The three fields have one writer, and it is the one that runs the RSVP rules.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
