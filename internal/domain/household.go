package domain

import "time"

// GuestKind separates adults from children. It is not derived from age: age is
// recorded for children only, and a household that skips it must still produce a
// correct adult headcount for the caterer.
type GuestKind string

const (
	GuestKindAdult GuestKind = "adult"
	GuestKindChild GuestKind = "child"
)

// GuestOrigin records who put this person on the list. Households may add
// plus-ones and children themselves (F4), and the admin views show what was added
// separately from what we seeded.
type GuestOrigin string

const (
	GuestOriginSeeded     GuestOrigin = "seeded"
	GuestOriginGuestAdded GuestOrigin = "guest_added"
)

// Household is the unit of authentication and of RSVP. One printed code, one
// session, one set of answers.
type Household struct {
	ID          int64
	DisplayName string
	// Code is the printed login code in stored form. It is the only secret this
	// application has, kept in plaintext because a household that loses its card
	// has to be told the code again.
	//
	// It must never appear in a response. That is not a matter of care in this
	// struct but of the DTO rule: nothing here is ever serialized, and
	// web/dto builds guest-facing bodies field by field.
	Code string
	// LastLoginAt is NULL until the household first redeems its code, which is what
	// answers "did they even see the invitation?" and drives the nudge list.
	LastLoginAt *time.Time
}

// Guest is one person, belonging to exactly one household.
//
// F3-B01 adds the RSVP fields (attending, meal choice, portion, seating need); the
// login and bootstrap responses need only who somebody is.
type Guest struct {
	ID          int64
	HouseholdID int64
	FirstName   string
	LastName    string
	Kind        GuestKind
	Origin      GuestOrigin
}
