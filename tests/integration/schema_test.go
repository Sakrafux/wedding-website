package integration

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The schema is SQL with no Go code in front of it yet, so these tests are the only
// thing that proves the constraints exist. They are deliberately about the constraints
// rather than about queries: a missing CHECK does not fail anything else until a wrong
// enum value has already been stored.

func TestEnumColumnsRejectUnknownValues(t *testing.T) {
	database := newTestDatabase(t)
	householdID := insertHousehold(t, database.Write, "ABC234")

	// One case per CHECK-constrained column. English values only; a German label
	// reaching the database is exactly the mistake these constraints catch.
	cases := map[string]struct {
		statement string
		arguments []any
	}{
		"guest.kind":         {`INSERT INTO guest (household_id, first_name, last_name, kind, origin) VALUES (?, 'A', 'B', 'grown_up', 'seeded')`, []any{householdID}},
		"guest.origin":       {`INSERT INTO guest (household_id, first_name, last_name, kind, origin) VALUES (?, 'A', 'B', 'adult', 'imported')`, []any{householdID}},
		"guest.attending":    {`INSERT INTO guest (household_id, first_name, last_name, kind, origin, attending) VALUES (?, 'A', 'B', 'adult', 'seeded', 'maybe')`, []any{householdID}},
		"guest.meal_choice":  {`INSERT INTO guest (household_id, first_name, last_name, kind, origin, meal_choice) VALUES (?, 'A', 'B', 'adult', 'seeded', 'vegetarisch')`, []any{householdID}},
		"guest.portion":      {`INSERT INTO guest (household_id, first_name, last_name, kind, origin, portion) VALUES (?, 'A', 'B', 'adult', 'seeded', 'half')`, []any{householdID}},
		"guest.seating_need": {`INSERT INTO guest (household_id, first_name, last_name, kind, origin, seating_need) VALUES (?, 'A', 'B', 'adult', 'seeded', 'booster')`, []any{householdID}},
		// 'stroller' used to be a seating_need. It moved to household.has_stroller, so
		// the old value must now be refused rather than quietly stored.
		"guest.seating_need/stroller": {`INSERT INTO guest (household_id, first_name, last_name, kind, origin, seating_need) VALUES (?, 'A', 'B', 'child', 'seeded', 'stroller')`, []any{householdID}},
		"guest.age":                   {`INSERT INTO guest (household_id, first_name, last_name, kind, origin, age) VALUES (?, 'A', 'B', 'adult', 'seeded', 34)`, []any{householdID}},
		"seating_unit.venue":          {`INSERT INTO seating_unit (venue, label) VALUES ('garden', 'Tisch 1')`, nil},
		"budget_item.status":          {`INSERT INTO budget_item (category, title, status) VALUES ('Catering', 'Menü', 'gebucht')`, nil},
		"session.subject_type":        {`INSERT INTO session (id, subject_type, expires_at) VALUES ('hash', 'guest', '2027-07-17T00:00:00Z')`, nil},
		"audit_log.actor_type":        {`INSERT INTO audit_log (actor_type, entity, entity_id, action) VALUES ('robot', 'guest', 1, 'update')`, nil},
		"audit_log.action":            {`INSERT INTO audit_log (actor_type, entity, entity_id, action) VALUES ('admin', 'guest', 1, 'undelete')`, nil},
	}

	for column, testCase := range cases {
		t.Run(column, func(t *testing.T) {
			_, err := database.Write.Exec(testCase.statement, testCase.arguments...)

			require.Error(t, err, "%s accepted a value outside its enum", column)
			assert.Contains(t, err.Error(), "CHECK", "expected a CHECK violation, got: %v", err)
		})
	}
}

// TestDuplicateHouseholdCodeIsRejected is what lets code generation retry on conflict
// instead of checking first and racing.
func TestDuplicateHouseholdCodeIsRejected(t *testing.T) {
	database := newTestDatabase(t)
	insertHousehold(t, database.Write, "ABC234")

	_, err := database.Write.Exec(`INSERT INTO household (display_name, code) VALUES ('Familie Zwei', 'ABC234')`)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE")
}

func TestDeletingHouseholdCascadesToGuests(t *testing.T) {
	database := newTestDatabase(t)
	householdID := insertHousehold(t, database.Write, "ABC234")
	insertGuest(t, database.Write, householdID, "Anna")

	_, err := database.Write.Exec(`DELETE FROM household WHERE id = ?`, householdID)
	require.NoError(t, err)

	var guests int
	require.NoError(t, database.Read.Get(&guests, `SELECT count(*) FROM guest WHERE household_id = ?`, householdID))
	assert.Zero(t, guests)
}

// TestDeletingSeatingUnitWithAssignmentIsRefused covers the RESTRICT chain
// seating_unit → seat → seat_assignment: dropping a table out from under a finished
// plan must fail, not silently unseat people.
func TestDeletingSeatingUnitWithAssignmentIsRefused(t *testing.T) {
	database := newTestDatabase(t)
	householdID := insertHousehold(t, database.Write, "ABC234")
	guestID := insertGuest(t, database.Write, householdID, "Anna")
	unitID := insertSeatingUnit(t, database.Write, "party", "Tisch 1")
	seatID := insertSeat(t, database.Write, unitID, "party", "Platz 1")
	assignSeat(t, database.Write, seatID, "party", guestID)

	_, unitErr := database.Write.Exec(`DELETE FROM seating_unit WHERE id = ?`, unitID)
	_, seatErr := database.Write.Exec(`DELETE FROM seat WHERE id = ?`, seatID)

	require.Error(t, unitErr)
	assert.Contains(t, unitErr.Error(), "FOREIGN KEY")
	require.Error(t, seatErr)
	assert.Contains(t, seatErr.Error(), "FOREIGN KEY")
}

// TestGuestHoldsOneSeatPerVenue proves the venue-carrying composite FK does its job:
// the same guest may be seated once in the church and once at the party, never twice in
// the same venue.
func TestGuestHoldsOneSeatPerVenue(t *testing.T) {
	database := newTestDatabase(t)
	householdID := insertHousehold(t, database.Write, "ABC234")
	guestID := insertGuest(t, database.Write, householdID, "Anna")

	churchUnitID := insertSeatingUnit(t, database.Write, "church", "Bank 1")
	churchSeatID := insertSeat(t, database.Write, churchUnitID, "church", "Platz 1")
	partyUnitID := insertSeatingUnit(t, database.Write, "party", "Tisch 1")
	firstPartySeatID := insertSeat(t, database.Write, partyUnitID, "party", "Platz 1")
	secondPartySeatID := insertSeat(t, database.Write, partyUnitID, "party", "Platz 2")

	assignSeat(t, database.Write, churchSeatID, "church", guestID)
	assignSeat(t, database.Write, firstPartySeatID, "party", guestID)

	_, err := database.Write.Exec(
		`INSERT INTO seat_assignment (seat_id, venue, guest_id) VALUES (?, 'party', ?)`,
		secondPartySeatID, guestID,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE")
}

// TestSeatAssignmentVenueMustMatchTheSeat is the other half of that design: the venue
// column on seat_assignment is a copy, and the composite FK is what stops the copy from
// lying about which venue the seat belongs to.
func TestSeatAssignmentVenueMustMatchTheSeat(t *testing.T) {
	database := newTestDatabase(t)
	householdID := insertHousehold(t, database.Write, "ABC234")
	guestID := insertGuest(t, database.Write, householdID, "Anna")
	unitID := insertSeatingUnit(t, database.Write, "party", "Tisch 1")
	seatID := insertSeat(t, database.Write, unitID, "party", "Platz 1")

	_, err := database.Write.Exec(
		`INSERT INTO seat_assignment (seat_id, venue, guest_id) VALUES (?, 'church', ?)`,
		seatID, guestID,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "FOREIGN KEY")
}

// TestSeatVenueMustMatchItsUnit closes the same loop one level up.
func TestSeatVenueMustMatchItsUnit(t *testing.T) {
	database := newTestDatabase(t)
	unitID := insertSeatingUnit(t, database.Write, "party", "Tisch 1")

	_, err := database.Write.Exec(
		`INSERT INTO seat (seating_unit_id, venue, label, svg_element_id) VALUES (?, 'church', 'Platz 1', 'mismatch')`,
		unitID,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "FOREIGN KEY")
}

// TestGuestDefaultsMatchTheDataModel guards the values a household never sees a control
// for: an unanswered guest still has to produce a sane catering row.
func TestGuestDefaultsMatchTheDataModel(t *testing.T) {
	database := newTestDatabase(t)
	householdID := insertHousehold(t, database.Write, "ABC234")
	guestID := insertGuest(t, database.Write, householdID, "Anna")

	var guest struct {
		Portion       string  `db:"portion"`
		SeatingNeed   string  `db:"seating_need"`
		MidnightSnack bool    `db:"midnight_snack"`
		Attending     *string `db:"attending"`
		CreatedAt     string  `db:"created_at"`
	}
	require.NoError(t, database.Read.Get(&guest,
		`SELECT portion, seating_need, midnight_snack, attending, created_at FROM guest WHERE id = ?`, guestID))

	assert.Equal(t, "full", guest.Portion)
	assert.Equal(t, "normal", guest.SeatingNeed)
	assert.False(t, guest.MidnightSnack)
	assert.Nil(t, guest.Attending, "an unanswered guest must be NULL, not 'no'")
	// The default is a safety net for hand-written INSERTs, so its format is part of
	// the contract: RFC3339, UTC, sortable.
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, guest.CreatedAt)
}

func TestHouseholdLogisticsDefaultsAreEmpty(t *testing.T) {
	database := newTestDatabase(t)
	householdID := insertHousehold(t, database.Write, "ABC234")

	var household struct {
		Needed      int  `db:"transport_seats_needed"`
		Offered     int  `db:"transport_seats_offered"`
		HasStroller bool `db:"has_stroller"`
	}
	require.NoError(t, database.Read.Get(&household,
		`SELECT transport_seats_needed, transport_seats_offered, has_stroller FROM household WHERE id = ?`, householdID))

	assert.Zero(t, household.Needed)
	assert.Zero(t, household.Offered)
	assert.False(t, household.HasStroller)
}

func TestAppSettingsAreSeeded(t *testing.T) {
	database := newTestDatabase(t)

	settings := map[string]string{}
	rows, err := database.Read.Queryx(`SELECT key, value FROM app_setting`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var key, value string
		require.NoError(t, rows.Scan(&key, &value))
		settings[key] = value
	}
	require.NoError(t, rows.Err())

	assert.Equal(t, map[string]string{
		"rsvp_deadline":          "2027-05-17T21:59:59Z",
		"default_addition_limit": "2",
		"seating_published":      "false",
		"uploads_open":           "false",
		"gallery_visible":        "false",
	}, settings)
}

func insertHousehold(t *testing.T, pool *sqlx.DB, code string) int64 {
	t.Helper()

	result, err := pool.Exec(`INSERT INTO household (display_name, code) VALUES (?, ?)`, "Familie "+code, code)
	require.NoError(t, err)
	return lastInsertID(t, result)
}

func insertGuest(t *testing.T, pool *sqlx.DB, householdID int64, firstName string) int64 {
	t.Helper()

	// Only the columns without a default, so the defaults stay observable.
	result, err := pool.Exec(
		`INSERT INTO guest (household_id, first_name, last_name, kind, origin) VALUES (?, ?, 'Muster', 'adult', 'seeded')`,
		householdID, firstName,
	)
	require.NoError(t, err)
	return lastInsertID(t, result)
}

func insertSeatingUnit(t *testing.T, pool *sqlx.DB, venue, label string) int64 {
	t.Helper()

	result, err := pool.Exec(
		`INSERT INTO seating_unit (venue, label, svg_element_id) VALUES (?, ?, ?)`,
		venue, label, fmt.Sprintf("%s-%s", venue, label),
	)
	require.NoError(t, err)
	return lastInsertID(t, result)
}

func insertSeat(t *testing.T, pool *sqlx.DB, seatingUnitID int64, venue, label string) int64 {
	t.Helper()

	result, err := pool.Exec(
		`INSERT INTO seat (seating_unit_id, venue, label, svg_element_id) VALUES (?, ?, ?, ?)`,
		seatingUnitID, venue, label, fmt.Sprintf("%s-%d-%s", venue, seatingUnitID, label),
	)
	require.NoError(t, err)
	return lastInsertID(t, result)
}

func assignSeat(t *testing.T, pool *sqlx.DB, seatID int64, venue string, guestID int64) {
	t.Helper()

	_, err := pool.Exec(
		`INSERT INTO seat_assignment (seat_id, venue, guest_id) VALUES (?, ?, ?)`,
		seatID, venue, guestID,
	)
	require.NoError(t, err)
}

func lastInsertID(t *testing.T, result interface{ LastInsertId() (int64, error) }) int64 {
	t.Helper()

	id, err := result.LastInsertId()
	require.NoError(t, err)
	return id
}
