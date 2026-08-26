package persistence

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// These tests are in-package because the interesting cases — a broken migration, a
// duplicated version — must not be shippable files. migrate takes an fs.FS so they
// can be constructed in memory instead.

// newTestPool opens a real database on a temp file. Not configuration.OpenDatabase:
// the migration runner only ever touches the write handle, and this keeps the
// package's tests free of a dependency on the config layer.
func newTestPool(t *testing.T) *sqlx.DB {
	t.Helper()

	pool, err := sqlx.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "migration-test.db")+"?_pragma=foreign_keys(1)")
	require.NoError(t, err)
	pool.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, pool.Close())
	})

	return pool
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func appliedVersions(t *testing.T, pool *sqlx.DB) []string {
	t.Helper()

	var versions []string
	require.NoError(t, pool.Select(&versions, `SELECT version FROM schema_migration ORDER BY version`))
	return versions
}

func TestMigrateAppliesInVersionOrder(t *testing.T) {
	pool := newTestPool(t)
	// Deliberately out of order in the map: order must come from the file names.
	migrations := fstest.MapFS{
		"0002-add-child.sql":  {Data: []byte(`CREATE TABLE child (parent_id INTEGER NOT NULL REFERENCES parent(id))`)},
		"0001-add-parent.sql": {Data: []byte(`CREATE TABLE parent (id INTEGER PRIMARY KEY)`)},
	}

	require.NoError(t, migrate(context.Background(), pool, migrations, discardLogger()))

	assert.Equal(t, []string{"0001", "0002"}, appliedVersions(t, pool))

	var appliedAt string
	require.NoError(t, pool.Get(&appliedAt, `SELECT applied_at FROM schema_migration WHERE version = '0001'`))
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, appliedAt, "applied_at must be UTC RFC3339")
}

func TestMigrateIsIdempotent(t *testing.T) {
	pool := newTestPool(t)
	migrations := fstest.MapFS{
		"0001-add-parent.sql": {Data: []byte(`CREATE TABLE parent (id INTEGER PRIMARY KEY)`)},
	}
	ctx := context.Background()

	require.NoError(t, migrate(ctx, pool, migrations, discardLogger()))
	// A second run must not re-apply: CREATE TABLE without IF NOT EXISTS would fail,
	// which is what makes this assertion meaningful rather than decorative.
	require.NoError(t, migrate(ctx, pool, migrations, discardLogger()))

	assert.Equal(t, []string{"0001"}, appliedVersions(t, pool))
}

// TestMigrateRollsBackFailedMigration is the invariant the transaction exists for: a
// migration that fails halfway leaves neither its tables nor its version behind.
func TestMigrateRollsBackFailedMigration(t *testing.T) {
	pool := newTestPool(t)
	migrations := fstest.MapFS{
		"0001-add-parent.sql": {Data: []byte(`CREATE TABLE parent (id INTEGER PRIMARY KEY)`)},
		"0002-broken.sql": {Data: []byte(`
			CREATE TABLE half_applied (id INTEGER PRIMARY KEY);
			THIS IS NOT SQL;
		`)},
	}

	err := migrate(context.Background(), pool, migrations, discardLogger())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "0002-broken.sql")
	assert.Equal(t, []string{"0001"}, appliedVersions(t, pool), "the broken version must not be recorded")

	var tables int
	require.NoError(t, pool.Get(&tables, `SELECT count(*) FROM sqlite_master WHERE name = 'half_applied'`))
	assert.Zero(t, tables, "the failed migration's table must be rolled back")
}

// TestMigrateRefusesUnknownAppliedVersion covers the accidental rollback: an older
// binary started against a newer database.
func TestMigrateRefusesUnknownAppliedVersion(t *testing.T) {
	pool := newTestPool(t)
	migrations := fstest.MapFS{
		"0001-add-parent.sql": {Data: []byte(`CREATE TABLE parent (id INTEGER PRIMARY KEY)`)},
	}
	ctx := context.Background()
	require.NoError(t, migrate(ctx, pool, migrations, discardLogger()))

	_, err := pool.Exec(`INSERT INTO schema_migration (version, applied_at) VALUES ('0009', '2027-01-01T00:00:00Z')`)
	require.NoError(t, err)

	err = migrate(ctx, pool, migrations, discardLogger())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "0009")
	assert.Contains(t, err.Error(), "older than the database")
}

func TestMigrateRejectsMalformedNames(t *testing.T) {
	cases := map[string]fstest.MapFS{
		"missing prefix":       {"initial.sql": {Data: []byte(`SELECT 1`)}},
		"short prefix":         {"001-initial.sql": {Data: []byte(`SELECT 1`)}},
		"uppercase name":       {"0001-Initial.sql": {Data: []byte(`SELECT 1`)}},
		"duplicate version":    {"0001-a.sql": {Data: []byte(`SELECT 1`)}, "0001-b.sql": {Data: []byte(`SELECT 1`)}},
		"no migrations at all": {},
	}

	for description, migrations := range cases {
		t.Run(description, func(t *testing.T) {
			err := migrate(context.Background(), newTestPool(t), migrations, discardLogger())
			require.Error(t, err)
		})
	}
}

// TestEmbeddedMigrationsAreValid guards the shipped set, which the in-memory tests
// above never touch: names must parse, and versions must be unique.
func TestEmbeddedMigrationsAreValid(t *testing.T) {
	pool := newTestPool(t)

	require.NoError(t, Migrate(context.Background(), pool, discardLogger()))

	assert.NotEmpty(t, appliedVersions(t, pool))
}
