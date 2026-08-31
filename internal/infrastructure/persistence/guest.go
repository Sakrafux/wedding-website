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
	const insertGuest = `
		INSERT INTO guest (household_id, name, kind, age, origin, seating_need, dietary_note)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

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
