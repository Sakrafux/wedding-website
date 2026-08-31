package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// F3-B06: the answer we take down the phone must land in the same place as everybody
// else's, and be indistinguishable in the data — except in the audit log, where it
// must be distinguishable.

func adminRSVPPath(householdID int64) string {
	return fmt.Sprintf("/api/admin/households/%d/rsvp", householdID)
}

func TestAdminWritesAHouseholdAnswerTheGuestThenSees(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write,
		withCode("ABC234"), withAdult("Anna Müller"), withAdult("Bernd Müller"))
	anna, bernd := household.Guests[0], household.Guests[1]

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)

	body := submission(answerFor(anna.ID, "both"), answerFor(bernd.ID, "church_only"))
	body["rsvp_note"] = "Am Telefon durchgegeben."
	saved := admin.putJSON(adminRSVPPath(household.ID), body)
	require.Equal(t, http.StatusOK, saved.Status, saved.Body)

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	seenByTheGuest := app.get("/api/rsvp").rsvp()

	require.NotNil(t, seenByTheGuest.memberByID(t, anna.ID).Attending)
	assert.Equal(t, "both", *seenByTheGuest.memberByID(t, anna.ID).Attending)
	assert.Equal(t, "Am Telefon durchgegeben.", seenByTheGuest.Household.RSVPNote)
}

// The same use case, addressed by id: the two GET bodies must be byte-identical, so a
// field added to one and not the other fails a test rather than a screen.
func TestAdminAndGuestRSVPBodiesAreIdentical(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write,
		withCode("ABC234"), withAdminNote("Ruft nie zurück"), withAdult("Anna Müller"), withChild("Emma Müller", 6))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	full := submission(answerFor(household.Guests[0].ID, "both"), answerFor(household.Guests[1].ID, "party_only"))
	full["rsvp_note"] = "Wir bringen einen Kinderwagen mit."
	full["has_stroller"] = true
	require.Equal(t, http.StatusOK, app.putJSON("/api/rsvp", full).Status)

	guestBody := app.get("/api/rsvp").Body

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)
	adminResponse := admin.get(adminRSVPPath(household.ID))

	assert.Equal(t, guestBody, adminResponse.Body)
	// No admin-only fields here either: code and admin_note belong to the household
	// endpoints, and adding them would mean the shared form component had to know
	// which caller it was serving.
	adminResponse.assertNoLeak(household.Code, "Ruft nie zurück")
}

// The route exists for the late call, so the deadline must not stop it — while the
// guest's own route still refuses.
func TestAdminWritesAfterTheDeadlineAndTheGuestDoesNot(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"), withAdult("Anna Müller"))
	anna := household.Guests[0]
	setRSVPDeadline(t, app.Database.Write, time.Now().Add(-24*time.Hour))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	refused := app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both")))
	require.Equal(t, http.StatusConflict, refused.Status)
	assert.Equal(t, "rsvp_closed", refused.errorEnvelope().Code)

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)
	accepted := admin.putJSON(adminRSVPPath(household.ID), submission(answerFor(anna.ID, "both")))

	require.Equal(t, http.StatusOK, accepted.Status, accepted.Body)
	body := accepted.rsvp()
	// The honest report of the deadline, not a statement about this caller: the admin
	// page shows it as information ("Frist ist abgelaufen — du kannst trotzdem
	// speichern") rather than as a lock.
	assert.False(t, body.Editable)
	require.NotNil(t, body.memberByID(t, anna.ID).Attending)
}

// If we took the answer down the phone, the household has answered and must drop off
// the nudge list.
func TestAdminSaveSetsSubmittedAt(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withAdult("Anna Müller"))

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)

	body := admin.putJSON(adminRSVPPath(household.ID),
		submission(answerFor(household.Guests[0].ID, "both"))).rsvp()

	assert.NotNil(t, body.Household.RSVPSubmittedAt)
}

// The one thing that must differ between the two paths: who the log says did it.
// Recording our own typing as the household's answer would mislead at the exact
// moment the log is consulted.
func TestAdminRSVPSaveIsAuditedAsAdminAndTheGuestSaveAsHousehold(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"), withAdult("Anna Müller"))
	anna := household.Guests[0]

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)
	require.Equal(t, http.StatusOK,
		admin.putJSON(adminRSVPPath(household.ID), submission(answerFor(anna.ID, "both"))).Status)

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	require.Equal(t, http.StatusOK,
		app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "church_only"))).Status)

	rows := rsvpAuditRows(app)
	require.Len(t, rows, 2)

	byAdmin, byHousehold := rows[0], rows[1]
	assert.Equal(t, "admin", byAdmin.ActorType)
	assert.False(t, byAdmin.ActorID.Valid, "there is no admin row to point at")
	assert.Equal(t, anna.ID, byAdmin.EntityID, "the entity is still the guest that changed")

	assert.Equal(t, "household", byHousehold.ActorType)
	require.True(t, byHousehold.ActorID.Valid)
	assert.Equal(t, household.ID, byHousehold.ActorID.Int64)
}

func TestAdminRSVPRefusesAnUnknownHousehold(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)

	read := app.get(adminRSVPPath(4711))
	assert.Equal(t, http.StatusNotFound, read.Status)
	assert.Equal(t, "not_found", read.errorEnvelope().Code)

	written := app.putJSON(adminRSVPPath(4711), submission())
	assert.Equal(t, http.StatusNotFound, written.Status)
	assert.Equal(t, "not_found", written.errorEnvelope().Code)
}

// The stale-tab rule applies here too, and for the same reason: two people editing
// one household is exactly what the admin path adds.
func TestAdminRSVPRefusesAMemberSetThatDoesNotMatch(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withAdult("Anna Müller"), withAdult("Bernd Müller"))

	response := app.putJSON(adminRSVPPath(household.ID),
		submission(answerFor(household.Guests[0].ID, "both")))

	assert.Equal(t, http.StatusConflict, response.Status)
	assert.Equal(t, "member_set_mismatch", response.errorEnvelope().Code)
}
