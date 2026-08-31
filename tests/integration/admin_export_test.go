package integration

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
)

// utf8BOM is what German Excel needs in front of the bytes to read them as UTF-8
// rather than as Latin-1. Written out as an escape, because a literal U+FEFF in a
// source file is invisible in review.
const utf8BOM = "\ufeff"

// readCSV parses an export body the way a spreadsheet does: BOM stripped,
// semicolon-delimited.
func readCSV(t *testing.T, body string) [][]string {
	t.Helper()

	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(body, utf8BOM)))
	reader.Comma = ';'

	rows, err := reader.ReadAll()
	require.NoError(t, err)
	return rows
}

// The file the print shop gets: one row per household, in the form that goes on the
// card.
func TestCodesExportListsEveryHouseholdWithItsPrintedCode(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	first := seedHousehold(t, app.Database.Write, withDisplayName("Familie Müller"), withCode("ABC234"))
	second := seedHousehold(t, app.Database.Write, withDisplayName("Familie Albrecht"), withCode("DEF567"))

	response := app.get("/api/admin/export/codes.csv")
	require.Equal(t, http.StatusOK, response.Status)
	assert.Contains(t, response.ContentType, "text/csv")

	rows := readCSV(t, response.Body)
	require.Len(t, rows, 3)

	// German headers, uniquely in this application, because a print shop reads them.
	assert.Equal(t, []string{"haushalt", "code"}, rows[0])
	assert.Equal(t, []string{second.DisplayName, "DEF567"}, rows[1], "sorted by name")
	assert.Equal(t, []string{first.DisplayName, "ABC234"}, rows[2])
}

// The three encoding decisions, asserted on the bytes rather than on the parsed
// values: this is exactly what the BOM is for, and a test that parsed first would
// pass without it.
func TestExportsAreExcelReadableUTF8(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	seedHousehold(t, app.Database.Write, withDisplayName("Familie Müller"), withGuests(1))

	for _, file := range []string{"codes.csv", "guests.csv"} {
		response := app.get("/api/admin/export/" + file)
		require.Equal(t, http.StatusOK, response.Status)

		assert.Truef(t, strings.HasPrefix(response.Body, utf8BOM), "%s starts with a BOM", file)
		assert.Containsf(t, response.Body, ";", "%s is semicolon-delimited", file)
		assert.Containsf(t, response.Body, "\r\n", "%s uses CRLF", file)
		assert.Containsf(t, response.Body, `"`, "%s quotes every field", file)
		// The umlaut survives as UTF-8 bytes and is not mangled into Latin-1.
		assert.Containsf(t, response.Body, "Müller", "%s keeps umlauts", file)
		assert.NotContainsf(t, response.Body, "MÃ¼ller", "%s is not double-encoded", file)

		assert.Equal(t, `attachment; filename="`+file+`"`, response.Header.Get("Content-Disposition"))
	}
}

// A name with a semicolon and a quote in it must survive the round trip: the
// separator is what a German spreadsheet splits on, so an unescaped one would shift
// every following column of that row.
func TestExportedFieldsSurviveSeparatorsAndQuotes(t *testing.T) {
	t.Parallel()

	const awkward = `Familie "Groß"; von Müller`
	app := newAdminApp(t)
	seedHousehold(t, app.Database.Write, withDisplayName(awkward), withCode("ABC234"))

	rows := readCSV(t, app.get("/api/admin/export/codes.csv").Body)

	require.Len(t, rows, 2)
	assert.Equal(t, []string{awkward, "ABC234"}, rows[1])
}

// The file is a dump of two tables, and it has to stay one: a migration that adds a
// column and forgets this export fails here rather than producing a file that
// quietly omits it.
func TestGuestExportCarriesEveryColumnOfGuestAndHousehold(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)

	expected := []string{}
	for _, column := range tableColumns(t, app, "guest") {
		if column == "id" {
			// `guest.id` is `guest_id`, so it does not collide with the household's.
			expected = append(expected, "guest_id")
			continue
		}
		expected = append(expected, column)
	}
	for _, column := range tableColumns(t, app, "household") {
		prefixed := "household_" + column
		// `household.id` and `guest.household_id` are the same number and share one
		// column in the file.
		if !slices.Contains(expected, prefixed) {
			expected = append(expected, prefixed)
		}
	}

	actual := readCSV(t, app.get("/api/admin/export/guests.csv").Body)[0]

	slices.Sort(expected)
	sorted := slices.Clone(actual)
	slices.Sort(sorted)
	assert.Equal(t, expected, sorted)

	// Third, so it cannot be missed while scanning: this file includes removed people.
	require.Greater(t, len(actual), 3)
	assert.Equal(t, "deleted_at", actual[2])
	assert.Equal(t, persistence.GuestExportColumns, actual)
}

// Excluding removed guests would make the file disagree with the database it claims
// to dump — and a removed plus-one is exactly the row somebody eventually asks about.
func TestGuestExportIncludesRemovedPeopleWithTheirDeletionTime(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	household := seedHousehold(t, app.Database.Write,
		withDisplayName("Familie Müller"), withCode("ABC234"),
		withAdult("Anna Müller"), withChild("Emil Müller", 4))

	require.Equal(t, http.StatusNoContent,
		app.deleteRequest(fmt.Sprintf("/api/admin/guests/%d", household.Guests[1].ID)).Status)

	rows := readCSV(t, app.get("/api/admin/export/guests.csv").Body)
	require.Len(t, rows, 3, "both guests, the removed one included")

	header := rows[0]
	byColumn := func(row []string, column string) string {
		index := slices.Index(header, column)
		require.GreaterOrEqual(t, index, 0, column)
		return row[index]
	}

	assert.Empty(t, byColumn(rows[1], "deleted_at"))
	assert.NotEmpty(t, byColumn(rows[2], "deleted_at"), "Emil was removed and is still in the dump")
	assert.Equal(t, "Emil Müller", byColumn(rows[2], "name"))
	assert.Equal(t, "4", byColumn(rows[2], "age"))
	assert.Equal(t, "Familie Müller", byColumn(rows[2], "household_display_name"))
	assert.Equal(t, "ABC234", byColumn(rows[2], "household_code"))
}

// The RSVP columns exist before F3 fills them. Emitting them from the start means the
// file's shape does not change on the day the answers arrive, so nothing built on it
// breaks at once.
func TestGuestExportCarriesEmptyRSVPColumnsBeforeF3(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	seedHousehold(t, app.Database.Write, withAdult("Anna Müller"))

	rows := readCSV(t, app.get("/api/admin/export/guests.csv").Body)
	require.Len(t, rows, 2)

	for _, column := range []string{"attending", "meal_choice", "household_rsvp_submitted_at"} {
		index := slices.Index(rows[0], column)
		require.GreaterOrEqualf(t, index, 0, "%s is present", column)
		assert.Emptyf(t, rows[1][index], "%s is empty until F3", column)
	}
}

// An export is recorded as a log line and deliberately not as an audit row:
// audit_log.action has no `read` value, and the CHECK constraint is what keeps that
// table a record of changes.
func TestAnExportIsLoggedWithItsRowCount(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t)
	seedHousehold(t, app.Database.Write, withGuests(2))
	before := len(app.auditRows())

	require.Equal(t, http.StatusOK, app.get("/api/admin/export/codes.csv").Status)
	require.Equal(t, http.StatusOK, app.get("/api/admin/export/guests.csv").Status)

	logged := app.Logs.String()
	assert.Contains(t, logged, "csv export written")
	assert.Contains(t, logged, `"file":"codes.csv"`)
	assert.Contains(t, logged, `"file":"guests.csv"`)
	// Two rows for codes.csv (header plus one household), three for guests.csv.
	assert.Contains(t, logged, `"rows":2`)
	assert.Contains(t, logged, `"rows":3`)

	assert.Len(t, app.auditRows(), before, "a read is not a change")
}

// The two files are the largest disclosure in the product: codes.csv is the entire
// login-code list.
func TestExportsAreRefusedToEverybodyButTheAdmin(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	for _, file := range []string{"codes.csv", "guests.csv"} {
		anonymous := app.get("/api/admin/export/" + file)
		assert.Equalf(t, http.StatusUnauthorized, anonymous.Status, "anonymous %s", file)
		assert.Equalf(t, "unauthenticated", anonymous.errorEnvelope().Code, "anonymous %s", file)
	}

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	for _, file := range []string{"codes.csv", "guests.csv"} {
		household := app.get("/api/admin/export/" + file)
		assert.Equalf(t, http.StatusUnauthorized, household.Status, "household session %s", file)
		assert.NotContainsf(t, household.Body, "ABC234", "%s must not leak a code to a household", file)
	}
}

// tableColumns reads a table's real columns, so the export assertions compare against
// the schema rather than against a second list somebody has to update.
func tableColumns(t *testing.T, app *testApp, table string) []string {
	t.Helper()

	var columns []string
	require.NoError(t, app.Database.Read.Select(&columns,
		fmt.Sprintf(`SELECT name FROM pragma_table_info('%s')`, table)))
	require.NotEmpty(t, columns)
	return columns
}
