package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/security"
)

// newSessionFor builds a session for a household with an expiry the caller
// controls, which is what the expiry and purge tests need and what
// domain.NewSession — correctly — does not offer.
func newSessionFor(householdID int64, token string, expiresAt, lastSeenAt time.Time) domain.Session {
	return domain.Session{
		ID:          security.HashSessionToken(token),
		SubjectType: domain.SubjectTypeHousehold,
		SubjectID:   householdID,
		CreatedAt:   lastSeenAt,
		ExpiresAt:   expiresAt,
		LastSeenAt:  lastSeenAt,
		UserAgent:   "Mozilla/5.0 (Test)",
		IP:          "10.0.0.1",
	}
}

func TestSessionStoreRoundTripsASession(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	store := app.sessionStore()
	household := seedHousehold(t, app.Database.Write)

	token := security.NewSessionToken()
	created := domain.NewSession(security.HashSessionToken(token), domain.SubjectTypeHousehold, household.ID, time.Now(), "Mozilla/5.0 (Test)", "10.0.0.1")
	require.NoError(t, store.Create(context.Background(), created))

	// Looked up by hashing the presented token, exactly as a request does.
	found, err := store.FindByID(context.Background(), security.HashSessionToken(token))
	require.NoError(t, err)

	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, domain.SubjectTypeHousehold, found.SubjectType)
	assert.Equal(t, created.UserAgent, found.UserAgent)
	assert.Equal(t, created.IP, found.IP)

	householdID, isHousehold := found.HouseholdID()
	assert.True(t, isHousehold)
	assert.Equal(t, household.ID, householdID)

	// Second precision, because that is what the stored format carries. Asserting
	// on equality would fail on the nanoseconds time.Now() brings and the column
	// deliberately does not.
	assert.WithinDuration(t, created.ExpiresAt, found.ExpiresAt, time.Second)
}

// The point of hashing: a database that leaks must not hand over live sessions.
func TestSessionStoreNeverStoresTheRawToken(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	household := seedHousehold(t, app.Database.Write)

	token := security.NewSessionToken()
	require.NoError(t, app.sessionStore().Create(context.Background(),
		domain.NewSession(security.HashSessionToken(token), domain.SubjectTypeHousehold, household.ID, time.Now(), "", "")))

	// Every column of the row, concatenated, so this keeps working when a column
	// is added — a token copied into a new field would still be caught.
	var wholeRow string
	require.NoError(t, app.Database.Read.Get(&wholeRow,
		`SELECT id || '|' || subject_type || '|' || COALESCE(subject_id, '') || '|' || created_at
		 || '|' || expires_at || '|' || last_seen_at || '|' || COALESCE(user_agent, '') || '|' || COALESCE(ip, '')
		 FROM session`))

	assert.NotContains(t, wholeRow, token)
}

// An expired session must not be returned, and must not be left behind either:
// the row is deleted on the way out so a stale cookie stops costing a lookup.
func TestSessionStoreTreatsAnExpiredSessionAsAbsentAndDeletesIt(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	store := app.sessionStore()
	household := seedHousehold(t, app.Database.Write)

	token := security.NewSessionToken()
	expired := newSessionFor(household.ID, token, time.Now().Add(-time.Minute), time.Now().Add(-2*time.Hour))
	require.NoError(t, store.Create(context.Background(), expired))

	_, err := store.FindByID(context.Background(), expired.ID)

	assert.ErrorIs(t, err, persistence.ErrNotFound)
	assert.Equal(t, 0, app.countSessions(), "the expired row should be gone")
}

func TestSessionStoreRefreshMovesTheExpiry(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	store := app.sessionStore()
	household := seedHousehold(t, app.Database.Write)

	token := security.NewSessionToken()
	yesterday := time.Now().Add(-25 * time.Hour)
	session := newSessionFor(household.ID, token, yesterday.Add(365*24*time.Hour), yesterday)
	require.NoError(t, store.Create(context.Background(), session))

	refreshed := session.Refreshed(time.Now())
	require.NoError(t, store.Refresh(context.Background(), refreshed))

	found, err := store.FindByID(context.Background(), session.ID)
	require.NoError(t, err)

	assert.True(t, found.ExpiresAt.After(session.ExpiresAt), "expiry should have moved forward")
	assert.WithinDuration(t, refreshed.LastSeenAt, found.LastSeenAt, time.Second)
}

// Revocation has to be immediate and real. This is the property that makes opaque
// database sessions worth their write cost over a stateless token.
func TestSessionStoreDeleteRevokesImmediately(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	store := app.sessionStore()
	household := seedHousehold(t, app.Database.Write)

	token := security.NewSessionToken()
	session := domain.NewSession(security.HashSessionToken(token), domain.SubjectTypeHousehold, household.ID, time.Now(), "", "")
	require.NoError(t, store.Create(context.Background(), session))

	require.NoError(t, store.Delete(context.Background(), session.ID))

	_, err := store.FindByID(context.Background(), session.ID)
	assert.ErrorIs(t, err, persistence.ErrNotFound)

	// Deleting again is not an error: logging out twice is not a mistake.
	assert.NoError(t, store.Delete(context.Background(), session.ID))
}

func TestSessionStorePurgeRemovesOnlyExpiredRows(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	store := app.sessionStore()
	household := seedHousehold(t, app.Database.Write)

	live := newSessionFor(household.ID, security.NewSessionToken(), time.Now().Add(time.Hour), time.Now())
	expired := newSessionFor(household.ID, security.NewSessionToken(), time.Now().Add(-time.Hour), time.Now().Add(-2*time.Hour))
	require.NoError(t, store.Create(context.Background(), live))
	require.NoError(t, store.Create(context.Background(), expired))

	purged, err := store.PurgeExpired(context.Background(), time.Now())
	require.NoError(t, err)

	assert.Equal(t, int64(1), purged)
	assert.Equal(t, 1, app.countSessions())

	_, err = store.FindByID(context.Background(), live.ID)
	assert.NoError(t, err, "the live session must survive the purge")
}

// The admin has no household row to point at, so subject_id is NULL. Reading it
// back as household 0 would be the bug this stores against.
func TestSessionStoreStoresTheAdminWithoutASubjectID(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	store := app.sessionStore()

	token := security.NewSessionToken()
	admin := domain.NewSession(security.HashSessionToken(token), domain.SubjectTypeAdmin, 0, time.Now(), "", "")
	require.NoError(t, store.Create(context.Background(), admin))

	var subjectIDIsNull bool
	require.NoError(t, app.Database.Read.Get(&subjectIDIsNull, `SELECT subject_id IS NULL FROM session WHERE id = ?`, admin.ID))
	assert.True(t, subjectIDIsNull)

	found, err := store.FindByID(context.Background(), admin.ID)
	require.NoError(t, err)

	_, isHousehold := found.HouseholdID()
	assert.False(t, isHousehold)
}

func TestSessionStorePurgePeriodicallySweepsOnceAtStartupAndStops(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	store := app.sessionStore()
	household := seedHousehold(t, app.Database.Write)

	expired := newSessionFor(household.ID, security.NewSessionToken(), time.Now().Add(-time.Hour), time.Now().Add(-2*time.Hour))
	require.NoError(t, store.Create(context.Background(), expired))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		store.PurgeExpiredPeriodically(ctx, testLogger(app.Logs).Logger)
	}()

	// The sweep runs immediately rather than waiting out the first tick, so a
	// container that restarts more often than daily still cleans up.
	require.Eventually(t, func() bool { return app.countSessions() == 0 }, time.Second, 5*time.Millisecond)

	// And it stops on cancellation instead of outliving the process it belongs to.
	cancel()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("purge loop did not stop when its context was cancelled")
	}
}
