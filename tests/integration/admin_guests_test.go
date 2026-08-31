package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

func (response *testResponse) adminGuest() adminGuest {
	response.t.Helper()

	var body adminGuest
	response.decodeJSON(&body)
	return body
}

// members reads a household's members back through the detail endpoint, which is how
// the admin UI sees them.
func (app *testApp) members(householdID int64) []adminGuest {
	app.t.Helper()

	return app.get(fmt.Sprintf("/api/admin/households/%d", householdID)).adminHousehold().Members
}

// The bulk task the epic exists for: real names, entered, corrected and removed
// through the API alone.
func TestAdminAddsUpdatesAndRemovesAMember(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withDisplayName("Familie Müller"))

	created := app.postJSON(fmt.Sprintf("/api/admin/households/%d/guests", household.ID), map[string]any{
		"first_name":   "Emil",
		"last_name":    "Müller",
		"kind":         "child",
		"age":          4,
		"seating_need": "high_chair",
		"dietary_note": "Nussallergie",
	})
	require.Equal(t, http.StatusCreated, created.Status)

	guest := created.adminGuest()
	assert.Positive(t, guest.ID)
	assert.Equal(t, household.ID, guest.HouseholdID)
	require.NotNil(t, guest.Age)
	assert.Equal(t, 4, *guest.Age)
	assert.Equal(t, "high_chair", guest.SeatingNeed)
	assert.Equal(t, "Nussallergie", guest.DietaryNote)

	// Never guest_added: origin is what the admin delta view reads to answer "what
	// did the households add themselves", and this is not that.
	assert.Equal(t, string(domain.GuestOriginSeeded), guest.Origin)

	require.Len(t, app.members(household.ID), 1)

	path := fmt.Sprintf("/api/admin/guests/%d", guest.ID)
	updated := app.patchJSON(path, map[string]any{"dietary_note": "Nuss- und Sellerieallergie"})
	require.Equal(t, http.StatusOK, updated.Status)
	assert.Equal(t, "Nuss- und Sellerieallergie", updated.adminGuest().DietaryNote)
	assert.Equal(t, "Emil", updated.adminGuest().FirstName, "an absent field is left alone")

	require.Equal(t, http.StatusNoContent, app.deleteRequest(path).Status)
	assert.Empty(t, app.members(household.ID))
}

// Omitting the seating need means the ordinary case, so entering eighty guests does
// not mean answering the same question eighty times.
func TestAMemberCreatedWithoutASeatingNeedGetsTheDefault(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write)

	created := app.postJSON(fmt.Sprintf("/api/admin/households/%d/guests", household.ID), map[string]any{
		"first_name": "Anna",
		"last_name":  "Müller",
		"kind":       "adult",
	})

	require.Equal(t, http.StatusCreated, created.Status)
	assert.Equal(t, string(domain.SeatingNeedNormal), created.adminGuest().SeatingNeed)
	assert.Nil(t, created.adminGuest().Age)
}

// The pairing is a domain rule and not only a column CHECK, because a driver error is
// not a message a form can put next to a field.
func TestAnAgeOnAnAdultIsAFieldErrorAndNotADriverError(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write)

	response := app.postJSON(fmt.Sprintf("/api/admin/households/%d/guests", household.ID), map[string]any{
		"first_name": "Anna",
		"last_name":  "Müller",
		"kind":       "adult",
		"age":        30,
	})

	require.Equal(t, http.StatusBadRequest, response.Status)
	envelope := response.errorEnvelope()
	assert.Equal(t, "validation_failed", envelope.Code)
	assert.Contains(t, envelope.Fields, "age")
	assert.Contains(t, envelope.Fields["age"], "Kinder")
}

func TestAChildsAgeMustBeInRange(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write)

	response := app.postJSON(fmt.Sprintf("/api/admin/households/%d/guests", household.ID), map[string]any{
		"first_name": "Emil",
		"last_name":  "Müller",
		"kind":       "child",
		"age":        18,
	})

	require.Equal(t, http.StatusBadRequest, response.Status)
	assert.Contains(t, response.errorEnvelope().Fields["age"], "17")
}

// A stale age would violate the column CHECK and, worse than failing, would keep
// feeding a caterer bracket for somebody who is now an adult.
func TestPromotingAChildToAnAdultClearsTheAge(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withChild("Emil", "Müller", 17))
	guestID := household.Guests[0].ID

	response := app.patchJSON(fmt.Sprintf("/api/admin/guests/%d", guestID), map[string]any{"kind": "adult"})

	require.Equal(t, http.StatusOK, response.Status)
	assert.Equal(t, "adult", response.adminGuest().Kind)
	assert.Nil(t, response.adminGuest().Age)

	stored, err := app.guestStore().FindByID(t.Context(), guestID)
	require.NoError(t, err)
	assert.Nil(t, stored.Age)
}

// An explicit null clears the age; an absent key leaves it alone. Both are legitimate
// requests from the same form, which is why the DTO can tell them apart.
func TestANullAgeClearsItAndAnAbsentAgeDoesNot(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withChild("Emil", "Müller", 4))
	path := fmt.Sprintf("/api/admin/guests/%d", household.Guests[0].ID)

	untouched := app.patchJSON(path, map[string]any{"first_name": "Emilia"})
	require.Equal(t, http.StatusOK, untouched.Status)
	require.NotNil(t, untouched.adminGuest().Age)
	assert.Equal(t, 4, *untouched.adminGuest().Age)

	cleared := app.patchJSON(path, map[string]any{"age": nil})
	require.Equal(t, http.StatusOK, cleared.Status)
	assert.Nil(t, cleared.adminGuest().Age)
	assert.Equal(t, "child", cleared.adminGuest().Kind, "clearing the age does not make a child an adult")
}

// Soft, not hard: the guest was counted, may hold a seat, and appears in the audit
// trail. Erasing the row would leave those dangling and the history unexplainable.
func TestRemovingAMemberKeepsTheRowAndItsSeatAssignment(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withAdult("Anna", "Müller"))
	guestID := household.Guests[0].ID

	unit := insertSeatingUnit(t, app.Database.Write, "party", "Tisch 1")
	seat := insertSeat(t, app.Database.Write, unit, "party", "Platz 1")
	assignSeat(t, app.Database.Write, seat, "party", guestID)

	require.Equal(t, http.StatusNoContent, app.deleteRequest(fmt.Sprintf("/api/admin/guests/%d", guestID)).Status)

	assert.Empty(t, app.members(household.ID), "gone from the member list")

	var deletedAt *string
	require.NoError(t, app.Database.Read.Get(&deletedAt, `SELECT deleted_at FROM guest WHERE id = ?`, guestID))
	require.NotNil(t, deletedAt, "the row and its deleted_at remain")

	// Left in place on purpose: F7-B01 reports it as a stale assignment for a human
	// to resolve, because seats must not quietly vanish from a finished plan.
	var assignments int
	require.NoError(t, app.Database.Read.Get(&assignments,
		`SELECT COUNT(*) FROM seat_assignment WHERE guest_id = ?`, guestID))
	assert.Equal(t, 1, assignments)
}

// A removed guest is not editable any more, and the request must say so rather than
// silently writing to a row nothing reads.
func TestARemovedMemberCannotBeEditedOrRemovedAgain(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withAdult("Anna", "Müller"))
	path := fmt.Sprintf("/api/admin/guests/%d", household.Guests[0].ID)

	require.Equal(t, http.StatusNoContent, app.deleteRequest(path).Status)

	assert.Equal(t, http.StatusNotFound, app.patchJSON(path, map[string]any{"first_name": "Anne"}).Status)
	assert.Equal(t, http.StatusNotFound, app.deleteRequest(path).Status)
}

// An empty household is a real state: we know a name, we have not yet asked who is
// coming.
func TestRemovingTheLastMemberLeavesTheHouseholdIntact(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withAdult("Anna", "Müller"))

	require.Equal(t, http.StatusNoContent,
		app.deleteRequest(fmt.Sprintf("/api/admin/guests/%d", household.Guests[0].ID)).Status)

	detail := app.get(fmt.Sprintf("/api/admin/households/%d", household.ID))
	require.Equal(t, http.StatusOK, detail.Status)
	assert.Equal(t, 0, detail.adminHousehold().MemberCount)
}

// Guests are addressed by their own id and the admin owns all of them. Asserted so
// the route is not later "fixed" into a household-scoped one the frontend does not
// call.
func TestAGuestIsAddressedByTheirOwnIDAcrossHouseholds(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	first := seedHousehold(t, app.Database.Write, withAdult("Anna", "Müller"))
	second := seedHousehold(t, app.Database.Write, withAdult("Bernd", "Schmidt"))

	for _, household := range []seededHousehold{first, second} {
		response := app.patchJSON(fmt.Sprintf("/api/admin/guests/%d", household.Guests[0].ID),
			map[string]any{"seating_need": "wheelchair"})

		require.Equal(t, http.StatusOK, response.Status)
		assert.Equal(t, household.ID, response.adminGuest().HouseholdID)
		assert.Equal(t, "wheelchair", response.adminGuest().SeatingNeed)
	}
}

func TestAdminGuestValidationFailuresNameTheirFields(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write)

	response := app.postJSON(fmt.Sprintf("/api/admin/households/%d/guests", household.ID), map[string]any{
		"first_name": "",
		"last_name":  "Müller",
		"kind":       "erwachsen",
	})

	require.Equal(t, http.StatusBadRequest, response.Status)
	envelope := response.errorEnvelope()
	assert.Contains(t, envelope.Fields, "first_name")
	// The enum values are English on the wire; German exists only as a label in the
	// frontend, so a German value is not a value at all.
	assert.Contains(t, envelope.Fields, "kind")
}

func TestEachGuestMutationWritesExactlyOneAuditRow(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write)

	created := app.postJSON(fmt.Sprintf("/api/admin/households/%d/guests", household.ID), map[string]any{
		"first_name": "Emil", "last_name": "Müller", "kind": "child", "age": 4,
	})
	require.Equal(t, http.StatusCreated, created.Status)
	guest := created.adminGuest()

	path := fmt.Sprintf("/api/admin/guests/%d", guest.ID)
	require.Equal(t, http.StatusOK, app.patchJSON(path, map[string]any{"kind": "adult"}).Status)
	require.Equal(t, http.StatusNoContent, app.deleteRequest(path).Status)

	var actions []string
	for _, row := range app.auditRows() {
		if row.Entity == domain.AuditEntityGuest {
			actions = append(actions, row.Action)
			assert.Equal(t, "admin", row.ActorType)
			assert.Equal(t, guest.ID, row.EntityID)
			assert.False(t, row.ActorID.Valid, "there is one admin and no row describing them")
		}
	}
	assert.Equal(t, []string{"create", "update", "delete"}, actions)
}
