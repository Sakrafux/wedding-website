package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/application/households"
	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
)

// AdminHouseholds serves the admin's guest list: households, the people in them, and
// reissuing a login code.
//
// Every route it registers sits behind RequireAdmin, and its responses are the only
// ones in the API that carry a login code or our private note. Nothing here is
// reachable with a household session.
type AdminHouseholds struct {
	households *households.UseCase
}

func NewAdminHouseholds(useCase *households.UseCase) *AdminHouseholds {
	return &AdminHouseholds{households: useCase}
}

// List answers GET /api/admin/households with every household and its member count.
func (handler *AdminHouseholds) List(w http.ResponseWriter, r *http.Request) {
	overviews, err := handler.households.List(r.Context())
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	households := make([]dto.AdminHouseholdOverview, 0, len(overviews))
	for _, overview := range overviews {
		households = append(households, householdOverviewResponse(overview))
	}

	httpio.WriteJSON(w, r, http.StatusOK, dto.AdminHouseholdListResponse{Households: households})
}

// Show answers GET /api/admin/households/{id} with the household and its members.
func (handler *AdminHouseholds) Show(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	detail, err := handler.households.Detail(r.Context(), id)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, householdResponse(detail.Household, detail.Members))
}

// Create answers POST /api/admin/households. The new household comes back with the
// login code it was just assigned, because the admin is about to go and use it.
func (handler *AdminHouseholds) Create(w http.ResponseWriter, r *http.Request) {
	var request dto.AdminHouseholdCreateRequest
	if err := decodeAndValidate(w, r, &request); err != nil {
		respondAdminError(w, r, err)
		return
	}

	created, err := handler.households.Create(r.Context(), domain.Household{
		DisplayName: request.DisplayName,
		AdminNote:   request.AdminNote,
	})
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	// No members yet, and an empty list rather than null: the frontend renders the
	// same component for a fresh household as for one with four people in it.
	httpio.WriteJSON(w, r, http.StatusCreated, householdResponse(created, nil))
}

// Update answers PATCH /api/admin/households/{id}. Absent fields are left alone.
func (handler *AdminHouseholds) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	var request dto.AdminHouseholdPatchRequest
	if err := decodeAndValidate(w, r, &request); err != nil {
		respondAdminError(w, r, err)
		return
	}

	// domain.HouseholdPatch keeps its transport fields — PUT /rsvp patches through
	// them — so the three left nil here are the ones this endpoint no longer owns.
	detail, err := handler.households.Update(r.Context(), id, domain.HouseholdPatch{
		DisplayName: request.DisplayName,
		AdminNote:   request.AdminNote,
	})
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, householdResponse(detail.Household, detail.Members))
}

// Delete answers DELETE /api/admin/households/{id}. Guests and seat assignments go
// with it; the audit trail does not.
func (handler *AdminHouseholds) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	if err := handler.households.Delete(r.Context(), id); err != nil {
		respondAdminError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReissueCode answers POST /api/admin/households/{id}/code with a fresh login code.
//
// Takes no body: there is nothing to choose. The old code stops working immediately
// and the household's sessions are revoked — the warning about the printed card
// belongs in front of the human (F5-F02), not in the API.
func (handler *AdminHouseholds) ReissueCode(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	reissue, err := handler.households.ReissueCode(r.Context(), id)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, dto.AdminCodeReissueResponse{
		Code:            reissue.Code,
		RevokedSessions: reissue.RevokedSessions,
	})
}

// AddGuest answers POST /api/admin/households/{id}/guests.
func (handler *AdminHouseholds) AddGuest(w http.ResponseWriter, r *http.Request) {
	householdID, err := pathID(r)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	var request dto.AdminGuestCreateRequest
	if err := decodeAndValidate(w, r, &request); err != nil {
		respondAdminError(w, r, err)
		return
	}

	created, err := handler.households.AddGuest(r.Context(), householdID, domain.Guest{
		Name:        request.Name,
		Kind:        domain.GuestKind(request.Kind),
		Age:         request.Age,
		SeatingNeed: seatingNeedOrDefault(request.SeatingNeed),
		DietaryNote: request.DietaryNote,
	})
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusCreated, guestResponse(created))
}

// UpdateGuest answers PATCH /api/admin/guests/{id}. Guests are addressed by their
// own id: the admin owns all of them, so a household in the path would be an id
// nothing checks.
func (handler *AdminHouseholds) UpdateGuest(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	var request dto.AdminGuestPatchRequest
	if err := decodeAndValidate(w, r, &request); err != nil {
		respondAdminError(w, r, err)
		return
	}

	updated, err := handler.households.UpdateGuest(r.Context(), id, domain.GuestPatch{
		Name:        request.Name,
		Kind:        guestKindPointer(request.Kind),
		AgeSet:      request.Age.Present,
		Age:         request.Age.Value,
		SeatingNeed: seatingNeedPointer(request.SeatingNeed),
		DietaryNote: request.DietaryNote,
	})
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, guestResponse(updated))
}

// RemoveGuest answers DELETE /api/admin/guests/{id}. A soft delete: the person stays
// in the record, which is what keeps a past headcount explainable.
func (handler *AdminHouseholds) RemoveGuest(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		respondAdminError(w, r, err)
		return
	}

	if err := handler.households.RemoveGuest(r.Context(), id); err != nil {
		respondAdminError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// decodeAndValidate reads a JSON body and checks its field rules, so that every
// admin endpoint reports a malformed body and a rejected field the same way.
func decodeAndValidate(w http.ResponseWriter, r *http.Request, target any) error {
	if err := httpio.DecodeJSON(w, r, target); err != nil {
		return err
	}
	return httpio.Validate(target)
}

// pathID reads the {id} path parameter.
//
// A non-numeric id is a 404 and not a 400: /api/admin/households/abc is an address
// that does not exist, which is the same answer as an id that names no row. Reporting
// "invalid id" instead would describe our routing to a caller who guessed a URL.
func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, httpio.ErrNotFound
	}
	return id, nil
}

// respondAdminError answers a failure from the use case layer.
//
// Its one job beyond RespondError is translating the use case's "no such row" into
// the API's 404, and the domain's age rules into a field error. Everything else
// passes straight through — including a driver error, which becomes the generic 500
// because a message this layer does not recognise cannot be described to a caller.
func respondAdminError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, application.ErrNotFound) {
		httpio.RespondError(w, r, httpio.ErrNotFound)
		return
	}
	httpio.RespondError(w, r, httpio.GuestFieldValidationError(err))
}

// seatingNeedOrDefault maps an omitted seating need onto the column's default, so a
// create request need not restate the ordinary case.
func seatingNeedOrDefault(value string) domain.SeatingNeed {
	if value == "" {
		return domain.SeatingNeedNormal
	}
	return domain.SeatingNeed(value)
}

func guestKindPointer(value *string) *domain.GuestKind {
	if value == nil {
		return nil
	}
	kind := domain.GuestKind(*value)
	return &kind
}

func seatingNeedPointer(value *string) *domain.SeatingNeed {
	if value == nil {
		return nil
	}
	need := domain.SeatingNeed(*value)
	return &need
}

// householdOverviewResponse and householdResponse map field by field, which is where
// the privacy rule lives — even here, where `code` and `admin_note` are allowed
// through on purpose. Nothing serializes a domain struct.
func householdOverviewResponse(overview domain.HouseholdOverview) dto.AdminHouseholdOverview {
	return dto.AdminHouseholdOverview{
		ID:              overview.ID,
		DisplayName:     overview.DisplayName,
		Code:            overview.Code,
		MemberCount:     overview.MemberCount,
		LastLoginAt:     overview.LastLoginAt,
		RSVPSubmittedAt: overview.RSVPSubmittedAt,
	}
}

func householdResponse(household domain.Household, members []domain.Guest) dto.AdminHousehold {
	guests := make([]dto.AdminGuest, 0, len(members))
	for _, member := range members {
		guests = append(guests, guestResponse(member))
	}

	return dto.AdminHousehold{
		AdminHouseholdOverview: dto.AdminHouseholdOverview{
			ID:              household.ID,
			DisplayName:     household.DisplayName,
			Code:            household.Code,
			MemberCount:     len(guests),
			LastLoginAt:     household.LastLoginAt,
			RSVPSubmittedAt: household.RSVPSubmittedAt,
		},
		AdminNote:             household.AdminNote,
		TransportSeatsNeeded:  household.TransportSeatsNeeded,
		TransportSeatsOffered: household.TransportSeatsOffered,
		HasStroller:           household.HasStroller,
		Members:               guests,
	}
}

func guestResponse(guest domain.Guest) dto.AdminGuest {
	return dto.AdminGuest{
		ID:          guest.ID,
		HouseholdID: guest.HouseholdID,
		Name:        guest.Name,
		Kind:        string(guest.Kind),
		Age:         guest.Age,
		Origin:      string(guest.Origin),
		SeatingNeed: string(guest.SeatingNeed),
		DietaryNote: guest.DietaryNote,
	}
}
