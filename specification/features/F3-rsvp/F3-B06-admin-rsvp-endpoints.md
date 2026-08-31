# `F3-B06` — `GET|PUT /api/admin/households/{id}/rsvp`

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F3-B04`, `F3-B05`, `F5-B01`

## Story

As an admin, I want to record a household's answer myself, so that the call that comes in after the deadline ends up in the same place as everybody else's answer instead of on a note beside the laptop.

## Scope

**In:**

- `GET` and `PUT /api/admin/households/{id}/rsvp`, behind `RequireAdmin`.
- The **same** use case and the **same** DTOs as `F3-B02` / `F3-B03`, addressed by path id instead of by session.
- Deadline enforcement switched off for this path (`F3-B04`), audit as `admin` (`F3-B05`).

**Out:**

- Everything the admin may edit that is *not* an answer — name, private note, members — which is `F5`. The line is drawn in `F5-B01` and this story does not move it.
- A second RSVP DTO. If this story adds one, it has failed.

## Instructions

1. Reuse `RSVP.Load` and `RSVP.Save` unchanged. The only difference between the two routes is where the household id comes from and what the options say. That is the entire point of the arrangement — one form, one use case, one field set to keep in step (`TODO.md`, decided 2026-08-31).
2. Ownership: the guest handler passes the session's household, this one passes `chi.URLParam`. A household id in a path that the admin gate has already cleared needs no further check beyond existing; an unknown id is `not_found`.
3. `EnforceDeadline: false`. This path exists for the late call.
4. `editable` in the `GET` response stays the honest report of `Settings.RSVPOpen` — the admin frontend shows the deadline state as information ("Frist ist abgelaufen — du kannst trotzdem speichern") rather than as a lock. Do not overload the field to mean "this caller may write"; two meanings on one boolean is how the guest route eventually gets it wrong.
5. `member_set_mismatch` applies here too, for the same stale-tab reason.
6. An admin save **sets `rsvp_submitted_at`** if it was null. If we took the answer down the phone, the household has answered and must drop off the nudge list — decided 2026-08-31.
7. Audit rows carry `actor_type = 'admin'` with `actor_id` NULL (`F3-B05`).
8. Register both routes inside the existing `/admin/households/{id}` subtree in `router.go`, beside the code and guest routes.

## Contract

```http
GET /api/admin/households/{id}/rsvp
PUT /api/admin/households/{id}/rsvp
```

Request and response bodies are **identical** to `F3-B02` and `F3-B03`. No admin-only fields are added: `code` and `admin_note` belong to the household endpoints, and adding them here would mean the shared frontend component had to know which caller it was serving.

Errors: `not_found` → 404 · `validation_failed` → 400 with `fields` · `member_set_mismatch` → 409 · `unauthenticated` → 401. Notably **not** `rsvp_closed`.

## Test plan

- [ ] Integration: an admin reads and writes a household's answer by id, and the guest sees the result through `GET /api/rsvp`.
- [ ] Integration: with the deadline in the past, the admin `PUT` succeeds and the guest `PUT` does not.
- [ ] Integration: an admin save on a household that never answered sets `rsvp_submitted_at`.
- [ ] Integration: the audit row says `admin`, and the guest route's says `household`, for the same change.
- [ ] Integration: an unknown household id → 404.
- [ ] Integration: a household session on these routes → 401, added to the `F1-B07` table drive.
- [ ] Structural: the response bodies of the guest and admin `GET` are byte-identical for the same household. Assert it, so that a field added to one and not the other fails a test rather than a screen.

## Done when

- [ ] The answer we take on the phone is indistinguishable in the data from one a household typed, except in the audit log, where it must be distinguishable.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
