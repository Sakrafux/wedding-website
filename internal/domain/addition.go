package domain

// The plus-one rule: what a household may add to itself, and what it may take back
// off again. Pure functions over the household's living members, so the endpoint
// (F4-B02), the `can_add_plus_one` flag the form branches on and any later reader all
// ask the same question of the same code.
//
// The rule was tightened on 2026-08-31 from a numeric cap of two, and the looser
// version is the obvious thing for a later reader to restore — so, the reason: the
// single genuine unknown at the moment we address the envelopes is whether a guest
// invited alone is bringing somebody. Every other addition — a child, a second
// companion, a friend of a couple — lands on the caterer's count, the seating plan and
// the budget, and we would rather be told on the phone than discover it. The cost is
// accepted and real: households have almost no control over their own member list.

// ErrPlusOneNotAllowed reports that this household may not add anybody.
//
// One sentinel for all three reasons — the household is too big, it has already added
// somebody, the addition is not an adult — because the answer to the guest is the same
// sentence in every case (ring us) and the distinction would tell a stranger the shape
// of a household they cannot see.
var ErrPlusOneNotAllowed = NewError(CodePlusOneNotAllowed)

// ErrNotGuestAdded reports an attempt to remove somebody we seeded.
var ErrNotGuestAdded = NewError(CodeCannotRemoveMember)

// CanHouseholdAddPlusOne reports whether this household may add one, given its living
// members.
//
// Exactly one living member, and nothing else. The limit is structural rather than a
// counter: after the addition the household has two members, so the same check refuses
// the next one without anything having to record that an addition happened. A
// soft-deleted member does not appear in members, which is why a household that
// removes its own plus-one may add another — the one path back from a mistyped name.
func CanHouseholdAddPlusOne(members []Guest) error {
	if len(members) != 1 {
		return ErrPlusOneNotAllowed
	}
	return nil
}

// CanHouseholdRemove reports whether a household may remove this member itself.
//
// Only what it added. A member we seeded is our record of whom we invited and has to
// survive the answer: a household saying they are not coming answers `attending = 'no'`
// for them, which keeps the person on the list and the headcount explainable.
func CanHouseholdRemove(member Guest) error {
	if member.Origin != GuestOriginGuestAdded {
		return ErrNotGuestAdded
	}
	return nil
}

// NewPlusOne returns the guest row for a companion a household adds, given nothing but
// a name.
//
// An adult by construction: the guest path takes no kind and no age at all, because a
// child a household wants added is a phone call — which is also how we learn the age
// the caterer bills against. Attending stays nil rather than inheriting the
// household's scope: an added person is a new question, and a defaulted `both` would be
// an answer nobody gave.
func NewPlusOne(householdID int64, name string) Guest {
	return Guest{
		HouseholdID: householdID,
		Name:        name,
		Kind:        GuestKindAdult,
		Origin:      GuestOriginGuestAdded,
		SeatingNeed: SeatingNeedNormal,
		Portion:     PortionFull,
	}
}
