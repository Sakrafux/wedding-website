package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// HouseholdStore reads and writes households and their members.
type HouseholdStore struct {
	database *configuration.Database
	// generateCode is domain.GenerateCode in production. It is a field so that the
	// collision path below can be forced in a test: with 32^6 codes and sixty in
	// use it never runs otherwise, and a retry nobody has watched work is a retry
	// that may not.
	generateCode func() string
}

func NewHouseholdStore(database *configuration.Database) *HouseholdStore {
	return &HouseholdStore{database: database, generateCode: domain.GenerateCode}
}

// WithCodeGenerator returns a copy of the store that draws login codes from
// generate. For tests only — production wiring uses NewHouseholdStore.
func (store *HouseholdStore) WithCodeGenerator(generate func() string) *HouseholdStore {
	return &HouseholdStore{database: store.database, generateCode: generate}
}

// codeAssignmentAttempts bounds the retry on a colliding login code.
//
// A collision needs two of 32^6 codes to match, so five in a row is not bad luck —
// it is a broken generator, and looping forever would hide that behind a request
// that never answers.
const codeAssignmentAttempts = 5

// householdColumns is shared by every read so they cannot drift into returning
// differently populated structs.
//
// One projection for the whole application rather than one per epic: a partially
// populated domain.Household is a value that lies about the columns it left at their
// zero value, and the caller cannot tell which read produced it. rsvp_note_seen_at is
// the one column still missing, because nothing reads it yet — F6's note inbox is
// what adds it, here and nowhere else.
const householdColumns = `id, display_name, code, transport_seats_needed, transport_seats_offered,
	has_stroller, admin_note, rsvp_note, rsvp_submitted_at, rsvp_updated_at, last_login_at`

// FindByCode returns the household holding this login code, or ErrNotFound.
//
// The code must already be normalized: the column stores the canonical form, so a
// raw "abc-234" would simply not match and a guest would be told their correct
// code is unknown.
func (store *HouseholdStore) FindByCode(ctx context.Context, code string) (domain.Household, error) {
	return store.findHousehold(ctx, `SELECT `+householdColumns+` FROM household WHERE code = ?`, code)
}

// FindByID returns the household with this id, or ErrNotFound.
func (store *HouseholdStore) FindByID(ctx context.Context, id int64) (domain.Household, error) {
	return store.findHousehold(ctx, `SELECT `+householdColumns+` FROM household WHERE id = ?`, id)
}

func (store *HouseholdStore) findHousehold(ctx context.Context, query string, argument any) (domain.Household, error) {
	var row householdRow
	if err := store.database.Read.GetContext(ctx, &row, query, argument); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Household{}, ErrNotFound
		}
		return domain.Household{}, fmt.Errorf("reading household: %w", err)
	}
	return row.toDomain()
}

// List returns every household with the member count the admin list screen shows.
//
// One query with a LEFT JOIN rather than a count per row: sixty households is
// small, but a loop issuing sixty-one queries is a habit rather than a size.
//
// Ordered by display name, case-insensitively, because the admin scans this list
// looking for a name — insertion order is meaningless to that task. NOCASE is
// ASCII-only in SQLite, so "Ärzte" sorts after "Zimmer"; accepted, since the whole
// list fits on one screen and the frontend's search is what actually finds a name.
func (store *HouseholdStore) List(ctx context.Context) ([]domain.HouseholdOverview, error) {
	selectHouseholds := `
		SELECT ` + prefixedColumns("h", householdColumns) + `,
		       COUNT(g.id) AS member_count
		FROM household h
		LEFT JOIN guest g ON g.household_id = h.id AND g.deleted_at IS NULL
		GROUP BY h.id
		ORDER BY h.display_name COLLATE NOCASE`

	var rows []householdOverviewRow
	if err := store.database.Read.SelectContext(ctx, &rows, selectHouseholds); err != nil {
		return nil, fmt.Errorf("listing households: %w", err)
	}

	overviews := make([]domain.HouseholdOverview, 0, len(rows))
	for _, row := range rows {
		household, err := row.householdRow.toDomain()
		if err != nil {
			return nil, err
		}
		overviews = append(overviews, domain.HouseholdOverview{Household: household, MemberCount: row.MemberCount})
	}
	return overviews, nil
}

// Create inserts a household and assigns it a fresh login code.
//
// The code is assigned here rather than by the caller because uniqueness is the
// UNIQUE index and nothing else: the insert is attempted, and a rejected one means
// "generate another", not "fail". Asking "is this code taken?" first would be a
// race with a window, and redundant besides.
//
// A household without a code is a household nobody can log in as, and no screen
// would show you that — so there is no way to create one.
func (store *HouseholdStore) Create(ctx context.Context, household domain.Household) (domain.Household, error) {
	const insertHousehold = `
		INSERT INTO household (display_name, code, transport_seats_needed, transport_seats_offered,
		                       has_stroller, admin_note)
		VALUES (?, ?, ?, ?, ?, ?)`

	err := store.withGeneratedCode(func(code string) error {
		result, err := store.database.Write.ExecContext(ctx, insertHousehold,
			household.DisplayName, code, household.TransportSeatsNeeded, household.TransportSeatsOffered,
			household.HasStroller, household.AdminNote)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		household.ID, household.Code = id, code
		return nil
	})
	if err != nil {
		return domain.Household{}, fmt.Errorf("inserting household: %w", err)
	}
	return household, nil
}

// Update writes the columns an admin may edit. Everything else — the code, the RSVP
// answers, the timestamps — is not reachable from here by construction.
func (store *HouseholdStore) Update(ctx context.Context, household domain.Household) error {
	const updateHousehold = `
		UPDATE household
		SET display_name = ?, transport_seats_needed = ?, transport_seats_offered = ?,
		    has_stroller = ?, admin_note = ?
		WHERE id = ?`

	result, err := store.database.Write.ExecContext(ctx, updateHousehold,
		household.DisplayName, household.TransportSeatsNeeded, household.TransportSeatsOffered,
		household.HasStroller, household.AdminNote, household.ID)
	if err != nil {
		return fmt.Errorf("updating household: %w", err)
	}
	return requireOneRow(result, "updating household")
}

// Delete removes the household row.
//
// Hard, not soft, unlike a guest: only we ever delete a household, whereas a
// household removes its own plus-ones and a person who was once counted has to stay
// explicable. Guests cascade by foreign key and seat assignments cascade from them;
// audit_log outlives the row, which is the whole reason it is a separate table.
func (store *HouseholdStore) Delete(ctx context.Context, id int64) error {
	result, err := store.database.Write.ExecContext(ctx, `DELETE FROM household WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting household: %w", err)
	}
	return requireOneRow(result, "deleting household")
}

// AssignNewCode replaces the household's login code and returns the new one.
//
// Revoking the household's sessions is the caller's job and is not optional — see
// application.Households.ReissueCode. The only reason to reissue is that the old
// code must stop working, and a year-long session issued from it would outlive the
// code by months.
func (store *HouseholdStore) AssignNewCode(ctx context.Context, id int64) (string, error) {
	const updateCode = `UPDATE household SET code = ? WHERE id = ?`

	var assigned string
	err := store.withGeneratedCode(func(code string) error {
		result, err := store.database.Write.ExecContext(ctx, updateCode, code, id)
		if err != nil {
			return err
		}
		if err := requireOneRow(result, "assigning login code"); err != nil {
			return err
		}
		assigned = code
		return nil
	})
	if err != nil {
		return "", err
	}
	return assigned, nil
}

// withGeneratedCode runs write with fresh codes until one is accepted.
//
// The collision is detected specifically — a unique-constraint violation — and not
// as "any write error": retrying a disk-full error five times helps nobody, and a
// missing row must surface as ErrNotFound rather than as a collision.
func (store *HouseholdStore) withGeneratedCode(write func(code string) error) error {
	var err error
	for attempt := 1; attempt <= codeAssignmentAttempts; attempt++ {
		err = write(store.generateCode())
		if err == nil {
			return nil
		}
		if !isUniqueViolation(err) {
			return err
		}
	}
	return fmt.Errorf("%d colliding login codes in a row: %w", codeAssignmentAttempts, err)
}

// ListMembers returns the household's living members, in a stable order.
//
// Soft-deleted guests are excluded here rather than at every call site: a removed
// plus-one stays in the table so the audit trail still explains a headcount, but
// they are nobody's member any more.
//
// Ordered by id, which is insertion order: seeded members keep the order we typed
// them in, and members a household adds itself appear at the end where that
// household last saw them. Sorting by name would reshuffle the form under people
// as they fill it in.
func (store *HouseholdStore) ListMembers(ctx context.Context, householdID int64) ([]domain.Guest, error) {
	selectMembers := `
		SELECT ` + guestColumns + `
		FROM guest
		WHERE household_id = ? AND deleted_at IS NULL
		ORDER BY id`

	var rows []guestRow
	if err := store.database.Read.SelectContext(ctx, &rows, selectMembers, householdID); err != nil {
		return nil, fmt.Errorf("reading household members: %w", err)
	}

	members := make([]domain.Guest, 0, len(rows))
	for _, row := range rows {
		members = append(members, row.toDomain())
	}
	return members, nil
}

// TouchLastLogin records that this household redeemed its code.
//
// Unconditional, on every login rather than only the first: the admin nudge list
// asks "who has never appeared", and the delta between last login and the RSVP
// state is what tells us whether a household looked but did not answer.
func (store *HouseholdStore) TouchLastLogin(ctx context.Context, householdID int64, at time.Time) error {
	const touchLastLogin = `UPDATE household SET last_login_at = ? WHERE id = ?`

	if _, err := store.database.Write.ExecContext(ctx, touchLastLogin, formatTimestamp(at), householdID); err != nil {
		return fmt.Errorf("updating last login: %w", err)
	}
	return nil
}

type householdRow struct {
	ID                    int64          `db:"id"`
	DisplayName           string         `db:"display_name"`
	Code                  string         `db:"code"`
	TransportSeatsNeeded  int            `db:"transport_seats_needed"`
	TransportSeatsOffered int            `db:"transport_seats_offered"`
	HasStroller           bool           `db:"has_stroller"`
	AdminNote             string         `db:"admin_note"`
	RSVPNote              string         `db:"rsvp_note"`
	RSVPSubmittedAt       sql.NullString `db:"rsvp_submitted_at"`
	RSVPUpdatedAt         sql.NullString `db:"rsvp_updated_at"`
	LastLoginAt           sql.NullString `db:"last_login_at"`
}

// householdOverviewRow is a household plus its member count. Embedded rather than
// restated, so the two projections cannot drift apart.
type householdOverviewRow struct {
	householdRow
	MemberCount int `db:"member_count"`
}

func (row householdRow) toDomain() (domain.Household, error) {
	household := domain.Household{
		ID:                    row.ID,
		DisplayName:           row.DisplayName,
		Code:                  row.Code,
		TransportSeatsNeeded:  row.TransportSeatsNeeded,
		TransportSeatsOffered: row.TransportSeatsOffered,
		HasStroller:           row.HasStroller,
		AdminNote:             row.AdminNote,
		RSVPNote:              row.RSVPNote,
	}

	submittedAt, err := parseNullableTimestamp(row.RSVPSubmittedAt)
	if err != nil {
		return domain.Household{}, err
	}
	household.RSVPSubmittedAt = submittedAt

	updatedAt, err := parseNullableTimestamp(row.RSVPUpdatedAt)
	if err != nil {
		return domain.Household{}, err
	}
	household.RSVPUpdatedAt = updatedAt

	lastLoginAt, err := parseNullableTimestamp(row.LastLoginAt)
	if err != nil {
		return domain.Household{}, err
	}
	household.LastLoginAt = lastLoginAt

	return household, nil
}

// prefixedColumns qualifies a column list with a table alias, for the joins.
//
// Written out rather than selecting `h.*`: sqlx fails on a column with no
// destination field, so a star would break the next time a migration adds a column to
// household — which has nothing to do with this query and would fail inside it.
func prefixedColumns(alias, columns string) string {
	parts := strings.Split(columns, ",")
	for index, part := range parts {
		parts[index] = alias + "." + strings.TrimSpace(part)
	}
	return strings.Join(parts, ", ")
}
