# `F1-B04` — Login, logout, and `/api/me`

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** `F1-B01`, `F1-B03`

## Story

As a guest, I want to type the code from my card and be logged in as my household, so that I can see our invitation and answer it.

## Scope

**In:**

- `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/me`.
- Cookie issuance with the specified attributes.
- The DTOs the frontend will consume.

**Out:**

- Rate limiting → `F1-B05`. Audit logging → `F1-B06`. Admin login → `F1-B07`.

## Instructions

1. Normalise the submitted code with `NormalizeCode` before lookup. Malformed shape → the same generic failure as unknown code; do not tell a caller their code "looks valid but does not exist".
2. Constant-ish behaviour on failure: one generic error, comparable timing. Not a hard requirement at this threat level, but free to get right.
3. On success: create the session, set the cookie — `HttpOnly`, `Secure` (from `SESSION_COOKIE_SECURE`), `SameSite=Lax`, `Path=/`, `Max-Age` matching the session lifetime.
4. Update `household.last_login_at`. This is what answers "did they even see it?" and drives the nudge list.
5. `GET /api/me` returns the household, its members, and the runtime flags from `app_setting`. This is the frontend's bootstrap call — one request that tells the app who it is talking to and what is switched on.
6. **The response DTO must not contain `household.code` or `admin_note`.** Comment the omission in the DTO struct so nobody later "completes" it. This is the exact leak the DTO rule exists to prevent.
7. Logout deletes the session row and clears the cookie. Idempotent: logging out twice is a 204 both times.
8. Login while already logged in: issue a fresh session and delete the old one. A guest re-entering a code should end up on the household the code names, not the one the cookie remembers — this is how a wrong-household session gets corrected in practice.

## Contract

```http
POST /api/auth/login
```

Request:

```json
{ "code": "abc-234" }
```

Response `200`:

```json
{
  "household": { "id": 12, "display_name": "Familie Müller" },
  "members": [
    { "id": 30, "first_name": "Anna", "last_name": "Müller", "kind": "adult", "origin": "seeded" }
  ],
  "flags": { "rsvp_open": true, "seating_published": false, "gallery_visible": false, "uploads_open": false },
  "rsvp_deadline": "2027-05-16T21:59:00Z"
}
```

Errors:

- `unknown_login_code` → 401 → "Diesen Code kennen wir nicht. Schau bitte noch mal auf deine Karte — Groß- und Kleinschreibung ist egal."
- `rate_limited` → 429 → (see `F1-B05`)

```http
GET /api/me
```

Response `200`: identical body to login. `401` when anonymous.

```http
POST /api/auth/logout
```

Response `204`.

## Test plan

- [ ] Integration: valid code → 200, cookie set with `HttpOnly`, `SameSite=Lax`, correct `Max-Age`.
- [ ] Integration: lowercase, spaced and dashed variants of the same code all succeed.
- [ ] Integration: unknown code → 401 `unknown_login_code`.
- [ ] Integration: malformed code → the same 401, not a distinguishable error.
- [ ] Integration: `last_login_at` is updated.
- [ ] Integration: `/api/me` anonymous → 401; with cookie → the household.
- [ ] Integration: `assertNoLeak` on both login and `/api/me` responses — no `code`, no `admin_note`.
- [ ] Integration: logout → cookie cleared, subsequent `/api/me` is 401, session row gone.
- [ ] Integration: logging in as household B while holding A's cookie yields B, and A's session is deleted.
- [ ] Integration: `Secure` is absent when `SESSION_COOKIE_SECURE=false`, present otherwise.

## Done when

- [ ] `curl` with a real code returns the household and a usable cookie.
- [ ] Checkbox ticked in `README.md`.
