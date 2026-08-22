# `F1-B03` — Session middleware and auth context

**Epic:** F1 — Household login · **Layer:** backend · **Depends on:** `F1-B02`

## Story

As a developer, I want the authenticated subject resolved once, in middleware, so that no handler ever reads a cookie and no handler can forget to check who is calling.

## Scope

**In:**

- Middleware that resolves the cookie to a subject and puts it in the request context.
- `RequireHousehold` and `RequireAdmin` gates.
- Typed context accessors.

**Out:**

- Login and logout endpoints → `F1-B04`, `F1-B07`.
- Rate limiting → `F1-B05`.

## Instructions

1. Resolve middleware runs on all `/api` routes: no cookie or invalid cookie → anonymous, **not** an error. `GET /api/health` and the login endpoints must work unauthenticated.
2. Context value is a small struct: subject type (`household`/`admin`), subject id, session id. Use an unexported context key type — a string key is a collision waiting to happen.
3. Accessor `HouseholdFromContext(ctx) (int64, bool)`. **Handlers never read the raw context value**, so the shape can change without touching every handler.
4. `RequireHousehold` → 401 with the German "bitte melde dich an" message on failure. `RequireAdmin` → 401, and it must reject a *household* session as firmly as it rejects anonymous.
5. Mount `RequireAdmin` on the whole `/api/admin` subtree, once, at the router level. Per-handler checks are how one endpoint eventually ships unguarded.
6. Touch the session (rolling refresh) in the resolve middleware, after a successful lookup, subject to the 24h throttle from `F1-B02`.
7. On a *valid* session for a household that no longer exists — deleted while logged in — treat as anonymous and delete the session. Otherwise every downstream query has to defend against a dangling id.

## Test plan

- [ ] Integration: no cookie → anonymous; a public route still works.
- [ ] Integration: a garbage cookie value → anonymous, no 500.
- [ ] Integration: valid household cookie → `RequireHousehold` route passes.
- [ ] Integration: household cookie → `RequireAdmin` route returns 401.
- [ ] Integration: admin cookie → `RequireAdmin` passes; `RequireHousehold` route behaves as specified.
- [ ] Integration: expired cookie → 401 on a guarded route.
- [ ] Integration: session for a deleted household → treated as anonymous, row gone.
- [ ] Integration: a request refreshes `expires_at` at most once in 24h.

## Done when

- [ ] Adding a new admin route requires no authorisation code in the handler.
- [ ] Checkbox ticked in `README.md`.
