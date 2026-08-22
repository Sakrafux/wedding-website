# `E0-05` — Migration `0001`: full schema

**Epic:** E0 — Project setup · **Layer:** backend · **Depends on:** `E0-04`

## Story

As a developer, I want the entire schema from the data model created in one migration so that later feature stories add behaviour rather than columns, and so the field-set freeze has something concrete to freeze.

## Scope

**In:**

- `0001-initial-schema.sql` creating every table in [03-data-model](../../03-data-model.md): `household`, `guest`, `seating_table`, `seat_assignment`, `budget_item`, `photo`, `session`, `audit_log`, `app_setting`.
- `CHECK` constraints for every enum, indexes, and the `app_setting` seed rows.

**Out:**

- Any Go struct or store. This story is SQL only.

## Instructions

1. Transcribe the tables exactly as specified, including nullability. Where this story and the data model disagree, the data model is right — fix the code, or fix the document deliberately.
2. `CHECK` constraints on every enum column, with the **English** values: `guest.attending`, `guest.kind`, `guest.origin`, `guest.meal_choice`, `guest.portion`, `guest.seating_need`, `budget_item.status`, `session.subject_type`, `audit_log.actor_type`, `audit_log.action`.
3. `household.code` is `TEXT UNIQUE NOT NULL`. Uniqueness is what makes a collision on generation impossible rather than merely unlikely.
4. Foreign keys with explicit behaviour: `guest.household_id` → `ON DELETE CASCADE`; `seat_assignment.guest_id` → `ON DELETE CASCADE`; `seat_assignment.seating_table_id` → `ON DELETE RESTRICT`, because deleting a table that still seats people should be refused, not silently accepted.
5. Defaults per the data model: `portion` = `'full'`, `seating_need` = `'normal'`, `midnight_snack` = 0, transport seats = 0.
6. Indexes: `guest(household_id)`, `guest(deleted_at)`, `session(expires_at)`, `audit_log(entity, entity_id)`, `audit_log(at)`, `photo(uploaded_at)`, `seat_assignment(seating_table_id)`.
7. Seed `app_setting`: `rsvp_deadline`, `default_addition_limit` (pending the confirmed value — use 2), `seating_published` = false, `uploads_open` = false, `gallery_visible` = false.
8. Money columns are `INTEGER`. Timestamps are `TEXT`. No `REAL` anywhere in this file — a float in a money column is a bug waiting for a rounding error.

## Test plan

- [ ] Integration: migration applies to an empty database without error.
- [ ] Integration: inserting an invalid enum value is rejected for each enum column.
- [ ] Integration: inserting a duplicate `household.code` is rejected.
- [ ] Integration: deleting a household cascades to its guests.
- [ ] Integration: deleting a `seating_table` that holds an assignment is refused.
- [ ] Integration: a guest row inserted with only the required columns gets `portion = 'full'` and `seating_need = 'normal'`.
- [ ] Integration: the five `app_setting` keys exist after migration.

## Done when

- [ ] `sqlite3 wedding.db .schema` matches the data model, read side by side.
- [ ] Checkbox ticked in `README.md`.
