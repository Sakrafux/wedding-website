package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// F3-B04: the deadline has to hold against a request that bypasses the frontend
// entirely, and must not hold against us.

func TestRSVPWriteIsRefusedAfterTheDeadline(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	setRSVPDeadline(t, app.Database.Write, time.Now().Add(-24*time.Hour))

	response := app.putJSON("/api/rsvp", submission(answerFor(household.Guests[0].ID, "both")))

	require.Equal(t, http.StatusConflict, response.Status, response.Body)
	envelope := response.errorEnvelope()
	assert.Equal(t, "rsvp_closed", envelope.Code)
	assert.Equal(t,
		"Die Rückmeldefrist ist vorbei. Wenn sich etwas geändert hat, ruf uns bitte kurz an.",
		envelope.Message)

	stored := app.get("/api/rsvp").rsvp()
	assert.Nil(t, stored.Household.RSVPSubmittedAt, "nothing may be written after the deadline")
	assert.Nil(t, stored.memberByID(t, household.Guests[0].ID).Attending)
}

// Reading stays open forever: a household must be able to see what they answered,
// which is what the read-only view renders (F3-F05).
func TestRSVPReadStaysOpenAfterTheDeadline(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	require.Equal(t, http.StatusOK,
		app.putJSON("/api/rsvp", submission(answerFor(household.Guests[0].ID, "both"))).Status)

	deadline := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	setRSVPDeadline(t, app.Database.Write, deadline)

	response := app.get("/api/rsvp")
	require.Equal(t, http.StatusOK, response.Status)

	body := response.rsvp()
	assert.False(t, body.Editable)
	assert.Equal(t, deadline.Format(time.RFC3339), body.Deadline,
		"the form renders its deadline from this response, not from another endpoint's cache")
	require.NotNil(t, body.memberByID(t, household.Guests[0].ID).Attending)
}

func TestRSVPReportsEditableBeforeTheDeadline(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Anna Müller"))
	setRSVPDeadline(t, app.Database.Write, time.Now().Add(48*time.Hour))

	assert.True(t, app.get("/api/rsvp").rsvp().Editable)
}

// The boundary: Settings.RSVPOpen is a strict Before, so the deadline instant itself
// counts as closed. Asserted because "closes at 21:59:59" and "closes after 21:59:59"
// are one second apart in the code and a day apart in a guest's mind.
func TestRSVPDeadlineAtThisInstantCountsAsClosed(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	setRSVPDeadline(t, app.Database.Write, time.Now())

	response := app.putJSON("/api/rsvp", submission(answerFor(household.Guests[0].ID, "both")))

	assert.Equal(t, http.StatusConflict, response.Status)
	assert.Equal(t, "rsvp_closed", response.errorEnvelope().Code)
}

// A closed form says that it is closed rather than listing three field errors first:
// the guest's problem is the deadline, and fixing a radio button would not help.
func TestRSVPClosedFormOutranksAnInvalidBody(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	setRSVPDeadline(t, app.Database.Write, time.Now().Add(-time.Hour))

	// Both wrong at once: an unanswerable scope and a member set that does not match.
	response := app.putJSON("/api/rsvp", submission(answerFor(household.Guests[0].ID, "maybe"),
		answerFor(household.Guests[0].ID, "no")))

	require.Equal(t, http.StatusConflict, response.Status)
	assert.Equal(t, "rsvp_closed", response.errorEnvelope().Code)
}
