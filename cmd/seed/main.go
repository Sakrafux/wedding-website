// Command seed inserts households with working login codes into the local
// development database. It exists for the dev shell alone.
//
// # Development only
//
// This is not an administration tool and must never be pointed at the deployed
// volume. It writes obviously synthetic households and prints their login codes —
// the one secret this application has — in plain text to stdout. The Dockerfile
// builds ./cmd/wedding only, so the binary does not exist in the image; the whole
// guard against running it in production is that it is not there.
//
// Households and codes are created for real elsewhere: F5-B01 for the admin CRUD
// endpoints, F5-B03 for per-household code generation and regeneration.
//
// Usage:
//
//	make seed                 # one household, two members
//	make seed SEED_ARGS='-households 5 -guests 3'
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
)

// codeAttempts bounds the retry on a colliding login code. A collision needs two
// of 32^6 codes to match, so anything above a couple of attempts is a symptom of
// something else entirely — a broken generator, say — and looping forever would
// hide it.
const codeAttempts = 5

// firstNames supplies member names, cycled. Deliberately not random: a seeded
// household should look the same shape every run, and nothing here depends on the
// names being varied.
var firstNames = []string{"Anna", "Bernd", "Clara", "Dieter", "Erika", "Frank"}

func main() {
	households := flag.Int("households", 1, "number of households to insert")
	guests := flag.Int("guests", 2, "number of adult members per household")
	flag.Parse()

	if err := run(*households, *guests, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "seed failed:", err)
		os.Exit(1)
	}
}

// run inserts the households and writes the report to out.
//
// Separate from main so the test can call it with a buffer and without exiting the
// process. No structured logger anywhere in this command: its output is read by a
// person in a terminal, not by a log collector.
func run(households, guests int, out io.Writer) error {
	if households < 1 || guests < 1 {
		return fmt.Errorf("-households and -guests must both be at least 1, got %d and %d", households, guests)
	}

	// Same environment as the app, so DB_PATH cannot drift between `make run` and
	// `make seed` and quietly seed a different file than the one being served.
	config, err := configuration.Load()
	if err != nil {
		return err
	}

	database, err := configuration.OpenDatabase(config)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "closing database failed:", err)
		}
	}()

	ctx := context.Background()

	// Migrating here means the command works against a database file that has never
	// been served, which is the common case: clone, `make seed`, `make run`.
	// slog.DiscardHandler keeps the migration runner's per-file logging out of a
	// report whose whole value is being short enough to read.
	if err := persistence.Migrate(ctx, database.Write, slog.New(slog.DiscardHandler)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Seeding %d household(s) with %d member(s) each into %s\n", households, guests, config.DatabasePath)
	fmt.Fprintln(out, "DEVELOPMENT ONLY — the codes below are printed in plain text.")

	existing, err := countHouseholds(ctx, database.Write)
	if err != nil {
		return err
	}

	for index := 1; index <= households; index++ {
		// Numbered past what is already there, so a second run does not produce a
		// second "Familie Testhaushalt 1" and leave two identical-looking rows.
		name := fmt.Sprintf("Familie Testhaushalt %d", existing+int64(index))

		id, code, err := insertHousehold(ctx, database.Write, name, guests)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "  #%d  %-28s  %s\n", id, name, domain.FormatCode(code))
	}
	return nil
}

func countHouseholds(ctx context.Context, writePool *sqlx.DB) (int64, error) {
	var count int64
	if err := writePool.GetContext(ctx, &count, `SELECT COUNT(*) FROM household`); err != nil {
		return 0, fmt.Errorf("counting households: %w", err)
	}
	return count, nil
}

// insertHousehold inserts one household and its adult members in a single
// transaction, returning the new id and the stored (undashed) code.
//
// The SQL sits in this command rather than in persistence on purpose: F5-B01 gives
// HouseholdStore a create method for the admin endpoints, and this function calls
// that instead once it exists. Adding it early would mean guessing the signature
// that story needs.
//
// One transaction per household, not one for the whole run: an interrupted seed
// then leaves whole households behind rather than a household with half its
// members, which is a state no later code is written to expect.
func insertHousehold(ctx context.Context, writePool *sqlx.DB, name string, guests int) (int64, string, error) {
	const insertHouseholdRow = `INSERT INTO household (display_name, code) VALUES (?, ?)`
	const insertGuestRow = `INSERT INTO guest (household_id, first_name, last_name, kind, origin) VALUES (?, ?, ?, 'adult', 'seeded')`

	// The household surname doubles as the members' last name, number included:
	// "Anna Testhaushalt 3" says which household a seeded person belongs to in every
	// list that shows people without their household.
	lastName := strings.TrimPrefix(name, "Familie ")

	for attempt := 1; ; attempt++ {
		code := domain.GenerateCode()

		id, err := insertInTransaction(ctx, writePool, func(tx *sqlx.Tx) (int64, error) {
			result, err := tx.ExecContext(ctx, insertHouseholdRow, name, code)
			if err != nil {
				return 0, err
			}
			householdID, err := result.LastInsertId()
			if err != nil {
				return 0, err
			}

			for member := 0; member < guests; member++ {
				firstName := firstNames[member%len(firstNames)]
				if _, err := tx.ExecContext(ctx, insertGuestRow, householdID, firstName, lastName); err != nil {
					return 0, err
				}
			}
			return householdID, nil
		})

		if err == nil {
			return id, code, nil
		}
		// The UNIQUE index on household.code is the only authority on code
		// uniqueness — the generator never asks the database for a free code first,
		// because the answer would be stale by the time it inserted. A rejected
		// insert therefore means "try another code", not "fail".
		if !isUniqueViolation(err) {
			return 0, "", fmt.Errorf("inserting household %q: %w", name, err)
		}
		if attempt == codeAttempts {
			return 0, "", fmt.Errorf("inserting household %q: %d colliding codes in a row: %w", name, codeAttempts, err)
		}
	}
}

// insertInTransaction runs body in a transaction, rolling back on any error.
func insertInTransaction(ctx context.Context, writePool *sqlx.DB, body func(*sqlx.Tx) (int64, error)) (int64, error) {
	tx, err := writePool.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}

	id, err := body(tx)
	if err != nil {
		// Rollback error is joined rather than dropped: on a WAL database a failed
		// rollback means the connection is in a state worth seeing.
		return 0, errors.Join(err, tx.Rollback())
	}
	return id, tx.Commit()
}

// isUniqueViolation reports whether err is SQLite rejecting a duplicate key.
//
// Matched on the message rather than on a driver error type: modernc.org/sqlite
// exposes its codes only through its own error struct, and importing the driver
// package here to type-assert would couple this command to it for a single string
// comparison. A dev tool retrying a code is the cheapest possible caller of this
// distinction; if anything in the request path ever needs it, the mapping belongs
// in persistence next to ErrNotFound instead.
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
