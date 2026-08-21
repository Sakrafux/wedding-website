# 03 — Data Model

Status: draft · Last updated: 2026-08-21

SQLite. Conventions: `INTEGER` primary keys, `BOOLEAN` for flags (SQLite stores these as 0/1; the declared type is for readability and Go driver mapping), timestamps as `TEXT` in UTC RFC3339, money as `INTEGER` cents (never floats), soft deletes only where history matters. Foreign keys enforced (`PRAGMA foreign_keys = ON`), WAL mode.

**Enum values are English everywhere** — in the DB, in the API, in Go and TypeScript types. German exists only as display labels in the frontend, mapped from the enum. Enums are `CHECK`-constrained rather than lookup tables wherever the value set is small and fixed.

Timestamps are TEXT rather than epoch integers because at this scale performance is irrelevant and readability when inspecting the DB by hand with `sqlite3` is worth more. UTC + fixed format means lexicographic sort is chronological sort.

Informational page content (schedule, travel, dress code, FAQ, …) is **not** in the database — it is hardcoded in the React components. Only data that changes at runtime lives here.

## Entities

### `household`

The unit of authentication and of RSVP.

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `display_name` | TEXT | e.g. "Familie Müller", "Anna & Ben", "Oma Gertrud". Shown on the login confirmation screen. |
| `code` | TEXT UNIQUE | The printed login code, normalized (uppercase, no dashes). |
| `transport_seats_needed` | INTEGER | Default 0. Seats this household needs for the church → reception trip. Not derived from household size; some members drive themselves. |
| `transport_seats_offered` | INTEGER | Default 0. Spare seats this household can offer others. |
| `admin_note` | TEXT | Private note for us. Never sent to a guest session. |
| `rsvp_note` | TEXT | Free-text note from the household to us. The deliberate escape hatch: anything the structured fields do not cover goes here. No length limit worth enforcing beyond a sane cap. Must be surfaced prominently in the admin dashboard — an unread note is a missed request. |
| `rsvp_note_seen_at` | TEXT NULL | When an admin acknowledged the current note. Older than `rsvp_updated_at` → the note counts as unread again. |
| `rsvp_submitted_at` | TEXT NULL | First submission. Null = never answered → appears on the nudge list. |
| `rsvp_updated_at` | TEXT NULL | Last change. |
| `last_login_at` | TEXT NULL | Answers "did they even see it?". |
| `created_at` | TEXT | |

**Code storage:** plaintext, not hashed. Reason: we must be able to read a code back for a guest who lost the card, and login is a bare lookup by code with no second factor. Accepted risk — the DB sits on a personal server, the secret only unlocks non-sensitive guest content, and the threat model is trusted guests. Consequence: the SQLite file must be treated as a secret (see [06-privacy-security](06-privacy-security.md)).

### `guest`

A person. Belongs to exactly one household.

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `household_id` | INTEGER FK → household | |
| `first_name` | TEXT | |
| `last_name` | TEXT NULL | Often redundant within a household. |
| `kind` | TEXT | `adult` \| `child` |
| `age` | INTEGER NULL | Children only. **Age at the wedding date**, not at RSVP time — the UI asks it that way so the value does not drift over the months before the event. Feeds caterer pricing brackets and venue headcounts. |
| `origin` | TEXT | `seeded` (we created it) \| `guest_added` (household added it). Drives the admin delta view. |
| `attending` | TEXT NULL | `no` \| `church_only` \| `party_only` \| `both`. NULL = not answered. Attendance and scope in one field, so contradictory states (declined but scoped) cannot exist. No "maybe" by design. |
| `meal_choice` | TEXT NULL | `all` \| `vegetarian` \| `vegan`. `all` = eats everything. NULL when `portion = 'none'` or the guest is not at the party. |
| `portion` | TEXT | `none` \| `kids` \| `full`. Defaults to `full`. `none` covers infants and adults not eating (e.g. arriving after dinner). |
| `midnight_snack` | BOOLEAN | Wants the midnight snack. Independent of `portion` — someone skipping dinner may still want it. |
| `seating_need` | TEXT | `normal` \| `with_parent` (no own seat) \| `high_chair` \| `stroller` \| `wheelchair`. Defaults to `normal`; applies to adults as well as children. |
| `dietary_note` | TEXT | Allergies and intolerances, free text. |
| `created_at` | TEXT | |
| `deleted_at` | TEXT NULL | Soft delete. |

Invariants:

- A household may only delete `guest` rows where `origin = 'guest_added'`, and only before the RSVP deadline. Seeded guests are never deleted by a guest — they get `attending = 'no'`.
- `seating_need = 'with_parent'` means the guest consumes no seat and must not be assigned one.
- **`attending` scope gates the catering fields.** `meal_choice`, `portion`, `midnight_snack` and `seat_assignment` are only meaningful for `party_only` and `both`. For `church_only` and `no` they are ignored, and every derived count keys off the scope — not off "is attending". Getting this wrong means paying for meals nobody eats.
- Transport is only relevant to guests with `attending = 'both'` — `church_only` guests never travel to the reception and `party_only` guests arrive there directly.

**Attendance scope is per guest, not per household**, because the real exceptions are within households — a grandmother attending the ceremony but skipping the party, or children going only to the church. The RSVP form avoids the extra clicking with a household-level selector ("Wir kommen zu: Kirche / Feier / beidem") that sets all members at once, with a per-member override beneath it. That is a UI affordance, not a second column.

### `seating_table`

A table in the room. Named `seating_table` because `table` is reserved.

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `number` | INTEGER NULL | Explicit ordering. Avoids "Tisch 10" sorting before "Tisch 2". Null for named tables like the head table. |
| `label` | TEXT | e.g. "Tisch 4", "Brauttisch". |
| `capacity` | INTEGER | Drives over-assignment warnings, not a hard block. |
| `svg_element_id` | TEXT NULL | The `id` attribute of the corresponding shape in the hand-drawn floor-plan SVG. This is the entire link between data and visual. |

### `seat_assignment`

| Column | Type | Notes |
|---|---|---|
| `guest_id` | INTEGER PK, FK → guest | One seat per guest. |
| `seating_table_id` | INTEGER FK → seating_table | |
| `assigned_at` | TEXT | |

Invariants:

- Only guests with `attending IN ('party_only', 'both')`, `deleted_at IS NULL`, and `seating_need != 'with_parent'` may hold an assignment. Church-only guests are never seated.
- If a guest's RSVP flips to `no` or `church_only`, or a guest-added member is deleted, the assignment is **not** silently removed — it is reported to admins as a stale assignment to resolve. Prevents seats quietly vanishing from a finished plan.

### `budget_item`

Admin-only. Never reachable from a household session.

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `category` | TEXT | e.g. Catering, Location, Musik, Deko, Kleidung. |
| `title` | TEXT | |
| `vendor` | TEXT NULL | |
| `planned_cents` | INTEGER | |
| `actual_cents` | INTEGER NULL | Null until known. |
| `paid_cents` | INTEGER | Running total actually paid. |
| `external_cents` | INTEGER NULL | Portion covered by someone else. Excluded from *our* cost in the rollup. |
| `paid_by` | TEXT NULL | Who covers it and any conditions, e.g. "Eltern der Braut übernehmen die Blumen". |
| `per_head_cents` | INTEGER NULL | If set, planned cost = `per_head_cents × live attending headcount`. Recomputed on read, never frozen. |
| `due_date` | TEXT NULL | |
| `status` | TEXT | `planned` \| `booked` \| `partially_paid` \| `paid` \| `cancelled` |
| `note` | TEXT | |
| `created_at`, `updated_at` | TEXT | |

Per-head items read the live headcount, so a late plus-one moves the budget automatically. The RSVP deadline is therefore the point where cost stops moving.

### `photo`

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `uploader_household_id` | INTEGER FK NULL | NULL = uploaded by an admin (curated gallery). |
| `stored_filename` | TEXT UNIQUE | Content-addressed name on disk; never the user-supplied name. |
| `original_filename` | TEXT | For display and the ZIP export. |
| `mime_type` | TEXT | Determined by content sniffing, not extension. |
| `size_bytes` | INTEGER | |
| `width`, `height` | INTEGER NULL | |
| `taken_at` | TEXT NULL | Read from EXIF at ingest so ordering by capture time works. |
| `uploaded_at` | TEXT | |
| `hidden_at` | TEXT NULL | Admin hide. Row kept, file kept, not served. |
| `caption` | TEXT NULL | |

Files live on a mounted volume, not in SQLite. Thumbnails are derived and regenerable.

**Originals are stored byte-for-byte, EXIF intact.** Stripping metadata was considered and rejected: the gallery is login-gated among trusted family, naive stripping breaks orientation tags so photos render sideways, and it degrades the long-term archive. If a photo is ever published outside this site, strip GPS at that point instead.

### `session`

| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | Hash of the cookie token, not the token itself. |
| `subject_type` | TEXT | `household` \| `admin` |
| `subject_id` | INTEGER NULL | Household id, or NULL for the admin (there is no admin table — credentials come from environment variables). |
| `created_at`, `expires_at`, `last_seen_at` | TEXT | |
| `user_agent`, `ip` | TEXT NULL | For the audit trail and for "log out other devices". |

Household sessions live 365 days with rolling refresh. Admin sessions are short (hours) — different risk profile.

### `audit_log`

Append-only. The record of who changed what, which is what lets us settle "but I said we were coming".

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | |
| `at` | TEXT | |
| `actor_type` | TEXT | `household` \| `admin` \| `system` |
| `actor_id` | INTEGER NULL | |
| `entity` | TEXT | Table name. |
| `entity_id` | INTEGER | |
| `action` | TEXT | `create` \| `update` \| `delete` \| `login` \| `login_failed` |
| `before`, `after` | TEXT NULL | JSON snapshots of the changed fields only. |

### `app_setting`

Key/value for things we want to change without a redeploy.

| Key | Purpose |
|---|---|
| `rsvp_deadline` | After this, guest RSVP editing is read-only. |
| `default_addition_limit` | Soft cap on guest-added members per household. |
| `seating_published` | Gate the guest-facing seating view. |
| `uploads_open` | Gate guest photo uploads (post-wedding). |
| `gallery_visible` | Gate the gallery entirely. |

Everything else that is genuinely static (page text, wedding date, venue) is hardcoded in the frontend.

## Derived views the admin UI needs

- Headcount by scope: church attendees, party attendees, and the overlap. These are three different numbers and three different vendors care about them.
- Seats actually needed at the party (party attendees, excluding `with_parent`).
- Counts per meal choice and per portion — party attendees only.
- Midnight snack count — party attendees only.
- Children by caterer age bracket, computed from `age` (brackets configurable, since every caterer draws them differently).
- Transport balance: total `transport_seats_needed` vs. total `transport_seats_offered` across households, and the resulting gap — the number that tells us whether to hire a shuttle. Matching individual riders to drivers is done by us offline, not by the app.
- Special seating needs list: high chairs, strollers, wheelchairs — the things that must be physically arranged.
- Consolidated dietary list, grouped so the caterer gets one readable page.
- Delta list: all `origin = 'guest_added'` guests, newest first.
- Nudge list: households with `rsvp_submitted_at IS NULL`.
- Households with a non-empty `rsvp_note`, flagged unread until an admin marks it seen.
- Stale seat assignments: assignments whose guest is no longer attending or is deleted.
- Budget rollup: planned vs. actual vs. paid per category and overall, with per-head items resolved against the live headcount, and total cost shown separately from our own cost after `external_cents`.

## Rejected fields

| Field | Why not |
|---|---|
| `arrival_day` | Single-day wedding — no arrival planning needed. |
| `needs_accommodation` | Not our concern. Hotel suggestions are static page content. |
| `household.addition_limit` | Over-engineering; a global default covers it, and we can add a person ourselves in admin. |
| `meal_option` table | Only three fixed choices — collapsed to a `CHECK`-constrained column on `guest`. |
| `guest.age_band` | Replaced by a plain `age` integer (defined as age at the wedding date) plus `seating_need` and `portion`. Bands are derived at read time, so a caterer's bracket boundaries can change without a migration. |
| `guest.kids_menu` BOOLEAN | Two states were not enough — infants eat nothing. Replaced by the three-way `portion` enum. |
| `sort_order` | Only justified where display order differs from natural order; `seating_table.number` handles the one real case. |
| `admin_user` table | One admin, low stakes. Credentials live in environment variables; see [04-architecture](04-architecture.md). |
| `household.phone` | We already know these people. Transport counts are for estimating shuttle capacity, not for the app to match riders to drivers. |
| Separate `attending` + `scope` columns | Folded into one `attending` enum so "declined but attending the party" is unrepresentable rather than merely invalid. |

## Open questions

- Does a household need a "we are not sure yet" state distinct from "not answered", purely for our own tracking? (Guests never see it.)
- Do we track invitation send-out date per household, to time reminders?
- Photo retention and deletion policy after the site goes offline.
