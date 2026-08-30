package domain

import "time"

// Settings are the few runtime switches that can change without a redeploy. They
// live in the app_setting table; everything genuinely static — page text, the
// wedding date, the venue — is hardcoded in the frontend instead.
type Settings struct {
	// RSVPDeadline is the moment guest editing closes. Stored rather than
	// hardcoded because it is the one date that has ever plausibly moved.
	RSVPDeadline time.Time
	// DefaultAdditionLimit is the soft cap on members a household may add itself.
	// A hint, not a wall — F4-B01 owns what happens at the cap.
	DefaultAdditionLimit int
	SeatingPublished     bool
	GalleryVisible       bool
	UploadsOpen          bool
}

// RSVPOpen reports whether households may still edit their answers.
//
// Derived from the deadline rather than stored as a fourth flag, so the switch the
// frontend renders and the rule the server enforces (F3-B04) cannot disagree —
// there is no second value to forget to flip.
func (settings Settings) RSVPOpen(now time.Time) bool {
	return now.Before(settings.RSVPDeadline)
}
