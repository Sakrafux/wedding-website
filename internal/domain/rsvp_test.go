package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

func attending(value domain.Attending) *domain.Attending {
	return &value
}

func mealChoice(value domain.MealChoice) *domain.MealChoice {
	return &value
}

// answeredGuest is a party guest with every catering field set, so that a test about
// the scope gate can say which of them survived.
func answeredGuest(scope domain.Attending) domain.Guest {
	return domain.Guest{
		ID:            31,
		Name:          "Anna Müller",
		Kind:          domain.GuestKindAdult,
		Attending:     attending(scope),
		MealChoice:    mealChoice(domain.MealChoiceVegetarian),
		Portion:       domain.PortionKids,
		MidnightSnack: true,
		SeatingNeed:   domain.SeatingNeedWheelchair,
		DietaryNote:   "Nussallergie",
	}
}

// The scope gate: nobody who is not at the party carries a catering answer, because
// every derived count keys off the scope and the stored row has to agree with it.
func TestNormalizeGuestAnswerClearsCateringOutsideTheParty(t *testing.T) {
	t.Parallel()

	for _, scope := range []domain.Attending{domain.AttendingNo, domain.AttendingChurchOnly} {
		normalized := domain.NormalizeGuestAnswer(answeredGuest(scope))

		assert.Nilf(t, normalized.MealChoice, "meal choice at scope %q", scope)
		// The schema default, not `none`: `none` reads as an answer the household
		// never gave.
		assert.Equalf(t, domain.PortionFull, normalized.Portion, "portion at scope %q", scope)
		assert.Falsef(t, normalized.MidnightSnack, "midnight snack at scope %q", scope)
	}
}

func TestNormalizeGuestAnswerKeepsCateringForPartyGuests(t *testing.T) {
	t.Parallel()

	for _, scope := range []domain.Attending{domain.AttendingPartyOnly, domain.AttendingBoth} {
		guest := answeredGuest(scope)

		assert.Equalf(t, guest, domain.NormalizeGuestAnswer(guest), "scope %q", scope)
	}
}

// An unanswered guest is not at the party either, and must not carry catering from a
// half-filled form.
func TestNormalizeGuestAnswerClearsCateringForAnUnansweredGuest(t *testing.T) {
	t.Parallel()

	guest := answeredGuest(domain.AttendingBoth)
	guest.Attending = nil

	normalized := domain.NormalizeGuestAnswer(guest)

	assert.Nil(t, normalized.MealChoice)
	assert.Equal(t, domain.PortionFull, normalized.Portion)
	assert.False(t, normalized.MidnightSnack)
}

func TestNormalizeGuestAnswerIsIdempotent(t *testing.T) {
	t.Parallel()

	for _, scope := range []domain.Attending{
		domain.AttendingNo, domain.AttendingChurchOnly, domain.AttendingPartyOnly, domain.AttendingBoth,
	} {
		once := domain.NormalizeGuestAnswer(answeredGuest(scope))
		twice := domain.NormalizeGuestAnswer(once)

		assert.Equalf(t, once, twice, "scope %q", scope)
	}
}

// The one place the scope gate deliberately does not apply: a wheelchair space is
// needed in the pew, and an allergy matters wherever somebody eats.
func TestNormalizeGuestAnswerKeepsSeatingNeedAndDietaryNoteAtEveryScope(t *testing.T) {
	t.Parallel()

	for _, scope := range []domain.Attending{
		domain.AttendingNo, domain.AttendingChurchOnly, domain.AttendingPartyOnly, domain.AttendingBoth,
	} {
		normalized := domain.NormalizeGuestAnswer(answeredGuest(scope))

		assert.Equalf(t, domain.SeatingNeedWheelchair, normalized.SeatingNeed, "scope %q", scope)
		assert.Equalf(t, "Nussallergie", normalized.DietaryNote, "scope %q", scope)
	}
}

func householdWithSeats() domain.Household {
	return domain.Household{ID: 12, TransportSeatsNeeded: 2, TransportSeatsOffered: 3}
}

func memberAt(scope domain.Attending) domain.Guest {
	return domain.Guest{Attending: attending(scope)}
}

// The seat counts are the church → reception trip, which only a guest attending both
// makes. Keeping them otherwise would inflate the shuttle gap they exist to produce.
func TestNormalizeHouseholdAnswerZeroesSeatsWithoutAMemberAttendingBoth(t *testing.T) {
	t.Parallel()

	members := []domain.Guest{
		memberAt(domain.AttendingChurchOnly),
		memberAt(domain.AttendingPartyOnly),
		memberAt(domain.AttendingNo),
		{},
	}

	normalized := domain.NormalizeHouseholdAnswer(householdWithSeats(), members)

	assert.Zero(t, normalized.TransportSeatsNeeded)
	assert.Zero(t, normalized.TransportSeatsOffered)
}

func TestNormalizeHouseholdAnswerKeepsSeatsWhenAMemberAttendsBoth(t *testing.T) {
	t.Parallel()

	members := []domain.Guest{memberAt(domain.AttendingChurchOnly), memberAt(domain.AttendingBoth)}

	normalized := domain.NormalizeHouseholdAnswer(householdWithSeats(), members)

	assert.Equal(t, 2, normalized.TransportSeatsNeeded)
	assert.Equal(t, 3, normalized.TransportSeatsOffered)
}

// Exhaustive over the four values, so a fifth one added later fails here first —
// before it fails in a query that quietly counts the wrong people.
func TestScopePredicatesAgreeWithEveryScope(t *testing.T) {
	t.Parallel()

	expectations := map[domain.Attending]struct{ church, party, attends bool }{
		domain.AttendingNo:         {false, false, false},
		domain.AttendingChurchOnly: {true, false, true},
		domain.AttendingPartyOnly:  {false, true, true},
		domain.AttendingBoth:       {true, true, true},
	}

	for scope, expected := range expectations {
		assert.Equalf(t, expected.church, scope.AttendsChurch(), "AttendsChurch at %q", scope)
		assert.Equalf(t, expected.party, scope.AttendsParty(), "AttendsParty at %q", scope)
		assert.Equalf(t, expected.attends, scope.Attends(), "Attends at %q", scope)

		guest := domain.Guest{Attending: attending(scope)}
		assert.Equalf(t, expected.church, guest.AttendsChurch(), "Guest.AttendsChurch at %q", scope)
		assert.Equalf(t, expected.party, guest.AttendsParty(), "Guest.AttendsParty at %q", scope)
		assert.Equalf(t, expected.attends, guest.Attends(), "Guest.Attends at %q", scope)
	}
}

func TestScopePredicatesReportNothingForAnUnansweredGuest(t *testing.T) {
	t.Parallel()

	var guest domain.Guest

	assert.False(t, guest.AttendsChurch())
	assert.False(t, guest.AttendsParty())
	assert.False(t, guest.Attends())
}

func TestApplyGuestAnswerReportsOnlyTheChangedFields(t *testing.T) {
	t.Parallel()

	current := answeredGuest(domain.AttendingBoth)

	updated, changes, err := domain.ApplyGuestAnswer(current, domain.GuestAnswer{
		Attending:     domain.AttendingBoth,
		MealChoice:    mealChoice(domain.MealChoiceVegan),
		Portion:       current.Portion,
		MidnightSnack: current.MidnightSnack,
		SeatingNeed:   current.SeatingNeed,
		DietaryNote:   current.DietaryNote,
	})

	require.NoError(t, err)
	require.NotNil(t, updated.MealChoice)
	assert.Equal(t, domain.MealChoiceVegan, *updated.MealChoice)
	// A nullable enum is unwrapped to a plain string so that an absent value can be
	// nil; a non-nullable one keeps its type, as every other diff in domain does. The
	// JSON that reaches audit_log is the same either way.
	assert.Equal(t, map[string]any{"meal_choice": "vegetarian"}, changes.Before)
	assert.Equal(t, map[string]any{"meal_choice": "vegan"}, changes.After)
}

// An answer that changes nothing must produce no audit row at all — see F3-B03 and
// Households.Update, which made the same trade for the same reason.
func TestApplyGuestAnswerReportsNoChangeForAnIdenticalAnswer(t *testing.T) {
	t.Parallel()

	current := answeredGuest(domain.AttendingPartyOnly)

	_, changes, err := domain.ApplyGuestAnswer(current, domain.GuestAnswer{
		Attending:     *current.Attending,
		MealChoice:    current.MealChoice,
		Portion:       current.Portion,
		MidnightSnack: current.MidnightSnack,
		SeatingNeed:   current.SeatingNeed,
		DietaryNote:   current.DietaryNote,
	})

	require.NoError(t, err)
	assert.True(t, changes.IsEmpty())
}

// The catering an answer sends for a church-only guest is not an error — the form may
// still have it in state when somebody flips the scope last — it is simply not stored.
func TestApplyGuestAnswerNormalizesBeforeDiffing(t *testing.T) {
	t.Parallel()

	current := answeredGuest(domain.AttendingBoth)

	updated, changes, err := domain.ApplyGuestAnswer(current, domain.GuestAnswer{
		Attending:     domain.AttendingChurchOnly,
		MealChoice:    mealChoice(domain.MealChoiceVegan),
		Portion:       domain.PortionKids,
		MidnightSnack: true,
		SeatingNeed:   current.SeatingNeed,
		DietaryNote:   current.DietaryNote,
	})

	require.NoError(t, err)
	assert.Nil(t, updated.MealChoice)
	assert.Equal(t, domain.PortionFull, updated.Portion)
	assert.False(t, updated.MidnightSnack)
	assert.Equal(t, map[string]any{
		"attending":      "both",
		"meal_choice":    "vegetarian",
		"portion":        domain.PortionKids,
		"midnight_snack": true,
	}, changes.Before)
	assert.Equal(t, map[string]any{
		"attending":      "church_only",
		"meal_choice":    nil,
		"portion":        domain.PortionFull,
		"midnight_snack": false,
	}, changes.After)
}

// The F5-B02 age rules apply unchanged through the RSVP, against the stored kind: an
// answer cannot turn an adult into a child.
func TestApplyGuestAnswerReportsTheAgeSentinels(t *testing.T) {
	t.Parallel()

	adult := domain.Guest{Kind: domain.GuestKindAdult}
	_, _, err := domain.ApplyGuestAnswer(adult, domain.GuestAnswer{Attending: domain.AttendingBoth, Age: age(30)})
	assert.ErrorIs(t, err, domain.ErrAgeOnAdult)

	child := domain.Guest{Kind: domain.GuestKindChild}
	_, _, err = domain.ApplyGuestAnswer(child, domain.GuestAnswer{Attending: domain.AttendingBoth, Age: age(18)})
	assert.ErrorIs(t, err, domain.ErrAgeOutOfRange)
}

func TestApplyHouseholdAnswerZeroesSeatsAndReportsThem(t *testing.T) {
	t.Parallel()

	current := domain.Household{ID: 12, TransportSeatsNeeded: 2}

	updated, changes := domain.ApplyHouseholdAnswer(current, domain.HouseholdAnswer{
		TransportSeatsNeeded:  4,
		TransportSeatsOffered: 1,
		RSVPNote:              "Wir kommen erst nach der Zeremonie.",
	}, []domain.Guest{memberAt(domain.AttendingPartyOnly)})

	assert.Zero(t, updated.TransportSeatsNeeded)
	assert.Zero(t, updated.TransportSeatsOffered)
	assert.Equal(t, map[string]any{"transport_seats_needed": 2, "rsvp_note": ""}, changes.Before)
	assert.Equal(t, map[string]any{
		"transport_seats_needed": 0,
		"rsvp_note":              "Wir kommen erst nach der Zeremonie.",
	}, changes.After)
	assert.Equal(t, "Wir kommen erst nach der Zeremonie.", updated.RSVPNote)
}

// The note is in the payload deliberately: it is the household's own words, the log
// is admin-only, and a note edited to remove a request is why the log exists.
func TestApplyHouseholdAnswerRecordsTheNote(t *testing.T) {
	t.Parallel()

	current := domain.Household{ID: 12, RSVPNote: "Oma braucht einen Platz nah am Ausgang."}

	_, changes := domain.ApplyHouseholdAnswer(current, domain.HouseholdAnswer{RSVPNote: ""},
		[]domain.Guest{memberAt(domain.AttendingBoth)})

	assert.Equal(t, map[string]any{"rsvp_note": "Oma braucht einen Platz nah am Ausgang."}, changes.Before)
	assert.Equal(t, map[string]any{"rsvp_note": ""}, changes.After)
}

// The pair of counts feeds one subtraction, so a household on both sides of it is
// refused rather than normalized into whichever side we picked (F3-B07).
func TestValidateTransportSeatsRefusesBothDirectionsAtOnce(t *testing.T) {
	t.Parallel()

	require.ErrorIs(t, domain.ValidateTransportSeats(2, 3), domain.ErrTransportSeatsConflict)

	require.NoError(t, domain.ValidateTransportSeats(0, 0))
	require.NoError(t, domain.ValidateTransportSeats(2, 0))
	require.NoError(t, domain.ValidateTransportSeats(0, 3))
}

// The seating-need rule reaches the RSVP path too: the form hides the options
// (F3-F08), and this is why it does not matter if it does not.
func TestApplyGuestAnswerRejectsAChildOnlySeatingNeedOnAnAdult(t *testing.T) {
	t.Parallel()

	adult := domain.Guest{ID: 7, Kind: domain.GuestKindAdult, SeatingNeed: domain.SeatingNeedNormal}
	_, _, err := domain.ApplyGuestAnswer(adult, domain.GuestAnswer{
		Attending:   domain.AttendingBoth,
		Portion:     domain.PortionFull,
		SeatingNeed: domain.SeatingNeedHighChair,
	})

	require.ErrorIs(t, err, domain.ErrSeatingNeedOnAdult)
}
