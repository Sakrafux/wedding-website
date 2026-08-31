package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// F3-B05: every RSVP change is traceable to a who, a when and a what — and no
// unchanged save appears in the log, because a log of "somebody pressed save" is one
// nobody reads.

// rsvpAuditRows is the log with the login rows dropped, since every test here logs in
// first and the login entry is F1-B06's business.
func rsvpAuditRows(app *testApp) []auditRow {
	var rows []auditRow
	for _, row := range app.auditRows() {
		if row.Action == "update" {
			rows = append(rows, row)
		}
	}
	return rows
}

func changedFields(t *testing.T, payload string) map[string]any {
	t.Helper()

	var fields map[string]any
	require.NoError(t, json.Unmarshal([]byte(payload), &fields))
	return fields
}

// One row per entity, not one per save: the household and each member are separate
// rows with separate ids, and a single entry covering five people could not be read
// back against any of them.
func TestRSVPSaveAuditsOneRowPerChangedMember(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"), withAdult("Bernd Müller"))
	anna, bernd := household.Guests[0], household.Guests[1]

	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp",
		submission(answerFor(anna.ID, "both"), answerFor(bernd.ID, "both"))).Status)

	rows := rsvpAuditRows(app)
	require.Len(t, rows, 2, "two members changed, the household's own fields did not")
	for _, row := range rows {
		assert.Equal(t, "guest", row.Entity)
		assert.Equal(t, "household", row.ActorType)
		require.True(t, row.ActorID.Valid)
		assert.Equal(t, household.ID, row.ActorID.Int64)
	}
	assert.ElementsMatch(t, []int64{anna.ID, bernd.ID}, []int64{rows[0].EntityID, rows[1].EntityID})
}

func TestRSVPSaveOfTheNoteAloneAuditsOnlyTheHousehold(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both"))).Status)

	body := submission(answerFor(anna.ID, "both"))
	body["rsvp_note"] = "Oma braucht einen Platz nah am Ausgang."
	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp", body).Status)

	rows := rsvpAuditRows(app)
	require.Len(t, rows, 2, "one row for the first answer, one for the note")

	note := rows[1]
	assert.Equal(t, "household", note.Entity)
	assert.Equal(t, household.ID, note.EntityID)

	// The note itself is recorded: it is the household's own words to us, the log is
	// admin-only, and a note edited to remove a request is why the log exists.
	after := changedFields(t, note.After.String)
	assert.Equal(t, "Oma braucht einen Platz nah am Ausgang.", after["rsvp_note"])
	assert.Len(t, after, 1, "only the changed field")
}

func TestRSVPSaveThatChangesNothingIsNotAudited(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both"))).Status)
	before := len(rsvpAuditRows(app))

	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both"))).Status)

	assert.Len(t, rsvpAuditRows(app), before)
}

// Before and after carry the changed fields only, with the old and the new value:
// the log is a history and not a second copy of the guest table.
func TestRSVPAuditRecordsOnlyTheChangedFieldsWithBothValues(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	first := answerFor(anna.ID, "both")
	first["meal_choice"] = "vegetarian"
	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp", submission(first)).Status)

	second := answerFor(anna.ID, "both")
	second["meal_choice"] = "vegan"
	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp", submission(second)).Status)

	rows := rsvpAuditRows(app)
	require.Len(t, rows, 2)

	change := rows[1]
	assert.Equal(t, map[string]any{"meal_choice": "vegetarian"}, changedFields(t, change.Before.String))
	assert.Equal(t, map[string]any{"meal_choice": "vegan"}, changedFields(t, change.After.String))
}

// A broken audit table must not lose a household's answer. Forced by dropping the
// table out from under the running application, as the login suite does.
func TestRSVPSaveSurvivesAFailingAuditWrite(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))

	_, err := app.Database.Write.Exec(`DROP TABLE audit_log`)
	require.NoError(t, err)

	response := app.putJSON("/api/rsvp", submission(answerFor(household.Guests[0].ID, "both")))

	require.Equal(t, http.StatusOK, response.Status, "the answer matters more than the record of it")
	require.NotNil(t, response.rsvp().Household.RSVPSubmittedAt)
	assert.Contains(t, app.Logs.String(), "audit write failed", "the silence has to be visible somewhere")
}
