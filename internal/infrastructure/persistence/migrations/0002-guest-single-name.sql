-- 0002 — one name per guest, instead of first_name and last_name.
--
-- What we actually care about is the full name: it is what goes on a place card, in
-- the caterer's list and in "so haben wir euch notiert". The split bought nothing and
-- cost flexibility for the people filling the form in — a double first name, a guest
-- with no surname we know, "Oma Erika", or a plus-one somebody knows by one name.
-- F4 lets households add members themselves, and one field is one field to get wrong.
--
-- Accepted cost: nothing sorts by surname any more. guests.csv has no last_name
-- column, and place cards, the caterer sheet and the seating lists order by the full
-- name, which means by first name. At ~80 people grouped by household and by table,
-- surname order was not doing any work.
--
-- Done now rather than later because there is no real guest data yet (E-OPS-01 has
-- not run): after send-out this would be a migration against live answers.

ALTER TABLE guest ADD COLUMN name TEXT NOT NULL DEFAULT '';

-- The DEFAULT exists only so ADD COLUMN can fill the existing rows; a name is
-- required, and the application never writes an empty one.
UPDATE guest SET name = trim(first_name || ' ' || last_name);

ALTER TABLE guest DROP COLUMN first_name;
ALTER TABLE guest DROP COLUMN last_name;
