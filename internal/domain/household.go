package domain

import "time"

// Household is the unit of authentication and of RSVP. One printed code, one
// session, one set of answers.
type Household struct {
	ID          int64
	DisplayName string
	// Code is the printed login code in stored form. It is the only secret this
	// application has, kept in plaintext because a household that loses its card
	// has to be told the code again.
	//
	// It must never appear in a guest-facing response. That is not a matter of care
	// in this struct but of the DTO rule: nothing here is ever serialized, and
	// web/dto builds bodies field by field. The admin DTOs do carry it — that is
	// the one place in the API that may, and it sits behind RequireAdmin.
	Code string
	// TransportSeatsNeeded and TransportSeatsOffered are church → reception only,
	// and feed the shuttle capacity gap and nothing else. Ours to record as well as
	// the household's to answer: this is the kind of thing somebody says on the
	// phone.
	TransportSeatsNeeded  int
	TransportSeatsOffered int
	// HasStroller belongs to the household rather than to a child: a pram needs
	// floor space rather than a seat, and nobody brings two.
	HasStroller bool
	// AdminNote is our private note about the household, written on the assumption
	// that they will never read it. It must never reach a guest response.
	AdminNote string
	// RSVPSubmittedAt is NULL until the household answers, which is what puts them
	// on the nudge list. F3 owns writing it; F5 only shows it.
	RSVPSubmittedAt *time.Time
	// LastLoginAt is NULL until the household first redeems its code, which is what
	// answers "did they even see the invitation?" and drives the nudge list.
	LastLoginAt *time.Time
}

// HouseholdOverview is one row of the admin household list: the household plus the
// number of people in it.
//
// The count is carried here rather than on Household because it is not a property
// of the household — it is the result of a query about its members, and a
// Household loaded by any other path would have to leave it at zero and lie.
type HouseholdOverview struct {
	Household
	MemberCount int
}

// HouseholdPatch is a partial update: a nil field is one the request did not send
// and must be left alone.
//
// Pointers rather than plain values throughout, so that "not sent" and "sent as
// empty" stay distinguishable — with a plain string the two are the same value, and
// clearing an admin note would become impossible.
//
// Code is deliberately absent. Reissuing a code revokes sessions and kills a
// printed card, so it is its own endpoint; a code changed by a stray field in a
// form body is a code nobody knows changed.
type HouseholdPatch struct {
	DisplayName           *string
	AdminNote             *string
	TransportSeatsNeeded  *int
	TransportSeatsOffered *int
	HasStroller           *bool
}

// ApplyHouseholdPatch returns the household as the patch leaves it, plus the
// changed fields for the audit log.
func ApplyHouseholdPatch(current Household, patch HouseholdPatch) (Household, Changes) {
	updated := current

	if patch.DisplayName != nil {
		updated.DisplayName = *patch.DisplayName
	}
	if patch.AdminNote != nil {
		updated.AdminNote = *patch.AdminNote
	}
	if patch.TransportSeatsNeeded != nil {
		updated.TransportSeatsNeeded = *patch.TransportSeatsNeeded
	}
	if patch.TransportSeatsOffered != nil {
		updated.TransportSeatsOffered = *patch.TransportSeatsOffered
	}
	if patch.HasStroller != nil {
		updated.HasStroller = *patch.HasStroller
	}

	var changes Changes
	changes.compare("display_name", current.DisplayName, updated.DisplayName)
	changes.compare("admin_note", current.AdminNote, updated.AdminNote)
	changes.compare("transport_seats_needed", current.TransportSeatsNeeded, updated.TransportSeatsNeeded)
	changes.compare("transport_seats_offered", current.TransportSeatsOffered, updated.TransportSeatsOffered)
	changes.compare("has_stroller", current.HasStroller, updated.HasStroller)

	return updated, changes
}
