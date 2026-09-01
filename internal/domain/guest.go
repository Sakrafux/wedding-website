package domain

import (
	"errors"
	"slices"
)

// GuestKind separates adults from children. It is not derived from age: age is
// recorded for children only, and a household that skips it must still produce a
// correct adult headcount for the caterer.
type GuestKind string

const (
	GuestKindAdult GuestKind = "adult"
	GuestKindChild GuestKind = "child"
)

// GuestOrigin records who put this person on the list. A household may add exactly one
// adult plus-one to itself (F4-B01) and nobody else — everything further goes through
// us — and the admin views show what was added separately from what we seeded.
type GuestOrigin string

const (
	GuestOriginSeeded     GuestOrigin = "seeded"
	GuestOriginGuestAdded GuestOrigin = "guest_added"
)

// SeatingNeed is what a guest needs at the seat itself, adults included.
//
// SeatingNeedWithParent means the guest consumes no seat of their own and must not
// hold a seat assignment (F7). There is no value for a pram: it is parked rather
// than sat on and belongs to the household — see Household.HasStroller.
type SeatingNeed string

const (
	SeatingNeedNormal     SeatingNeed = "normal"
	SeatingNeedWithParent SeatingNeed = "with_parent"
	SeatingNeedHighChair  SeatingNeed = "high_chair"
	SeatingNeedWheelchair SeatingNeed = "wheelchair"
)

// childOnlySeatingNeeds are the needs that only make sense for a child.
//
// with_parent means the person consumes no seat of their own, and high_chair means a
// chair we hire per child; both feed counts we buy things with. An adult who needs
// something at the seat is what wheelchair and the household note are for.
var childOnlySeatingNeeds = []SeatingNeed{SeatingNeedWithParent, SeatingNeedHighChair}

// maxChildAge is the highest age still recorded as a child's. Eighteen is an
// adult, and the caterer's brackets are all drawn below that — the exact
// boundaries are derived at read time from age, so this number is about who counts
// as a child at all, not about pricing.
const maxChildAge = 17

// Guest is one person, belonging to exactly one household.
type Guest struct {
	ID          int64
	HouseholdID int64
	// Name is the whole name in one field, as the guest writes it. Deliberately not
	// split into first and last: what every output needs is the full name, and one
	// field lets a household enter "Oma Erika" or a double first name without
	// deciding which half is which. See migration 0002 for the accepted cost.
	Name string
	Kind GuestKind
	// Age is **age at the wedding date** and is set for children only. Asked that
	// way in the UI so the value does not drift over the months before the event.
	Age    *int
	Origin GuestOrigin
	// SeatingNeed and DietaryNote are ours to edit as well as the household's: both
	// are things somebody tells us on the phone in March as often as through the
	// RSVP form.
	SeatingNeed SeatingNeed
	DietaryNote string
	// Attending is the RSVP answer: attendance and scope in one value. A nil pointer
	// is "has not answered", which is a state the nudge list is built from — see
	// Attending, and read it through Attends/AttendsChurch/AttendsParty rather than
	// by comparing it.
	Attending *Attending
	// MealChoice, Portion and MidnightSnack are catering, and catering is gated by
	// scope rather than by attendance: they mean nothing for a guest who is not at
	// the party, and NormalizeGuestAnswer is what keeps the stored row saying so.
	MealChoice    *MealChoice
	Portion       Portion
	MidnightSnack bool
}

// ErrAgeOnAdult reports an age recorded against an adult.
//
// A sentinel rather than a domain.Error, because the caller has to be able to put
// the message next to the `age` field of the form — and field names are a shape the
// domain does not know. The schema has the same rule as a CHECK constraint; this
// exists so the answer is a field error and not a driver error.
var ErrAgeOnAdult = errors.New("age is recorded for children only")

// ErrAgeOutOfRange reports a child's age outside 0..maxChildAge.
var ErrAgeOutOfRange = errors.New("a child's age must be between 0 and 17")

// ErrSeatingNeedOnAdult reports a child-only seating need recorded against an adult.
//
// A sentinel for the same reason ErrAgeOnAdult is one: the answer has to land next to
// the seating-need field of a form, and field names are a shape the domain does not
// know.
var ErrSeatingNeedOnAdult = errors.New("with_parent and high_chair are recorded for children only")

// ResolveAge returns the age to store for a guest of this kind.
//
// An adult's age is refused rather than silently dropped: a stale age would feed a
// caterer age bracket for somebody who is not a child, and the request said
// something that is not true. Switching an existing guest from child to adult is
// the one case where an age *is* dropped, and that happens in ApplyGuestPatch,
// where the kind is what changed.
func ResolveAge(kind GuestKind, age *int) (*int, error) {
	if age == nil {
		return nil, nil
	}
	if kind != GuestKindChild {
		return nil, ErrAgeOnAdult
	}
	if *age < 0 || *age > maxChildAge {
		return nil, ErrAgeOutOfRange
	}
	return age, nil
}

// ResolveSeatingNeed returns the seating need to store for a guest of this kind.
//
// Refused rather than silently reset, exactly like ResolveAge: a request that asks
// for a high chair for an adult has said something that cannot be true, and quietly
// storing `normal` instead would answer a question nobody asked. The case this bites
// in practice is a child being turned into an adult while a high chair is still
// recorded — and the person who just made that change is the only one who can say
// what that guest now needs.
func ResolveSeatingNeed(kind GuestKind, need SeatingNeed) (SeatingNeed, error) {
	if kind == GuestKindChild || !slices.Contains(childOnlySeatingNeeds, need) {
		return need, nil
	}
	return "", ErrSeatingNeedOnAdult
}

// GuestPatch is a partial update of a guest: a nil field is one the request did not
// send and must be left alone.
//
// Age needs two fields rather than one because a JSON null and an absent key mean
// different things here — see AgeSet.
type GuestPatch struct {
	Name *string
	Kind *GuestKind
	// AgeSet says the request carried an `age` key at all; Age is its value, nil for
	// an explicit null. A single *int could not express "clear the age", since that
	// is exactly what a null decodes to.
	AgeSet      bool
	Age         *int
	SeatingNeed *SeatingNeed
	DietaryNote *string
}

// ApplyGuestPatch returns the guest as the patch leaves it, plus the changed fields
// for the audit log.
//
// Switching kind from child to adult clears the age in the same result. Leaving it
// behind would violate the schema's CHECK constraint and, worse than failing, would
// keep feeding a caterer bracket for somebody who is now an adult.
func ApplyGuestPatch(current Guest, patch GuestPatch) (Guest, Changes, error) {
	updated := current
	var changes Changes

	if patch.Name != nil {
		updated.Name = *patch.Name
	}
	if patch.Kind != nil {
		updated.Kind = *patch.Kind
	}
	if patch.AgeSet {
		updated.Age = patch.Age
	}
	if patch.SeatingNeed != nil {
		updated.SeatingNeed = *patch.SeatingNeed
	}
	if patch.DietaryNote != nil {
		updated.DietaryNote = *patch.DietaryNote
	}

	// After the kind is settled, so that "child with an age" → "adult" drops the age
	// instead of being reported as an age on an adult. An age sent *together with*
	// adult is still an error: the request asked for a state that cannot exist.
	if updated.Kind != GuestKindChild && !patch.AgeSet {
		updated.Age = nil
	}

	age, err := ResolveAge(updated.Kind, updated.Age)
	if err != nil {
		return Guest{}, Changes{}, err
	}
	updated.Age = age

	// Also after the kind is settled: turning a child with a high chair into an adult
	// is the case this refuses, and it can only be seen once both fields are known.
	seatingNeed, err := ResolveSeatingNeed(updated.Kind, updated.SeatingNeed)
	if err != nil {
		return Guest{}, Changes{}, err
	}
	updated.SeatingNeed = seatingNeed

	changes.compare("name", current.Name, updated.Name)
	changes.compare("kind", current.Kind, updated.Kind)
	changes.compareOptionalInt("age", current.Age, updated.Age)
	changes.compare("seating_need", current.SeatingNeed, updated.SeatingNeed)
	changes.compare("dietary_note", current.DietaryNote, updated.DietaryNote)

	return updated, changes, nil
}
