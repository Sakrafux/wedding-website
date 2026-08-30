package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessfulLoginIsAudited(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)

	rows := app.auditRows()
	require.Len(t, rows, 1)

	assert.Equal(t, "household", rows[0].ActorType)
	assert.Equal(t, household.ID, rows[0].ActorID.Int64)
	assert.Equal(t, "household", rows[0].Entity)
	assert.Equal(t, household.ID, rows[0].EntityID)
	assert.Equal(t, "login", rows[0].Action)
	assert.False(t, rows[0].Before.Valid, "a login changed nothing, so there is no before")

	at, err := time.Parse(time.RFC3339, rows[0].At)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), at, time.Minute)
}

// Nobody has been identified, which is what "failed" means — so the actor is the
// system rather than a household we would only be guessing at.
func TestFailedLoginIsAuditedAsSystem(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusUnauthorized, app.logIn("ZZZZZZ").Status)

	rows := app.auditRows()
	require.Len(t, rows, 1)

	assert.Equal(t, "system", rows[0].ActorType)
	assert.False(t, rows[0].ActorID.Valid, "a failed login has no actor to name")
	assert.Equal(t, "login_failed", rows[0].Action)
	assert.Equal(t, "household", rows[0].Entity, "which door was tried")
	assert.Equal(t, int64(0), rows[0].EntityID)
}

// A malformed code never reaches the database, but it is still an attempt and must
// still be recorded — otherwise the log would show only the near-misses.
func TestMalformedLoginIsAuditedToo(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	require.Equal(t, http.StatusUnauthorized, app.logIn("nonsense!").Status)

	rows := app.auditRows()
	require.Len(t, rows, 1)
	assert.Equal(t, "login_failed", rows[0].Action)
}

// The assertion this story exists for. A log of near-misses is a partial key list,
// and a log of typos will eventually hold another household's real code — typed by
// a guest holding the wrong card.
func TestNoAuditRowEverContainsASubmittedCode(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write, withCode("ABC234"))

	submitted := []string{"ABC234", "abc-234", "ABC235", "ZZZZZZ", "wrong-one"}
	for _, code := range submitted {
		app.logIn(code)
	}

	// Every column of every row, concatenated: a code copied into a field nobody
	// thought to check would still be caught.
	var wholeTable string
	require.NoError(t, app.Database.Read.Get(&wholeTable,
		`SELECT COALESCE(GROUP_CONCAT(at || '|' || actor_type || '|' || COALESCE(actor_id, '') || '|' || entity
		 || '|' || entity_id || '|' || action || '|' || COALESCE(before, '') || '|' || COALESCE(after, '')), '')
		 FROM audit_log`))

	require.NotEmpty(t, wholeTable, "the attempts should have been recorded at all")
	for _, code := range submitted {
		assert.NotContainsf(t, wholeTable, code, "the audit log must never carry a submitted code (%q)", code)
	}
	assert.NotContains(t, wholeTable, household.Code)
}

// Consulted by hand, months later, in sqlite3. Payloads that are not valid JSON
// would make that unreadable exactly when it matters.
func TestAuditPayloadsAreValidJSONOrNull(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	require.Equal(t, http.StatusOK, app.logIn("ABC234").Status)
	require.Equal(t, http.StatusUnauthorized, app.logIn("ZZZZZZ").Status)
	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)

	rows := app.auditRows()
	require.Len(t, rows, 3)

	for _, row := range rows {
		for name, payload := range map[string]struct {
			Valid  bool
			String string
		}{"before": {row.Before.Valid, row.Before.String}, "after": {row.After.Valid, row.After.String}} {
			if !payload.Valid {
				continue
			}
			assert.Truef(t, json.Valid([]byte(payload.String)), "%s is not valid JSON: %s", name, payload.String)
		}

		// The connection details are what makes a run of failures legible later.
		var after map[string]any
		require.NoError(t, json.Unmarshal([]byte(row.After.String), &after))
		assert.Contains(t, after, "ip")
		assert.Contains(t, after, "user_agent")
	}
}

// A broken audit table must not stop a guest from logging in. Forced by dropping
// the table out from under the running application, which is as close to a real
// write failure as a test can get without a fake store.
func TestLoginSurvivesAFailingAuditWrite(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	seedHousehold(t, app.Database.Write, withCode("ABC234"))

	_, err := app.Database.Write.Exec(`DROP TABLE audit_log`)
	require.NoError(t, err)

	response := app.logIn("ABC234")

	assert.Equal(t, http.StatusOK, response.Status, "a login must not depend on the audit table")
	assert.Equal(t, 1, app.countSessions())
}

func TestAdminLoginIsAudited(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	require.Equal(t, http.StatusOK, app.logInAsAdmin().Status)

	rows := app.auditRows()
	require.Len(t, rows, 1)

	assert.Equal(t, "admin", rows[0].ActorType)
	assert.False(t, rows[0].ActorID.Valid, "there is no admin row to point at")
	assert.Equal(t, "admin", rows[0].Entity)
	assert.Equal(t, "login", rows[0].Action)
}
