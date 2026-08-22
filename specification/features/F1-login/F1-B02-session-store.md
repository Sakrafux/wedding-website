# `F1-B02` — Session store

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** `E0-05`

## Story

As a guest, I want to stay logged in for a year on my own phone, so that I never have to find the card again — while an admin can still revoke any session immediately.

## Scope

**In:**

- `internal/infrastructure/security`: token generation and hashing.
- `internal/infrastructure/persistence`: session create, lookup, touch, delete, purge.
- Rolling refresh, throttled.
- Startup and daily purge of expired rows.

**Out:**

- Middleware → `F1-B03`.
- Cookie handling → `F1-B04`.

## Instructions

1. Token: 32 bytes from `crypto/rand`, base64url without padding. This is the value that goes in the cookie and is never stored.
2. Store `session.id = SHA-256(token)` in hex. Look up by hashing the presented token. A leaked database therefore yields no usable sessions — defence in depth, per [06-privacy-security](../../06-privacy-security.md).
3. Lifetimes: `household` = 365 days, `admin` = 8 hours. Take the lifetime from the subject type at creation; do not let callers pass an arbitrary duration.
4. Rolling refresh for household sessions only, and **at most once per 24 hours**. A write on every request turns a read-heavy app into a write-heavy one against a single-writer database, for no benefit.
5. Admin sessions do **not** roll. Short by design.
6. Record `user_agent` and `ip` at creation for the audit trail.
7. Purge expired rows at startup and once a day. A simple `time.Ticker` goroutine is enough; no scheduler dependency.
8. Lookup returns "not found" for an expired session and deletes it opportunistically. Never return an expired session and let a caller decide — that check will eventually be forgotten in one call site.

## Test plan

- [ ] Unit: two generated tokens differ; length and encoding are as specified.
- [ ] Integration: create → look up by raw token → returns the session.
- [ ] Integration: the raw token does not appear anywhere in the `session` table.
- [ ] Integration: an expired household session is not returned and is removed.
- [ ] Integration: touching twice within 24h performs one update, not two — assert on `expires_at` being unchanged the second time.
- [ ] Integration: an admin session does not roll.
- [ ] Integration: delete → subsequent lookup fails. Revocation is immediate.
- [ ] Integration: purge removes only expired rows.

## Done when

- [ ] Sessions can be created, validated, refreshed and revoked through the store API.
- [ ] Checkbox ticked in `README.md`.
