package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// GuestStore reads and writes individual guests. Listing a household's members
// belongs to HouseholdStore, which owns the "who is in this household" question.
type GuestStore struct {
	database *configuration.Database
}

func NewGuestStore(database *configuration.Database) *GuestStore {
	return &GuestStore{database: database}
}

// guestColumns is shared by every read of a guest, RSVP answers included.
//
// One projection rather than one per epic, for the same reason as householdColumns:
// a Guest with an empty Attending because of *which query loaded it* is a value no
// caller can reason about, and "is this person coming" is asked from F6, F7 and F8
// alike.
const guestColumns = `id, household_id, name, kind, age, origin, seating_need, dietary_note,
	attending, meal_choice, portion, midnight_snack`

// insertGuest is shared by the admin path and the household's own plus-one path, which
// differ only in the origin they pass and in the rule checked before the insert.
const insertGuest = `
	INSERT INTO guest (household_id, name, kind, age, origin, seating_need, dietary_note)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

// selectLivingMembers is the household's members, soft-deleted rows excluded, ordered
// by id — insertion order, so a household's form does not reshuffle under them. Shared
// with HouseholdStore.ListMembers so that "who is in this household" is one query, and
// the rule in CreateIfHouseholdAllows counts the same people the form shows.
const selectLivingMembers = `
	SELECT ` + guestColumns + `
	FROM guest
	WHERE household_id = ? AND deleted_at IS NULL
	ORDER BY id`

// FindByID returns the guest with this id, or ErrNotFound. A soft-deleted guest is
// not found: they are still in the table so the audit trail explains a headcount,
// but nothing may edit them any more.
func (store *GuestStore) FindByID(ctx context.Context, id int64) (domain.Guest, error) {
	selectGuest := `SELECT ` + guestColumns + ` FROM guest WHERE id = ? AND deleted_at IS NULL`

	var row guestRow
	if err := store.database.Read.GetContext(ctx, &row, selectGuest, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Guest{}, ErrNotFound
		}
		return domain.Guest{}, fmt.Errorf("reading guest: %w", err)
	}
	return row.toDomain(), nil
}

// Create inserts a guest into a household.
//
// The origin comes from the guest passed in rather than being fixed here: the admin
// path seeds ('seeded') and F4's household path adds ('guest_added'), and which one
// it is decides what the admin delta view shows.
func (store *GuestStore) Create(ctx context.Context, guest domain.Guest) (domain.Guest, error) {
	result, err := store.database.Write.ExecContext(ctx, insertGuest,
		guest.HouseholdID, guest.Name, string(guest.Kind), guest.Age,
		string(guest.Origin), string(guest.SeatingNeed), guest.DietaryNote)
	if err != nil {
		return domain.Guest{}, fmt.Errorf("inserting guest: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Guest{}, fmt.Errorf("inserting guest: %w", err)
	}
	guest.ID = id
	return guest, nil
}

// CreateIfHouseholdAllows inserts a guest only if allows accepts the household's
// living members, checked inside the same write transaction as the insert.
//
// The rule is the caller's — domain.CanHouseholdAddPlusOne (F4-B01) — and is passed in
// rather than written here, so this store still knows no business rules. What it owns
// is the *moment* the rule is asked: reading the members before the transaction and
// inserting after it would let two phones submitting at once turn a household of one
// into a household of three. Unlikely, and cheap to exclude.
func (store *GuestStore) CreateIfHouseholdAllows(
	ctx context.Context,
	guest domain.Guest,
	allows func(members []domain.Guest) error,
) (domain.Guest, error) {
	err := inTransaction(ctx, store.database, func(transaction *sqlx.Tx) error {
		var rows []guestRow
		if err := transaction.SelectContext(ctx, &rows, selectLivingMembers, guest.HouseholdID); err != nil {
			return fmt.Errorf("reading household members: %w", err)
		}
		if err := allows(guestsFrom(rows)); err != nil {
			return err
		}

		result, err := transaction.ExecContext(ctx, insertGuest,
			guest.HouseholdID, guest.Name, string(guest.Kind), guest.Age,
			string(guest.Origin), string(guest.SeatingNeed), guest.DietaryNote)
		if err != nil {
			return fmt.Errorf("inserting guest: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("inserting guest: %w", err)
		}
		guest.ID = id
		return nil
	})
	if err != nil {
		return domain.Guest{}, err
	}
	return guest, nil
}

// Update writes the columns an admin or a household may edit. The household id is
// not among them: a guest does not move between households, and the one case that
// looks like it — a plus-one counted in the wrong household — is a removal and an
// addition, which is also what the audit trail should show.
func (store *GuestStore) Update(ctx context.Context, guest domain.Guest) error {
	const updateGuest = `
		UPDATE guest
		SET name = ?, kind = ?, age = ?, seating_need = ?, dietary_note = ?
		WHERE id = ? AND deleted_at IS NULL`

	result, err := store.database.Write.ExecContext(ctx, updateGuest,
		guest.Name, string(guest.Kind), guest.Age,
		string(guest.SeatingNeed), guest.DietaryNote, guest.ID)
	if err != nil {
		return fmt.Errorf("updating guest: %w", err)
	}
	return requireOneRow(result, "updating guest")
}

// SoftDelete marks the guest as removed, keeping the row.
//
// Soft on purpose: the guest was counted, may hold a seat assignment, and appears in
// the audit trail. Erasing the row would leave those dangling and the history
// unexplainable. The seat assignment is deliberately left in place — F7-B01 reports
// it as stale for a human to resolve, because seats must not quietly vanish from a
// finished plan.
func (store *GuestStore) SoftDelete(ctx context.Context, id int64, at time.Time) error {
	const softDeleteGuest = `UPDATE guest SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`

	result, err := store.database.Write.ExecContext(ctx, softDeleteGuest, formatTimestamp(at), id)
	if err != nil {
		return fmt.Errorf("deleting guest: %w", err)
	}
	return requireOneRow(result, "deleting guest")
}

// guestsFrom maps a row slice to domain guests, so every caller that lists members
// produces the same values from the same projection.
func guestsFrom(rows []guestRow) []domain.Guest {
	guests := make([]domain.Guest, 0, len(rows))
	for _, row := range rows {
		guests = append(guests, row.toDomain())
	}
	return guests
}

type guestRow struct {
	ID          int64         `db:"id"`
	HouseholdID int64         `db:"household_id"`
	Name        string        `db:"name"`
	Kind        string        `db:"kind"`
	Age         sql.NullInt64 `db:"age"`
	Origin      string        `db:"origin"`
	SeatingNeed string        `db:"seating_need"`
	DietaryNote string        `db:"dietary_note"`
	// Attending and MealChoice are nullable: NULL attending means the household has
	// not answered for this person, which is a state, not a missing value.
	Attending     sql.NullString `db:"attending"`
	MealChoice    sql.NullString `db:"meal_choice"`
	Portion       string         `db:"portion"`
	MidnightSnack bool           `db:"midnight_snack"`
}

func (row guestRow) toDomain() domain.Guest {
	guest := domain.Guest{
		ID:            row.ID,
		HouseholdID:   row.HouseholdID,
		Name:          row.Name,
		Kind:          domain.GuestKind(row.Kind),
		Origin:        domain.GuestOrigin(row.Origin),
		SeatingNeed:   domain.SeatingNeed(row.SeatingNeed),
		DietaryNote:   row.DietaryNote,
		Portion:       domain.Portion(row.Portion),
		MidnightSnack: row.MidnightSnack,
	}

	if row.Age.Valid {
		age := int(row.Age.Int64)
		guest.Age = &age
	}
	if row.Attending.Valid {
		attending := domain.Attending(row.Attending.String)
		guest.Attending = &attending
	}
	if row.MealChoice.Valid {
		mealChoice := domain.MealChoice(row.MealChoice.String)
		guest.MealChoice = &mealChoice
	}
	return guest
}
