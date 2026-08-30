# `F5-B03` — Code assignment and regeneration

**Epic:** F5 — Admin: households & guests · **Layer:** backend · **Depends on:** `F1-B01`, `F5-B01`

## Story

As an admin, I want every household to get a unique code automatically, and to be able to replace one, so that a lost or leaked card can be recovered from without touching the database by hand.

## Scope

**In:**

- Collision-safe assignment of `domain.GenerateCode()` output on household creation.
- `POST /api/admin/households/{id}/code` to regenerate.
- Revoking the household's sessions when its code changes.

**Out:**

- Generating the code string → `F1-B01`, done.
- The print export → `F5-B04`.

## Instructions

1. Uniqueness comes from the `UNIQUE` index on `household.code`, not from checking first. Insert, and on a unique-constraint violation generate another and retry. Asking "is this code taken?" and then inserting is a race with a window, and the window is not the interesting part — the check is simply redundant when the database already enforces it.
2. Cap the retries at a small number (5) and fail loudly if they are all exhausted. With 32⁶ codes and sixty in use, five collisions in a row is not bad luck; it is a broken generator, and looping forever would hide that behind a hung request.
3. Detect the collision specifically — a unique-constraint violation on `household.code` — and not "any insert error". Retrying a disk-full error five times helps nobody.
4. Regeneration **revokes that household's sessions**: `DELETE FROM session WHERE subject_type = 'household' AND subject_id = ?`. The only reason to regenerate is that the old code should stop working, and a year-long session issued from it would outlive the code by months. Before send-out this deletes nothing, which is the harmless case.
5. Regeneration is audited as `update` on the household. The payload records **that** the code changed and never either value — see `F1-B06`.
6. The response returns the new code in both stored and formatted form, because the admin is about to read it to somebody or paste it into a document.
7. Regenerating after the cards are printed invalidates a printed card. The API does not prevent it; `F5-F02` puts the warning in front of the human, which is where a warning belongs.

## Contract

```http
POST /api/admin/households/{id}/code
```

Request: no body.

Response `200`:

```json
{ "code": "DEF567", "formatted_code": "DEF-567", "revoked_sessions": 1 }
```

`revoked_sessions` is there so the frontend can say "der alte Code funktioniert jetzt nicht mehr, ein Gerät wurde abgemeldet" rather than leaving the admin guessing what they just did.

Errors: `not_found` → 404 · `unauthenticated` → 401

## Test plan

- [ ] Integration: a created household has a code satisfying `domain.ValidateCode`.
- [ ] Integration: many households created in a row all get distinct codes.
- [ ] Unit or integration: a forced collision on the first attempt retries and succeeds — seed a household with a known code and stub the generator to return it once. This is the path that never runs in practice and therefore never gets exercised by accident.
- [ ] Integration: exhausting the retries returns a 500 and logs, rather than hanging.
- [ ] Integration: regenerating changes the code, and the old code no longer logs in while the new one does.
- [ ] Integration: regenerating deletes that household's sessions and **only** that household's — seed two logged-in households and assert the other survives.
- [ ] Integration: no audit row contains either the old or the new code.

## Done when

- [ ] A household can be issued a fresh code, and the old one stops working immediately and everywhere.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
