package persistence

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// ExportStore reads the flat rows the CSV exports stream.
//
// Its own store rather than a method on HouseholdStore, because what it produces is
// not a domain object: guests.csv is a dump of two tables joined, and mapping it
// through domain structs would mean adding a field to Guest for every column the
// dump carries and nothing else reads.
type ExportStore struct {
	database *configuration.Database
}

func NewExportStore(database *configuration.Database) *ExportStore {
	return &ExportStore{database: database}
}

// GuestExportColumns are the headers of guests.csv, in order.
//
// The names are the schema's column names verbatim, English, prefixed `household_`
// where they come from the household. The prefix is what stops `created_at` from
// being ambiguous, and matching the schema exactly means a question about a value is
// answered by reading 03-data-model rather than by guessing what a friendly header
// meant.
//
// `household_id` appears once and serves both `guest.household_id` and
// `household.id`: they are the same number, and a second copy of it would be a
// column nobody can explain.
//
// The order is not the schema's. The three identifying columns come first, and
// **deleted_at is third** so it cannot be missed while scanning — this file includes
// removed people, and anything counted from it must filter them out. The remaining
// guest columns follow, then the household's.
//
// An integration test compares this list against the actual columns of both tables,
// so a migration that adds a field and forgets this file fails there rather than
// producing a dump that quietly omits it.
var GuestExportColumns = []string{
	"guest_id",
	"household_id",
	"deleted_at",
	"household_display_name",
	"household_code",
	"first_name",
	"last_name",
	"kind",
	"age",
	"origin",
	"attending",
	"meal_choice",
	"portion",
	"midnight_snack",
	"seating_need",
	"dietary_note",
	"created_at",
	"household_transport_seats_needed",
	"household_transport_seats_offered",
	"household_has_stroller",
	"household_admin_note",
	"household_rsvp_note",
	"household_rsvp_note_seen_at",
	"household_rsvp_submitted_at",
	"household_rsvp_updated_at",
	"household_last_login_at",
	"household_created_at",
}

// guestExportExpressions are the SQL expressions behind GuestExportColumns, in the
// same order. Kept as a parallel list so the header row and the SELECT cannot drift:
// they are zipped into `expression AS header` below, and a length mismatch is a
// panic at startup of the query rather than a silently shifted column.
var guestExportExpressions = []string{
	"g.id",
	"g.household_id",
	"g.deleted_at",
	"h.display_name",
	"h.code",
	"g.first_name",
	"g.last_name",
	"g.kind",
	"g.age",
	"g.origin",
	"g.attending",
	"g.meal_choice",
	"g.portion",
	"g.midnight_snack",
	"g.seating_need",
	"g.dietary_note",
	"g.created_at",
	"h.transport_seats_needed",
	"h.transport_seats_offered",
	"h.has_stroller",
	"h.admin_note",
	"h.rsvp_note",
	"h.rsvp_note_seen_at",
	"h.rsvp_submitted_at",
	"h.rsvp_updated_at",
	"h.last_login_at",
	"h.created_at",
}

// StreamGuests calls yield once per guest, with the values in GuestExportColumns
// order.
//
// Rows are handed over one at a time rather than returned as a slice, so the response
// is written as the database produces it. Sixty rows would fit in memory; F10-B03's
// photo ZIP will not, and this is the shape it copies.
//
// Soft-deleted guests are included. Excluding them would make the file disagree with
// the database it claims to dump, and a removed plus-one is exactly the row somebody
// eventually wants to see. It is therefore a dump and not a headcount — see the
// download label on the admin household list.
func (store *ExportStore) StreamGuests(ctx context.Context, yield func(values []any) error) error {
	if len(guestExportExpressions) != len(GuestExportColumns) {
		return fmt.Errorf("guest export has %d columns and %d expressions",
			len(GuestExportColumns), len(guestExportExpressions))
	}

	projection := make([]string, 0, len(GuestExportColumns))
	for index, expression := range guestExportExpressions {
		projection = append(projection, expression+" AS "+GuestExportColumns[index])
	}

	// Ordered by household so the file reads like the guest list it is, and by guest
	// id within it, which is the order the members were entered in.
	query := `SELECT ` + strings.Join(projection, ", ") + `
		FROM guest g
		JOIN household h ON h.id = g.household_id
		ORDER BY h.display_name COLLATE NOCASE, g.id`

	rows, err := store.database.Read.QueryxContext(ctx, query)
	if err != nil {
		return fmt.Errorf("reading guest export: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		values, err := rows.SliceScan()
		if err != nil {
			return fmt.Errorf("scanning guest export row: %w", err)
		}
		if err := yield(values); err != nil {
			return err
		}
	}
	return rows.Err()
}
