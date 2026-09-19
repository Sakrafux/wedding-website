package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

func household() domain.Household {
	return domain.Household{
		ID:                    12,
		Name:                  "Familie Müller",
		Addressee:             "Hans & Erika",
		Code:                  "ABC234",
		TransportSeatsNeeded:  0,
		TransportSeatsOffered: 4,
		AdminNote:             "Kommen mit dem Zug",
	}
}

func TestApplyHouseholdPatchLeavesAbsentFieldsAlone(t *testing.T) {
	t.Parallel()

	name := "Familie Müller-Schmidt"
	updated, changes := domain.ApplyHouseholdPatch(household(), domain.HouseholdPatch{Name: &name})

	assert.Equal(t, name, updated.Name)
	assert.Equal(t, "Hans & Erika", updated.Addressee, "the two names move independently")
	assert.Equal(t, "Kommen mit dem Zug", updated.AdminNote)
	assert.Equal(t, 4, updated.TransportSeatsOffered)
	assert.Equal(t, []string{"name"}, changedFields(changes))
}

// The reason the patch fields are pointers: an empty value has to be able to clear a
// column, which a plain string could not express.
func TestApplyHouseholdPatchClearsAFieldSentEmpty(t *testing.T) {
	t.Parallel()

	empty := ""
	updated, changes := domain.ApplyHouseholdPatch(household(), domain.HouseholdPatch{AdminNote: &empty})

	assert.Empty(t, updated.AdminNote)
	assert.Equal(t, map[string]any{"admin_note": "Kommen mit dem Zug"}, changes.Before)
	assert.Equal(t, map[string]any{"admin_note": ""}, changes.After)
}

// The code is not patchable at all: it is not a field of HouseholdPatch, so a form
// body cannot reach it. Asserted anyway, because "add the missing field" is exactly
// the well-meant change this guards against.
func TestApplyHouseholdPatchNeverTouchesTheCode(t *testing.T) {
	t.Parallel()

	name := "Familie Schmidt"
	updated, changes := domain.ApplyHouseholdPatch(household(), domain.HouseholdPatch{Name: &name})

	assert.Equal(t, "ABC234", updated.Code)
	assert.NotContains(t, changes.Before, "code")
	assert.NotContains(t, changes.After, "code")
}

// The internal name and the guest-facing addressee are separate columns and separate
// audit entries: renaming the household must not quietly reprint the cards.
func TestApplyHouseholdPatchTracksNameAndAddresseeSeparately(t *testing.T) {
	t.Parallel()

	addressee := "Luki & Paddi"
	updated, changes := domain.ApplyHouseholdPatch(household(), domain.HouseholdPatch{Addressee: &addressee})

	assert.Equal(t, "Familie Müller", updated.Name)
	assert.Equal(t, addressee, updated.Addressee)
	assert.Equal(t, []string{"addressee"}, changedFields(changes))
}

func TestApplyHouseholdPatchReportsNoChangesWhenNothingDiffers(t *testing.T) {
	t.Parallel()

	sameName := "Familie Müller"
	seats := 4
	_, changes := domain.ApplyHouseholdPatch(household(),
		domain.HouseholdPatch{Name: &sameName, TransportSeatsOffered: &seats})

	assert.True(t, changes.IsEmpty())
}
