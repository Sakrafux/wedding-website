package dto

import "time"

// LoginRequest is the body of POST /api/auth/login.
//
// The code arrives exactly as the guest typed it — lower case, spaces, dashes and
// all. Normalizing is the server's job, so that every client normalizes it the
// same way, which is to say once.
type LoginRequest struct {
	Code string `json:"code"`
}

// BootstrapResponse is the body of both POST /api/auth/login and GET /api/me.
//
// One shape for both, because the frontend must not learn different things from
// "I just logged in" and "I was already logged in" — a second shape is a second
// set of fields to keep in step.
type BootstrapResponse struct {
	Household HouseholdSummary `json:"household"`
	Members   []Member         `json:"members"`
	Flags     Flags            `json:"flags"`
	// RSVPDeadline lets the frontend show the date and disable the form on its own.
	// The server enforces it regardless (F3-B04); this is for the sentence above
	// the form, not for the rule.
	RSVPDeadline time.Time `json:"rsvp_deadline"`
}

// HouseholdSummary is the household as a guest may see it.
//
// Deliberately omitted, and not an oversight to be "completed" later:
//
//   - code — the login code, the one secret in this application. It is printed on
//     the card and stored in plaintext; putting it in a response would publish it
//     to anything that can read a network log or a browser cache.
//   - admin_note — our private note about the household. Written on the assumption
//     that they will never read it.
type HouseholdSummary struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
}

// Member is one person in the household.
//
// Deliberately omitted: dietary_note and every RSVP field. This response is the
// "who are you" bootstrap; the answers live behind GET /api/rsvp (F3-B02), so that
// there is one endpoint that owns the RSVP shape rather than two that must agree.
type Member struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	// Kind and Origin are the English enum values from the database. German labels
	// are the frontend's business, mapped in web/src/lib/labels.ts.
	Kind   string `json:"kind"`
	Origin string `json:"origin"`
}

// Flags are the runtime switches the frontend renders from: which sections exist
// and whether the RSVP form still accepts input.
type Flags struct {
	// RSVPOpen is derived from the deadline rather than stored, so it cannot
	// disagree with what the server enforces.
	RSVPOpen         bool `json:"rsvp_open"`
	SeatingPublished bool `json:"seating_published"`
	GalleryVisible   bool `json:"gallery_visible"`
	UploadsOpen      bool `json:"uploads_open"`
}
