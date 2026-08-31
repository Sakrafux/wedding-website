package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// F4-B02 and F4-B03: a guest invited alone adds the person they are bringing, and can
// take them off again. Everything else is a phone call, and these tests are what say
// so — the rule is only worth anything if it holds against a request that never went
// through the form.

// addMemberBody mirrors dto.RSVPAddMemberResponse rather than importing it, so a
// rename in the DTO fails here as a wire-format assertion instead of compiling.
type addMemberBody struct {
	Member        rsvpMember `json:"member"`
	CanAddPlusOne bool       `json:"can_add_plus_one"`
}

func (response *testResponse) addedMember() addMemberBody {
	response.t.Helper()

	var body addMemberBody
	response.decodeJSON(&body)
	return body
}

func memberPath(id int64) string {
	return fmt.Sprintf("/api/rsvp/members/%d", id)
}

// deletedAt reads the soft-delete marker straight from the table: the whole point of
// a soft delete is that the row survives, and only the table can show that.
func deletedAt(t *testing.T, app *testApp, guestID int64) *string {
	t.Helper()

	var at *string
	require.NoError(t, app.Database.Read.Get(&at, `SELECT deleted_at FROM guest WHERE id = ?`, guestID))
	return at
}

func TestASingleGuestAddsTheirCompanion(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	require.True(t, app.get("/api/rsvp").rsvp().CanAddPlusOne)

	response := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"})
	require.Equal(t, http.StatusCreated, response.Status, response.Body)
	response.assertNoLeak(household.Code)

	added := response.addedMember()
	assert.Equal(t, "Isabella Michelbacher", added.Member.Name)
	assert.Equal(t, "adult", added.Member.Kind)
	assert.Equal(t, "guest_added", added.Member.Origin)
	assert.Nil(t, added.Member.Age)
	// An added person is a new question: a defaulted scope would be an answer nobody
	// gave.
	assert.Nil(t, added.Member.Attending)
	assert.False(t, added.CanAddPlusOne, "the household now has two members")

	form := app.get("/api/rsvp").rsvp()
	require.Len(t, form.Members, 2)
	assert.Equal(t, added.Member, form.memberByID(t, added.Member.ID))
	assert.False(t, form.CanAddPlusOne)
}

// The limit is structural rather than counted: after the first addition the household
// has two members, and the rule that refuses the second is the same one.
func TestASecondPlusOneIsRefused(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"))
	require.Equal(t, http.StatusCreated,
		app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).Status)

	response := app.postJSON("/api/rsvp/members", map[string]any{"name": "Noch Jemand"})

	require.Equal(t, http.StatusConflict, response.Status, response.Body)
	envelope := response.errorEnvelope()
	assert.Equal(t, "plus_one_not_allowed", envelope.Code)
	assert.Equal(t,
		"Weitere Personen tragen wir gern für euch ein — ruf uns bitte kurz an: +43 650 9408100.",
		envelope.Message)
	assert.Len(t, app.get("/api/rsvp").rsvp().Members, 2, "nothing may be written")
}

func TestAHouseholdSeededWithTwoMayNotAddAnybody(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"), withAdult("Bernd Müller"))
	require.False(t, app.get("/api/rsvp").rsvp().CanAddPlusOne)

	response := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"})

	require.Equal(t, http.StatusConflict, response.Status, response.Body)
	assert.Equal(t, "plus_one_not_allowed", response.errorEnvelope().Code)
	assert.Len(t, app.get("/api/rsvp").rsvp().Members, 2)
}

// The body is one field, and a body carrying more is *rejected* rather than silently
// stripped — DecodeJSON refuses unknown fields for the whole API. Asserted here
// because "which of the two" was an open question in F4-B02, and because the property
// that matters either way is the one on the second half: no guest path produces a
// child.
func TestAnAddedMemberIsAnAdultWhateverTheBodySays(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"))

	refused := app.postJSON("/api/rsvp/members", map[string]any{
		"name": "Emma Michelbacher", "kind": "child", "age": 6, "origin": "seeded",
	})

	require.Equal(t, http.StatusBadRequest, refused.Status, refused.Body)
	assert.Equal(t, "validation_failed", refused.errorEnvelope().Code)
	assert.Empty(t, app.get("/api/rsvp").rsvp().Members[1:], "no child may be created here")

	added := app.postJSON("/api/rsvp/members", map[string]any{"name": "Emma Michelbacher"}).addedMember()
	assert.Equal(t, "adult", added.Member.Kind)
	assert.Nil(t, added.Member.Age)
	assert.Equal(t, "guest_added", added.Member.Origin)
}

func TestAddingAMemberRequiresAName(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"))

	response := app.postJSON("/api/rsvp/members", map[string]any{"name": ""})

	require.Equal(t, http.StatusBadRequest, response.Status, response.Body)
	envelope := response.errorEnvelope()
	assert.Equal(t, "validation_failed", envelope.Code)
	assert.Contains(t, envelope.Fields, "name")
}

// A closed form is a closed form regardless of who is being added, and rsvp_closed
// wins over plus_one_not_allowed when both apply.
func TestAddingAMemberIsRefusedAfterTheDeadline(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"), withAdult("Bernd Müller"))
	setRSVPDeadline(t, app.Database.Write, time.Now().Add(-24*time.Hour))

	response := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"})

	require.Equal(t, http.StatusConflict, response.Status, response.Body)
	assert.Equal(t, "rsvp_closed", response.errorEnvelope().Code)
	assert.Len(t, app.get("/api/rsvp").rsvp().Members, 2)
}

// The admin path has no rule at all beyond validity. That asymmetry is the design:
// every addition we do not automate is one we want to hear about, and then enter.
func TestTheAdminAddsToAHouseholdThatHasUsedItsPlusOne(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	require.Equal(t, http.StatusCreated,
		app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).Status)

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)

	created := admin.postJSON(fmt.Sprintf("/api/admin/households/%d/guests", household.ID),
		map[string]any{"name": "Emma Müller", "kind": "child", "age": 6})

	require.Equal(t, http.StatusCreated, created.Status, created.Body)
	assert.Len(t, app.get("/api/rsvp").rsvp().Members, 3)
}

// The row that later answers "where did this person come from".
func TestAddingAMemberWritesOneAuditRow(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))

	added := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).addedMember()

	var rows []auditRow
	for _, row := range app.auditRows() {
		if row.Action == "create" {
			rows = append(rows, row)
		}
	}
	require.Len(t, rows, 1)
	assert.Equal(t, "guest", rows[0].Entity)
	assert.Equal(t, added.Member.ID, rows[0].EntityID)
	assert.Equal(t, "household", rows[0].ActorType)
	require.True(t, rows[0].ActorID.Valid)
	assert.Equal(t, household.ID, rows[0].ActorID.Int64)

	after := changedFields(t, rows[0].After.String)
	assert.Equal(t, "Isabella Michelbacher", after["name"])
	assert.Equal(t, "guest_added", after["origin"])
	assert.Equal(t, "adult", after["kind"])
}

func TestAHouseholdRemovesItsOwnAddition(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"))
	added := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).addedMember()

	response := app.deleteRequest(memberPath(added.Member.ID))

	require.Equal(t, http.StatusNoContent, response.Status, response.Body)
	form := app.get("/api/rsvp").rsvp()
	assert.Len(t, form.Members, 1, "the removed member is nobody's member any more")
	// The one path back from a mistyped name.
	assert.True(t, form.CanAddPlusOne)
	assert.NotNil(t, deletedAt(t, app, added.Member.ID), "the row survives for the audit trail")

	require.Equal(t, http.StatusCreated,
		app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).Status)
}

// Removing somebody who said yes is a real correction, not a mistake to guard against.
func TestAnAnsweredAdditionMayStillBeRemoved(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	added := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).addedMember()
	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp",
		submission(answerFor(household.Guests[0].ID, "both"), answerFor(added.Member.ID, "both"))).Status)

	assert.Equal(t, http.StatusNoContent, app.deleteRequest(memberPath(added.Member.ID)).Status)
}

// A household never deletes somebody we put on the list — they answer "no", which is
// exactly what the message says.
func TestASeededMemberCannotBeRemoved(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))

	response := app.deleteRequest(memberPath(household.Guests[0].ID))

	require.Equal(t, http.StatusConflict, response.Status, response.Body)
	envelope := response.errorEnvelope()
	assert.Equal(t, "cannot_remove_member", envelope.Code)
	assert.Equal(t,
		"Diese Person haben wir eingetragen. Wenn sie nicht kommt, wähl bitte «Kommt nicht» aus.",
		envelope.Message)
	assert.Nil(t, deletedAt(t, app, household.Guests[0].ID))
}

// Not found rather than forbidden: a household must not be able to learn which ids
// exist by reading the difference between the two answers.
func TestAnotherHouseholdsMemberIsNotFound(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"))
	stranger := seedHousehold(t, app.Database.Write, withAdult("Fremde Person"))

	response := app.deleteRequest(memberPath(stranger.Guests[0].ID))

	require.Equal(t, http.StatusNotFound, response.Status, response.Body)
	assert.Equal(t, "not_found", response.errorEnvelope().Code)
	assert.Nil(t, deletedAt(t, app, stranger.Guests[0].ID))
}

func TestRemovingAMemberIsRefusedAfterTheDeadline(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"))
	added := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).addedMember()
	setRSVPDeadline(t, app.Database.Write, time.Now().Add(-24*time.Hour))

	response := app.deleteRequest(memberPath(added.Member.ID))

	require.Equal(t, http.StatusConflict, response.Status, response.Body)
	assert.Equal(t, "rsvp_closed", response.errorEnvelope().Code)
	assert.Nil(t, deletedAt(t, app, added.Member.ID))
}

// The row that explains a headcount that went down.
func TestRemovingAMemberWritesOneAuditRow(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	added := app.postJSON("/api/rsvp/members", map[string]any{"name": "Isabella Michelbacher"}).addedMember()

	require.Equal(t, http.StatusNoContent, app.deleteRequest(memberPath(added.Member.ID)).Status)

	var rows []auditRow
	for _, row := range app.auditRows() {
		if row.Action == "delete" {
			rows = append(rows, row)
		}
	}
	require.Len(t, rows, 1)
	assert.Equal(t, "guest", rows[0].Entity)
	assert.Equal(t, added.Member.ID, rows[0].EntityID)
	assert.Equal(t, "household", rows[0].ActorType)
	require.True(t, rows[0].ActorID.Valid)
	assert.Equal(t, household.ID, rows[0].ActorID.Int64)

	before := changedFields(t, rows[0].Before.String)
	assert.Equal(t, "Isabella Michelbacher", before["name"])
	assert.Equal(t, "guest_added", before["origin"])
}
