package persistence

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// inTransaction runs work inside one write transaction, committing on success and
// rolling back on any failure.
//
// On the write pool, which is capped at a single connection: concurrent writes queue
// in Go rather than racing in SQLite. The DSN opens transactions as IMMEDIATE, so the
// write lock is taken up front and a busy timeout can actually resolve it — see
// configuration.writeDataSourceName.
//
// A rollback failure is joined onto the original error rather than replacing it: what
// went wrong is the interesting half, and losing it to "rollback failed" is how a
// broken save becomes unexplainable.
func inTransaction(ctx context.Context, database *configuration.Database, work func(transaction *sqlx.Tx) error) error {
	transaction, err := database.Write.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	if err := work(transaction); err != nil {
		if rollbackErr := transaction.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w (rollback failed: %w)", err, rollbackErr)
		}
		return err
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
