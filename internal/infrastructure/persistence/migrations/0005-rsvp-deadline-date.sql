-- 0005 — fix the RSVP deadline at 2027-05-01.
--
-- 0001 seeded a placeholder two months before the wedding (2027-05-17) while the real
-- date was still open. It is now decided: guests answer by **1 May 2027**, which is
-- what gets printed on the invitation card ("bitte bis 1. Mai 2027"). Two and a half
-- months before the wedding rather than two — the caterer and the seating plan both
-- want the headcount earlier than the original guess allowed, and a date at the start
-- of a month is easier to remember than one in the middle of it.
--
-- Stored as the last second of that day in Berlin/Vienna time (CEST, UTC+2), so an
-- answer saved late on 1 May still counts.
--
-- An UPDATE rather than a re-seed: the row exists, and app_setting is editable in a
-- pinch, so a deployment where somebody already moved the date by hand must not have
-- that overwritten silently — hence the guard on the old placeholder value.

UPDATE app_setting
   SET value      = '2027-05-01T21:59:59Z',
       updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
 WHERE key   = 'rsvp_deadline'
   AND value = '2027-05-17T21:59:59Z';
