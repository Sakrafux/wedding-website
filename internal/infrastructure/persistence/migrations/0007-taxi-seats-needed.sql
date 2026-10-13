-- 0007 — a seat count for the taxi home, beside the shuttle counts.
--
-- The reception is in Wilhelmsburg and most guests sleep in or near St. Pölten,
-- so the evening ends with a taxi ride for whoever did not bring a car. Whether
-- we reserve one (or several) depends on the headcount, and that headcount is
-- exactly the kind of thing the RSVP exists to gather.
--
-- One column, not a needed/offered pair: the shuttle counts feed a subtraction
-- (the capacity gap, F3-B07) because guests drive each other both ways — but a
-- taxi is something *we* would order for the trip away from the venue, and
-- "offering" taxi seats to strangers is not a thing a household can do. So
-- there is no conflict to refuse and no gap to compute, just a count.
--
-- Reception → city only, which is why it is gated on the party scopes rather
-- than on "attends both" like the shuttle pair: a guest who comes only to the
-- party still needs the ride home. Zeroed by NormalizeHouseholdAnswer whenever
-- nobody in the household attends the party — the same rule the shuttle counts
-- follow for their own trip.

ALTER TABLE household ADD COLUMN taxi_seats_needed INTEGER NOT NULL DEFAULT 0 CHECK (taxi_seats_needed >= 0);
