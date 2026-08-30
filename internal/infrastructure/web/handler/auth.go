package handler

import (
	"net"
	"net/http"
	"time"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

// Auth serves logging in, logging out, and the bootstrap call the frontend makes
// on every load.
type Auth struct {
	auth *application.Auth
	// cookieSecure mirrors SESSION_COOKIE_SECURE, and is passed straight through to
	// the cookie helpers so that issuing and clearing agree on the attribute.
	cookieSecure bool
}

func NewAuth(auth *application.Auth, cookieSecure bool) *Auth {
	return &Auth{auth: auth, cookieSecure: cookieSecure}
}

// LogIn redeems the household code from the request body and sets the session
// cookie.
//
// The response body is the same as GET /api/me, so a client that has just logged
// in needs no second request to render.
func (handler *Auth) LogIn(w http.ResponseWriter, r *http.Request) {
	var request dto.LoginRequest
	if err := httpio.DecodeJSON(w, r, &request); err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	// A session already on the request is the one being replaced. Reading it from
	// the context rather than from the cookie means only a *valid* session is
	// deleted here — a stale cookie names nothing, and there is nothing to revoke.
	var previousSessionID string
	if session, isAuthenticated := middleware.SessionFromContext(r.Context()); isAuthenticated {
		previousSessionID = session.ID
	}

	login, err := handler.auth.LogInHousehold(r.Context(), request.Code, r.UserAgent(), clientIP(r), previousSessionID)
	if err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	middleware.WriteSessionCookie(w, login.Token, time.Until(login.Session.ExpiresAt), handler.cookieSecure)

	httpio.WriteJSON(w, r, http.StatusOK, bootstrapResponse(login.Bootstrap))
}

// LogOut revokes the current session and clears the cookie.
//
// Idempotent, and 204 either way: an anonymous caller asking to be logged out has
// got what they wanted. Answering 401 would mean the frontend has to special-case
// the one request whose whole purpose is to end up unauthenticated.
func (handler *Auth) LogOut(w http.ResponseWriter, r *http.Request) {
	if session, isAuthenticated := middleware.SessionFromContext(r.Context()); isAuthenticated {
		if err := handler.auth.LogOut(r.Context(), session.ID); err != nil {
			httpio.RespondError(w, r, err)
			return
		}
	}

	middleware.ClearSessionCookie(w, handler.cookieSecure)

	w.WriteHeader(http.StatusNoContent)
}

// Me returns the household the session belongs to, its members and the runtime
// flags. It is the frontend's bootstrap call: one request that says who the app is
// talking to and what is switched on.
//
// Mounted behind RequireHousehold, so the household id comes from the session and
// never from the request.
func (handler *Auth) Me(w http.ResponseWriter, r *http.Request) {
	householdID, isHousehold := middleware.HouseholdFromContext(r.Context())
	if !isHousehold {
		// Unreachable behind the gate; kept so that a future route registration
		// that forgets the gate fails closed instead of panicking on a zero id.
		httpio.RespondError(w, r, httpio.ErrNotFound)
		return
	}

	bootstrap, err := handler.auth.BootstrapFor(r.Context(), householdID)
	if err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, bootstrapResponse(bootstrap))
}

// bootstrapResponse maps the use case result onto the wire shape.
//
// Field by field on purpose: this is the boundary the privacy rule lives on, and
// the household's login code and admin note stop here because nothing copies them
// across. See dto.HouseholdSummary for what is left out and why.
func bootstrapResponse(bootstrap application.Bootstrap) dto.BootstrapResponse {
	members := make([]dto.Member, 0, len(bootstrap.Members))
	for _, member := range bootstrap.Members {
		members = append(members, dto.Member{
			ID:        member.ID,
			FirstName: member.FirstName,
			LastName:  member.LastName,
			Kind:      string(member.Kind),
			Origin:    string(member.Origin),
		})
	}

	return dto.BootstrapResponse{
		Household: dto.HouseholdSummary{
			ID:          bootstrap.Household.ID,
			DisplayName: bootstrap.Household.DisplayName,
		},
		Members: members,
		Flags: dto.Flags{
			RSVPOpen:         bootstrap.Settings.RSVPOpen(time.Now()),
			SeatingPublished: bootstrap.Settings.SeatingPublished,
			GalleryVisible:   bootstrap.Settings.GalleryVisible,
			UploadsOpen:      bootstrap.Settings.UploadsOpen,
		},
		RSVPDeadline: bootstrap.Settings.RSVPDeadline,
	}
}

// clientIP is the address recorded on a session for the audit trail.
//
// F1-B05 replaces this with resolution against TRUSTED_PROXY_CIDRS. Until then it
// is the direct peer, which behind the reverse proxy is the proxy itself — wrong
// for the audit trail, and deliberately not guessed at from X-Forwarded-For, since
// believing that header from an untrusted source is exactly the bug F1-B05 exists
// to avoid.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
