# `F3-B02` — `GET /api/rsvp`

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F3-B01`, `F1-B04`

## Story

As a guest, I want the form to open showing exactly what my household answered last time, so that changing one person's meal does not mean filling the whole thing in again.

## Scope

**In:**

- The `RSVP` use case in `application`, with `Load(ctx, householdID)` as its read side.
- `GET /api/rsvp`, behind `RequireHousehold`, reading the household from the session.
- `dto.RSVPResponse`: the household's answer fields, every living member with their answers, the deadline and whether the form is editable.
- Reading the RSVP answer columns in `GuestStore` / `HouseholdStore`.

**Out:**

- Writing → `F3-B03`. Deadline *enforcement* → `F3-B04`; this response only reports the state.
- The admin-addressed variant → `F3-B06`, which reuses `Load` with a path id.
- `can_add_plus_one` → `F4-B02` extends this response; do not invent the field here.

## Instructions

1. One use case, `application/rsvp`, taking a household id. The guest handler passes `session.SubjectID`; the admin handler in `F3-B06` passes the path parameter. Ownership is therefore checked in one place, and the two routes cannot drift.
   This is the second use case in `application`, which is the moment the package split noted in `TODO.md` earns itself: `application/auth`, `application/households`, `application/rsvp`. Do it here or record why not.
2. The response carries the whole form's data in one body. A second request for the members would let the screen render a household with nobody in it.
3. Members are the living ones, ordered as `ListMembers` already orders them, so the form and the admin detail page list people in the same order. A form that reshuffles between visits makes an unconfident guest doubt they are looking at the right household.
4. `attending` is `null` for an unanswered guest and the frontend renders "Noch keine Antwort" from it. Do not substitute a default — a defaulted `no` on the wire is an answer nobody gave.
5. `editable` is `Settings.RSVPOpen(now)`. It exists so the frontend can render the read-only view (`F3-F05`) without doing date arithmetic against a timezone it does not have. `F3-B04` is what actually refuses a write; this field is a hint and the DTO comment says so.
6. `deadline` is repeated here even though `/api/me` already carries it, because this response is what the form renders from and a form that reads its deadline from a different endpoint's cache can show a stale date after we move it.
7. `rsvp_submitted_at` and `rsvp_updated_at` are in the response: the frontend uses the first to decide between "Zusagen" and "Antwort ändern" (`05-design`, nav rule), and the second for "zuletzt geändert am …" above the form.
8. **`code`, `admin_note` and `rsvp_note_seen_at` are omitted, deliberately.** The first two for the reasons `dto.HouseholdSummary` already documents; `rsvp_note_seen_at` because whether we have read a note is our business and a guest who saw an unread marker would reasonably start chasing us. Comment all three in the DTO.

## Contract

```http
GET /api/rsvp
```

Response `200`:

```json
{
  "household": {
    "id": 12,
    "display_name": "Familie Müller",
    "transport_seats_needed": 0,
    "transport_seats_offered": 2,
    "has_stroller": false,
    "rsvp_note": "Wir kommen erst nach der Zeremonie.",
    "rsvp_submitted_at": "2026-11-09T19:04:00Z",
    "rsvp_updated_at": "2026-12-01T08:12:00Z"
  },
  "members": [
    {
      "id": 31,
      "name": "Anna Müller",
      "kind": "adult",
      "age": null,
      "origin": "seeded",
      "attending": "both",
      "meal_choice": "vegetarian",
      "portion": "full",
      "midnight_snack": true,
      "seating_need": "normal",
      "dietary_note": "Nussallergie"
    }
  ],
  "deadline": "2027-05-17T21:59:59Z",
  "editable": true
}
```

Errors: `unauthenticated` → 401

## Test plan

- [ ] Integration: a household with answers reads them back unchanged.
- [ ] Integration: a household that has never answered gets `attending: null` for every member and `rsvp_submitted_at: null`.
- [ ] Integration: soft-deleted members are absent.
- [ ] Integration: `editable` is false when the deadline has passed — drive it by writing `rsvp_deadline` in `app_setting`, not by waiting.
- [ ] Integration: an admin session gets 401 here. The admin route is `F3-B06`; the guest route resolves its household from the session and an admin session has none.
- [ ] Privacy: `assertNoLeak` over this response — no `code`, no `admin_note`, no `rsvp_note_seen_at`.

## Done when

- [ ] The form can be rendered from one request, including for a household that has never answered.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
