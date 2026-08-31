-- 0003 — drop the default_addition_limit setting.
--
-- What a household may add stopped being a number on 2026-08-31 (F4-B01): one adult,
-- and only if we seeded the household as a single person. That rule is structural —
-- after the addition the household has two members and the same check refuses the
-- next one — so there is nothing left for a number to configure.
--
-- Removed rather than left in place: a setting nothing reads is a setting somebody
-- eventually changes and expects to matter, and the person changing it would be
-- reaching for exactly the looser rule we decided against.

DELETE FROM app_setting WHERE key = 'default_addition_limit';
