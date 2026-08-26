package integration

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// concurrentQueries is above the read pool's connection cap, so the pragma checks
// below are forced across several distinct connections rather than reusing one.
const concurrentQueries = 8

func TestPragmasApplyToEveryConnection(t *testing.T) {
	database := newTestDatabase(t)

	pools := map[string]*sqlx.DB{"write": database.Write, "read": database.Read}
	for name, pool := range pools {
		t.Run(name, func(t *testing.T) {
			var waitGroup sync.WaitGroup
			for range concurrentQueries {
				waitGroup.Add(1)
				go func() {
					defer waitGroup.Done()

					var foreignKeys int
					// Errors are asserted, not fataled: t.Fatal from a goroutine
					// does not stop the test it belongs to.
					if assert.NoError(t, pool.Get(&foreignKeys, "PRAGMA foreign_keys")) {
						assert.Equal(t, 1, foreignKeys, "foreign_keys must be ON")
					}

					var journalMode string
					if assert.NoError(t, pool.Get(&journalMode, "PRAGMA journal_mode")) {
						assert.Equal(t, "wal", journalMode)
					}

					var busyTimeout int
					if assert.NoError(t, pool.Get(&busyTimeout, "PRAGMA busy_timeout")) {
						assert.Equal(t, 5000, busyTimeout)
					}
				}()
			}
			waitGroup.Wait()
		})
	}
}

// TestForeignKeyViolationIsRejected is the real proof that foreign_keys applies per
// connection: reading the pragma only shows what one connection was told, while a
// rejected insert shows the setting has teeth.
func TestForeignKeyViolationIsRejected(t *testing.T) {
	database := newTestDatabase(t)
	createParentAndChildTables(t, database.Write)

	_, err := database.Write.Exec(`INSERT INTO child (parent_id) VALUES (999)`)

	require.Error(t, err)
	assert.Contains(t, strings.ToUpper(err.Error()), "FOREIGN KEY")
}

// TestConcurrentWritesSerialise covers the failure this design exists to prevent:
// two households submitting an RSVP at the same second must both succeed, queued by
// the one-connection write pool, never rejected with SQLITE_BUSY.
func TestConcurrentWritesSerialise(t *testing.T) {
	database := newTestDatabase(t)
	createParentAndChildTables(t, database.Write)

	var waitGroup sync.WaitGroup
	for writer := range concurrentQueries {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			_, err := database.Write.Exec(`INSERT INTO parent (id, label) VALUES (?, ?)`, writer, fmt.Sprintf("writer-%d", writer))
			assert.NoError(t, err)
		}()
	}
	waitGroup.Wait()

	var rows int
	require.NoError(t, database.Write.Get(&rows, `SELECT count(*) FROM parent`))
	assert.Equal(t, concurrentQueries, rows)
}

// TestReadPoolCannotWrite guards the point of a separate read-only handle: a query
// accidentally routed through it fails here instead of silently occupying the single
// writer slot.
func TestReadPoolCannotWrite(t *testing.T) {
	database := newTestDatabase(t)
	createParentAndChildTables(t, database.Write)

	_, err := database.Read.Exec(`INSERT INTO parent (id, label) VALUES (1, 'nope')`)

	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "readonly")
}

// TestReadPoolSeesCommittedWrites is the check that the two handles are the same
// database — an easy thing to get wrong with a mistyped path or an in-memory DSN.
func TestReadPoolSeesCommittedWrites(t *testing.T) {
	database := newTestDatabase(t)
	createParentAndChildTables(t, database.Write)

	_, err := database.Write.Exec(`INSERT INTO parent (id, label) VALUES (1, 'visible')`)
	require.NoError(t, err)

	var label string
	require.NoError(t, database.Read.Get(&label, `SELECT label FROM parent WHERE id = 1`))
	assert.Equal(t, "visible", label)
}

func TestReadyReportsDatabaseReachable(t *testing.T) {
	server := newTestServer(t)

	status, contentType, body := get(t, server.URL, "/api/ready")

	assert.Equal(t, http.StatusOK, status)
	assert.True(t, strings.HasPrefix(contentType, "application/json"), "Content-Type = %q", contentType)
	assert.JSONEq(t, `{"status":"ok","database":"ok"}`, body)
}

// TestReadyReportsUnavailableWhenDatabaseClosed exercises the 503 path without
// deleting the file: a closed pool is the same observable condition as an unmounted
// volume, and it is the only one a test can produce deterministically.
func TestReadyReportsUnavailableWhenDatabaseClosed(t *testing.T) {
	database := newTestDatabase(t)
	server := newTestServerWithDatabase(t, database)
	require.NoError(t, database.Close())

	status, _, body := get(t, server.URL, "/api/ready")

	assert.Equal(t, http.StatusServiceUnavailable, status)
	assert.Contains(t, body, `"code":"not_ready"`)
	// The reason belongs in the log, not in an unauthenticated response: a driver
	// error carries the database path.
	assert.NotContains(t, body, "wedding-test.db")
}

// createParentAndChildTables sets up the smallest schema that can demonstrate a
// foreign-key violation. Written inline because the real schema arrives in E0-05.
func createParentAndChildTables(t *testing.T, pool *sqlx.DB) {
	t.Helper()

	_, err := pool.Exec(`
		CREATE TABLE parent (
			id    INTEGER PRIMARY KEY,
			label TEXT NOT NULL
		);
		CREATE TABLE child (
			id        INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL REFERENCES parent(id)
		);
	`)
	require.NoError(t, err)
}

// TestMigrationsApplyOnFreshDatabase covers the startup order main relies on: the
// harness opens an empty file and migrates it, exactly as the binary does, so a
// missing migration shows up here rather than as a missing table in a later story.
func TestMigrationsApplyOnFreshDatabase(t *testing.T) {
	database := newTestDatabase(t)

	var versions []string
	require.NoError(t, database.Read.Select(&versions, `SELECT version FROM schema_migration ORDER BY version`))

	assert.NotEmpty(t, versions, "a fresh database must come up migrated")
	// E0-05 grows this set; the assertion is on "the runner ran", not on a count.
	assert.Equal(t, "0001", versions[0])
}
