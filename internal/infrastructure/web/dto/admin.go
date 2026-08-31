package dto

import (
	"bytes"
	"encoding/json"
	"time"
)

// The admin's view of the guest list.
//
// These are the one place in the API that may carry `code` and `admin_note`, and
// they are reachable only behind RequireAdmin. Everything guest-facing builds its
// own bodies — see HouseholdSummary for what a guest gets and why.

// AdminHouseholdListResponse is the body of GET /api/admin/households.
type AdminHouseholdListResponse struct {
	Households []AdminHouseholdOverview `json:"households"`
}

// AdminHouseholdOverview is one row of the admin household list.
//
// Deliberately omitted: admin_note and the transport counts. The list is scanned,
// not read, and a note written for one household would push the columns that answer
// "did they log in, did they answer" off a laptop screen. They are in
// AdminHousehold, which is what the detail page loads.
type AdminHouseholdOverview struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
	// Code in stored form, which is exactly the form printed on the card. There is
	// no other form: the group separator was dropped, so what is shown here can be
	// compared against a card character by character.
	Code        string `json:"code"`
	MemberCount int    `json:"member_count"`
	// LastLoginAt is null until the household redeems its code — the column the
	// nudge calls before send-out are made from.
	LastLoginAt *time.Time `json:"last_login_at"`
	// RSVPSubmittedAt is null until they answer. F3 fills it; F5 only shows it.
	RSVPSubmittedAt *time.Time `json:"rsvp_submitted_at"`
}

// AdminHousehold is one household in full: the overview fields plus everything the
// detail page edits.
//
// One shape for GET, POST and PATCH alike, so the frontend has a single type to
// render from and cannot learn different things depending on which request it just
// made. Members is empty rather than absent on a fresh household.
type AdminHousehold struct {
	AdminHouseholdOverview
	AdminNote             string       `json:"admin_note"`
	TransportSeatsNeeded  int          `json:"transport_seats_needed"`
	TransportSeatsOffered int          `json:"transport_seats_offered"`
	HasStroller           bool         `json:"has_stroller"`
	Members               []AdminGuest `json:"members"`
}

// AdminGuest is one person as the admin sees them.
//
// Deliberately omitted: every RSVP answer (`attending`, `meal_choice`, `portion`,
// `midnight_snack`). Those belong to F3's endpoint, addressed by household id, so
// that there is one shape owning the RSVP field set rather than two that must agree
// — which is the whole point of Gate 1.
type AdminGuest struct {
	ID          int64  `json:"id"`
	HouseholdID int64  `json:"household_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	// Kind, Origin and SeatingNeed are the English enum values from the database.
	// German labels are the frontend's business, mapped in web/src/lib/labels.ts.
	Kind string `json:"kind"`
	// Age is age at the wedding date, and null for an adult.
	Age         *int   `json:"age"`
	Origin      string `json:"origin"`
	SeatingNeed string `json:"seating_need"`
	DietaryNote string `json:"dietary_note"`
}

// AdminHouseholdCreateRequest is the body of POST /api/admin/households.
//
// Plain values rather than pointers: on a create an absent field means the column
// default, and there is nothing to distinguish it from.
type AdminHouseholdCreateRequest struct {
	DisplayName string `json:"display_name" validate:"required,max=120"`
	AdminNote   string `json:"admin_note" validate:"max=2000"`
	// A household is not a coach: twenty seats is far past any real answer and
	// catches a stray keystroke rather than a plausible one.
	TransportSeatsNeeded  int  `json:"transport_seats_needed" validate:"gte=0,lte=20"`
	TransportSeatsOffered int  `json:"transport_seats_offered" validate:"gte=0,lte=20"`
	HasStroller           bool `json:"has_stroller"`
}

// AdminHouseholdPatchRequest is the body of PATCH /api/admin/households/{id}. Any
// subset of the fields.
//
// Pointers throughout, so "not sent" and "sent as empty" stay distinguishable: with
// a plain string the two are the same value and clearing an admin note would be
// impossible. `omitnil` rather than `omitempty` for the same reason — an empty
// display name has to fail the rule rather than skip it.
//
// Code is absent on purpose. Reissuing is its own endpoint, because it kills a
// printed card and signs devices out; a code changed by a stray field in a form body
// is a code nobody knows changed.
type AdminHouseholdPatchRequest struct {
	DisplayName           *string `json:"display_name" validate:"omitnil,min=1,max=120"`
	AdminNote             *string `json:"admin_note" validate:"omitnil,max=2000"`
	TransportSeatsNeeded  *int    `json:"transport_seats_needed" validate:"omitnil,gte=0,lte=20"`
	TransportSeatsOffered *int    `json:"transport_seats_offered" validate:"omitnil,gte=0,lte=20"`
	HasStroller           *bool   `json:"has_stroller"`
}

// AdminCodeReissueResponse is the body of POST /api/admin/households/{id}/code.
//
// RevokedSessions is there so the admin UI can say "der alte Code funktioniert
// jetzt nicht mehr, ein Gerät wurde abgemeldet" instead of leaving the admin to
// guess what they just did.
type AdminCodeReissueResponse struct {
	Code            string `json:"code"`
	RevokedSessions int64  `json:"revoked_sessions"`
}

// AdminGuestCreateRequest is the body of POST /api/admin/households/{id}/guests.
//
// Origin is absent by design: a guest created here is always `seeded`, and letting a
// request claim otherwise would corrupt the one column that answers "what did the
// households add themselves".
type AdminGuestCreateRequest struct {
	FirstName string `json:"first_name" validate:"required,max=80"`
	LastName  string `json:"last_name" validate:"required,max=80"`
	Kind      string `json:"kind" validate:"required,oneof=adult child"`
	// Age carries no validation tag: the kind/age pairing and the range are one
	// domain rule (domain.ResolveAge), and splitting it across a struct tag and a
	// domain function would be two places to keep in step.
	Age *int `json:"age"`
	// Empty means the column default, `normal`.
	SeatingNeed string `json:"seating_need" validate:"omitempty,oneof=normal with_parent high_chair wheelchair"`
	// An allergy list, not an essay — it has to fit on the sheet the caterer gets.
	DietaryNote string `json:"dietary_note" validate:"max=500"`
}

// AdminGuestPatchRequest is the body of PATCH /api/admin/guests/{id}. Any subset.
type AdminGuestPatchRequest struct {
	FirstName *string `json:"first_name" validate:"omitnil,min=1,max=80"`
	LastName  *string `json:"last_name" validate:"omitnil,min=1,max=80"`
	Kind      *string `json:"kind" validate:"omitnil,oneof=adult child"`
	// Age is Optional rather than *int because a null and an absent key mean
	// different things here: null clears the age, absent leaves it alone. Every
	// other field can express "clear" as an empty value.
	Age         Optional[int] `json:"age"`
	SeatingNeed *string       `json:"seating_need" validate:"omitnil,oneof=normal with_parent high_chair wheelchair"`
	DietaryNote *string       `json:"dietary_note" validate:"omitnil,max=500"`
}

// Optional distinguishes a JSON field that was absent from one sent as null.
//
// A *T cannot: both decode to nil, so an endpoint using one can never offer
// "clear this value" on a nullable column. Present is set by UnmarshalJSON, which
// the decoder calls only for a key that is actually in the body.
//
// It exists for `age` and is generic only so the next nullable number does not need
// a second copy of it.
type Optional[T any] struct {
	Present bool
	Value   *T
}

// UnmarshalJSON records that the key was present and decodes its value.
//
// DisallowUnknownFields is on for every request body (see httpio.DecodeJSON), so a
// value of the wrong type is still an error — this only widens what "absent" means,
// never what is accepted.
func (optional *Optional[T]) UnmarshalJSON(data []byte) error {
	optional.Present = true

	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		optional.Value = nil
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	optional.Value = &value
	return nil
}
