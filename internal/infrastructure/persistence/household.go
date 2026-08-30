package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// HouseholdStore reads and writes households and their members.
//
// F5-B01 and F5-B02 widen this into the full admin CRUD. What is here is what
// login and the bootstrap response need: find a household, list its living
// members, record that somebody logged in.
type HouseholdStore struct {
	database *configuration.Database
}

func NewHouseholdStore(database *configuration.Database) *HouseholdStore {
	return &HouseholdStore{database: database}
}

// householdColumns is shared by the two finders so they cannot drift into
// returning differently populated structs.
const householdColumns = `id, display_name, code, last_login_at`

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
	const selectMembers = `
		SELECT id, household_id, first_name, last_name, kind, origin
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
	ID          int64          `db:"id"`
	DisplayName string         `db:"display_name"`
	Code        string         `db:"code"`
	LastLoginAt sql.NullString `db:"last_login_at"`
}

func (row householdRow) toDomain() (domain.Household, error) {
	household := domain.Household{
		ID:          row.ID,
		DisplayName: row.DisplayName,
		Code:        row.Code,
	}

	if row.LastLoginAt.Valid {
		lastLoginAt, err := parseTimestamp(row.LastLoginAt.String)
		if err != nil {
			return domain.Household{}, err
		}
		household.LastLoginAt = &lastLoginAt
	}
	return household, nil
}

type guestRow struct {
	ID          int64  `db:"id"`
	HouseholdID int64  `db:"household_id"`
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	Kind        string `db:"kind"`
	Origin      string `db:"origin"`
}

func (row guestRow) toDomain() domain.Guest {
	return domain.Guest{
		ID:          row.ID,
		HouseholdID: row.HouseholdID,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		Kind:        domain.GuestKind(row.Kind),
		Origin:      domain.GuestOrigin(row.Origin),
	}
}
