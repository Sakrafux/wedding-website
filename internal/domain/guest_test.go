package domain_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

func age(value int) *int {
	return &value
}

// An age on an adult is refused rather than dropped: the request said something that
// cannot be true, and a silently ignored field is an answer somebody gave and lost.
func TestResolveAgeRejectsAnAgeOnAnAdult(t *testing.T) {
	t.Parallel()

	_, err := domain.ResolveAge(domain.GuestKindAdult, age(30))

	require.ErrorIs(t, err, domain.ErrAgeOnAdult)
}

func TestResolveAgeAcceptsTheChildRange(t *testing.T) {
	t.Parallel()

	for _, value := range []int{0, 4, 17} {
		resolved, err := domain.ResolveAge(domain.GuestKindChild, age(value))

		require.NoError(t, err)
		require.NotNil(t, resolved)
		assert.Equal(t, value, *resolved)
	}
}

// Eighteen is an adult, and every caterer bracket is drawn below that.
func TestResolveAgeRejectsAnAgeOutsideTheChildRange(t *testing.T) {
	t.Parallel()

	for _, value := range []int{-1, 18, 120} {
		_, err := domain.ResolveAge(domain.GuestKindChild, age(value))

		assert.ErrorIsf(t, err, domain.ErrAgeOutOfRange, "age %d", value)
	}
}

func TestResolveAgeLeavesAnAbsentAgeAbsent(t *testing.T) {
	t.Parallel()

	for _, kind := range []domain.GuestKind{domain.GuestKindAdult, domain.GuestKindChild} {
		resolved, err := domain.ResolveAge(kind, nil)

		require.NoError(t, err)
		assert.Nil(t, resolved)
	}
}

func child() domain.Guest {
	return domain.Guest{
		ID:          31,
		HouseholdID: 12,
		Name:        "Emil Müller",
		Kind:        domain.GuestKindChild,
		Age:         age(4),
		Origin:      domain.GuestOriginSeeded,
		SeatingNeed: domain.SeatingNeedHighChair,
	}
}

// The case the caterer's bill depends on: a child who turns out to be an adult must
// not keep an age that still feeds an age bracket.
//
// The seating need travels with the kind, because a high chair cannot: see
// TestApplyGuestPatchRejectsAChildOnlySeatingNeedOnAnAdult.
func TestApplyGuestPatchClearsTheAgeWhenAChildBecomesAnAdult(t *testing.T) {
	t.Parallel()

	adult := domain.GuestKindAdult
	normal := domain.SeatingNeedNormal
	updated, changes, err := domain.ApplyGuestPatch(child(), domain.GuestPatch{Kind: &adult, SeatingNeed: &normal})

	require.NoError(t, err)
	assert.Equal(t, domain.GuestKindAdult, updated.Kind)
	assert.Nil(t, updated.Age)
	assert.Equal(t,
		map[string]any{"kind": domain.GuestKindChild, "age": 4, "seating_need": domain.SeatingNeedHighChair},
		changes.Before)
	assert.Equal(t,
		map[string]any{"kind": domain.GuestKindAdult, "age": nil, "seating_need": domain.SeatingNeedNormal},
		changes.After)
}

// A high chair on an adult is refused rather than reset: the admin who just changed
// the kind is the only one who can say what that guest now needs (F3-B08).
func TestApplyGuestPatchRejectsAChildOnlySeatingNeedOnAnAdult(t *testing.T) {
	t.Parallel()

	adult := domain.GuestKindAdult
	_, _, err := domain.ApplyGuestPatch(child(), domain.GuestPatch{Kind: &adult})

	require.ErrorIs(t, err, domain.ErrSeatingNeedOnAdult)
}

func TestResolveSeatingNeedAllowsChildOnlyValuesForChildrenOnly(t *testing.T) {
	t.Parallel()

	childOnly := []domain.SeatingNeed{domain.SeatingNeedWithParent, domain.SeatingNeedHighChair}
	for _, need := range childOnly {
		resolved, err := domain.ResolveSeatingNeed(domain.GuestKindChild, need)
		require.NoErrorf(t, err, "need %q", need)
		assert.Equal(t, need, resolved)

		_, err = domain.ResolveSeatingNeed(domain.GuestKindAdult, need)
		require.ErrorIsf(t, err, domain.ErrSeatingNeedOnAdult, "need %q", need)
	}

	// A wheelchair space is needed by adults as often as by children, and `normal` is
	// the default for everybody — neither is gated.
	for _, need := range []domain.SeatingNeed{domain.SeatingNeedNormal, domain.SeatingNeedWheelchair} {
		for _, kind := range []domain.GuestKind{domain.GuestKindAdult, domain.GuestKindChild} {
			resolved, err := domain.ResolveSeatingNeed(kind, need)
			require.NoErrorf(t, err, "kind %q, need %q", kind, need)
			assert.Equal(t, need, resolved)
		}
	}
}

// Asking for adult *and* an age is a contradiction, and reporting it is what lets the
// form put the message next to the age field.
func TestApplyGuestPatchRejectsAnAgeSentWithAdult(t *testing.T) {
	t.Parallel()

	adult := domain.GuestKindAdult
	_, _, err := domain.ApplyGuestPatch(child(), domain.GuestPatch{Kind: &adult, AgeSet: true, Age: age(4)})

	require.ErrorIs(t, err, domain.ErrAgeOnAdult)
}

// A null age is not an absent one: absent leaves the value alone, null clears it.
func TestApplyGuestPatchDistinguishesAnAbsentAgeFromAnExplicitNull(t *testing.T) {
	t.Parallel()

	untouched, _, err := domain.ApplyGuestPatch(child(), domain.GuestPatch{})
	require.NoError(t, err)
	require.NotNil(t, untouched.Age)
	assert.Equal(t, 4, *untouched.Age)

	cleared, changes, err := domain.ApplyGuestPatch(child(), domain.GuestPatch{AgeSet: true, Age: nil})
	require.NoError(t, err)
	assert.Nil(t, cleared.Age)
	assert.Equal(t, map[string]any{"age": nil}, changes.After)
}

func TestApplyGuestPatchLeavesAbsentFieldsAlone(t *testing.T) {
	t.Parallel()

	note := "Nussallergie"
	updated, changes, err := domain.ApplyGuestPatch(child(), domain.GuestPatch{DietaryNote: &note})

	require.NoError(t, err)
	assert.Equal(t, note, updated.DietaryNote)
	assert.Equal(t, "Emil Müller", updated.Name)
	assert.Equal(t, domain.SeatingNeedHighChair, updated.SeatingNeed)
	assert.Equal(t, []string{"dietary_note"}, changedFields(changes))
}

// A patch that changes nothing writes no audit row, which is what keeps every row in
// the log a change rather than a record of somebody pressing save.
func TestApplyGuestPatchReportsNoChangesWhenNothingDiffers(t *testing.T) {
	t.Parallel()

	sameName := "Emil Müller"
	_, changes, err := domain.ApplyGuestPatch(child(), domain.GuestPatch{Name: &sameName})

	require.NoError(t, err)
	assert.True(t, changes.IsEmpty())
}

// changedFields lists the keys of a Changes pair, sorted, for asserting on which
// fields an audit row would carry.
func changedFields(changes domain.Changes) []string {
	fields := make([]string, 0, len(changes.After))
	for field := range changes.After {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}
