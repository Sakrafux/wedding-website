# `F1-B07` — Admin login and the `/api/admin` gate

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** `F1-B03`, `F1-B05`

## Story

As an admin, I want a separate short-lived login, so that the planning data — and the budget above all — is genuinely unreachable from a guest session rather than merely unlinked.

This is the one hard security boundary in the product. Everything else is calibrated to trusted guests; this is not.

## Scope

**In:**

- `POST /api/auth/admin/login`.
- `GET /api/admin/me` — confirms an existing admin session. Pulled in here because `F1-F04`'s admin guard cannot be written without it: `/api/me` answers 401 for an admin, so nothing else reports that the cookie in hand is still an admin session, and a client-side flag would put "am I the admin" in a second place.
- The `/api/admin` subtree gate.
- Constant-time credential comparison.

**Out:**

- Any admin feature endpoint → F5, F6, F8.

## Instructions

1. Compare both username and password with `subtle.ConstantTimeCompare`. Compare **both** even when the username already mismatched — an early return leaks which half was wrong via timing, and the cost of not returning early is nil.
2. Credentials come from `ADMIN_USER` / `ADMIN_PASSWORD`. No `admin_user` table, no reset flow.
3. On success create a session with `subject_type = 'admin'`, `subject_id = NULL`, 8-hour lifetime, no rolling refresh.
4. Same cookie name and attributes as a household session. The subject type distinguishes them; two cookie names would double the ways to get this wrong.
5. Mount `RequireAdmin` on the whole `/api/admin` subtree at the router, once.
6. Failed admin logins are audit-logged and rate-limited at the stricter admin threshold from `F1-B05`.
7. Generic error message. Do not distinguish unknown user from wrong password.
8. Logging in as admin while holding a household cookie replaces the session. Logging in as a household while holding an admin cookie likewise. One cookie, one subject.

## Contract

```http
POST /api/auth/admin/login
```

Request:

```json
{ "user": "…", "password": "…" }
```

Response `200`:

```json
{ "subject_type": "admin" }
```

Errors: `invalid_credentials` → 401 → "Anmeldung fehlgeschlagen." · `rate_limited` → 429

## Test plan

- [ ] Integration: correct credentials → 200, admin cookie set, 8-hour expiry.
- [ ] Integration: wrong password → 401; wrong user → the identical 401 body.
- [ ] Integration: admin session reaches an `/api/admin` route.
- [ ] Integration: **household session on every `/api/admin` route → 401.** Table-drive this across all admin routes that exist, so a route added later without the guard fails the suite.
- [ ] Integration: anonymous on `/api/admin` → 401.
- [ ] Integration: admin session does not roll — `expires_at` unchanged after use.
- [ ] Integration: 6th failed admin attempt in an hour → 429.
- [ ] Integration: failures appear in `audit_log`; the password never does.

## Done when

- [ ] No admin route is reachable without an admin session, proven by a test that enumerates the routes.
- [ ] Checkbox ticked in `README.md`.
