-- 0001 — initial schema.
--
-- The whole data model in one migration, transcribed from specification/03-data-model.md.
-- Later feature stories add behaviour, not columns: the field set is frozen here so the
-- API and DTO shapes have something concrete to be built against.
--
-- Conventions enforced throughout this file:
--   * Enum columns are TEXT with a CHECK constraint and **English** values. German
--     exists only as display labels in the frontend.
--   * Money is INTEGER cents. There is no REAL in this file — a float in a money
--     column is a rounding error waiting to happen.
--   * Timestamps are TEXT, UTC, RFC3339 (`2027-07-17T14:00:00Z`). Fixed format plus UTC
--     means lexicographic sort is chronological sort.
--   * Flags are declared BOOLEAN. SQLite stores 0/1; the declared type is for
--     readability and for the Go driver's bool mapping.
--   * Timestamp columns that are always set on insert carry a `now` default. The
--     application always writes them explicitly; the default exists so a hand-written
--     INSERT in `sqlite3` cannot introduce a row with a NULL or a wrongly formatted
--     timestamp.

-- household — the unit of authentication and of RSVP.
CREATE TABLE household (
    id                      INTEGER PRIMARY KEY,
    display_name            TEXT    NOT NULL,
    -- The printed login code, normalized (uppercase, no dashes). Stored in plaintext
    -- because a guest who loses their card must be told their code again; see
    -- 03-data-model.md for the accepted risk. UNIQUE is what makes a generation
    -- collision impossible rather than merely unlikely — code generation retries on
    -- a failed insert rather than asking first and racing between the question and
    -- the answer.
    code                    TEXT    NOT NULL UNIQUE,
    -- Church → reception only. Not derived from household size: some members drive
    -- themselves. Feeds the shuttle capacity gap, nothing else.
    transport_seats_needed  INTEGER NOT NULL DEFAULT 0 CHECK (transport_seats_needed >= 0),
    transport_seats_offered INTEGER NOT NULL DEFAULT 0 CHECK (transport_seats_offered >= 0),
    -- Brings a pram. A flag, not a count and not a per-guest value: the pram belongs to
    -- the household, needs floor space rather than a seat, and nobody brings two.
    has_stroller            BOOLEAN NOT NULL DEFAULT 0,
    -- Private note for us. Never leaves the server in a guest response.
    admin_note              TEXT    NOT NULL DEFAULT '',
    -- Free-text note from the household to us: the escape hatch for everything the
    -- structured fields do not cover.
    rsvp_note               TEXT    NOT NULL DEFAULT '',
    -- Older than rsvp_updated_at → the note counts as unread again.
    rsvp_note_seen_at       TEXT,
    -- NULL = never answered → the household appears on the nudge list.
    rsvp_submitted_at       TEXT,
    rsvp_updated_at         TEXT,
    last_login_at           TEXT,
    created_at              TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- guest — a person, belonging to exactly one household.
CREATE TABLE guest (
    id            INTEGER PRIMARY KEY,
    household_id  INTEGER NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    first_name    TEXT    NOT NULL,
    -- Required, not inherited from the household: plenty of households are couples with
    -- different surnames, and the caterer's and seating lists need the real name.
    last_name     TEXT    NOT NULL,
    kind          TEXT    NOT NULL CHECK (kind IN ('adult', 'child')),
    -- Children only, and **age at the wedding date** — the UI asks it that way so the
    -- value does not drift over the months before the event. Caterer age brackets are
    -- derived from this at read time, never stored, so a caterer changing their
    -- boundaries is not a migration.
    age           INTEGER CHECK (age IS NULL OR kind = 'child'),
    -- Drives the admin delta view of what households added themselves.
    origin        TEXT    NOT NULL CHECK (origin IN ('seeded', 'guest_added')),
    -- Attendance and scope in one column, so "declined but coming to the party" is
    -- unrepresentable rather than merely invalid. NULL = not answered. No "maybe".
    attending     TEXT    CHECK (attending IN ('no', 'church_only', 'party_only', 'both')),
    -- Catering fields below are gated by scope, not by "is attending": they mean
    -- something only for party_only and both. Every derived count keys off the scope,
    -- otherwise we pay for meals nobody eats.
    meal_choice   TEXT    CHECK (meal_choice IN ('all', 'vegetarian', 'vegan')),
    portion       TEXT    NOT NULL DEFAULT 'full' CHECK (portion IN ('none', 'kids', 'full')),
    -- Independent of portion: someone skipping dinner may still want the snack.
    midnight_snack BOOLEAN NOT NULL DEFAULT 0,
    -- Physical arrangements at the seat, adults included. 'with_parent' means the guest
    -- consumes no seat and must not hold a seat_assignment.
    --
    -- No 'stroller' value: a pram is parked, not sat on, and it belongs to the
    -- household rather than to one child — see household.has_stroller.
    seating_need  TEXT    NOT NULL DEFAULT 'normal'
                          CHECK (seating_need IN ('normal', 'with_parent', 'high_chair', 'wheelchair')),
    -- Allergies and intolerances, free text.
    dietary_note  TEXT    NOT NULL DEFAULT '',
    created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    -- Soft delete: a household removing a plus-one must not erase the audit trail of
    -- them having been counted.
    deleted_at    TEXT
);

CREATE INDEX guest_household_id_index ON guest (household_id);
-- Every guest-facing read filters deleted_at IS NULL.
CREATE INDEX guest_deleted_at_index ON guest (deleted_at);

-- seating_unit — one group of seats: a table at the reception, a pew at the church.
-- Both venues are seated, so the table is named for the role rather than the furniture
-- (and `table` is a reserved word anyway).
--
-- No `capacity` column: seats are enumerated in `seat`, so the capacity is COUNT(seat)
-- and over-assignment is unrepresentable rather than merely warned about.
CREATE TABLE seating_unit (
    id             INTEGER PRIMARY KEY,
    venue          TEXT    NOT NULL CHECK (venue IN ('church', 'party')),
    -- Explicit ordering, so "Tisch 10" does not sort before "Tisch 2". NULL for named
    -- units like the head table.
    number         INTEGER,
    label          TEXT    NOT NULL,
    -- The `id` attribute of the matching shape in the hand-drawn floor-plan SVG. This
    -- column is the entire link between data and visual; the app colours and labels
    -- existing shapes and never positions them.
    svg_element_id TEXT    UNIQUE,
    -- Redundant given the primary key, but a composite FK needs a matching unique index
    -- on its target: it is what lets `seat` inherit the venue instead of copying it.
    UNIQUE (id, venue)
);

-- seat — one physical place. Transcribed by hand from the seat shapes in the SVG, which
-- stays the single source of truth for the layout.
CREATE TABLE seat (
    id              INTEGER PRIMARY KEY,
    seating_unit_id INTEGER NOT NULL,
    -- Copied from the unit, and kept honest by the composite foreign key below rather
    -- than by application code. Carrying the venue down this chain is what allows
    -- "one seat per guest per venue" to be a plain UNIQUE constraint on
    -- seat_assignment, with no cross-table trigger.
    venue           TEXT    NOT NULL,
    -- Shown on the place card and in the guest's "you sit here", e.g. "Platz 4".
    label           TEXT    NOT NULL,
    -- The seat's own shape in the SVG. Required: a seat the plan cannot point at is a
    -- seat nobody can be shown.
    svg_element_id  TEXT    NOT NULL UNIQUE,
    -- Deleting a unit takes its seats with it — but only if none of them is assigned,
    -- because seat_assignment restricts the delete of a seat.
    FOREIGN KEY (seating_unit_id, venue) REFERENCES seating_unit (id, venue) ON DELETE CASCADE,
    UNIQUE (id, venue)
);

CREATE INDEX seat_seating_unit_id_index ON seat (seating_unit_id);

-- seat_assignment — which guest sits on which seat. Church and party are assigned
-- independently: a guest attending both gets two rows, one per venue.
CREATE TABLE seat_assignment (
    -- One guest per seat, straight from the primary key.
    seat_id     INTEGER PRIMARY KEY,
    venue       TEXT    NOT NULL,
    guest_id    INTEGER NOT NULL REFERENCES guest (id) ON DELETE CASCADE,
    assigned_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    -- RESTRICT, not CASCADE: deleting a seat that still holds someone should be refused
    -- loudly, not quietly unseat a finished plan.
    FOREIGN KEY (seat_id, venue) REFERENCES seat (id, venue) ON DELETE RESTRICT,
    -- One seat per guest per venue. Without this a guest could be seated at two party
    -- tables at once and every derived count would disagree with the room.
    UNIQUE (guest_id, venue)
);

-- budget_item — admin-only. Never reachable from a household session; enforced in the
-- application layer, since SQLite has no row-level security.
CREATE TABLE budget_item (
    id              INTEGER PRIMARY KEY,
    category        TEXT    NOT NULL,
    title           TEXT    NOT NULL,
    vendor          TEXT,
    planned_cents   INTEGER NOT NULL DEFAULT 0,
    -- NULL until the real number is known.
    actual_cents    INTEGER,
    paid_cents      INTEGER NOT NULL DEFAULT 0,
    -- Portion covered by someone else. Excluded from *our* cost in the rollup, while
    -- still counting towards the total.
    external_cents  INTEGER,
    -- Who covers it and under what condition, e.g. "Eltern der Braut übernehmen die Blumen".
    paid_by         TEXT,
    -- If set, planned cost = per_head_cents × live attending headcount, recomputed on
    -- read and never frozen. A late plus-one therefore moves the budget by itself.
    per_head_cents  INTEGER,
    due_date        TEXT,
    status          TEXT    NOT NULL
                            CHECK (status IN ('planned', 'booked', 'partially_paid', 'paid', 'cancelled')),
    note            TEXT    NOT NULL DEFAULT '',
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- photo — metadata only. The files live on the PHOTO_DIR volume, originals kept
-- byte-for-byte with EXIF intact; thumbnails are derived and regenerable.
CREATE TABLE photo (
    id                    INTEGER PRIMARY KEY,
    -- NULL = uploaded by an admin (the curated gallery).
    uploader_household_id INTEGER REFERENCES household (id) ON DELETE SET NULL,
    -- Content-addressed name on disk, never the user-supplied one: a guest-controlled
    -- filename must not decide where we write.
    stored_filename       TEXT    NOT NULL UNIQUE,
    -- Kept for display and for the ZIP export.
    original_filename     TEXT    NOT NULL,
    -- Determined by sniffing the content, not by trusting the extension.
    mime_type             TEXT    NOT NULL,
    size_bytes            INTEGER NOT NULL CHECK (size_bytes >= 0),
    width                 INTEGER,
    height                INTEGER,
    -- Read from EXIF at ingest, so the gallery can order by capture time rather than
    -- by the order people got round to uploading.
    taken_at              TEXT,
    uploaded_at           TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    -- Admin hide: row kept, file kept, not served.
    hidden_at             TEXT,
    caption               TEXT
);

CREATE INDEX photo_uploaded_at_index ON photo (uploaded_at);

-- session — server-side sessions. The cookie carries a token; this table stores only
-- its hash, so a stolen database dump does not hand over live sessions.
CREATE TABLE session (
    -- Hash of the cookie token, not the token itself.
    id           TEXT    PRIMARY KEY,
    subject_type TEXT    NOT NULL CHECK (subject_type IN ('household', 'admin')),
    -- Household id, or NULL for the admin: there is no admin table, credentials come
    -- from environment variables. No FK, because NULL-plus-FK would still allow a
    -- household id under subject_type = 'admin'; the pairing is enforced in code.
    subject_id   INTEGER,
    created_at   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    -- Household sessions live 365 days with rolling refresh; admin sessions hours.
    -- Different risk profiles, same table.
    expires_at   TEXT    NOT NULL,
    last_seen_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    -- For the audit trail only, read via the DB. There is no session list in the UI and
    -- no "log out other devices" — one household shares one code, so a device list would
    -- mean nothing to them.
    user_agent   TEXT,
    ip           TEXT
);

-- Serves both validation ("is this session still alive") and the purge sweep.
CREATE INDEX session_expires_at_index ON session (expires_at);

-- audit_log — append-only. The record that lets us settle "but I said we were coming".
CREATE TABLE audit_log (
    id         INTEGER PRIMARY KEY,
    at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    actor_type TEXT    NOT NULL CHECK (actor_type IN ('household', 'admin', 'system')),
    actor_id   INTEGER,
    -- Table name. No FK on entity_id: the row it points at may be gone, and the log
    -- outliving the record is the whole point of keeping it.
    entity     TEXT    NOT NULL,
    entity_id  INTEGER NOT NULL,
    action     TEXT    NOT NULL CHECK (action IN ('create', 'update', 'delete', 'login', 'login_failed')),
    -- JSON snapshots of the changed fields only, not whole rows.
    before     TEXT,
    after      TEXT
);

CREATE INDEX audit_log_entity_index ON audit_log (entity, entity_id);
CREATE INDEX audit_log_at_index ON audit_log (at);

-- app_setting — key/value for the few things we want to change without a redeploy.
-- Everything genuinely static (page text, wedding date, venue) is hardcoded in the
-- frontend instead.
CREATE TABLE app_setting (
    key        TEXT PRIMARY KEY,
    -- Untyped TEXT, parsed by the reader. Booleans are the strings 'true'/'false' and
    -- timestamps are RFC3339, both chosen so `SELECT * FROM app_setting` is readable
    -- by hand — the point of this table is being editable in a pinch.
    value      TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

INSERT INTO app_setting (key, value) VALUES
    -- End of 2027-05-17 Berlin time: exactly two months before the wedding, and the
    -- same day of the month so it reads as intended on a printed card. After this,
    -- guest RSVP editing goes read-only.
    ('rsvp_deadline', '2027-05-17T21:59:59Z'),
    -- Soft cap on guest-added members per household, not a hard wall. Provisional
    -- value pending the confirmed number (TODO.md).
    ('default_addition_limit', '2'),
    -- All three gates start closed: the seating plan and the gallery only open when we
    -- say so, and uploads are a post-wedding affair.
    ('seating_published', 'false'),
    ('uploads_open', 'false'),
    ('gallery_visible', 'false');
