package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/security"
)

// Auth is the use case layer for logging in, staying logged in and logging out.
//
// It exists so that neither a handler nor a middleware ever assembles the sequence
// itself: normalize, look up, issue, refresh, revoke. The middleware and the login
// endpoint are two entry points into the same rules, and a second copy of them is
// how one of the two ends up not deleting the session of a household that was
// removed while logged in.
type Auth struct {
	sessions   *persistence.SessionStore
	households *persistence.HouseholdStore
	settings   *persistence.SettingStore
}

func NewAuth(sessions *persistence.SessionStore, households *persistence.HouseholdStore, settings *persistence.SettingStore) *Auth {
	return &Auth{sessions: sessions, households: households, settings: settings}
}

// Bootstrap is everything the frontend needs on the first render: who it is
// talking to, who is in that household, and what is currently switched on.
//
// One value for both POST /api/auth/login and GET /api/me, because the two return
// the same body — the app must not learn different things depending on whether it
// just logged in or was already logged in.
type Bootstrap struct {
	Household domain.Household
	Members   []domain.Guest
	Settings  domain.Settings
}

// HouseholdLogin is the outcome of a successful login: the token for the cookie,
// the session it stands for, and the body to answer with.
//
// The session is returned so the handler can derive the cookie's Max-Age from the
// row's own expiry instead of recomputing the lifetime. Two places computing "365
// days" is two places to get it wrong, and the failure — a cookie that outlives its
// session — looks to a guest like being logged out at random.
type HouseholdLogin struct {
	Token     string
	Session   domain.Session
	Bootstrap Bootstrap
}

// LogInHousehold redeems a login code and issues a session.
//
// previousSessionID is the session the request already carried, or empty. It is
// deleted on success: a guest who re-enters a code must end up on the household
// that code names, not the one their cookie remembered. That is how a wrong
// household — a code typed on a shared family tablet, say — gets corrected in
// practice, and it leaves no orphaned session behind.
//
// Every failure is CodeUnknownLoginCode. A malformed code and an unknown one are
// indistinguishable to the caller by design: "that is a valid code, just not one
// of ours" is a sentence that turns guessing from hopeless into merely slow.
func (auth *Auth) LogInHousehold(ctx context.Context, submittedCode, userAgent, ip, previousSessionID string) (HouseholdLogin, error) {
	code := domain.NormalizeCode(submittedCode)

	// Shape is checked before the query, so a mistyped code costs no database
	// round trip. This is observable in the response time, and deliberately so:
	// the difference tells an attacker only what they could already see by
	// looking at what they typed, unlike a timing gap between "no such code" and
	// "that code exists" — which the single lookup below does not have.
	if err := domain.ValidateCode(code); err != nil {
		return HouseholdLogin{}, domain.NewError(domain.CodeUnknownLoginCode)
	}

	household, err := auth.households.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			return HouseholdLogin{}, domain.NewError(domain.CodeUnknownLoginCode)
		}
		return HouseholdLogin{}, fmt.Errorf("looking up household: %w", err)
	}

	now := time.Now()

	token, session, err := auth.issueSession(ctx, domain.SubjectTypeHousehold, household.ID, now, userAgent, ip)
	if err != nil {
		return HouseholdLogin{}, err
	}

	// After the new session exists, so a failure here leaves the guest logged in
	// rather than logged out of both.
	if previousSessionID != "" {
		if err := auth.sessions.Delete(ctx, previousSessionID); err != nil {
			return HouseholdLogin{}, err
		}
	}

	if err := auth.households.TouchLastLogin(ctx, household.ID, now); err != nil {
		return HouseholdLogin{}, err
	}
	household.LastLoginAt = &now

	bootstrap, err := auth.bootstrapFor(ctx, household)
	if err != nil {
		return HouseholdLogin{}, err
	}
	return HouseholdLogin{Token: token, Session: session, Bootstrap: bootstrap}, nil
}

// issueSession creates a session and returns the raw token for the cookie. The
// token is returned and never stored; what reaches the database is its hash.
func (auth *Auth) issueSession(ctx context.Context, subjectType domain.SubjectType, subjectID int64, now time.Time, userAgent, ip string) (string, domain.Session, error) {
	token := security.NewSessionToken()

	session := domain.NewSession(security.HashSessionToken(token), subjectType, subjectID, now, userAgent, ip)
	if err := auth.sessions.Create(ctx, session); err != nil {
		return "", domain.Session{}, err
	}
	return token, session, nil
}

// ResolveSession turns a cookie token into the session it stands for, reporting
// false for anything that is not a live session right now.
//
// It also performs the two side effects that must not be left to a caller:
//
//   - A household session whose household no longer exists is deleted and reported
//     as absent. Without that, every query downstream would have to defend against
//     a dangling household id, and the guest would meet errors instead of the
//     login screen.
//   - A household session past its refresh interval is rolled forward, so a guest
//     who visits at least once a year never has to find their card again.
func (auth *Auth) ResolveSession(ctx context.Context, token string) (domain.Session, bool, error) {
	session, err := auth.sessions.FindByID(ctx, security.HashSessionToken(token))
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			return domain.Session{}, false, nil
		}
		return domain.Session{}, false, err
	}

	if householdID, isHousehold := session.HouseholdID(); isHousehold {
		if _, err := auth.households.FindByID(ctx, householdID); err != nil {
			if !errors.Is(err, persistence.ErrNotFound) {
				return domain.Session{}, false, err
			}
			if err := auth.sessions.Delete(ctx, session.ID); err != nil {
				return domain.Session{}, false, err
			}
			return domain.Session{}, false, nil
		}
	}

	now := time.Now()
	if session.NeedsRefresh(now) {
		refreshed := session.Refreshed(now)
		if err := auth.sessions.Refresh(ctx, refreshed); err != nil {
			return domain.Session{}, false, err
		}
		session = refreshed
	}

	return session, true, nil
}

// LogOut revokes one session. Deleting a row that is not there is not an error, so
// logging out twice behaves the same as logging out once.
func (auth *Auth) LogOut(ctx context.Context, sessionID string) error {
	return auth.sessions.Delete(ctx, sessionID)
}

// BootstrapFor assembles the bootstrap body for an already-authenticated
// household. Used by GET /api/me, where the session is what identifies the
// household — never an id the frontend sent.
func (auth *Auth) BootstrapFor(ctx context.Context, householdID int64) (Bootstrap, error) {
	household, err := auth.households.FindByID(ctx, householdID)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			// ResolveSession deletes the session of a deleted household before any
			// handler runs, so reaching this is a bug in the ordering rather than a
			// state a guest can produce.
			return Bootstrap{}, fmt.Errorf("session household %d no longer exists", householdID)
		}
		return Bootstrap{}, err
	}
	return auth.bootstrapFor(ctx, household)
}

func (auth *Auth) bootstrapFor(ctx context.Context, household domain.Household) (Bootstrap, error) {
	members, err := auth.households.ListMembers(ctx, household.ID)
	if err != nil {
		return Bootstrap{}, err
	}

	settings, err := auth.settings.Load(ctx)
	if err != nil {
		return Bootstrap{}, err
	}

	return Bootstrap{Household: household, Members: members, Settings: settings}, nil
}
