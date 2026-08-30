package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// AuditStore appends to audit_log. There is no update and no delete: the table is
// append-only, which is the only property that makes it worth consulting later.
type AuditStore struct {
	database *configuration.Database
}

func NewAuditStore(database *configuration.Database) *AuditStore {
	return &AuditStore{database: database}
}

// Write appends one entry.
//
// Callers treat a failure here as non-fatal — a broken audit table must not stop a
// guest from logging in — so the error is returned for logging rather than for
// aborting. That decision belongs to the caller and not to this method, which would
// otherwise be a store that silently swallows its own failures.
func (store *AuditStore) Write(ctx context.Context, entry domain.AuditEntry) error {
	const insertEntry = `
		INSERT INTO audit_log (at, actor_type, actor_id, entity, entity_id, action, before, after)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	before, err := auditPayload(entry.Before)
	if err != nil {
		return err
	}
	after, err := auditPayload(entry.After)
	if err != nil {
		return err
	}

	_, err = store.database.Write.ExecContext(ctx, insertEntry,
		formatTimestamp(entry.At),
		string(entry.ActorType),
		entry.ActorID,
		entry.Entity,
		entry.EntityID,
		string(entry.Action),
		before,
		after,
	)
	if err != nil {
		return fmt.Errorf("writing audit entry: %w", err)
	}
	return nil
}

// auditPayload encodes a snapshot, mapping an absent one to SQL NULL rather than to
// the string "null" — so that `WHERE before IS NULL` means what a reader expects
// when they are picking through this table by hand at some point in 2027.
func auditPayload(fields map[string]any) (any, error) {
	if fields == nil {
		return nil, nil
	}

	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("encoding audit payload: %w", err)
	}
	return string(encoded), nil
}
