package domain

import "time"

// SubjectType is who a session belongs to. The values are the ones the
// session.subject_type CHECK constraint allows, English as every enum in this
// project is.
type SubjectType string

const (
	// SubjectTypeHousehold is a guest session, created by redeeming a login code.
	SubjectTypeHousehold SubjectType = "household"
	// SubjectTypeAdmin is the single admin, whose credentials come from the
	// environment. There is no admin row anywhere to point at.
	SubjectTypeAdmin SubjectType = "admin"
)

// Session lifetimes differ by an order of four because the risk profiles do. A
// household holds one shared printed code and must never be asked to find the card
// again — "log in once, ever" is a stated product goal, and the alternative is a
// phone call. The admin session reaches the budget and every household's data, so
// it is short enough that an unlocked laptop stops being a standing key.
const (
	householdSessionLifetime = 365 * 24 * time.Hour
	adminSessionLifetime     = 8 * time.Hour
)

// sessionRefreshInterval throttles rolling refresh. Extending a 365-day session on
// every request would turn a read-heavy app into a write-heavy one against a
// database with exactly one writer, and buys nothing: a session that was refreshed
// yesterday still has 364 days left.
const sessionRefreshInterval = 24 * time.Hour

// Session is an issued session as it is stored. The raw token is deliberately not a
// field — it exists only in the cookie and in the request that presents it, so
// there is nothing here that a leaked database could replay.
type Session struct {
	// ID is the SHA-256 hash of the token, hex encoded. It is the primary key.
	ID          string
	SubjectType SubjectType
	// SubjectID is household.id for a household session, and 0 for the admin, who
	// has no row to reference. Read it through HouseholdID rather than directly,
	// so that "0" can never be mistaken for a household.
	SubjectID  int64
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
	// UserAgent and IP are recorded once, at creation, for the audit trail. They are
	// not updated on use: that would be the per-request write the refresh throttle
	// exists to avoid, and the value at creation is the one worth having.
	UserAgent string
	IP        string
}

// NewSession builds a session for subjectType, with the lifetime that type gets.
//
// The lifetime is not a parameter on purpose. A caller that could pass a duration
// is a caller that can issue a 365-day admin session, and nothing downstream would
// notice — the two lifetimes are policy, and policy belongs here.
func NewSession(id string, subjectType SubjectType, subjectID int64, now time.Time, userAgent, ip string) Session {
	return Session{
		ID:          id,
		SubjectType: subjectType,
		SubjectID:   subjectID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(SessionLifetime(subjectType)),
		LastSeenAt:  now,
		UserAgent:   userAgent,
		IP:          ip,
	}
}

// SessionLifetime returns how long a freshly issued session of this type lives.
// An unknown subject type gets the shorter lifetime: failing towards less access
// is the only safe direction for a value that reached us from a database column.
func SessionLifetime(subjectType SubjectType) time.Duration {
	if subjectType == SubjectTypeHousehold {
		return householdSessionLifetime
	}
	return adminSessionLifetime
}

// HouseholdID returns the household this session acts as, and false for the admin.
//
// The pair return is what stops an admin session — whose SubjectID is 0 — from
// being read as household 0 by a caller that forgot to check the type first.
func (session Session) HouseholdID() (int64, bool) {
	if session.SubjectType != SubjectTypeHousehold {
		return 0, false
	}
	return session.SubjectID, true
}

// IsExpired reports whether the session has passed its expiry.
func (session Session) IsExpired(now time.Time) bool {
	return !now.Before(session.ExpiresAt)
}

// NeedsRefresh reports whether the session's expiry should be pushed out now.
//
// Household sessions only: an admin session that rolled on use would never end,
// which is precisely the property the short admin lifetime exists to deny.
func (session Session) NeedsRefresh(now time.Time) bool {
	if session.SubjectType != SubjectTypeHousehold {
		return false
	}
	return now.Sub(session.LastSeenAt) >= sessionRefreshInterval
}

// Refreshed returns the session with its expiry and last-seen moved to now.
//
// LastSeenAt is therefore "when we last extended this session", accurate to within
// a day rather than to the request — the deliberate cost of not writing to the
// database on every read.
func (session Session) Refreshed(now time.Time) Session {
	session.ExpiresAt = now.Add(SessionLifetime(session.SubjectType))
	session.LastSeenAt = now
	return session
}
