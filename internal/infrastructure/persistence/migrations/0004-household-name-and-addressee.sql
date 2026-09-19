-- 0004 — split household.display_name into `name` and `addressee`.
--
-- One field was doing two jobs. `display_name` was both the label the admin list is
-- read and searched by and the text every guest-facing greeting prints — and those
-- want different strings. We address most households by first name ("Luki & Paddi"),
-- which is precisely the form that makes an admin list of sixty households
-- ambiguous: several households share a first name, and the list needs the surname
-- to be scannable at all.
--
--   * `name`      — Haushalt. Internal only: admin list, search, ordering, guests.csv.
--   * `addressee` — Anschrift. Guest-facing: every greeting, and codes.csv, because
--                   that file is what gets printed onto the invitation cards.
--
-- The existing values become `name`: that is how they were actually used (admin list
-- labels, "Familie Müller"), and the addressee is the new information that has to be
-- entered per household. Seeded from `name` so both columns are NOT NULL from the
-- start and no read site needs a fallback — an addressee nobody edited is merely the
-- household name, which is the behaviour before this migration.

ALTER TABLE household RENAME COLUMN display_name TO name;

-- The DEFAULT exists only so ADD COLUMN can fill the existing rows; an addressee is
-- required, and the application never writes an empty one.
ALTER TABLE household ADD COLUMN addressee TEXT NOT NULL DEFAULT '';

UPDATE household SET addressee = name;
