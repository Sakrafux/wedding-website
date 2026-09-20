package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// F4-B04: a household writes the names on its own card, and we keep the name we
// posted the invitation to.

func TestRSVPRenameKeepsTheSeededNameForTheAdmin(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Erika Huber"))
	erika := household.Guests[0]

	renamed := answerFor(erika, "both")
	renamed["name"] = "  Oma Erika  "
	saved := app.putJSON("/api/rsvp", submission(renamed))
	require.Equal(t, http.StatusOK, saved.Status, saved.Body)

	// Trimmed, and the stored name is what comes back — the form renders the answer
	// as saved, not as typed.
	assert.Equal(t, "Oma Erika", saved.rsvp().memberByID(t, erika.ID).Name)

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)

	member := admin.members(household.ID)[0]
	assert.Equal(t, "Oma Erika", member.Name)
	assert.Equal(t, "Erika Huber", member.SeededName)
}

// The seeded name is admin-only history. A guest response carrying it would hand the
// household back a name they deliberately replaced.
func TestTheGuestRSVPBodyCarriesNoSeededName(t *testing.T) {
	t.Parallel()

	app, _ := newHouseholdApp(t, withAdult("Erika Huber"))

	assert.NotContains(t, app.get("/api/rsvp").Body, "seeded_name")
}

// An admin rename is us fixing our own typo, so it moves both names and the detail
// page stops claiming the household renamed anybody.
func TestAdminRenameMovesTheSeededName(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withAdult("Anna Müler"))

	path := fmt.Sprintf("/api/admin/guests/%d", household.Guests[0].ID)
	require.Equal(t, http.StatusOK, app.patchJSON(path, map[string]any{"name": "Anna Müller"}).Status)

	member := app.members(household.ID)[0]
	assert.Equal(t, "Anna Müller", member.Name)
	assert.Equal(t, "Anna Müller", member.SeededName)
}

// A plus-one has no name of ours behind them, so there is nothing to keep.
func TestAPlusOneHasNoSeededName(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	added := app.postJSON("/api/rsvp/members", map[string]any{"name": "Tom Berger"})
	require.Equal(t, http.StatusCreated, added.Status, added.Body)

	admin := app.onANewDevice()
	require.Equal(t, http.StatusOK, admin.logInAsAdmin().Status)

	members := admin.members(household.ID)
	require.Len(t, members, 2)
	assert.Empty(t, members[1].SeededName)
}

func TestRSVPRefusesABlankName(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	blank := answerFor(anna, "both")
	blank["name"] = "   "
	response := app.putJSON("/api/rsvp", submission(blank))

	require.Equal(t, http.StatusBadRequest, response.Status)
	assert.Contains(t, response.errorEnvelope().Fields, keyFor(anna.ID, "name"))
	assert.Equal(t, "Anna Müller", app.get("/api/rsvp").rsvp().memberByID(t, anna.ID).Name)
}

// The name is capped like every other one we store, so a paste accident is a field
// error rather than a row nothing can render.
func TestRSVPRefusesAnOverlongName(t *testing.T) {
	t.Parallel()

	app, household := newHouseholdApp(t, withAdult("Anna Müller"))
	anna := household.Guests[0]

	overlong := answerFor(anna, "both")
	overlong["name"] = strings.Repeat("a", 161)

	require.Equal(t, http.StatusBadRequest, app.putJSON("/api/rsvp", submission(overlong)).Status)
}
