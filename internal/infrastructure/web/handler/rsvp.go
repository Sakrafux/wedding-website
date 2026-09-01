package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/application/rsvp"
	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/dto"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/httpio"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/web/middleware"
)

// RSVP serves a household's answer, on two pairs of routes: the guest's own
// /api/rsvp and the admin's /api/admin/households/{id}/rsvp.
//
// Both pairs go through the same use case and answer with the same body. The only
// differences are where the household id comes from and what the save options say —
// which is what keeps one field set instead of two that must agree (F3-B06).
type RSVP struct {
	rsvp *rsvp.UseCase
}

func NewRSVP(useCase *rsvp.UseCase) *RSVP {
	return &RSVP{rsvp: useCase}
}

// Show answers GET /api/rsvp for the household in the session.
//
// The household is read from the session and never from the request: an id a client
// could send would be an id this handler had to check.
func (handler *RSVP) Show(w http.ResponseWriter, r *http.Request) {
	householdID, isHousehold := middleware.HouseholdFromContext(r.Context())
	if !isHousehold {
		// RequireHousehold has already refused anybody else, so reaching this is a
		// routing mistake rather than a state a caller can produce.
		httpio.RespondError(w, r, domain.NewError(domain.CodeUnauthenticated))
		return
	}

	handler.respondWithAnswer(w, r, householdID)
}

// Save answers PUT /api/rsvp with the answer as stored.
func (handler *RSVP) Save(w http.ResponseWriter, r *http.Request) {
	householdID, isHousehold := middleware.HouseholdFromContext(r.Context())
	if !isHousehold {
		httpio.RespondError(w, r, domain.NewError(domain.CodeUnauthenticated))
		return
	}

	handler.save(w, r, householdID, rsvp.GuestOptions())
}

// AddMember answers POST /api/rsvp/members with the created companion.
//
// The household comes from the session and the name from the body; nothing else is
// read, because nothing else is an input (F4-B02).
func (handler *RSVP) AddMember(w http.ResponseWriter, r *http.Request) {
	householdID, isHousehold := middleware.HouseholdFromContext(r.Context())
	if !isHousehold {
		httpio.RespondError(w, r, domain.NewError(domain.CodeUnauthenticated))
		return
	}

	var request dto.RSVPAddMemberRequest
	if err := decodeAndValidate(w, r, &request); err != nil {
		respondRSVPError(w, r, err)
		return
	}

	addition, err := handler.rsvp.AddPlusOne(r.Context(), householdID, request.Name, rsvp.GuestOptions())
	if err != nil {
		respondRSVPError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusCreated, dto.RSVPAddMemberResponse{
		Member:        rsvpMember(addition.Member),
		CanAddPlusOne: addition.CanAddPlusOne,
	})
}

// RemoveMember answers DELETE /api/rsvp/members/{id} with 204.
//
// No body: the frontend already knows which row it removed, and refetching the whole
// form to learn it is gone is a round trip for nothing (F4-B03).
func (handler *RSVP) RemoveMember(w http.ResponseWriter, r *http.Request) {
	householdID, isHousehold := middleware.HouseholdFromContext(r.Context())
	if !isHousehold {
		httpio.RespondError(w, r, domain.NewError(domain.CodeUnauthenticated))
		return
	}

	memberID, err := pathID(r)
	if err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	if err := handler.rsvp.RemoveMember(r.Context(), householdID, memberID, rsvp.GuestOptions()); err != nil {
		respondRSVPError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminShow answers GET /api/admin/households/{id}/rsvp.
//
// The same body as the guest route, byte for byte, so that the shared form component
// does not have to know which caller it is serving.
func (handler *RSVP) AdminShow(w http.ResponseWriter, r *http.Request) {
	householdID, err := pathID(r)
	if err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	handler.respondWithAnswer(w, r, householdID)
}

// AdminSave answers PUT /api/admin/households/{id}/rsvp — the answer we take down the
// phone. The deadline does not apply here, and the audit log says the change was ours.
func (handler *RSVP) AdminSave(w http.ResponseWriter, r *http.Request) {
	householdID, err := pathID(r)
	if err != nil {
		httpio.RespondError(w, r, err)
		return
	}

	handler.save(w, r, householdID, rsvp.AdminOptions())
}

func (handler *RSVP) respondWithAnswer(w http.ResponseWriter, r *http.Request, householdID int64) {
	answer, err := handler.rsvp.Load(r.Context(), householdID)
	if err != nil {
		respondRSVPError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, rsvpResponse(answer))
}

func (handler *RSVP) save(w http.ResponseWriter, r *http.Request, householdID int64, options rsvp.SaveOptions) {
	// Asked before the body is even read, so that a guest who is past the deadline is
	// told exactly that instead of being handed a list of field errors for a form that
	// cannot be saved either way. Save checks again and is the authority.
	if err := handler.rsvp.EnsureWritable(r.Context(), options); err != nil {
		respondRSVPError(w, r, err)
		return
	}

	var request dto.RSVPSaveRequest
	if err := httpio.DecodeJSON(w, r, &request); err != nil {
		respondRSVPError(w, r, err)
		return
	}
	if err := httpio.ValidatePaths(&request, memberFieldKeys(request.Members)); err != nil {
		respondRSVPError(w, r, err)
		return
	}

	answer, err := handler.rsvp.Save(r.Context(), householdID, submissionFrom(request), options)
	if err != nil {
		respondRSVPError(w, r, err)
		return
	}

	httpio.WriteJSON(w, r, http.StatusOK, rsvpResponse(answer))
}

// memberFieldKeys rewrites a validation path that addresses a member by array index
// into one that addresses them by guest id: `members[0].attending` becomes
// `members.31.attending`.
//
// The id and not the index, because the frontend renders member cards keyed by id and
// an index would break the moment the list is filtered — the key shape is part of the
// contract F3-F04 renders from.
func memberFieldKeys(members []dto.RSVPMemberRequest) func(path string) string {
	return func(path string) string {
		rest, hasPrefix := strings.CutPrefix(path, "members[")
		if !hasPrefix {
			return path
		}

		index, field, hasField := strings.Cut(rest, "]")
		if !hasField {
			return path
		}

		position, err := strconv.Atoi(index)
		if err != nil || position < 0 || position >= len(members) {
			// Cannot happen for a path the validator produced from this very slice.
			// Left as it came rather than guessed at: a wrong id would put the message
			// on somebody else's card.
			return path
		}

		return fmt.Sprintf("members.%d%s", members[position].ID, field)
	}
}

// submissionFrom maps the request onto the use case's submission shape.
//
// Field by field, in the direction that matters least for privacy and most for
// clarity: the enums are re-typed here, so a value the DTO's `oneof` did not cover
// cannot reach the domain as a valid-looking string.
func submissionFrom(request dto.RSVPSaveRequest) rsvp.Submission {
	members := make([]rsvp.MemberSubmission, 0, len(request.Members))
	for _, member := range request.Members {
		members = append(members, rsvp.MemberSubmission{
			ID: member.ID,
			Answer: domain.GuestAnswer{
				Attending:     domain.Attending(member.Attending),
				MealChoice:    mealChoicePointer(member.MealChoice),
				Portion:       domain.Portion(member.Portion),
				MidnightSnack: member.MidnightSnack,
				SeatingNeed:   domain.SeatingNeed(member.SeatingNeed),
				DietaryNote:   member.DietaryNote,
				Age:           member.Age,
			},
		})
	}

	return rsvp.Submission{
		Household: domain.HouseholdAnswer{
			TransportSeatsNeeded:  request.TransportSeatsNeeded,
			TransportSeatsOffered: request.TransportSeatsOffered,
			HasStroller:           request.HasStroller,
			RSVPNote:              request.RSVPNote,
		},
		Members: members,
	}
}

// rsvpResponse maps the use case result onto the wire shape, field by field. This is
// where the privacy rule lives: the household's code, our admin note and
// rsvp_note_seen_at stop here because nothing copies them across — see dto.RSVPHousehold.
func rsvpResponse(answer rsvp.Answer) dto.RSVPResponse {
	members := make([]dto.RSVPMember, 0, len(answer.Members))
	for _, member := range answer.Members {
		members = append(members, rsvpMember(member))
	}

	return dto.RSVPResponse{
		Household: dto.RSVPHousehold{
			ID:                    answer.Household.ID,
			DisplayName:           answer.Household.DisplayName,
			TransportSeatsNeeded:  answer.Household.TransportSeatsNeeded,
			TransportSeatsOffered: answer.Household.TransportSeatsOffered,
			HasStroller:           answer.Household.HasStroller,
			RSVPNote:              answer.Household.RSVPNote,
			RSVPSubmittedAt:       answer.Household.RSVPSubmittedAt,
			RSVPUpdatedAt:         answer.Household.RSVPUpdatedAt,
		},
		Members:       members,
		Deadline:      answer.Deadline,
		Editable:      answer.Editable,
		CanAddPlusOne: answer.CanAddPlusOne,
	}
}

// rsvpMember maps one guest onto the wire shape. Shared by the form response and by
// the addition response, so a plus-one is rendered as the card the form already knows
// how to draw rather than as a second, nearly identical shape.
func rsvpMember(member domain.Guest) dto.RSVPMember {
	return dto.RSVPMember{
		ID:            member.ID,
		Name:          member.Name,
		Kind:          string(member.Kind),
		Age:           member.Age,
		Origin:        string(member.Origin),
		Attending:     enumString(member.Attending),
		MealChoice:    enumString(member.MealChoice),
		Portion:       string(member.Portion),
		MidnightSnack: member.MidnightSnack,
		SeatingNeed:   string(member.SeatingNeed),
		DietaryNote:   member.DietaryNote,
	}
}

// respondRSVPError answers a failure from the RSVP use case.
//
// Three translations beyond RespondError: the use case's "no such row" into the API's
// 404, a rejected age or seating need into a field error on the member it belongs to,
// and a contradictory transport answer into a field error on both counts (F3-B07).
func respondRSVPError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, application.ErrNotFound) {
		httpio.RespondError(w, r, httpio.ErrNotFound)
		return
	}

	if memberError, isMemberError := errors.AsType[rsvp.MemberAnswerError](err); isMemberError {
		httpio.RespondError(w, r, httpio.GuestFieldValidationErrorUnder(
			fmt.Sprintf("members.%d.", memberError.MemberID), memberError.Err))
		return
	}

	httpio.RespondError(w, r, httpio.TransportSeatsValidationError(err))
}

// enumString renders a nullable domain enum for the wire: a JSON null for "not
// answered", which is a state and not a missing value.
func enumString[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	rendered := string(*value)
	return &rendered
}

func mealChoicePointer(value *string) *domain.MealChoice {
	if value == nil {
		return nil
	}
	choice := domain.MealChoice(*value)
	return &choice
}
