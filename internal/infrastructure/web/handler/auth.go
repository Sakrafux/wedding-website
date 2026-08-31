package handler

import (
	"net/http"
	"time"

	"github.com/Sakrafux/wedding-website/internal/application/auth"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

// Auth serves logging in, logging out, and the bootstrap call the frontend makes
// on every load.
type Auth struct {
	auth *auth.UseCase
	// cookieSecure mirrors SESSION_COOKIE_SECURE, and is passed straight through to
	// the cookie helpers so that issuing and clearing agree on the attribute.
	cookieSecure bool
}

func NewAuth(useCase *auth.UseCase, cookieSecure bool) *Auth {
	return &Auth{auth: useCase, cookieSecure: cookieSecure}
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

	login, err := handler.auth.LogInHousehold(r.Context(), request.Code, r.UserAgent(),
		middleware.ClientIPFromContext(r.Context()), currentSessionID(r))
	if err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	middleware.WriteSessionCookie(w, login.Token, time.Until(login.Session.ExpiresAt), handler.cookieSecure)

	httpio.WriteJSON(w, r, http.StatusOK, bootstrapResponse(login.Bootstrap))
}

// AdminLogIn checks the configured credentials and sets an admin session cookie.
//
// Same cookie name and attributes as a household session — the subject type in the
// session table is what distinguishes them. Two cookie names would double the ways
// to get this wrong, and would let a browser hold both at once.
func (handler *Auth) AdminLogIn(w http.ResponseWriter, r *http.Request) {
	var request dto.AdminLoginRequest
	if err := httpio.DecodeJSON(w, r, &request); err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	login, err := handler.auth.LogInAdmin(r.Context(), request.User, request.Password,
		r.UserAgent(), middleware.ClientIPFromContext(r.Context()), currentSessionID(r))
	if err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	middleware.WriteSessionCookie(w, login.Token, time.Until(login.Session.ExpiresAt), handler.cookieSecure)

	httpio.WriteJSON(w, r, http.StatusOK, dto.AdminLoginResponse{SubjectType: string(login.Session.SubjectType)})
}

// AdminMe confirms that the caller holds an admin session.
//
// It carries no data, and that is the whole design: the admin frontend needs to
// know whether the cookie it already has is still an admin session, and there is
// nothing else to tell it — /api/me answers 401 for an admin, since an admin is
// not a household. Mounted behind RequireAdmin, so reaching the body at all is the
// answer.
//
// It exists because the admin frontend's route guard would otherwise have nothing
// to ask, and a client-side "I logged in earlier" flag would have put "am I the
// admin" in a second place — the one that goes stale.
func (handler *Auth) AdminMe(w http.ResponseWriter, r *http.Request) {
	session, isAuthenticated := middleware.SessionFromContext(r.Context())
	if !isAuthenticated {
		// Unreachable behind the gate; fails closed if a future registration
		// forgets it.
		httpio.RespondError(w, r, httpio.ErrNotFound)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, dto.AdminLoginResponse{SubjectType: string(session.SubjectType)})
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

// currentSessionID is the session the request already carried, or empty.
//
// Read from the context rather than from the cookie, so only a *valid* session is
// named: a stale cookie points at nothing, and there is nothing to revoke.
func currentSessionID(r *http.Request) string {
	session, isAuthenticated := middleware.SessionFromContext(r.Context())
	if !isAuthenticated {
		return ""
	}
	return session.ID
}

// bootstrapResponse maps the use case result onto the wire shape.
//
// Field by field on purpose: this is the boundary the privacy rule lives on, and
// the household's login code and admin note stop here because nothing copies them
// across. See dto.HouseholdSummary for what is left out and why.
func bootstrapResponse(bootstrap auth.Bootstrap) dto.BootstrapResponse {
	members := make([]dto.Member, 0, len(bootstrap.Members))
	for _, member := range bootstrap.Members {
		members = append(members, dto.Member{
			ID:     member.ID,
			Name:   member.Name,
			Kind:   string(member.Kind),
			Origin: string(member.Origin),
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
