package domain

import "errors"

// Attending carries attendance *and* scope in one value, so that "declined but
// coming to the party" is unrepresentable rather than merely invalid.
//
// A nil *Attending on a Guest means the question has not been answered, and that is
// a real state: it is what the nudge list is built from. There is deliberately no
// `unanswered` member — the column is nullable, and a fifth value would be a second
// way to say the same thing, which is one way for two readers to disagree.
type Attending string

const (
	AttendingNo         Attending = "no"
	AttendingChurchOnly Attending = "church_only"
	AttendingPartyOnly  Attending = "party_only"
	AttendingBoth       Attending = "both"
)

// MealChoice is what a guest eats at the reception. Relevant only for a guest whose
// scope covers the party — see NormalizeGuestAnswer.
type MealChoice string

const (
	MealChoiceAll        MealChoice = "all"
	MealChoiceVegetarian MealChoice = "vegetarian"
	MealChoiceVegan      MealChoice = "vegan"
)

// Portion is how much food a guest is served, orthogonal to MealChoice: a vegan
// child eats a kids' portion of the vegan dish.
//
// PortionNone covers an infant and an adult who is not eating at all, which is why
// it is not the same question as "is this person attending".
type Portion string

const (
	PortionNone Portion = "none"
	PortionKids Portion = "kids"
	PortionFull Portion = "full"
)

// AttendsChurch reports whether this scope includes the ceremony.
func (attending Attending) AttendsChurch() bool {
	return attending == AttendingChurchOnly || attending == AttendingBoth
}

// AttendsParty reports whether this scope includes the reception, and is therefore
// what gates every catering field and all party seating.
func (attending Attending) AttendsParty() bool {
	return attending == AttendingPartyOnly || attending == AttendingBoth
}

// Attends reports whether the guest is coming to any part of the day.
func (attending Attending) Attends() bool {
	return attending.AttendsChurch() || attending.AttendsParty()
}

// AttendsChurch, AttendsParty and Attends on Guest are the nil-safe form: an
// unanswered guest attends nothing yet.
//
// Every later epic reads these rather than comparing attending to a string — F6's
// counts, F7's assignment validity, F8's per-head resolution. A switch on the raw
// value in a query file is the thing they exist to prevent, because that switch is
// where "is attending" quietly replaces "is at the party" and we pay for meals
// nobody eats.
func (guest Guest) AttendsChurch() bool {
	return guest.Attending != nil && guest.Attending.AttendsChurch()
}

func (guest Guest) AttendsParty() bool {
	return guest.Attending != nil && guest.Attending.AttendsParty()
}

func (guest Guest) Attends() bool {
	return guest.Attending != nil && guest.Attending.Attends()
}

// NormalizeGuestAnswer returns the answer as it will be stored: the scope gate, in
// one place.
//
// A guest whose scope does not cover the party is not asked about food, so the
// catering fields are reset to their schema defaults — no meal choice, PortionFull,
// no midnight snack. Reset rather than preserved, because a leftover meal choice is
// a value every later reader has to remember to ignore and the one that forgets pays
// for a plate. Reset to the defaults rather than to PortionNone, because `none`
// reads as an answer the household never gave; the neutral default is the honest
// record of "not asked". The scope is the single source of truth and the stored row
// agrees with it.
//
// SeatingNeed and DietaryNote are deliberately **not** gated. A wheelchair space is
// needed in the church pew as much as at the table, and an allergy is worth knowing
// wherever somebody eats — their absence from the reset above is the rule, not an
// oversight.
//
// Idempotent: normalizing an already-normalized answer changes nothing.
func NormalizeGuestAnswer(guest Guest) Guest {
	if guest.AttendsParty() {
		return guest
	}

	guest.MealChoice = nil
	guest.Portion = PortionFull
	guest.MidnightSnack = false
	return guest
}

// NormalizeHouseholdAnswer returns the household's answer as it will be stored, given
// the answers of its members.
//
// The transport seat counts are church → reception, so they only mean anything for a
// household with at least one member attending both. A household whose members all
// attend one half and which still carries a seat count would inflate the shuttle
// capacity gap — the one number those two fields exist to produce. The form hides
// them in that case (F3-F03); this is why it does not matter if it does not.
func NormalizeHouseholdAnswer(household Household, members []Guest) Household {
	for _, member := range members {
		if member.Attending != nil && *member.Attending == AttendingBoth {
			return household
		}
	}

	household.TransportSeatsNeeded = 0
	household.TransportSeatsOffered = 0
	return household
}

// GuestAnswer is the RSVP answer for one guest, as a submission carries it.
//
// Age is here because a child's age is editable through the RSVP; Kind is not,
// because a household turning an adult into a child changes a caterer bracket by
// pressing a radio button, and the case that would serve — we typed the wrong thing
// — is ours to fix in F5-F02.
type GuestAnswer struct {
	Attending     Attending
	MealChoice    *MealChoice
	Portion       Portion
	MidnightSnack bool
	SeatingNeed   SeatingNeed
	DietaryNote   string
	Age           *int
}

// ApplyGuestAnswer returns the guest with the answer stored on them, normalized, plus
// the changed fields for the audit log.
//
// The age rules are F5-B02's (ErrAgeOnAdult, ErrAgeOutOfRange) and the seating-need
// rule is F3-B08's (ErrSeatingNeedOnAdult), both against the guest's stored kind,
// which the answer cannot change.
func ApplyGuestAnswer(current Guest, answer GuestAnswer) (Guest, Changes, error) {
	updated := current
	updated.Attending = &answer.Attending
	updated.MealChoice = answer.MealChoice
	updated.Portion = answer.Portion
	updated.MidnightSnack = answer.MidnightSnack
	updated.SeatingNeed = answer.SeatingNeed
	updated.DietaryNote = answer.DietaryNote

	age, err := ResolveAge(current.Kind, answer.Age)
	if err != nil {
		return Guest{}, Changes{}, err
	}
	updated.Age = age

	seatingNeed, err := ResolveSeatingNeed(current.Kind, answer.SeatingNeed)
	if err != nil {
		return Guest{}, Changes{}, err
	}
	updated.SeatingNeed = seatingNeed

	updated = NormalizeGuestAnswer(updated)

	var changes Changes
	compareOptional(&changes, "attending", current.Attending, updated.Attending)
	compareOptional(&changes, "meal_choice", current.MealChoice, updated.MealChoice)
	changes.compare("portion", current.Portion, updated.Portion)
	changes.compare("midnight_snack", current.MidnightSnack, updated.MidnightSnack)
	changes.compare("seating_need", current.SeatingNeed, updated.SeatingNeed)
	changes.compare("dietary_note", current.DietaryNote, updated.DietaryNote)
	changes.compareOptionalInt("age", current.Age, updated.Age)

	return updated, changes, nil
}

// ErrTransportSeatsConflict reports a household that both needs seats and offers them.
//
// A sentinel rather than a domain.Error, because the answer belongs next to the two
// fields that produced it — see httpio.GuestAnswerValidationError.
var ErrTransportSeatsConflict = errors.New("a household needs seats or offers them, never both")

// ValidateTransportSeats refuses a transport answer that points both ways.
//
// The pair of counts feeds one subtraction — the shuttle capacity gap in F6 — and a
// household on both sides of it inflates the demand and the supply at once. Refused
// rather than normalized: the body said two things that cannot both be true, and
// picking one of them for the household would be answering for them.
//
// There is no direction field, and deliberately so: the pair of counts *is* the
// direction, and a third column would be a second way to say the same thing.
func ValidateTransportSeats(needed, offered int) error {
	if needed > 0 && offered > 0 {
		return ErrTransportSeatsConflict
	}
	return nil
}

// HouseholdAnswer is the household's own half of an RSVP submission: the fields that
// belong to the group rather than to a person.
type HouseholdAnswer struct {
	TransportSeatsNeeded  int
	TransportSeatsOffered int
	HasStroller           bool
	RSVPNote              string
}

// ApplyHouseholdAnswer returns the household with the answer stored on it, normalized
// against the members' scopes, plus the changed fields for the audit log.
//
// members are the answers as they will be stored — normalize the guests first, or the
// transport rule reads a scope that is about to change.
func ApplyHouseholdAnswer(current Household, answer HouseholdAnswer, members []Guest) (Household, Changes) {
	updated := current
	updated.TransportSeatsNeeded = answer.TransportSeatsNeeded
	updated.TransportSeatsOffered = answer.TransportSeatsOffered
	updated.HasStroller = answer.HasStroller
	updated.RSVPNote = answer.RSVPNote

	updated = NormalizeHouseholdAnswer(updated, members)

	var changes Changes
	changes.compare("transport_seats_needed", current.TransportSeatsNeeded, updated.TransportSeatsNeeded)
	changes.compare("transport_seats_offered", current.TransportSeatsOffered, updated.TransportSeatsOffered)
	changes.compare("has_stroller", current.HasStroller, updated.HasStroller)
	// The note is recorded in the audit payload on purpose: it is the household's own
	// words to us, the log is admin-only, and a note edited to remove a request is
	// exactly the case the log exists for (F3-B05).
	changes.compare("rsvp_note", current.RSVPNote, updated.RSVPNote)

	return updated, changes
}
