package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// RSVPStore writes a household's whole answer.
//
// Reading is HouseholdStore's — FindByID plus ListMembers already return the answer
// columns, and a second read path would be a second projection to keep in step. What
// this store owns is the one thing neither of those can do: writing the household row
// and every member row inside a single transaction.
type RSVPStore struct {
	database *configuration.Database
}

func NewRSVPStore(database *configuration.Database) *RSVPStore {
	return &RSVPStore{database: database}
}

// SaveAnswer writes the household's answer and its members' answers as one commit.
//
// A partial write is a household that told us four people are coming and one meal, so
// the transaction is not an optimisation. Every member UPDATE is scoped to the
// household id as well as to the guest id: the use case has already checked that the
// set matches, and this is the second lock on the one mistake that would let a
// household write into somebody else's row.
//
// Members not listed are left untouched — the caller passes the complete set, which is
// what F3-B03's member_set_mismatch rule guarantees.
func (store *RSVPStore) SaveAnswer(ctx context.Context, household domain.Household, members []domain.Guest) error {
	const updateHousehold = `
		UPDATE household
		SET transport_seats_needed = ?, transport_seats_offered = ?, has_stroller = ?,
		    rsvp_note = ?, rsvp_submitted_at = ?, rsvp_updated_at = ?
		WHERE id = ?`

	const updateMember = `
		UPDATE guest
		SET attending = ?, meal_choice = ?, portion = ?, midnight_snack = ?,
		    seating_need = ?, dietary_note = ?, age = ?
		WHERE id = ? AND household_id = ? AND deleted_at IS NULL`

	return store.inTransaction(ctx, func(transaction *sqlx.Tx) error {
		result, err := transaction.ExecContext(ctx, updateHousehold,
			household.TransportSeatsNeeded, household.TransportSeatsOffered, household.HasStroller,
			household.RSVPNote, nullableTimestamp(household.RSVPSubmittedAt), nullableTimestamp(household.RSVPUpdatedAt),
			household.ID)
		if err != nil {
			return fmt.Errorf("updating household answer: %w", err)
		}
		if err := requireOneRow(result, "updating household answer"); err != nil {
			return err
		}

		for _, member := range members {
			result, err := transaction.ExecContext(ctx, updateMember,
				nullableEnum(member.Attending), nullableEnum(member.MealChoice),
				string(member.Portion), member.MidnightSnack,
				string(member.SeatingNeed), member.DietaryNote, member.Age,
				member.ID, household.ID)
			if err != nil {
				return fmt.Errorf("updating guest answer: %w", err)
			}
			if err := requireOneRow(result, "updating guest answer"); err != nil {
				return err
			}
		}
		return nil
	})
}

// inTransaction runs work inside one write transaction, committing on success and
// rolling back on any failure.
//
// On the write pool, which is capped at a single connection: concurrent saves queue in
// Go rather than racing in SQLite. The DSN opens transactions as IMMEDIATE, so the
// write lock is taken up front and a busy timeout can actually resolve it — see
// configuration.writeDataSourceName.
//
// A rollback failure is joined onto the original error rather than replacing it: what
// went wrong is the interesting half, and losing it to "rollback failed" is how a
// broken save becomes unexplainable.
func (store *RSVPStore) inTransaction(ctx context.Context, work func(transaction *sqlx.Tx) error) error {
	transaction, err := store.database.Write.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	if err := work(transaction); err != nil {
		if rollbackErr := transaction.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w (rollback failed: %w)", err, rollbackErr)
		}
		return err
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// nullableEnum renders a nullable enum pointer for storage: SQL NULL for nil.
//
// Without it a *domain.Attending would be handed to the driver as a pointer to a
// named string type, which modernc.org/sqlite refuses rather than dereferences.
func nullableEnum[T ~string](value *T) any {
	if value == nil {
		return nil
	}
	return string(*value)
}

// nullableTimestamp renders a nullable timestamp for storage. Kept beside
// formatTimestamp's other callers so that the format is decided in exactly one place.
func nullableTimestamp(at *time.Time) any {
	if at == nil {
		return nil
	}
	return formatTimestamp(*at)
}
