package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// SessionCookieName carries the raw session token. Not prefixed with `__Host-`,
// which would be stricter but also mandates the Secure attribute — and
// SESSION_COOKIE_SECURE=false has to keep working for local development over plain
// HTTP, which is the one case a cookie prefix would break.
const SessionCookieName = "wedding_session"

// sessionContextKey is the key the resolved session is stored under.
//
// Its own unexported type, so that no other package can produce a value equal to
// it: a plain string key is shared with every other library writing to the same
// context, and the collision is silent when it happens.
type sessionContextKey struct{}

// SessionGate resolves session cookies and guards routes by subject type.
type SessionGate struct {
	auth *application.Auth
	// cookieSecure mirrors SESSION_COOKIE_SECURE. Held here so that the attributes
	// of the cookie we set, clear and read are decided in one place — a Secure flag
	// that disagreed between issuing and clearing would leave a cookie behind that
	// the browser will not overwrite.
	cookieSecure bool
}

func NewSessionGate(auth *application.Auth, cookieSecure bool) *SessionGate {
	return &SessionGate{auth: auth, cookieSecure: cookieSecure}
}

// Resolve puts the caller's session into the request context, and is mounted on
// the whole /api tree.
//
// No cookie, an unknown cookie, a garbage cookie or an expired one all mean
// anonymous — not an error. The login endpoints and the health probes are behind
// this middleware too, and they must work for a caller who by definition has no
// session yet. Refusing a request is RequireHousehold's and RequireAdmin's job.
//
// A failure of the database itself is a different matter and is answered as one:
// treating "the disk is gone" as "not logged in" would show the login screen to
// somebody whose session is perfectly valid, and they would type their code in
// vain over and over.
func (gate *SessionGate) Resolve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		session, isAuthenticated, err := gate.auth.ResolveSession(r.Context(), cookie.Value)
		if err != nil {
			httpio.RespondError(w, r, err)
			return
		}
		if !isAuthenticated {
			// The cookie names a session that no longer exists. Clearing it saves
			// the browser from presenting it on every future request for a year.
			ClearSessionCookie(w, gate.cookieSecure)
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r.WithContext(withSession(r.Context(), session)))
	})
}

// RequireHousehold rejects anyone who is not a logged-in household.
//
// The admin is rejected too. An admin session is not a household and has no
// members, no RSVP and no seat — an endpoint under this gate would have nothing to
// answer with, and letting the admin through would mean every one of those
// handlers needs its own "which household, though?" branch.
func RequireHousehold(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, isHousehold := HouseholdFromContext(r.Context()); !isHousehold {
			httpio.RespondError(w, r, domain.NewError(domain.CodeUnauthenticated))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin rejects anyone who is not the admin, household sessions included.
//
// Mount it once on the whole /api/admin subtree rather than per handler. Every
// admin-only rule in this application — the budget above all — rests on this one
// call, and a per-handler check is how one endpoint eventually ships without one.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, isAuthenticated := SessionFromContext(r.Context())
		if !isAuthenticated || session.SubjectType != domain.SubjectTypeAdmin {
			httpio.RespondError(w, r, domain.NewError(domain.CodeUnauthenticated))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SessionFromContext returns the caller's session, and false when the request is
// anonymous.
//
// Handlers read the session through this and through HouseholdFromContext, never
// through the context value itself, so what is stored there can change shape
// without touching every handler.
func SessionFromContext(ctx context.Context) (domain.Session, bool) {
	session, isPresent := ctx.Value(sessionContextKey{}).(domain.Session)
	return session, isPresent
}

// HouseholdFromContext returns the household the caller is acting as, and false
// for the admin or an anonymous request. This is what a guest-facing handler uses
// to decide whose data it is looking at — never an id from the request.
func HouseholdFromContext(ctx context.Context) (int64, bool) {
	session, isAuthenticated := SessionFromContext(ctx)
	if !isAuthenticated {
		return 0, false
	}
	return session.HouseholdID()
}

func withSession(ctx context.Context, session domain.Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, session)
}

// WriteSessionCookie issues the session cookie for a freshly created session.
//
// HttpOnly: no script needs the token, and the site embeds no third-party code, so
// there is nothing legitimate to read it. SameSite=Lax: the API is same-origin and
// every mutation is a POST/PUT/DELETE, so Lax blocks cross-site submission while
// still letting a guest follow a link from a chat app into a logged-in page.
// Path=/: the app owns its own subdomain, so a root-scoped cookie reaches nothing
// else.
//
// Max-Age is derived from the session's own lifetime rather than set to a constant,
// so the cookie and the row in the session table cannot expire on different days.
func WriteSessionCookie(w http.ResponseWriter, token string, lifetime time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(lifetime.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie expires the session cookie.
//
// Every attribute except Value and MaxAge repeats what WriteSessionCookie set:
// browsers match a cookie for replacement on name, domain and path, so a clear
// that disagreed on Path would add a second, empty cookie and leave the real one
// in place.
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
