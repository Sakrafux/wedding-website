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
// Households and codes are created for real through the admin endpoints; this
// command drives the same stores they do, so a seeded household is
// indistinguishable from one an admin typed in.
//
// Usage:
//
//	make seed                 # one household, two members
//	make seed SEED_ARGS='-households 5 -guests 3'
package main

import (
	"context"
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

// firstNames supplies the given-name half of a seeded member's name, cycled. Deliberately not random: a seeded
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

	// The same stores the admin endpoints use, so this command cannot drift into
	// creating households the application would not: the login code, the origin and
	// the defaults all come from one place.
	householdStore := persistence.NewHouseholdStore(database)
	guestStore := persistence.NewGuestStore(database)

	existing, err := countHouseholds(ctx, database.Write)
	if err != nil {
		return err
	}

	for index := 1; index <= households; index++ {
		// Numbered past what is already there, so a second run does not produce a
		// second "Familie Testhaushalt 1" and leave two identical-looking rows.
		name := fmt.Sprintf("Familie Testhaushalt %d", existing+int64(index))

		household, err := insertHousehold(ctx, householdStore, guestStore, name, guests)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "  #%d  %-28s  %s\n", household.ID, household.DisplayName, household.Code)
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

// insertHousehold inserts one household and its adult members through the stores.
//
// Not transactional across the two: the stores write a row at a time, which is what
// the admin endpoints need, and an interrupted seed therefore leaves a household with
// fewer members than asked for. For a dev tool that is a re-run, whereas a
// transactional creation path used by nothing else would be a second way to create a
// household — and the one that quietly falls behind.
func insertHousehold(
	ctx context.Context,
	households *persistence.HouseholdStore,
	guests *persistence.GuestStore,
	name string,
	members int,
) (domain.Household, error) {
	household, err := households.Create(ctx, domain.Household{DisplayName: name})
	if err != nil {
		return domain.Household{}, fmt.Errorf("inserting household %q: %w", name, err)
	}

	// The household name, minus its prefix, doubles as the members' surname, number
	// included: "Anna Testhaushalt 3" says which household a seeded person belongs to
	// in every list that shows people without their household. Only this command may
	// derive a name that way — it invents the household names in the first place. The
	// admin form does not, because a real household is named anything at all.
	surname := strings.TrimPrefix(name, "Familie ")

	for member := range members {
		_, err := guests.Create(ctx, domain.Guest{
			HouseholdID: household.ID,
			Name:        firstNames[member%len(firstNames)] + " " + surname,
			Kind:        domain.GuestKindAdult,
			Origin:      domain.GuestOriginSeeded,
			SeatingNeed: domain.SeatingNeedNormal,
		})
		if err != nil {
			return domain.Household{}, fmt.Errorf("inserting a member of %q: %w", name, err)
		}
	}

	return household, nil
}
