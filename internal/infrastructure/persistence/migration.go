package persistence

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// migrationFiles carries the schema with the binary, so the schema a container runs
// against can never drift from the code that queries it. Deploying is a restart.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrationFileName matches `NNNN-name.sql`. The four-digit zero-padded prefix is
// what makes lexical order the same as numeric order; anything else in the directory
// is a mistake worth failing on rather than skipping silently.
var migrationFileName = regexp.MustCompile(`^(\d{4})-[a-z0-9-]+\.sql$`)

// Migrate applies every embedded migration the database has not seen yet.
//
// Called from main before the server listens: a half-migrated database serving
// requests is worse than a container that refuses to start.
func Migrate(ctx context.Context, writePool *sqlx.DB, logger *slog.Logger) error {
	embedded, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("reading embedded migrations: %w", err)
	}
	return migrate(ctx, writePool, embedded, logger)
}

// migrate is Migrate over any filesystem, which is how the tests feed it a broken
// migration without shipping one.
func migrate(ctx context.Context, writePool *sqlx.DB, migrations fs.FS, logger *slog.Logger) error {
	if err := createSchemaMigrationTable(ctx, writePool); err != nil {
		return err
	}

	available, err := readAvailableMigrations(migrations)
	if err != nil {
		return err
	}

	applied, err := readAppliedVersions(ctx, writePool)
	if err != nil {
		return err
	}

	// A version in the database that the binary does not know is a rollback nobody
	// meant to perform: the running code is older than the schema it is about to
	// query. Failing here turns a subtle "column does not exist" into a clear
	// startup error.
	if unknown := unknownVersions(applied, available); len(unknown) > 0 {
		return fmt.Errorf("database has migration versions this binary does not know (%s): the binary is older than the database", strings.Join(unknown, ", "))
	}

	pending := 0
	for _, migration := range available {
		if slices.Contains(applied, migration.version) {
			continue
		}
		if err := applyMigration(ctx, writePool, migrations, migration); err != nil {
			return err
		}

		pending++
		logger.Info("migration applied", "version", migration.version, "file", migration.fileName)
	}

	// Silence on startup is indistinguishable from a bug, so the no-op case says so.
	if pending == 0 {
		logger.Info("schema is current", "version", latestVersion(available))
	}
	return nil
}

// migration is one embedded file, identified by its numeric prefix.
type migration struct {
	// version is the four-digit prefix only, not the whole file name: it is what
	// lands in schema_migration, and keying on the prefix lets a file be renamed for
	// clarity without the database considering it a different, unapplied migration.
	version  string
	fileName string
}

// createSchemaMigrationTable is step zero and runs outside the migration sequence —
// the table has to exist before anything can be recorded as applied.
func createSchemaMigrationTable(ctx context.Context, writePool *sqlx.DB) error {
	const createTable = `
		CREATE TABLE IF NOT EXISTS schema_migration (
			version    TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		)`

	if _, err := writePool.ExecContext(ctx, createTable); err != nil {
		return fmt.Errorf("creating schema_migration: %w", err)
	}
	return nil
}

// readAvailableMigrations lists the embedded migrations in version order, rejecting
// a malformed name or a duplicated version rather than guessing what was meant.
func readAvailableMigrations(migrations fs.FS) ([]migration, error) {
	names, err := fs.Glob(migrations, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("listing migrations: %w", err)
	}
	// Glob returns sorted names, and the zero-padded prefix makes that numeric order.
	slices.Sort(names)

	available := make([]migration, 0, len(names))
	for _, name := range names {
		match := migrationFileName.FindStringSubmatch(name)
		if match == nil {
			return nil, fmt.Errorf("migration %q does not match NNNN-name.sql", name)
		}

		version := match[1]
		if index := slices.IndexFunc(available, func(other migration) bool { return other.version == version }); index >= 0 {
			return nil, fmt.Errorf("migrations %q and %q share version %s", available[index].fileName, name, version)
		}

		available = append(available, migration{version: version, fileName: name})
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no migrations found")
	}
	return available, nil
}

func readAppliedVersions(ctx context.Context, writePool *sqlx.DB) ([]string, error) {
	var applied []string
	if err := writePool.SelectContext(ctx, &applied, `SELECT version FROM schema_migration`); err != nil {
		return nil, fmt.Errorf("reading applied migrations: %w", err)
	}
	return applied, nil
}

func unknownVersions(applied []string, available []migration) []string {
	var unknown []string
	for _, version := range applied {
		if !slices.ContainsFunc(available, func(candidate migration) bool { return candidate.version == version }) {
			unknown = append(unknown, version)
		}
	}
	slices.Sort(unknown)
	return unknown
}

// applyMigration runs one file and records it in the same transaction.
//
// Both together or neither: a migration that half-applied must not be recordable as
// applied, because the next start would skip the rest of it and the schema would be
// wrong in a way nothing detects.
func applyMigration(ctx context.Context, writePool *sqlx.DB, migrations fs.FS, migration migration) error {
	statements, err := fs.ReadFile(migrations, migration.fileName)
	if err != nil {
		return fmt.Errorf("reading migration %s: %w", migration.fileName, err)
	}

	transaction, err := writePool.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction for migration %s: %w", migration.fileName, err)
	}
	// Rollback after a successful commit is a no-op error we deliberately ignore;
	// the alternative is a named error variable in every exit path below.
	defer transaction.Rollback() //nolint:errcheck

	if _, err := transaction.ExecContext(ctx, string(statements)); err != nil {
		return fmt.Errorf("applying migration %s: %w", migration.fileName, err)
	}

	const recordMigration = `INSERT INTO schema_migration (version, applied_at) VALUES (?, ?)`
	if _, err := transaction.ExecContext(ctx, recordMigration, migration.version, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("recording migration %s: %w", migration.fileName, err)
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing migration %s: %w", migration.fileName, err)
	}
	return nil
}

func latestVersion(available []migration) string {
	return available[len(available)-1].version
}
