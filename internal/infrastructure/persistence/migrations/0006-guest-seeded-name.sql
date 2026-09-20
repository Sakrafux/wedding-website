-- 0006 — keep the name we seeded a guest under, now that households may rename them.
--
-- F4-B04 lets a household edit the names on its own RSVP card. That is what the
-- guests actually want — we type "Oma Erika" and "Fam. Müller +1" off an address
-- book, and the person themselves knows how they are called — but it breaks the one
-- thing `guest.name` was also doing for us: being the name on the invitation we
-- posted. A household that renames "Erika Huber" to "Omi" leaves the admin list, the
-- caterer sheet and our own memory of who that card went to with no anchor.
--
-- So the seeded name is kept beside the editable one. `name` stays the single name
-- every output prints; `seeded_name` is admin-only history, shown as "früher: …" on
-- the household detail page when the two differ, and never sent to a guest.
--
-- Empty for `guest_added` rows, and that is the honest value rather than a copy of
-- `name`: a plus-one has no name we gave them, and a reader comparing the two columns
-- must not be told a household renamed somebody it invented.

ALTER TABLE guest ADD COLUMN seeded_name TEXT NOT NULL DEFAULT '';

UPDATE guest SET seeded_name = name WHERE origin = 'seeded';
