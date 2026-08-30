package integration

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// Fixture builders. Every field has a sensible default and every default can be
// overridden, because a fixture that demands all twenty columns makes a test
// unreadable and — worse — hides which field the test is actually about.

// seededHousehold is what a test needs to talk about the row it just created: the
// identifiers to query by, and the code it must never find in a response.
type seededHousehold struct {
	ID          int64
	Code        string
	DisplayName string
	Guests      []seededGuest
}

type seededGuest struct {
	ID        int64
	FirstName string
	LastName  string
	Kind      string
	Age       *int
}

type householdOption func(*householdSpec)

type householdSpec struct {
	code        string
	displayName string
	adminNote   string
	guests      []guestSpec
}

type guestSpec struct {
	firstName string
	lastName  string
	kind      string
	age       *int
}

// withCode pins the login code, for a test that logs in or asserts on the code
// itself. Left alone, every household gets a unique generated one.
func withCode(code string) householdOption {
	return func(spec *householdSpec) { spec.code = code }
}

func withDisplayName(displayName string) householdOption {
	return func(spec *householdSpec) { spec.displayName = displayName }
}

// withAdminNote sets the private note. Mostly used to give assertNoLeak something
// real to find.
func withAdminNote(note string) householdOption {
	return func(spec *householdSpec) { spec.adminNote = note }
}

// withGuests adds count adults with generated names, for a test that cares about
// how many people are in a household rather than who they are.
func withGuests(count int) householdOption {
	return func(spec *householdSpec) {
		for range count {
			spec.guests = append(spec.guests, guestSpec{
				// Numbered from the current length, so repeated options and the
				// named builders below keep producing Gast1, Gast2, Gast3.
				firstName: fmt.Sprintf("Gast%d", len(spec.guests)+1),
				lastName:  "Muster",
				kind:      "adult",
			})
		}
	}
}

func withAdult(firstName, lastName string) householdOption {
	return func(spec *householdSpec) {
		spec.guests = append(spec.guests, guestSpec{firstName: firstName, lastName: lastName, kind: "adult"})
	}
}

// withChild adds a child. The age is age at the wedding date, as the schema means it.
func withChild(firstName, lastName string, age int) householdOption {
	return func(spec *householdSpec) {
		spec.guests = append(spec.guests, guestSpec{firstName: firstName, lastName: lastName, kind: "child", age: &age})
	}
}

// seedHousehold inserts a household and its guests, returning what the test needs to
// address them. Guests are seeded, never guest_added: a fixture describes the state
// before the household has touched anything.
func seedHousehold(t *testing.T, pool *sqlx.DB, options ...householdOption) seededHousehold {
	t.Helper()

	spec := householdSpec{code: nextTestCode()}
	for _, option := range options {
		option(&spec)
	}
	if spec.displayName == "" {
		// Deliberately not derived from the code. A default that embedded the
		// login code in a *displayed* field would make every household's secret
		// appear legitimately in guest-facing JSON, and assertNoLeak — which
		// searches the body for the code as a value — would fire on every test
		// that did not override it.
		spec.displayName = fmt.Sprintf("Familie Muster %d", testHouseholdCounter.Add(1))
	}

	result, err := pool.Exec(
		`INSERT INTO household (display_name, code, admin_note) VALUES (?, ?, ?)`,
		spec.displayName, spec.code, spec.adminNote,
	)
	require.NoError(t, err)

	household := seededHousehold{
		ID:          lastInsertID(t, result),
		Code:        spec.code,
		DisplayName: spec.displayName,
	}

	for _, guest := range spec.guests {
		guestResult, err := pool.Exec(
			`INSERT INTO guest (household_id, first_name, last_name, kind, age, origin) VALUES (?, ?, ?, ?, ?, 'seeded')`,
			household.ID, guest.firstName, guest.lastName, guest.kind, guest.age,
		)
		require.NoError(t, err)

		household.Guests = append(household.Guests, seededGuest{
			ID:        lastInsertID(t, guestResult),
			FirstName: guest.firstName,
			LastName:  guest.lastName,
			Kind:      guest.kind,
			Age:       guest.age,
		})
	}

	return household
}

// codeAlphabet duplicates the login-code alphabet from package domain, which keeps
// its own copy unexported. Fixtures use it so a seeded code is shaped like a real
// one — a test that accidentally depends on the shape then fails here rather than
// in production.
const codeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

var (
	testCodeCounter      atomic.Int64
	testHouseholdCounter atomic.Int64
)

// nextTestCode returns a unique six-character code. Counted rather than random: the
// column is UNIQUE, and a random generator would make a rare collision into a rare
// flake, which is the worst kind of test failure to chase.
func nextTestCode() string {
	value := testCodeCounter.Add(1)

	code := make([]byte, 6)
	for index := len(code) - 1; index >= 0; index-- {
		code[index] = codeAlphabet[value%int64(len(codeAlphabet))]
		value /= int64(len(codeAlphabet))
	}

	return string(code)
}

// The insert helpers below stay deliberately low-level: the schema tests use them to
// probe one constraint at a time, and defaults they do not set are defaults they are
// checking.

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
