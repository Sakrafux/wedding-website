# `F3-B05` — Audit logging for every RSVP mutation

**Epic:** F3 — RSVP · **Layer:** backend · **Depends on:** `F3-B03`

## Story

As an admin, I want every RSVP change recorded with who made it and what changed, so that "but I said we were coming" can be answered from the data instead of from memory.

## Scope

**In:**

- One audit row per changed entity per save: the household row when its own fields changed, one row per member whose answer changed.
- `actor_type` reflecting who actually saved — `household` for the guest route, `admin` for `F3-B06`.
- `before` / `after` carrying only the changed fields, as `F1-B06` established.

**Out:**

- Showing the audit trail in the UI. There is no screen for it and does not need to be: `sqlite3` is the reader, and a log nobody has built a page for is still the record that settles the argument.
- Member additions and removals → `F4`, which audits its own.

## Instructions

1. Reuse `domain.NewAdminChangeEntry` / the household equivalent and the diffing helpers in `domain/changes.go`. A second diffing implementation for RSVP fields is a second place for a field to be forgotten.
2. **One row per entity, not one per save.** The household and each member are separate entities with separate ids, and a single row covering five people could not be read back against any of them.
3. A save that changed one member's meal writes **one** row. This is the reason `F3-B03` compares before writing: an audit log where most rows are "somebody pressed save" is a log nobody reads.
4. `actor_type = 'admin'` when we take the answer down the phone (`F3-B06`), never `household`. Recording our own typing as the household's answer would mislead at the exact moment the log is consulted — see `TODO.md`, decided 2026-08-31.
5. `actor_id` is the household id for a guest save and NULL for an admin save; there is no admin table, and the `entity_id` already names the household the change was about.
6. An audit write that fails must not fail the save. Log it through the use case's `logger`, exactly as `Auth` and `Households` already do — the answer matters more than the record of it, and silent breakage is what the log line is for.
7. Never put `code` in a payload. It cannot appear in an RSVP body at all, which is why this is a review note rather than a filter, but the rule holds across every audit call site.
8. `rsvp_note` **is** recorded in the payload. It is the household's own words to us, the audit log is admin-only, and a note that was edited to remove a request is exactly the case the log exists for.

## Test plan

- [ ] Integration: a save changing two members writes exactly two `guest` rows and no `household` row.
- [ ] Integration: a save changing only `rsvp_note` writes exactly one `household` row.
- [ ] Integration: a no-op save writes nothing.
- [ ] Integration: `before`/`after` contain only the changed fields, with the old and new values.
- [ ] Integration: the admin route records `actor_type = 'admin'` with `actor_id` NULL; the guest route records `household` with the household id.
- [ ] Integration: with the audit store made to fail, the save still succeeds and the failure is logged.

## Done when

- [ ] Every RSVP change in the database can be traced to a who, a when and a what, and no unchanged save appears in the log.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
