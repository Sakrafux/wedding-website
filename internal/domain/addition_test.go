package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

func member(origin domain.GuestOrigin) domain.Guest {
	return domain.Guest{ID: 1, HouseholdID: 7, Name: "Paddi", Kind: domain.GuestKindAdult, Origin: origin}
}

func TestAHouseholdOfOneMayAddAPlusOne(t *testing.T) {
	t.Parallel()

	err := domain.CanHouseholdAddPlusOne([]domain.Guest{member(domain.GuestOriginSeeded)})

	require.NoError(t, err)
}

// Both the household seeded as two and the household that has already used its
// plus-one are the same case here, which is the point of a structural limit: nothing
// counts additions, the member list is the counter.
func TestAHouseholdOfTwoMayNotAddAPlusOne(t *testing.T) {
	t.Parallel()

	for name, members := range map[string][]domain.Guest{
		"seeded as two": {member(domain.GuestOriginSeeded), member(domain.GuestOriginSeeded)},
		"already added": {member(domain.GuestOriginSeeded), member(domain.GuestOriginGuestAdded)},
	} {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, domain.CanHouseholdAddPlusOne(members), domain.ErrPlusOneNotAllowed)
		})
	}
}

// A household with nobody in it is a real state (F5 allows it) and is not a household
// of one — the plus-one exists for a person we invited alone, and there is no such
// person here.
func TestAnEmptyHouseholdMayNotAddAPlusOne(t *testing.T) {
	t.Parallel()

	assert.ErrorIs(t, domain.CanHouseholdAddPlusOne(nil), domain.ErrPlusOneNotAllowed)
}

func TestAHouseholdRemovesOnlyWhatItAdded(t *testing.T) {
	t.Parallel()

	assert.ErrorIs(t, domain.CanHouseholdRemove(member(domain.GuestOriginSeeded)), domain.ErrNotGuestAdded)
	assert.NoError(t, domain.CanHouseholdRemove(member(domain.GuestOriginGuestAdded)))
}

// Removing somebody who has already answered is a real correction, not a mistake to
// be guarded against: the guest is not coming after all, and they were ours to remove.
func TestAnAnsweredGuestAddedMemberMayStillBeRemoved(t *testing.T) {
	t.Parallel()

	answered := member(domain.GuestOriginGuestAdded)
	attending := domain.AttendingBoth
	answered.Attending = &attending

	assert.NoError(t, domain.CanHouseholdRemove(answered))
}

func TestANewPlusOneIsAnUnansweredAdult(t *testing.T) {
	t.Parallel()

	plusOne := domain.NewPlusOne(7, "Isabella Michelbacher")

	assert.Equal(t, int64(7), plusOne.HouseholdID)
	assert.Equal(t, "Isabella Michelbacher", plusOne.Name)
	assert.Equal(t, domain.GuestKindAdult, plusOne.Kind)
	assert.Equal(t, domain.GuestOriginGuestAdded, plusOne.Origin)
	assert.Equal(t, domain.SeatingNeedNormal, plusOne.SeatingNeed)
	assert.Equal(t, domain.PortionFull, plusOne.Portion)
	assert.Nil(t, plusOne.Age)
	// Not pre-filled with the household's scope: an added person is a new question.
	assert.Nil(t, plusOne.Attending)
}
