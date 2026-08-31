package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
)

// withDevEnvironment points the command at a fresh temp-file database and gives it
// the rest of the environment configuration.Load insists on. A real file, never
// :memory:, because the command opens two handles on the path and an in-memory
// database would give each of them its own empty schema.
func withDevEnvironment(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "seed-test.db")
	t.Setenv("DB_PATH", path)
	t.Setenv("PHOTO_DIR", filepath.Join(t.TempDir(), "photos"))
	t.Setenv("ADMIN_USER", "test-admin")
	t.Setenv("ADMIN_PASSWORD", "test-admin-password")
	return path
}

// openSeededDatabase opens the file the command just wrote, for assertions.
func openSeededDatabase(t *testing.T, path string) *configuration.Database {
	t.Helper()

	database, err := configuration.OpenDatabase(configuration.Config{DatabasePath: path})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	return database
}

func TestRunSeedsRequestedHouseholdsAndMembers(t *testing.T) {
	path := withDevEnvironment(t)

	var out bytes.Buffer
	require.NoError(t, run(3, 2, &out))

	database := openSeededDatabase(t, path)

	var households, guests int
	require.NoError(t, database.Read.Get(&households, `SELECT COUNT(*) FROM household`))
	require.NoError(t, database.Read.Get(&guests, `SELECT COUNT(*) FROM guest`))
	require.Equal(t, 3, households)
	require.Equal(t, 6, guests)

	// Every code must be storable and loginable exactly as printed. A code stored
	// lower case would let the tool report something POST /api/auth/login rejects.
	var codes []string
	require.NoError(t, database.Read.Select(&codes, `SELECT code FROM household`))
	for _, code := range codes {
		require.NoError(t, domain.ValidateCode(code))
		require.Contains(t, out.String(), code)
	}

	require.Contains(t, out.String(), "DEVELOPMENT ONLY")
}

// A printed code must resolve back to its household through the same normalisation
// the login endpoint applies — including the dash a guest may add out of habit,
// which nothing prints but everything accepts.
func TestPrintedCodesResolveToTheirHousehold(t *testing.T) {
	path := withDevEnvironment(t)
	require.NoError(t, run(2, 1, new(bytes.Buffer)))

	database := openSeededDatabase(t, path)
	households := persistence.NewHouseholdStore(database)

	var codes []string
	require.NoError(t, database.Read.Select(&codes, `SELECT code FROM household`))
	require.Len(t, codes, 2)

	for _, code := range codes {
		typedByAGuest := domain.NormalizeCode(code[:3] + "-" + code[3:])
		require.NoError(t, domain.ValidateCode(typedByAGuest))

		household, err := households.FindByCode(context.Background(), typedByAGuest)
		require.NoError(t, err)
		require.Equal(t, code, household.Code)
	}
}

// A second run is the normal case: the database from yesterday is still there.
func TestRunAppendsToAnExistingDatabase(t *testing.T) {
	path := withDevEnvironment(t)

	require.NoError(t, run(2, 1, new(bytes.Buffer)))
	require.NoError(t, run(2, 1, new(bytes.Buffer)))

	database := openSeededDatabase(t, path)

	var names []string
	require.NoError(t, database.Read.Select(&names, `SELECT display_name FROM household ORDER BY id`))
	require.Equal(t, []string{
		"Familie Testhaushalt 1",
		"Familie Testhaushalt 2",
		"Familie Testhaushalt 3",
		"Familie Testhaushalt 4",
	}, names)
}

func TestRunRejectsNonPositiveCounts(t *testing.T) {
	path := withDevEnvironment(t)

	for name, arguments := range map[string][2]int{
		"no households": {0, 2},
		"no guests":     {1, 0},
	} {
		t.Run(name, func(t *testing.T) {
			err := run(arguments[0], arguments[1], new(bytes.Buffer))
			require.ErrorContains(t, err, "must both be at least 1")
		})
	}

	// The count check runs before the database is touched, so a usage error leaves
	// no file behind to confuse the next run.
	require.NoFileExists(t, path)
}

// Members belong to their household and are seeded adults — the state before the
// household has answered anything, which is what the login flow expects to find.
func TestSeededMembersAreAdultsOfTheirHousehold(t *testing.T) {
	path := withDevEnvironment(t)
	require.NoError(t, run(1, 2, new(bytes.Buffer)))

	database := openSeededDatabase(t, path)

	type member struct {
		Name      string `db:"name"`
		Kind      string `db:"kind"`
		Origin    string `db:"origin"`
		Attending *string
	}
	var members []member
	require.NoError(t, database.Read.Select(&members,
		`SELECT g.name, g.kind, g.origin, g.attending
		 FROM guest g JOIN household h ON h.id = g.household_id
		 WHERE h.display_name = 'Familie Testhaushalt 1'
		 ORDER BY g.id`))

	require.Len(t, members, 2)
	for _, m := range members {
		// Numbered, so a member's own name names their household too — the seeder
		// derives it, which is something only the seeder may do.
		require.Regexp(t, `^\S+ Testhaushalt 1$`, m.Name)
		require.Equal(t, string(domain.GuestKindAdult), m.Kind)
		require.Equal(t, string(domain.GuestOriginSeeded), m.Origin)
		require.Nil(t, m.Attending)
		require.NotEmpty(t, strings.TrimSpace(m.Name))
	}
}

// The command must work against a database file that has never been served, since
// clone → make seed → make run is the order a new checkout is set up in.
func TestRunMigratesAFreshDatabase(t *testing.T) {
	path := withDevEnvironment(t)
	require.NoFileExists(t, path)

	require.NoError(t, run(1, 1, new(bytes.Buffer)))

	database := openSeededDatabase(t, path)

	var version int
	require.NoError(t, database.Read.GetContext(context.Background(), &version,
		`SELECT MAX(version) FROM schema_migration`))
	require.Positive(t, version)
}
