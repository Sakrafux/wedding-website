package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
)

// newAdminApp starts an application with an admin already logged in, which is the
// precondition of every test in this file.
func newAdminApp(t *testing.T, options ...testAppOption) *testApp {
	t.Helper()

	app := newTestApp(t, options...)
	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)
	return app
}

// adminHousehold mirrors dto.AdminHousehold rather than importing it, so a rename in
// the DTO surfaces here as a failing wire-format assertion instead of compiling
// silently.
type adminHousehold struct {
	ID                    int64        `json:"id"`
	DisplayName           string       `json:"display_name"`
	Code                  string       `json:"code"`
	MemberCount           int          `json:"member_count"`
	LastLoginAt           *string      `json:"last_login_at"`
	RSVPSubmittedAt       *string      `json:"rsvp_submitted_at"`
	AdminNote             string       `json:"admin_note"`
	TransportSeatsNeeded  int          `json:"transport_seats_needed"`
	TransportSeatsOffered int          `json:"transport_seats_offered"`
	HasStroller           bool         `json:"has_stroller"`
	Members               []adminGuest `json:"members"`
}

type adminGuest struct {
	ID          int64  `json:"id"`
	HouseholdID int64  `json:"household_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Kind        string `json:"kind"`
	Age         *int   `json:"age"`
	Origin      string `json:"origin"`
	SeatingNeed string `json:"seating_need"`
	DietaryNote string `json:"dietary_note"`
}

func (response *testResponse) adminHousehold() adminHousehold {
	response.t.Helper()

	var body adminHousehold
	response.decodeJSON(&body)
	return body
}

func (response *testResponse) adminHouseholds() []adminHousehold {
	response.t.Helper()

	var body struct {
		Households []adminHousehold `json:"households"`
	}
	response.decodeJSON(&body)
	return body.Households
}

// The whole point of the epic: the real guest list can be entered through the API
// alone.
func TestAdminCreatesReadsUpdatesAndDeletesAHousehold(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)

	created := app.postJSON("/api/admin/households", map[string]any{
		"display_name":            "Familie Müller",
		"admin_note":              "Kommen mit dem Zug",
		"transport_seats_offered": 4,
	})
	require.Equal(t, http.StatusCreated, created.Status)

	household := created.adminHousehold()
	assert.Positive(t, household.ID)
	assert.Equal(t, "Familie Müller", household.DisplayName)
	assert.Equal(t, "Kommen mit dem Zug", household.AdminNote)
	assert.Equal(t, 4, household.TransportSeatsOffered)
	assert.Empty(t, household.Members, "a fresh household has an empty member list, not a null one")
	assert.Nil(t, household.LastLoginAt)

	// A household without a code is one nobody can log in as, and no screen would
	// show you that — so creation assigns one.
	require.NoError(t, domain.ValidateCode(household.Code))

	listed := app.get("/api/admin/households").adminHouseholds()
	require.Len(t, listed, 1)
	assert.Equal(t, household.ID, listed[0].ID)
	assert.Equal(t, household.Code, listed[0].Code)
	assert.Equal(t, 0, listed[0].MemberCount)

	path := fmt.Sprintf("/api/admin/households/%d", household.ID)
	assert.Equal(t, "Familie Müller", app.get(path).adminHousehold().DisplayName)

	updated := app.patchJSON(path, map[string]any{"display_name": "Familie Müller-Schmidt"})
	require.Equal(t, http.StatusOK, updated.Status)
	assert.Equal(t, "Familie Müller-Schmidt", updated.adminHousehold().DisplayName)

	require.Equal(t, http.StatusNoContent, app.deleteRequest(path).Status)
	assert.Equal(t, http.StatusNotFound, app.get(path).Status)
	assert.Empty(t, app.get("/api/admin/households").adminHouseholds())
}

// The admin scans this list looking for a name, so it arrives sorted by name and not
// in insertion order.
func TestAdminHouseholdListIsSortedByNameWithMemberCounts(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	seedHousehold(t, app.Database.Write, withDisplayName("Zimmermann"), withGuests(1))
	seedHousehold(t, app.Database.Write, withDisplayName("albrecht"), withGuests(3))
	müller := seedHousehold(t, app.Database.Write, withDisplayName("Müller"), withGuests(2))

	// A soft-deleted member is nobody's member any more, and must not be counted.
	require.NoError(t, app.guestStore().SoftDelete(t.Context(), müller.Guests[0].ID, time.Now()))

	listed := app.get("/api/admin/households").adminHouseholds()
	require.Len(t, listed, 3)

	assert.Equal(t, []string{"albrecht", "Müller", "Zimmermann"}, []string{
		listed[0].DisplayName, listed[1].DisplayName, listed[2].DisplayName,
	}, "sorted case-insensitively by name")
	assert.Equal(t, []int{3, 1, 1}, []int{listed[0].MemberCount, listed[1].MemberCount, listed[2].MemberCount})
}

func TestAdminHouseholdDetailEmbedsItsMembers(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write,
		withAdult("Anna", "Müller"), withChild("Emil", "Müller", 4))

	detail := app.get(fmt.Sprintf("/api/admin/households/%d", household.ID)).adminHousehold()

	require.Len(t, detail.Members, 2)
	assert.Equal(t, "Anna", detail.Members[0].FirstName)
	assert.Nil(t, detail.Members[0].Age)
	assert.Equal(t, "child", detail.Members[1].Kind)
	require.NotNil(t, detail.Members[1].Age)
	assert.Equal(t, 4, *detail.Members[1].Age)
	assert.Equal(t, "normal", detail.Members[1].SeatingNeed)
}

// Absent means "leave alone" and present-but-empty means "clear" — the distinction
// the pointer fields exist for.
func TestAdminHouseholdPatchLeavesAbsentFieldsAloneAndClearsEmptyOnes(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write,
		withDisplayName("Familie Müller"), withAdminNote("Kommen mit dem Zug"))
	path := fmt.Sprintf("/api/admin/households/%d", household.ID)

	patched := app.patchJSON(path, map[string]any{"has_stroller": true}).adminHousehold()
	assert.True(t, patched.HasStroller)
	assert.Equal(t, "Familie Müller", patched.DisplayName)
	assert.Equal(t, "Kommen mit dem Zug", patched.AdminNote)

	cleared := app.patchJSON(path, map[string]any{"admin_note": ""}).adminHousehold()
	assert.Empty(t, cleared.AdminNote)
	assert.True(t, cleared.HasStroller, "the earlier change survives the next patch")
}

// A code changed by a stray field in a form body is a code nobody knows changed. The
// field does not exist on the request, and an unknown field is rejected outright.
func TestAdminHouseholdPatchCannotChangeTheCode(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	response := app.patchJSON(fmt.Sprintf("/api/admin/households/%d", household.ID),
		map[string]any{"display_name": "Familie Müller", "code": "ZZZ999"})

	assert.Equal(t, http.StatusBadRequest, response.Status)
	assert.Equal(t, "validation_failed", response.errorEnvelope().Code)
	assert.Equal(t, "ABC234", app.storedCode(household.ID))
}

func TestAdminHouseholdValidationFailuresNameTheirFieldsInGerman(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)

	response := app.postJSON("/api/admin/households", map[string]any{
		"display_name":           "",
		"transport_seats_needed": 40,
	})
	require.Equal(t, http.StatusBadRequest, response.Status)

	envelope := response.errorEnvelope()
	assert.Equal(t, "validation_failed", envelope.Code)
	assert.Contains(t, envelope.Message, "Felder")
	assert.Contains(t, envelope.Fields, "display_name", "keyed by the JSON name, not the Go one")
	assert.Contains(t, envelope.Fields, "transport_seats_needed")
	assert.Contains(t, envelope.Fields["display_name"], "Bitte")
}

// Deleting a household takes its guests and their seats with it, by foreign key. What
// it does not take is the audit trail, which is the whole reason that table outlives
// the row.
func TestDeletingAHouseholdRemovesItsGuestsAndSeatAssignments(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withAdult("Anna", "Müller"))
	unit := insertSeatingUnit(t, app.Database.Write, "party", "Tisch 1")
	seat := insertSeat(t, app.Database.Write, unit, "party", "Platz 1")
	assignSeat(t, app.Database.Write, seat, "party", household.Guests[0].ID)

	require.Equal(t, http.StatusNoContent,
		app.deleteRequest(fmt.Sprintf("/api/admin/households/%d", household.ID)).Status)

	var guests, assignments int
	require.NoError(t, app.Database.Read.Get(&guests, `SELECT COUNT(*) FROM guest WHERE household_id = ?`, household.ID))
	require.NoError(t, app.Database.Read.Get(&assignments, `SELECT COUNT(*) FROM seat_assignment`))
	assert.Equal(t, 0, guests)
	assert.Equal(t, 0, assignments)

	assert.NotEmpty(t, app.auditRows(), "the audit trail outlives the household")
}

// A logged-in household whose row is deleted must meet the login screen, not an
// error — and their session must not linger for a year.
func TestASessionOfADeletedHouseholdResolvesAsAnonymous(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	guest := app.onANewDevice()
	require.Equal(t, http.StatusOK, guest.logIn("ABC234").Status)
	require.Equal(t, 1, app.countHouseholdSessions(household.ID))

	require.Equal(t, http.StatusNoContent,
		app.deleteRequest(fmt.Sprintf("/api/admin/households/%d", household.ID)).Status)

	assert.Equal(t, http.StatusUnauthorized, guest.get("/api/me").Status)
	assert.Equal(t, 0, app.countHouseholdSessions(household.ID))
}

// Every mutation, exactly one row, and never the code. audit_log is append-only and
// nobody deletes from it, so a code written here would be a permanent second copy of
// the key list.
func TestAdminHouseholdMutationsAreAuditedWithoutTheCode(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)

	created := app.postJSON("/api/admin/households", map[string]any{"display_name": "Familie Müller"})
	require.Equal(t, http.StatusCreated, created.Status)
	household := created.adminHousehold()
	path := fmt.Sprintf("/api/admin/households/%d", household.ID)

	require.Equal(t, http.StatusOK, app.patchJSON(path, map[string]any{"admin_note": "Ruft an"}).Status)
	require.Equal(t, http.StatusOK, app.post(path+"/code").Status)
	newCode := app.storedCode(household.ID)
	require.Equal(t, http.StatusNoContent, app.deleteRequest(path).Status)

	// The admin login is audited too, so the household rows are the ones after it.
	var actions []string
	for _, row := range app.auditRows() {
		if row.Entity == domain.AuditEntityHousehold {
			actions = append(actions, row.Action)
			assert.Equal(t, "admin", row.ActorType)
			assert.Equal(t, household.ID, row.EntityID)
		}
	}
	assert.Equal(t, []string{"create", "update", "update", "delete"}, actions,
		"one row per mutation: create, the patch, the reissued code, delete")

	wholeTable := app.auditPayloads()
	assert.NotContains(t, wholeTable, household.Code, "the original code must not be in the log")
	assert.NotContains(t, wholeTable, newCode, "nor the new one")
	assert.Contains(t, wholeTable, "code_reissued", "that it changed is recorded; neither value is")
}

// A patch that changes nothing is not a change, and the log is worth reading because
// every row in it is one.
func TestAPatchThatChangesNothingWritesNoAuditRow(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withDisplayName("Familie Müller"))
	before := len(app.auditRows())

	response := app.patchJSON(fmt.Sprintf("/api/admin/households/%d", household.ID),
		map[string]any{"display_name": "Familie Müller"})

	require.Equal(t, http.StatusOK, response.Status)
	assert.Len(t, app.auditRows(), before)
}

func TestAdminHouseholdRoutesAnswer404ForAnUnknownID(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)

	paths := map[string]*testResponse{
		"read":    app.get("/api/admin/households/9999"),
		"patch":   app.patchJSON("/api/admin/households/9999", map[string]any{"display_name": "X"}),
		"delete":  app.deleteRequest("/api/admin/households/9999"),
		"code":    app.post("/api/admin/households/9999/code"),
		"guest":   app.postJSON("/api/admin/households/9999/guests", map[string]any{"first_name": "A", "last_name": "B", "kind": "adult"}),
		"garbage": app.get("/api/admin/households/nicht-numerisch"),
	}

	for name, response := range paths {
		assert.Equalf(t, http.StatusNotFound, response.Status, "%s", name)
		assert.Equalf(t, "not_found", response.errorEnvelope().Code, "%s", name)
	}
}

/* ----------------------------------------------------------------------- *
 * F5-B03 — code assignment and reissue
 * ----------------------------------------------------------------------- */

func TestEveryCreatedHouseholdGetsADistinctCode(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)

	codes := map[string]bool{}
	for index := range 25 {
		response := app.postJSON("/api/admin/households",
			map[string]any{"display_name": fmt.Sprintf("Familie %d", index)})
		require.Equal(t, http.StatusCreated, response.Status)

		code := response.adminHousehold().Code
		require.NoError(t, domain.ValidateCode(code))
		require.Falsef(t, codes[code], "code %s was issued twice", code)
		codes[code] = true
	}
}

// The path that never runs in practice, and therefore never gets exercised by
// accident: the UNIQUE index rejects the insert and the store tries another code.
func TestACollidingCodeIsRetried(t *testing.T) {
	t.Parallel()

	const taken = "ABC234"
	attempt := 0
	app := newAdminApp(t, withCodeGenerator(func() string {
		attempt++
		if attempt == 1 {
			return taken
		}
		return "DEF567"
	}))
	seedHousehold(t, app.Database.Write, withCode(taken))

	response := app.postJSON("/api/admin/households", map[string]any{"display_name": "Familie Müller"})

	require.Equal(t, http.StatusCreated, response.Status)
	assert.Equal(t, "DEF567", response.adminHousehold().Code)
	assert.Equal(t, 2, attempt, "the first code collided and a second was generated")
}

// Five collisions in a row is a broken generator, not bad luck. It must fail loudly
// rather than loop, because a request that never answers is the one failure nobody
// can diagnose from the outside.
func TestAGeneratorThatOnlyEverCollidesFailsLoudly(t *testing.T) {
	t.Parallel()

	const taken = "ABC234"
	attempts := 0
	app := newAdminApp(t, withCodeGenerator(func() string {
		attempts++
		return taken
	}))
	seedHousehold(t, app.Database.Write, withCode(taken))

	response := app.postJSON("/api/admin/households", map[string]any{"display_name": "Familie Müller"})

	assert.Equal(t, http.StatusInternalServerError, response.Status)
	assert.Equal(t, "internal_error", response.errorEnvelope().Code)
	assert.Equal(t, 5, attempts, "bounded, not endless")

	var households int
	require.NoError(t, app.Database.Read.Get(&households, `SELECT COUNT(*) FROM household WHERE display_name = ?`,
		"Familie Müller"))
	assert.Equal(t, 0, households)
}

// The reason the endpoint exists: a lost or leaked card can be recovered from, and
// the old code stops working the moment it is.
func TestReissuingACodeReplacesItAndTheOldOneStopsWorking(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	response := app.post(fmt.Sprintf("/api/admin/households/%d/code", household.ID))
	require.Equal(t, http.StatusOK, response.Status)

	var body struct {
		Code            string `json:"code"`
		RevokedSessions int64  `json:"revoked_sessions"`
	}
	response.decodeJSON(&body)

	require.NoError(t, domain.ValidateCode(body.Code))
	assert.NotEqual(t, "ABC234", body.Code)
	assert.Equal(t, body.Code, app.storedCode(household.ID))
	assert.Equal(t, int64(0), body.RevokedSessions, "nobody was logged in")

	guest := app.onANewDevice()
	assert.Equal(t, http.StatusUnauthorized, guest.logIn("ABC234").Status)
	assert.Equal(t, http.StatusOK, guest.logIn(body.Code).Status)
}

// Scoped to the one household: reissuing a code must not sign the other seventy-nine
// out.
func TestReissuingACodeRevokesOnlyThatHouseholdsSessions(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	reissued := seedHousehold(t, app.Database.Write, withCode("ABC234"))
	untouched := seedHousehold(t, app.Database.Write, withCode("DEF567"))

	first, second := app.onANewDevice(), app.onANewDevice()
	require.Equal(t, http.StatusOK, first.logIn("ABC234").Status)
	require.Equal(t, http.StatusOK, second.logIn("DEF567").Status)

	response := app.post(fmt.Sprintf("/api/admin/households/%d/code", reissued.ID))
	require.Equal(t, http.StatusOK, response.Status)

	var body struct {
		RevokedSessions int64 `json:"revoked_sessions"`
	}
	response.decodeJSON(&body)

	assert.Equal(t, int64(1), body.RevokedSessions, "reported, so the admin sees what they just did")
	assert.Equal(t, 0, app.countHouseholdSessions(reissued.ID))
	assert.Equal(t, 1, app.countHouseholdSessions(untouched.ID))
	assert.Equal(t, http.StatusOK, second.get("/api/me").Status, "the other household stays logged in")
}

/* ----------------------------------------------------------------------- *
 * Privacy
 * ----------------------------------------------------------------------- */

// Admin responses legitimately carry `code` and `admin_note`, so they do not call
// assertNoLeak. The guard is therefore made stronger here rather than weaker: every
// non-admin route is walked and its body checked, so adding an endpoint that leaks
// one of them fails without anybody having to remember this rule.
func TestNoNonAdminRouteLeaksPrivateData(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write,
		withCode("ABC234"), withAdminNote("Kommen mit dem Zug"), withGuests(2))
	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	for _, route := range nonAdminAPIRoutes(t, app) {
		response := app.requestWith(route.method, route.path, nil, nil)

		response.assertNoLeak(household.Code, "Kommen mit dem Zug")
	}
}

// nonAdminAPIRoutes walks the real router for every /api route outside the admin
// subtree — a list nobody has to maintain, which is the only kind that stays complete.
func nonAdminAPIRoutes(t *testing.T, app *testApp) []registeredRoute {
	t.Helper()

	var routes []registeredRoute
	for _, route := range allRoutes(t, app) {
		if strings.HasPrefix(route.path, "/api") && !strings.HasPrefix(route.path, "/api/admin") {
			routes = append(routes, route)
		}
	}
	require.NotEmpty(t, routes)
	return routes
}
