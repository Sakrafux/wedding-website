# `F1-B06` — Audit logging for authentication

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** `F1-B04`

## Story

As an admin, I want every login and failed attempt recorded, so that after the fact I can tell whether a household actually saw the site and whether anything odd happened.

## Scope

**In:**

- `login` and `login_failed` rows in `audit_log`.
- A reusable audit-write helper for the later feature epics.

**Out:**

- RSVP and admin-change auditing → `F3-B05` and the admin stories.
- Any UI → F6.

## Instructions

1. Helper signature roughly `audit.Write(ctx, actorType, actorID, entity, entityID, action, before, after)`. Every later epic uses it; getting the shape right once is the point of doing it here.
2. On success: `actor_type = 'household'` (or `admin`), `actor_id`, `entity = 'session'`, `action = 'login'`.
3. On failure: `actor_type = 'system'`, `actor_id = NULL`, `action = 'login_failed'`, with the IP and user agent in the `after` JSON.
4. **Never record the attempted code.** A log of near-misses is a partial key list, and a log of typos will eventually contain another household's real code. State this in a comment at the write site, because "log the input for debugging" is exactly the change someone makes later in good faith.
5. Audit writes must not fail the request. Log the error and continue — a broken audit table must not stop a guest from logging in.
6. Writes go through the single-writer pool like everything else.

## Test plan

- [ ] Integration: successful login writes one `login` row with the right household.
- [ ] Integration: failed login writes one `login_failed` row with `actor_type = 'system'`.
- [ ] Integration: **no audit row anywhere contains the submitted code string.** Assert directly against the table contents.
- [ ] Integration: a forced audit-write failure does not fail the login.
- [ ] Integration: `before`/`after` are valid JSON or NULL.

## Done when

- [ ] The audit table shows a readable history of logins after a manual session.
- [ ] Checkbox ticked in `README.md`.
