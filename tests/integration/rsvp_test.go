package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rsvpBody mirrors dto.RSVPResponse rather than importing it, so a rename in the DTO
// surfaces here as a failing wire-format assertion instead of compiling silently.
type rsvpBody struct {
	Household rsvpHousehold `json:"household"`
	Members   []rsvpMember  `json:"members"`
	Deadline  string        `json:"deadline"`
	Editable  bool          `json:"editable"`
}

type rsvpHousehold struct {
	ID                    int64   `json:"id"`
	DisplayName           string  `json:"display_name"`
	TransportSeatsNeeded  int     `json:"transport_seats_needed"`
	TransportSeatsOffered int     `json:"transport_seats_offered"`
	HasStroller           bool    `json:"has_stroller"`
	RSVPNote              string  `json:"rsvp_note"`
	RSVPSubmittedAt       *string `json:"rsvp_submitted_at"`
	RSVPUpdatedAt         *string `json:"rsvp_updated_at"`
}

type rsvpMember struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Kind          string  `json:"kind"`
	Age           *int    `json:"age"`
	Origin        string  `json:"origin"`
	Attending     *string `json:"attending"`
	MealChoice    *string `json:"meal_choice"`
	Portion       string  `json:"portion"`
	MidnightSnack bool    `json:"midnight_snack"`
	SeatingNeed   string  `json:"seating_need"`
	DietaryNote   string  `json:"dietary_note"`
}

func (response *testResponse) rsvp() rsvpBody {
	response.t.Helper()

	var body rsvpBody
	response.decodeJSON(&body)
	return body
}

// memberByID is how a test addresses one card in the response, since the assertions
// are about a person rather than about a position.
func (body rsvpBody) memberByID(t *testing.T, id int64) rsvpMember {
	t.Helper()

	for _, member := range body.Members {
		if member.ID == id {
			return member
		}
	}
	t.Fatalf("member %d is not in the response", id)
	return rsvpMember{}
}

// answerFor is a submitted answer for one member, with the ordinary defaults. Tests
// override the field they are about, which keeps the interesting value visible.
func answerFor(id int64, attending string) map[string]any {
	return map[string]any{
		"id":             id,
		"attending":      attending,
		"portion":        "full",
		"midnight_snack": false,
		"seating_need":   "normal",
		"dietary_note":   "",
	}
}

// submission is a complete household body — the only shape PUT /api/rsvp accepts.
func submission(members ...map[string]any) map[string]any {
	return map[string]any{
		"transport_seats_needed":  0,
		"transport_seats_offered": 0,
		"has_stroller":            false,
		"rsvp_note":               "",
		"members":                 members,
	}
}

// newHouseholdApp starts an application with one household logged in, which is the
// precondition of nearly every test in this file.
func newHouseholdApp(t *testing.T, options ...householdOption) (*testApp, seededHousehold) {
	t.Helper()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, append([]householdOption{withCode("ABC234")}, options...)...)
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	return app, household
}

// F3-B02: the form must open on exactly what was answered last time, or changing one
// person's meal means filling the whole thing in again.
func TestRSVPReadsBackWhatWasSaved(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"), withChild("Emma Müller", 6))
	anna, emma := household.Guests[0], household.Guests[1]

	saved := app.putJSON("/api/rsvp", map[string]any{
		"transport_seats_needed":  1,
		"transport_seats_offered": 2,
		"has_stroller":            true,
		"rsvp_note":               "Wir kommen erst nach der Zeremonie.",
		"members": []map[string]any{
			{
				"id": anna.ID, "attending": "both", "meal_choice": "vegetarian", "portion": "full",
				"midnight_snack": true, "seating_need": "wheelchair", "dietary_note": "Nussallergie",
			},
			{
				"id": emma.ID, "attending": "both", "meal_choice": "all", "portion": "kids",
				"midnight_snack": false, "seating_need": "high_chair", "dietary_note": "", "age": 7,
			},
		},
	})
	require.Equal(t, http.StatusOK, saved.Status, saved.Body)

	reloaded := app.get("/api/rsvp")
	require.Equal(t, http.StatusOK, reloaded.Status)

	body := reloaded.rsvp()
	assert.Equal(t, 1, body.Household.TransportSeatsNeeded)
	assert.Equal(t, 2, body.Household.TransportSeatsOffered)
	assert.True(t, body.Household.HasStroller)
	assert.Equal(t, "Wir kommen erst nach der Zeremonie.", body.Household.RSVPNote)

	annaAnswer := body.memberByID(t, anna.ID)
	require.NotNil(t, annaAnswer.Attending)
	assert.Equal(t, "both", *annaAnswer.Attending)
	require.NotNil(t, annaAnswer.MealChoice)
	assert.Equal(t, "vegetarian", *annaAnswer.MealChoice)
	assert.True(t, annaAnswer.MidnightSnack)
	assert.Equal(t, "wheelchair", annaAnswer.SeatingNeed)
	assert.Equal(t, "Nussallergie", annaAnswer.DietaryNote)

	emmaAnswer := body.memberByID(t, emma.ID)
	assert.Equal(t, "kids", emmaAnswer.Portion)
	assert.Equal(t, "high_chair", emmaAnswer.SeatingNeed)
	require.NotNil(t, emmaAnswer.Age, "a child's age is editable through the RSVP")
	assert.Equal(t, 7, *emmaAnswer.Age)

	// The response is the whole form's data, so the members come in the order
	// ListMembers uses — a form that reshuffled between visits would make an
	// unconfident guest doubt they are looking at the right household.
	assert.Equal(t, []int64{anna.ID, emma.ID}, []int64{body.Members[0].ID, body.Members[1].ID})
	reloaded.assertNoLeak(household.Code)
}

// A household that has never answered must be distinguishable from one that declined:
// a defaulted "no" on the wire would be an answer nobody gave.
func TestRSVPReportsAnUnansweredHouseholdAsUnanswered(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withGuests(3))

	body := app.get("/api/rsvp").rsvp()

	require.Len(t, body.Members, 3)
	for _, member := range body.Members {
		assert.Nilf(t, member.Attending, "%s should have no answer yet", member.Name)
		assert.Nilf(t, member.MealChoice, "%s should have no meal choice yet", member.Name)
		// The column defaults, which is what "not asked" looks like in the row.
		assert.Equal(t, "full", member.Portion)
		assert.Equal(t, "normal", member.SeatingNeed)
	}
	assert.Nil(t, body.Household.RSVPSubmittedAt)
	assert.Nil(t, body.Household.RSVPUpdatedAt)
	assert.NotEmpty(t, body.Household.DisplayName)
	assert.Equal(t, household.ID, body.Household.ID)
}

func TestRSVPOmitsSoftDeletedMembers(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"), withAdult("Bernd Müller"))

	_, err := app.Database.Write.Exec(
		`UPDATE guest SET deleted_at = '2026-09-01T10:00:00Z' WHERE id = ?`, household.Guests[1].ID)
	require.NoError(t, err)

	body := app.get("/api/rsvp").rsvp()

	require.Len(t, body.Members, 1)
	assert.Equal(t, household.Guests[0].ID, body.Members[0].ID)
}

// The admin has no household, so the guest route has nothing to answer with. The
// admin's own route is /api/admin/households/{id}/rsvp.
func TestRSVPRefusesAnAdminSession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)

	response := app.get("/api/rsvp")

	assert.Equal(t, http.StatusUnauthorized, response.Status)
	assert.Equal(t, "unauthenticated", response.errorEnvelope().Code)
}

// F3-B03: the scope gate, through the API. A church-only guest must never carry a
// meal choice, whatever the form sends.
func TestRSVPStoresNoCateringForAGuestOutsideTheParty(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Oma Erika"))
	oma := household.Guests[0]

	response := app.putJSON("/api/rsvp", submission(map[string]any{
		"id": oma.ID, "attending": "church_only", "meal_choice": "vegan", "portion": "kids",
		"midnight_snack": true, "seating_need": "wheelchair", "dietary_note": "Laktose",
	}))
	require.Equal(t, http.StatusOK, response.Status, response.Body)

	// The response is the stored answer, so the client sees what was kept rather than
	// what it sent.
	answer := response.rsvp().memberByID(t, oma.ID)
	assert.Nil(t, answer.MealChoice)
	assert.Equal(t, "full", answer.Portion, "the schema default, not `none`")
	assert.False(t, answer.MidnightSnack)
	// The two fields the scope gate deliberately does not touch: a wheelchair space is
	// needed in the pew, and an allergy matters wherever somebody eats.
	assert.Equal(t, "wheelchair", answer.SeatingNeed)
	assert.Equal(t, "Laktose", answer.DietaryNote)
}

// The pair of timestamps is what F6 reads for "answered" versus "changed since we
// last looked", so the first must never move.
func TestRSVPSetsSubmittedOnceAndUpdatedOnEverySave(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	first := app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both"))).rsvp()
	require.NotNil(t, first.Household.RSVPSubmittedAt)
	require.NotNil(t, first.Household.RSVPUpdatedAt)

	// A whole second, because the stored format has second precision: without it the
	// second save could legitimately land on the same timestamp.
	time.Sleep(1100 * time.Millisecond)

	second := app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "church_only"))).rsvp()

	require.NotNil(t, second.Household.RSVPSubmittedAt)
	assert.Equal(t, *first.Household.RSVPSubmittedAt, *second.Household.RSVPSubmittedAt,
		"the first answer's timestamp must not move")
	require.NotNil(t, second.Household.RSVPUpdatedAt)
	assert.NotEqual(t, *first.Household.RSVPUpdatedAt, *second.Household.RSVPUpdatedAt)
}

// A no-op save must move nothing: F6 compares rsvp_note_seen_at against
// rsvp_updated_at to decide whether a note is unread, so re-saving an unchanged form
// would silently re-flag a note we have already read.
func TestRSVPSaveThatChangesNothingMovesNoTimestamp(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	first := app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both"))).rsvp()
	time.Sleep(1100 * time.Millisecond)

	repeated := app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both")))
	require.Equal(t, http.StatusOK, repeated.Status)

	body := repeated.rsvp()
	assert.Equal(t, *first.Household.RSVPUpdatedAt, *body.Household.RSVPUpdatedAt)
	assert.Equal(t, *first.Household.RSVPSubmittedAt, *body.Household.RSVPSubmittedAt)
}

// The stale-tab case: a body describing a member set that no longer exists must be
// refused rather than written, because writing it would lose an answer.
func TestRSVPRefusesAMemberSetThatDoesNotMatch(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write,
		withCode("ABC234"), withAdult("Anna Müller"), withAdult("Bernd Müller"))
	other := seedHousehold(t, app.Database.Write, withAdult("Fremde Person"))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	anna, bernd := household.Guests[0], household.Guests[1]

	bodies := map[string]map[string]any{
		"a member of another household": submission(
			answerFor(anna.ID, "both"), answerFor(other.Guests[0].ID, "both")),
		"a missing member": submission(answerFor(anna.ID, "both")),
		"a duplicated member": submission(
			answerFor(anna.ID, "both"), answerFor(anna.ID, "no")),
		"an extra member": submission(
			answerFor(anna.ID, "both"), answerFor(bernd.ID, "both"), answerFor(other.Guests[0].ID, "no")),
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			response := app.putJSON("/api/rsvp", body)

			assert.Equal(t, http.StatusConflict, response.Status, response.Body)
			envelope := response.errorEnvelope()
			assert.Equal(t, "member_set_mismatch", envelope.Code)
			assert.Equal(t, "Die Liste der Personen hat sich geändert. Bitte lade die Seite neu.", envelope.Message)

			// Nothing written: not our household's answer, and certainly not the
			// other household's member.
			stored := app.get("/api/rsvp").rsvp()
			assert.Nil(t, stored.Household.RSVPSubmittedAt)
			for _, member := range stored.Members {
				assert.Nil(t, member.Attending)
			}
			foreign, err := app.guestStore().FindByID(t.Context(), other.Guests[0].ID)
			require.NoError(t, err)
			assert.Nil(t, foreign.Attending, "another household's answer must be untouched")
		})
	}
}

// There is no way to store half an answer: rsvp_submitted_at means "they have told us
// who is coming", and a save leaving somebody at null would make the nudge list lie.
func TestRSVPRequiresAnAnswerForEveryMember(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"), withAdult("Bernd Müller"))
	anna, bernd := household.Guests[0], household.Guests[1]

	unanswered := answerFor(bernd.ID, "both")
	delete(unanswered, "attending")

	response := app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "both"), unanswered))

	require.Equal(t, http.StatusBadRequest, response.Status, response.Body)
	envelope := response.errorEnvelope()
	assert.Equal(t, "validation_failed", envelope.Code)
	// Keyed by member id, not by array index: the frontend renders cards by id, and an
	// index breaks the moment the list is filtered.
	assert.Contains(t, envelope.Fields, keyFor(bernd.ID, "attending"))
	assert.NotContains(t, envelope.Fields, keyFor(anna.ID, "attending"))
}

func TestRSVPReportsAnInvalidScopeOnTheRightMember(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	response := app.putJSON("/api/rsvp", submission(answerFor(anna.ID, "maybe")))

	require.Equal(t, http.StatusBadRequest, response.Status)
	assert.Contains(t, response.errorEnvelope().Fields, keyFor(anna.ID, "attending"))
}

// A failure partway through must leave nothing behind: the household row and n member
// rows are one answer, and a partial write is a household that told us four people are
// coming and one meal.
func TestRSVPSaveIsAllOrNothing(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"), withChild("Emma Müller", 6))
	anna, emma := household.Guests[0], household.Guests[1]

	// The invalid age is on the last member, so the first member's answer and the
	// household's own fields would already have been written by a save without a
	// transaction.
	invalidAge := answerFor(emma.ID, "both")
	invalidAge["age"] = 40

	response := app.putJSON("/api/rsvp", map[string]any{
		"transport_seats_needed":  3,
		"transport_seats_offered": 0,
		"has_stroller":            true,
		"rsvp_note":               "Bitte einen Platz nah am Ausgang.",
		"members":                 []map[string]any{answerFor(anna.ID, "both"), invalidAge},
	})

	require.Equal(t, http.StatusBadRequest, response.Status, response.Body)
	assert.Contains(t, response.errorEnvelope().Fields, keyFor(emma.ID, "age"))

	stored := app.get("/api/rsvp").rsvp()
	assert.Nil(t, stored.Household.RSVPSubmittedAt)
	assert.Zero(t, stored.Household.TransportSeatsNeeded)
	assert.False(t, stored.Household.HasStroller)
	assert.Empty(t, stored.Household.RSVPNote)
	assert.Nil(t, stored.memberByID(t, anna.ID).Attending)
}

// F3-B01's transport rule, through the API: the church → reception trip only exists
// for a guest attending both, and a stray seat count would inflate the shuttle gap.
func TestRSVPZeroesTransportSeatsWithoutAMemberAttendingBoth(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	response := app.putJSON("/api/rsvp", map[string]any{
		"transport_seats_needed":  2,
		"transport_seats_offered": 2,
		"has_stroller":            false,
		"rsvp_note":               "",
		"members":                 []map[string]any{answerFor(anna.ID, "church_only")},
	})
	require.Equal(t, http.StatusOK, response.Status, response.Body)

	body := response.rsvp()
	assert.Zero(t, body.Household.TransportSeatsNeeded)
	assert.Zero(t, body.Household.TransportSeatsOffered)
}

func TestRSVPRejectsSeatCountsAboveTheBound(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))

	response := app.putJSON("/api/rsvp", map[string]any{
		"transport_seats_needed":  21,
		"transport_seats_offered": 0,
		"has_stroller":            false,
		"rsvp_note":               "",
		"members":                 []map[string]any{answerFor(household.Guests[0].ID, "both")},
	})

	require.Equal(t, http.StatusBadRequest, response.Status)
	assert.Contains(t, response.errorEnvelope().Fields, "transport_seats_needed")
}

func TestRSVPResponseLeaksNothing(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write,
		withCode("ABC234"), withAdminNote("Ruft nie zurück"), withAdult("Anna Müller"))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	// The seen marker is ours, and a household that saw an unread flag would
	// reasonably start chasing us.
	_, err := app.Database.Write.Exec(
		`UPDATE household SET rsvp_note_seen_at = '2026-09-01T10:00:00Z' WHERE id = ?`, household.ID)
	require.NoError(t, err)

	read := app.get("/api/rsvp")
	read.assertNoLeak(household.Code, "Ruft nie zurück")
	assert.NotContains(t, read.Body, "rsvp_note_seen_at")

	saved := app.putJSON("/api/rsvp", submission(answerFor(household.Guests[0].ID, "both")))
	saved.assertNoLeak(household.Code, "Ruft nie zurück")
	assert.NotContains(t, saved.Body, "rsvp_note_seen_at")
}

// keyFor is the per-member field error key the API promises and F3-F04 renders from.
func keyFor(memberID int64, field string) string {
	return fmt.Sprintf("members.%d.%s", memberID, field)
}
