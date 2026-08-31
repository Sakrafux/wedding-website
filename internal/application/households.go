package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
)

// Households is the admin use case for the guest list: households and the people in
// them.
//
// Both live here rather than in two use cases because they are one task — entering
// eighty guests — and because every guest mutation has to be able to answer "which
// household is this" for the audit trail. What it deliberately does *not* do is
// touch the household's RSVP answers: those are F3's, reached through the same use
// case the guest form uses, addressed by household id. See F5-B01 for the line.
type Households struct {
	households *persistence.HouseholdStore
	guests     *persistence.GuestStore
	sessions   *persistence.SessionStore
	audit      *persistence.AuditStore

	// logger records audit writes that failed, exactly as Auth does and for the same
	// reason: a broken audit table must not fail an admin's edit, but the silence
	// must be visible somewhere.
	logger *slog.Logger
}

func NewHouseholds(
	households *persistence.HouseholdStore,
	guests *persistence.GuestStore,
	sessions *persistence.SessionStore,
	audit *persistence.AuditStore,
	logger *slog.Logger,
) *Households {
	return &Households{households: households, guests: guests, sessions: sessions, audit: audit, logger: logger}
}

// HouseholdDetail is one household with the people in it — the admin detail screen's
// whole payload in one value, so the handler makes one call and cannot render a
// household whose members failed to load.
type HouseholdDetail struct {
	Household domain.Household
	Members   []domain.Guest
}

// CodeReissue is the outcome of replacing a login code: the new code, and how many
// sessions stopped working because of it.
type CodeReissue struct {
	Code            string
	RevokedSessions int64
}

// List returns every household with its member count, ordered by name.
func (useCase *Households) List(ctx context.Context) ([]domain.HouseholdOverview, error) {
	return useCase.households.List(ctx)
}

// Detail returns one household and its living members, or ErrNotFound.
func (useCase *Households) Detail(ctx context.Context, id int64) (HouseholdDetail, error) {
	household, err := useCase.households.FindByID(ctx, id)
	if err != nil {
		return HouseholdDetail{}, translateNotFound(err)
	}

	return useCase.withMembers(ctx, household)
}

// withMembers loads the household's members and pairs them with it, so that the
// detail body is assembled the same way whichever endpoint produced the household.
func (useCase *Households) withMembers(ctx context.Context, household domain.Household) (HouseholdDetail, error) {
	members, err := useCase.households.ListMembers(ctx, household.ID)
	if err != nil {
		return HouseholdDetail{}, err
	}
	return HouseholdDetail{Household: household, Members: members}, nil
}

// Create inserts a household, which assigns it a login code in the process.
func (useCase *Households) Create(ctx context.Context, draft domain.Household) (domain.Household, error) {
	created, err := useCase.households.Create(ctx, draft)
	if err != nil {
		return domain.Household{}, err
	}

	useCase.recordAudit(ctx, domain.NewAdminChangeEntry(
		domain.AuditEntityHousehold, created.ID, domain.AuditActionCreate, time.Now(),
		// The code is created here too and is deliberately not recorded. See
		// NewAdminChangeEntry.
		domain.CreatedChanges(map[string]any{
			"display_name":            created.DisplayName,
			"admin_note":              created.AdminNote,
			"transport_seats_needed":  created.TransportSeatsNeeded,
			"transport_seats_offered": created.TransportSeatsOffered,
			"has_stroller":            created.HasStroller,
		}),
	))

	return created, nil
}

// Update applies a partial change to a household and returns it, members and all, as
// it now stands.
//
// A patch that changes nothing writes no audit row: an entry whose before and after
// are identical says only that somebody pressed save, and the log is worth reading
// precisely because every row in it is a change.
func (useCase *Households) Update(ctx context.Context, id int64, patch domain.HouseholdPatch) (HouseholdDetail, error) {
	current, err := useCase.households.FindByID(ctx, id)
	if err != nil {
		return HouseholdDetail{}, translateNotFound(err)
	}

	updated, changes := domain.ApplyHouseholdPatch(current, patch)
	if changes.IsEmpty() {
		return useCase.withMembers(ctx, current)
	}

	if err := useCase.households.Update(ctx, updated); err != nil {
		return HouseholdDetail{}, translateNotFound(err)
	}

	useCase.recordAudit(ctx, domain.NewAdminChangeEntry(
		domain.AuditEntityHousehold, id, domain.AuditActionUpdate, time.Now(), changes))

	return useCase.withMembers(ctx, updated)
}

// Delete removes a household, its guests and their seat assignments.
//
// Deleting a household with answered RSVPs is allowed. The frontend confirms it and
// names what is lost; the API does not second-guess a decision it cannot
// know the reason for. The audit row is what outlives the household, which is the
// answer to the fear that makes deleting feel dangerous.
func (useCase *Households) Delete(ctx context.Context, id int64) error {
	household, err := useCase.households.FindByID(ctx, id)
	if err != nil {
		return translateNotFound(err)
	}

	if err := useCase.households.Delete(ctx, id); err != nil {
		return translateNotFound(err)
	}

	useCase.recordAudit(ctx, domain.NewAdminChangeEntry(
		domain.AuditEntityHousehold, id, domain.AuditActionDelete, time.Now(),
		domain.DeletedChanges(map[string]any{"display_name": household.DisplayName}),
	))

	return nil
}

// ReissueCode gives the household a new login code and revokes its sessions.
//
// The revocation is the point, not a side effect: the only reason to reissue is that
// the old code should stop working, and a 365-day session issued from it would
// outlive the code by months. Before send-out this revokes nothing, which is the
// harmless case.
func (useCase *Households) ReissueCode(ctx context.Context, id int64) (CodeReissue, error) {
	code, err := useCase.households.AssignNewCode(ctx, id)
	if err != nil {
		return CodeReissue{}, translateNotFound(err)
	}

	revoked, err := useCase.sessions.DeleteForHousehold(ctx, id)
	if err != nil {
		return CodeReissue{}, err
	}

	useCase.recordAudit(ctx, domain.NewAdminChangeEntry(
		domain.AuditEntityHousehold, id, domain.AuditActionUpdate, time.Now(),
		// That the code changed, never either value: audit_log is append-only and
		// would otherwise become a permanent second copy of the key list.
		domain.CreatedChanges(map[string]any{"code_reissued": true, "revoked_sessions": revoked}),
	))

	return CodeReissue{Code: code, RevokedSessions: revoked}, nil
}

// AddGuest puts a person into a household.
//
// Guests added here are always seeded, never guest_added: origin is what the admin
// delta view reads to answer "what did the households add themselves", and an
// admin-created guest is not that. F4-B02 owns the other path, with its soft cap.
func (useCase *Households) AddGuest(ctx context.Context, householdID int64, draft domain.Guest) (domain.Guest, error) {
	// Checked first so a request naming a household that does not exist answers 404
	// rather than surfacing a foreign-key error as a 500.
	if _, err := useCase.households.FindByID(ctx, householdID); err != nil {
		return domain.Guest{}, translateNotFound(err)
	}

	age, err := domain.ResolveAge(draft.Kind, draft.Age)
	if err != nil {
		return domain.Guest{}, err
	}

	draft.HouseholdID = householdID
	draft.Age = age
	draft.Origin = domain.GuestOriginSeeded

	created, err := useCase.guests.Create(ctx, draft)
	if err != nil {
		return domain.Guest{}, err
	}

	useCase.recordAudit(ctx, domain.NewAdminChangeEntry(
		domain.AuditEntityGuest, created.ID, domain.AuditActionCreate, time.Now(),
		domain.CreatedChanges(map[string]any{
			"household_id": created.HouseholdID,
			"name":         created.Name,
			"kind":         string(created.Kind),
			"age":          created.Age,
			"seating_need": string(created.SeatingNeed),
			"dietary_note": created.DietaryNote,
		}),
	))

	return created, nil
}

// UpdateGuest applies a partial change to one guest, addressed by their own id.
//
// Not scoped to a household: the admin owns every guest, and a guest's id is the
// whole address. A household-scoped route would put an id in the path that nothing
// checks and that the frontend would have to carry for no reason.
func (useCase *Households) UpdateGuest(ctx context.Context, id int64, patch domain.GuestPatch) (domain.Guest, error) {
	current, err := useCase.guests.FindByID(ctx, id)
	if err != nil {
		return domain.Guest{}, translateNotFound(err)
	}

	updated, changes, err := domain.ApplyGuestPatch(current, patch)
	if err != nil {
		return domain.Guest{}, err
	}
	if changes.IsEmpty() {
		return current, nil
	}

	if err := useCase.guests.Update(ctx, updated); err != nil {
		return domain.Guest{}, translateNotFound(err)
	}

	useCase.recordAudit(ctx, domain.NewAdminChangeEntry(
		domain.AuditEntityGuest, id, domain.AuditActionUpdate, time.Now(), changes))

	return updated, nil
}

// RemoveGuest soft-deletes a guest.
//
// Removing the last member of a household is allowed: an empty household is a real
// state — we know a name and have not yet asked who is coming.
func (useCase *Households) RemoveGuest(ctx context.Context, id int64) error {
	guest, err := useCase.guests.FindByID(ctx, id)
	if err != nil {
		return translateNotFound(err)
	}

	if err := useCase.guests.SoftDelete(ctx, id, time.Now()); err != nil {
		return translateNotFound(err)
	}

	useCase.recordAudit(ctx, domain.NewAdminChangeEntry(
		domain.AuditEntityGuest, id, domain.AuditActionDelete, time.Now(),
		domain.DeletedChanges(map[string]any{
			"household_id": guest.HouseholdID,
			"name":         guest.Name,
		}),
	))

	return nil
}

// recordAudit appends an entry, and logs rather than propagates a failure. Same
// trade as Auth.recordAudit: a broken audit table must not fail the operation, and
// the log line is what keeps the resulting gap from reading like an event that never
// happened.
func (useCase *Households) recordAudit(ctx context.Context, entry domain.AuditEntry) {
	if err := useCase.audit.Write(ctx, entry); err != nil {
		useCase.logger.Error("audit write failed",
			"action", entry.Action, "entity", entry.Entity, "entity_id", entry.EntityID, "error", err)
	}
}

// translateNotFound maps the store's miss onto the use case layer's own sentinel,
// and passes everything else through untouched.
func translateNotFound(err error) error {
	if errors.Is(err, persistence.ErrNotFound) {
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return err
}
