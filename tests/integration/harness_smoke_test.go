package integration

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The harness is the one piece of test code with no production code behind it, so it
// gets tests of its own. A broken fixture makes every feature test lie, and a leak
// detector nobody has watched fail is not a control.

func TestHarnessServesTheRealApplication(t *testing.T) {
	app := newTestApp(t)

	response := app.get("/api/health")

	assert.Equal(t, http.StatusOK, response.Status)
	assert.JSONEq(t, `{"status":"ok"}`, response.Body)
}

// TestHarnessUsesATempFileDatabase guards the deliberate choice of a real file:
// :memory: gives each connection its own schema, which would make the two pools
// invisible to each other and hide the bugs these tests exist to catch.
func TestHarnessUsesATempFileDatabase(t *testing.T) {
	app := newTestApp(t)

	info, err := os.Stat(app.DatabasePath)
	require.NoError(t, err)
	assert.False(t, info.IsDir())
}

// TestHarnessRemovesItsDatabaseAfterTheTest inspects the path after the subtest that
// owns it has finished, which is the only moment its cleanup has definitely run.
func TestHarnessRemovesItsDatabaseAfterTheTest(t *testing.T) {
	var databasePath string
	t.Run("inner", func(t *testing.T) {
		databasePath = newTestApp(t).DatabasePath
		require.FileExists(t, databasePath)
	})

	_, err := os.Stat(databasePath)
	assert.ErrorIs(t, err, os.ErrNotExist, "the temp database outlived its test")
}

// TestHarnessIsolatesParallelTests seeds the same login code in both subtests. The
// column is UNIQUE, so a shared database would fail one insert outright — and the
// count assertion catches the subtler case of one test seeing the other's rows.
func TestHarnessIsolatesParallelTests(t *testing.T) {
	for _, name := range []string{"first", "second"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(t)
			seedHousehold(t, app.Database.Write, withCode("ABC234"), withGuests(2))

			var households, guests int
			require.NoError(t, app.Database.Read.Get(&households, `SELECT count(*) FROM household`))
			require.NoError(t, app.Database.Read.Get(&guests, `SELECT count(*) FROM guest`))

			assert.Equal(t, 1, households)
			assert.Equal(t, 2, guests)
		})
	}
}

func TestSeedHouseholdDefaultsAreUsable(t *testing.T) {
	app := newTestApp(t)

	first := seedHousehold(t, app.Database.Write)
	second := seedHousehold(t, app.Database.Write)

	assert.Len(t, first.Code, 6)
	assert.NotEqual(t, first.Code, second.Code, "generated codes must not collide")
	assert.NotContains(t, first.DisplayName, first.Code, "the default name must not carry the household's secret")
	assert.NotEqual(t, first.DisplayName, second.DisplayName)
	assert.Empty(t, first.Guests)
}

func TestSeedHouseholdAppliesOptions(t *testing.T) {
	app := newTestApp(t)

	household := seedHousehold(t, app.Database.Write,
		withCode("XYZ789"),
		withDisplayName("Familie Beispiel"),
		withAdminNote("zahlt bar"),
		withAdult("Anna Beispiel"),
		withChild("Emil Beispiel", 4),
	)

	assert.Equal(t, "XYZ789", household.Code)
	require.Len(t, household.Guests, 2)
	assert.Equal(t, "adult", household.Guests[0].Kind)
	assert.Equal(t, "child", household.Guests[1].Kind)

	var age int
	require.NoError(t, app.Database.Read.Get(&age, `SELECT age FROM guest WHERE id = ?`, household.Guests[1].ID))
	assert.Equal(t, 4, age)

	var note string
	require.NoError(t, app.Database.Read.Get(&note, `SELECT admin_note FROM household WHERE id = ?`, household.ID))
	assert.Equal(t, "zahlt bar", note)
}

// TestFindLeakDetectsPrivateFields tests the test: assertNoLeak is only worth having
// if it actually fires, and every case below is a field that reaching a guest would
// be a real privacy failure.
func TestFindLeakDetectsPrivateFields(t *testing.T) {
	leaky := []struct {
		name string
		body string
	}{
		{"household code", `{"household":{"id":1,"code":"ABC234"}}`},
		{"admin note", `{"household":{"admin_note":"zahlt bar"}}`},
		{"budget amount", `{"items":[{"title":"Catering","planned_cents":250000}]}`},
		{"budget vendor", `{"items":[{"vendor":"Catering Müller"}]}`},
		{"nested in a list", `{"guests":[{"name":"Anna"},{"name":"Emil","admin_note":"x"}]}`},
	}

	for _, testCase := range leaky {
		t.Run(testCase.name, func(t *testing.T) {
			assert.NotEmpty(t, findLeak(testCase.body))
		})
	}
}

func TestFindLeakAllowsHarmlessBodies(t *testing.T) {
	clean := []struct {
		name string
		body string
	}{
		// The envelope's own "code" is the error kind, not a login code — the one
		// exception the detector has to make, and therefore the one worth pinning.
		{"error envelope", `{"error":{"code":"not_found","message":"Nicht gefunden.","request_id":"ABC"}}`},
		{"guest DTO", `{"household":{"id":1,"display_name":"Familie Beispiel"}}`},
		{"not JSON at all", `<!doctype html><html lang="de"></html>`},
	}

	for _, testCase := range clean {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Empty(t, findLeak(testCase.body))
		})
	}
}
