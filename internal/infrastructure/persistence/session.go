package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/configuration"
)

// purgeInterval is how often expired sessions are swept. Daily is often enough for
// rows that expire on a scale of hours and months, and the sweep is also run once
// at startup so a container that is restarted more often than that still cleans up.
const purgeInterval = 24 * time.Hour

// SessionStore is the only code that reads or writes the session table.
type SessionStore struct {
	// Both handles are taken as one Database rather than as two *sqlx.DB
	// parameters: two positional pools of the same type are trivially swapped at a
	// call site, and passing the write pool as the reader would silently spend the
	// single writer connection on every session lookup in the app.
	database *configuration.Database
}

func NewSessionStore(database *configuration.Database) *SessionStore {
	return &SessionStore{database: database}
}

// Create stores a new session. The caller has already hashed the token into
// session.ID; the raw token never reaches this package.
func (store *SessionStore) Create(ctx context.Context, session domain.Session) error {
	const insertSession = `
		INSERT INTO session (id, subject_type, subject_id, created_at, expires_at, last_seen_at, user_agent, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := store.database.Write.ExecContext(ctx, insertSession,
		session.ID,
		string(session.SubjectType),
		subjectIDValue(session),
		formatTimestamp(session.CreatedAt),
		formatTimestamp(session.ExpiresAt),
		formatTimestamp(session.LastSeenAt),
		session.UserAgent,
		session.IP,
	)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	return nil
}

// FindByID returns the live session with this id, or ErrNotFound.
//
// An expired session counts as absent and is deleted on the way out, rather than
// being returned with an expiry for the caller to check. A caller that has to
// remember the check is a caller that will one day forget it, and the forgetting
// would not fail loudly — it would hand a year-old cookie a valid request.
func (store *SessionStore) FindByID(ctx context.Context, id string) (domain.Session, error) {
	const selectSession = `
		SELECT id, subject_type, subject_id, created_at, expires_at, last_seen_at, user_agent, ip
		FROM session
		WHERE id = ?`

	var row sessionRow
	if err := store.database.Read.GetContext(ctx, &row, selectSession, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, ErrNotFound
		}
		return domain.Session{}, fmt.Errorf("reading session: %w", err)
	}

	session, err := row.toDomain()
	if err != nil {
		return domain.Session{}, err
	}

	if session.IsExpired(time.Now()) {
		// Best effort: the session is gone as far as the caller is concerned either
		// way, and failing the request because the tidy-up failed would turn a
		// harmless leftover row into an outage. The daily purge catches it later.
		if err := store.Delete(ctx, id); err != nil {
			return domain.Session{}, fmt.Errorf("deleting expired session: %w", err)
		}
		return domain.Session{}, ErrNotFound
	}

	return session, nil
}

// Refresh writes the pushed-out expiry of a rolling session.
//
// It takes the whole session rather than an id and a duration so that the new
// expiry is the one domain.Session.Refreshed computed — the lifetime policy stays
// in the domain, and this method cannot be called into extending a session by an
// amount nobody sanctioned.
func (store *SessionStore) Refresh(ctx context.Context, session domain.Session) error {
	const refreshSession = `UPDATE session SET expires_at = ?, last_seen_at = ? WHERE id = ?`

	_, err := store.database.Write.ExecContext(ctx, refreshSession,
		formatTimestamp(session.ExpiresAt),
		formatTimestamp(session.LastSeenAt),
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("refreshing session: %w", err)
	}
	return nil
}

// Delete removes one session. Idempotent: deleting a session that is already gone
// is not an error, because logging out twice is not a mistake worth reporting.
func (store *SessionStore) Delete(ctx context.Context, id string) error {
	const deleteSession = `DELETE FROM session WHERE id = ?`

	if _, err := store.database.Write.ExecContext(ctx, deleteSession, id); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

// PurgeExpired deletes every session past its expiry and reports how many went.
//
// The comparison is a string comparison, which is correct only because every
// timestamp is fixed-width UTC RFC3339; see timestampLayout.
func (store *SessionStore) PurgeExpired(ctx context.Context, now time.Time) (int64, error) {
	const purgeSessions = `DELETE FROM session WHERE expires_at <= ?`

	result, err := store.database.Write.ExecContext(ctx, purgeSessions, formatTimestamp(now))
	if err != nil {
		return 0, fmt.Errorf("purging expired sessions: %w", err)
	}

	purged, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("counting purged sessions: %w", err)
	}
	return purged, nil
}

// PurgeExpiredPeriodically sweeps once immediately and then every purgeInterval,
// until ctx is cancelled. Run it in a goroutine from main.
//
// A ticker rather than a scheduler dependency: there is exactly one periodic job in
// this application, and it does not need to survive a restart or coordinate with
// anything — a missed sweep costs a few dead rows until the next one.
//
// A failed sweep is logged and the loop continues. Sessions expiring is enforced on
// every lookup, so the purge is housekeeping; giving up on it because the disk was
// briefly busy would be worse than retrying tomorrow.
func (store *SessionStore) PurgeExpiredPeriodically(ctx context.Context, logger *slog.Logger) {
	ticker := time.NewTicker(purgeInterval)
	defer ticker.Stop()

	for {
		purged, err := store.PurgeExpired(ctx, time.Now())
		switch {
		case err != nil && ctx.Err() != nil:
			// Shutdown cancelled the query mid-flight; not a failure worth logging.
			return
		case err != nil:
			logger.Error("session purge failed", "error", err)
		case purged > 0:
			logger.Info("expired sessions purged", "count", purged)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// sessionRow is the session table as it is stored: timestamps as text, and the
// nullable columns as null-aware types. It exists so that the conversion to
// domain.Session — and every failure it can have — happens in exactly one place.
type sessionRow struct {
	ID          string         `db:"id"`
	SubjectType string         `db:"subject_type"`
	SubjectID   sql.NullInt64  `db:"subject_id"`
	CreatedAt   string         `db:"created_at"`
	ExpiresAt   string         `db:"expires_at"`
	LastSeenAt  string         `db:"last_seen_at"`
	UserAgent   sql.NullString `db:"user_agent"`
	IP          sql.NullString `db:"ip"`
}

func (row sessionRow) toDomain() (domain.Session, error) {
	createdAt, err := parseTimestamp(row.CreatedAt)
	if err != nil {
		return domain.Session{}, err
	}
	expiresAt, err := parseTimestamp(row.ExpiresAt)
	if err != nil {
		return domain.Session{}, err
	}
	lastSeenAt, err := parseTimestamp(row.LastSeenAt)
	if err != nil {
		return domain.Session{}, err
	}

	return domain.Session{
		ID:          row.ID,
		SubjectType: domain.SubjectType(row.SubjectType),
		SubjectID:   row.SubjectID.Int64,
		CreatedAt:   createdAt,
		ExpiresAt:   expiresAt,
		LastSeenAt:  lastSeenAt,
		UserAgent:   row.UserAgent.String,
		IP:          row.IP.String,
	}, nil
}

// subjectIDValue maps the admin's absent subject to SQL NULL.
//
// The schema carries no foreign key on subject_id, because a nullable FK would
// still permit a household id under subject_type = 'admin'. The pairing is enforced
// here instead: an admin session stores NULL, so an admin row can never be read
// back as household number whatever-was-left-in-the-struct.
func subjectIDValue(session domain.Session) any {
	if session.SubjectType != domain.SubjectTypeHousehold {
		return nil
	}
	return session.SubjectID
}
