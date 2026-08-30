package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// A fixed instant rather than time.Now(), so a failure is reproducible and no test
// here depends on how long the suite took to get this far.
var testNow = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

func TestSessionLifetimeDependsOnSubjectType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 365*24*time.Hour, SessionLifetime(SubjectTypeHousehold))
	assert.Equal(t, 8*time.Hour, SessionLifetime(SubjectTypeAdmin))
}

// An unrecognised subject type can only come from a database column, and the safe
// reading of one is the shorter life. The assertion is against the admin lifetime
// rather than "not a year" so that it keeps saying something if the values change.
func TestSessionLifetimeOfUnknownSubjectTypeIsTheShortOne(t *testing.T) {
	t.Parallel()

	assert.Equal(t, adminSessionLifetime, SessionLifetime(SubjectType("something else")))
}

func TestNewSessionExpiresAfterItsTypesLifetime(t *testing.T) {
	t.Parallel()

	household := NewSession("hash", SubjectTypeHousehold, 12, testNow, "Firefox", "10.0.0.1")
	admin := NewSession("hash", SubjectTypeAdmin, 0, testNow, "Firefox", "10.0.0.1")

	assert.Equal(t, testNow.Add(365*24*time.Hour), household.ExpiresAt)
	assert.Equal(t, testNow.Add(8*time.Hour), admin.ExpiresAt)
	assert.Equal(t, testNow, household.LastSeenAt)
}

func TestHouseholdIDIsOnlyReadableForHouseholdSessions(t *testing.T) {
	t.Parallel()

	household := NewSession("hash", SubjectTypeHousehold, 12, testNow, "", "")
	householdID, isHousehold := household.HouseholdID()
	assert.True(t, isHousehold)
	assert.Equal(t, int64(12), householdID)

	// The admin's SubjectID is 0, which is exactly the value a careless caller
	// would treat as a household id. The second return is what stops that.
	admin := NewSession("hash", SubjectTypeAdmin, 0, testNow, "", "")
	_, isHousehold = admin.HouseholdID()
	assert.False(t, isHousehold)
}

func TestIsExpiredIncludesTheExpiryInstantItself(t *testing.T) {
	t.Parallel()

	session := NewSession("hash", SubjectTypeAdmin, 0, testNow, "", "")

	assert.False(t, session.IsExpired(session.ExpiresAt.Add(-time.Second)))
	// The boundary counts as expired: a session valid at exactly its expiry would
	// be a session that outlives the moment it was promised to end.
	assert.True(t, session.IsExpired(session.ExpiresAt))
	assert.True(t, session.IsExpired(session.ExpiresAt.Add(time.Second)))
}

func TestNeedsRefreshIsThrottledToOnceADay(t *testing.T) {
	t.Parallel()

	session := NewSession("hash", SubjectTypeHousehold, 12, testNow, "", "")

	assert.False(t, session.NeedsRefresh(testNow), "a session created now has just been seen")
	assert.False(t, session.NeedsRefresh(testNow.Add(23*time.Hour)))
	assert.True(t, session.NeedsRefresh(testNow.Add(24*time.Hour)))
	assert.True(t, session.NeedsRefresh(testNow.Add(72*time.Hour)))
}

// An admin session that rolled on use would never end, which is the property the
// eight-hour lifetime exists to deny.
func TestAdminSessionsNeverNeedRefresh(t *testing.T) {
	t.Parallel()

	session := NewSession("hash", SubjectTypeAdmin, 0, testNow, "", "")

	assert.False(t, session.NeedsRefresh(testNow.Add(365*24*time.Hour)))
}

func TestRefreshedMovesExpiryAndLastSeen(t *testing.T) {
	t.Parallel()

	session := NewSession("hash", SubjectTypeHousehold, 12, testNow, "Firefox", "10.0.0.1")
	later := testNow.Add(30 * 24 * time.Hour)

	refreshed := session.Refreshed(later)

	assert.Equal(t, later.Add(365*24*time.Hour), refreshed.ExpiresAt)
	assert.Equal(t, later, refreshed.LastSeenAt)
	// Everything identifying stays put: refreshing is not re-issuing.
	assert.Equal(t, session.ID, refreshed.ID)
	assert.Equal(t, session.CreatedAt, refreshed.CreatedAt)
	assert.Equal(t, session.UserAgent, refreshed.UserAgent)
	// And the receiver is untouched, since Session is a value type and a caller
	// that ignores the return must not have mutated anything.
	assert.Equal(t, testNow, session.LastSeenAt)
}
