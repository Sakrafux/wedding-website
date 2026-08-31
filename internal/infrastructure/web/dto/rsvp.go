package dto

import "time"

// The RSVP wire shapes, shared by the guest routes and the admin routes.
//
// One set of shapes for both, deliberately: the admin answers a household's RSVP on
// the same form the household would have used (F3-B06), so a field added for one
// caller and not the other would be a field the shared frontend component cannot
// render. No admin-only fields are added here — `code` and `admin_note` belong to the
// admin household endpoints.

// RSVPResponse is the body of GET and PUT /api/rsvp, and of the admin routes.
//
// The whole form's data in one body: a second request for the members would let the
// screen render a household with nobody in it.
type RSVPResponse struct {
	Household RSVPHousehold `json:"household"`
	Members   []RSVPMember  `json:"members"`
	// Deadline is repeated here even though /api/me carries it, because this response
	// is what the form renders from — a form reading its deadline out of another
	// endpoint's cache can show a stale date after we move it.
	Deadline time.Time `json:"deadline"`
	// Editable is the server's report of whether the deadline has passed, so the
	// frontend can render the read-only view (F3-F05) without doing date arithmetic
	// against a timezone it does not have. It is a hint about the *deadline*, not a
	// statement about whether the caller may write: F3-B04 is what refuses a write,
	// and the admin page renders inputs regardless (F3-F06).
	Editable bool `json:"editable"`
}

// RSVPHousehold is the household's own half of the answer.
//
// Deliberately omitted, and not a gap to be "completed" later:
//
//   - code — the login code, for the reasons dto.HouseholdSummary states.
//   - admin_note — our private note about the household.
//   - rsvp_note_seen_at — whether we have read their note is our business, and a
//     household that saw an unread marker would reasonably start chasing us.
type RSVPHousehold struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
	// Church → reception only. Zero when nobody in the household attends both, which
	// the server enforces rather than trusting the form to hide the fields.
	TransportSeatsNeeded  int    `json:"transport_seats_needed"`
	TransportSeatsOffered int    `json:"transport_seats_offered"`
	HasStroller           bool   `json:"has_stroller"`
	RSVPNote              string `json:"rsvp_note"`
	// RSVPSubmittedAt is null until the household has answered once. The frontend
	// reads it to choose between "Zusagen" and "Antwort ändern" in the navigation.
	RSVPSubmittedAt *time.Time `json:"rsvp_submitted_at"`
	// RSVPUpdatedAt is when the answer last changed — "zuletzt geändert am …" above
	// the form. Null for a household that has never answered.
	RSVPUpdatedAt *time.Time `json:"rsvp_updated_at"`
}

// RSVPMember is one person with their answer.
type RSVPMember struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Kind, Origin, Attending, MealChoice, Portion and SeatingNeed are the English
	// enum values from the database. German labels are the frontend's business,
	// mapped in web/src/lib/labels.ts.
	Kind string `json:"kind"`
	// Age is age at the wedding date, and null for an adult.
	Age    *int   `json:"age"`
	Origin string `json:"origin"`
	// Attending is null for a guest the household has not answered for yet, and the
	// frontend renders "Noch keine Antwort" from that. Never defaulted on the wire:
	// a defaulted "no" would be an answer nobody gave.
	Attending *string `json:"attending"`
	// MealChoice is null both for "not answered" and for a guest whose scope does not
	// cover the party — the server clears it, so the two are indistinguishable here on
	// purpose. Attending is what says which case it is.
	MealChoice    *string `json:"meal_choice"`
	Portion       string  `json:"portion"`
	MidnightSnack bool    `json:"midnight_snack"`
	SeatingNeed   string  `json:"seating_need"`
	DietaryNote   string  `json:"dietary_note"`
}

// RSVPSaveRequest is the body of PUT /api/rsvp and of the admin PUT.
//
// A complete answer, not a patch: the form is one screen with one save button, so a
// partial body is a shape no client produces, and a full replace makes the request
// idempotent on a phone that retries.
//
// Members must list exactly the household's living members — a missing, duplicated or
// foreign id is `member_set_mismatch` (409) rather than a validation error, because
// nothing in the body is malformed: the world moved.
type RSVPSaveRequest struct {
	// A household is not a coach: the same bound as the admin household endpoints, and
	// for the same reason — twenty catches a stray keystroke rather than a plausible
	// answer.
	TransportSeatsNeeded  int  `json:"transport_seats_needed" validate:"gte=0,lte=20"`
	TransportSeatsOffered int  `json:"transport_seats_offered" validate:"gte=0,lte=20"`
	HasStroller           bool `json:"has_stroller"`
	// The cap matches admin_note's. Not a product rule — a protection against a paste
	// accident, since the note is shown in full in the admin inbox.
	RSVPNote string `json:"rsvp_note" validate:"max=2000"`
	// No `required`, deliberately: a household with no living members is a real state
	// (F5 allows one), and its complete answer is an empty list. A body that omits
	// members a household *does* have is the mismatch case, which answers 409 and asks
	// for a reload — not a field error suggesting the list could be typed in.
	Members []RSVPMemberRequest `json:"members" validate:"dive"`
}

// RSVPMemberRequest is one member's answer.
//
// Kind is absent on purpose: a household turning an adult into a child would change a
// caterer bracket with a radio button, and the case that would serve — we typed the
// wrong thing — is ours to fix in F5-F02.
type RSVPMemberRequest struct {
	ID int64 `json:"id" validate:"required"`
	// Required, because there is no way to store half an answer: rsvp_submitted_at
	// means "they have told us who is coming", and a save leaving two people at null
	// while setting it would make the nudge list lie.
	Attending string `json:"attending" validate:"required,oneof=no church_only party_only both"`
	// Sent by the form even for a guest whose scope excludes the party — the state may
	// legitimately still hold a meal choice when somebody flips the scope last — and
	// dropped by the server's normalization rather than refused. Validation rejects
	// what is malformed; normalization decides what is meaningless.
	MealChoice    *string `json:"meal_choice" validate:"omitnil,oneof=all vegetarian vegan"`
	Portion       string  `json:"portion" validate:"required,oneof=none kids full"`
	MidnightSnack bool    `json:"midnight_snack"`
	SeatingNeed   string  `json:"seating_need" validate:"required,oneof=normal with_parent high_chair wheelchair"`
	// An allergy list, not an essay — it has to fit on the sheet the caterer gets.
	DietaryNote string `json:"dietary_note" validate:"max=500"`
	// Age carries no validation tag: the kind/age pairing and the range are one domain
	// rule (domain.ResolveAge), and splitting it across a struct tag and a domain
	// function would be two places to keep in step.
	Age *int `json:"age"`
}
