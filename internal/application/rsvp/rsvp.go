// Package rsvp is the use case for a household's answer: reading it, and writing it
// whole.
//
// One use case for both callers. The guest route passes the household from the
// session, the admin route passes the id from the path (F3-B06), and the difference
// between them is two arguments — never a second copy of the rules. Ownership is
// therefore checked in exactly one place, and the two routes cannot drift apart in
// which fields they accept.
package rsvp

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/Sakrafux/wedding-website/internal/application"
	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
)

// UseCase reads and writes RSVP answers.
type UseCase struct {
	households *persistence.HouseholdStore
	guests     *persistence.GuestStore
	answers    *persistence.RSVPStore
	settings   *persistence.SettingStore
	audit      *persistence.AuditStore

	// logger records audit writes that failed, as Auth and Households do: the answer
	// matters more than the record of it, but the silence has to be visible somewhere.
	logger *slog.Logger
}

func New(
	households *persistence.HouseholdStore,
	guests *persistence.GuestStore,
	answers *persistence.RSVPStore,
	settings *persistence.SettingStore,
	audit *persistence.AuditStore,
	logger *slog.Logger,
) *UseCase {
	return &UseCase{
		households: households,
		guests:     guests,
		answers:    answers,
		settings:   settings,
		audit:      audit,
		logger:     logger,
	}
}

// Answer is a household's whole RSVP state: the household's own fields, its living
// members with their answers, and the deadline that governs both.
//
// Everything the form renders from, in one value, so a screen can never show a
// household with nobody in it because a second query failed.
type Answer struct {
	Household domain.Household
	Members   []domain.Guest
	Deadline  time.Time
	// Editable is Settings.RSVPOpen at the moment of reading — the honest report of
	// the deadline, and not a statement about whether *this* caller may write. The
	// admin writes after it turns false (F3-B06).
	Editable bool
	// CanAddPlusOne is domain.CanHouseholdAddPlusOne over these members. A boolean
	// rather than the counts behind it, because the form has exactly one decision to
	// make and two integers would invite it to re-derive the rule and get it wrong
	// (F4-B02).
	CanAddPlusOne bool
}

// Submission is a household's complete answer as a caller sends it.
//
// Complete, never partial: the form is one screen with one save button, so a partial
// body is a shape no client produces, and a full replace is idempotent on a phone
// that retries. Members must be exactly the household's living members — see
// matchSubmissionToMembers.
type Submission struct {
	Household domain.HouseholdAnswer
	Members   []MemberSubmission
}

// MemberSubmission is one person's answer, addressed by their guest id.
type MemberSubmission struct {
	ID     int64
	Answer domain.GuestAnswer
}

// SaveOptions is who is saving and under which rules.
//
// EnforceDeadline is an argument rather than a rule inside Save because the admin
// path exists precisely for the answer that arrives late (F3-B04). ActorType decides
// what the audit log records, and the two are set together on purpose: an admin save
// that enforced the deadline, or a guest save recorded as ours, would each be a
// contradiction nobody would notice until the log was being read.
type SaveOptions struct {
	EnforceDeadline bool
	ActorType       domain.ActorType
}

// GuestOptions are the rules for a save a household makes itself.
func GuestOptions() SaveOptions {
	return SaveOptions{EnforceDeadline: true, ActorType: domain.ActorTypeHousehold}
}

// AdminOptions are the rules for an answer we take down the phone: no deadline, and
// recorded as ours (F3-B05).
func AdminOptions() SaveOptions {
	return SaveOptions{EnforceDeadline: false, ActorType: domain.ActorTypeAdmin}
}

// MemberAgeError reports a rejected age together with the member it belongs to, so
// the web layer can key the field error to the right card.
//
// The rule itself is the domain's (domain.ResolveAge) and the wording is the web
// layer's; the member id is the only part that is this layer's to add.
type MemberAgeError struct {
	MemberID int64
	Err      error
}

func (e MemberAgeError) Error() string {
	return fmt.Sprintf("member %d: %s", e.MemberID, e.Err)
}

func (e MemberAgeError) Unwrap() error {
	return e.Err
}

// Load returns the household's answer as it stands, or application.ErrNotFound.
//
// Reading is never refused by the deadline: a household must be able to see what they
// answered, which is exactly what the read-only view renders (F3-F05).
func (useCase *UseCase) Load(ctx context.Context, householdID int64) (Answer, error) {
	household, err := useCase.households.FindByID(ctx, householdID)
	if err != nil {
		return Answer{}, application.TranslateNotFound(err)
	}

	return useCase.answerFor(ctx, household, time.Now())
}

// EnsureWritable reports rsvp_closed when the deadline governs this caller and has
// passed, and nothing otherwise.
//
// It exists for the *order* of the answer, not for the rule: a closed form must say
// that it is closed rather than list three field errors first (F3-B04), and field
// errors are the web layer's, produced before a use case is reached at all. The rule
// itself still lives in Save, which checks again and is the only authority — a caller
// that forgets this guard gets a worse message, never a write it should not have.
func (useCase *UseCase) EnsureWritable(ctx context.Context, options SaveOptions) error {
	if !options.EnforceDeadline {
		return nil
	}

	settings, err := useCase.settings.Load(ctx)
	if err != nil {
		return err
	}
	if !settings.RSVPOpen(time.Now()) {
		return domain.NewError(domain.CodeRSVPClosed)
	}
	return nil
}

// Save writes the household's whole answer and returns it as stored.
//
// The order is deliberate: deadline, then the member set, then the field rules. A
// closed form should say that it is closed rather than list three field errors first
// (F3-B04), and a body describing a member set that no longer exists must not have
// its fields validated as if the answer could be stored.
//
// A save that changes nothing writes nothing — neither row, nor timestamp, nor audit
// entry. Same rule as Households.Update, and here it has a second reason: F6 compares
// rsvp_note_seen_at against rsvp_updated_at to decide whether a note is unread, so a
// no-op save would silently re-flag a note we have already read.
func (useCase *UseCase) Save(ctx context.Context, householdID int64, submission Submission, options SaveOptions) (Answer, error) {
	// Once per request, so that everything in this save — the deadline check and both
	// timestamps — agrees about what "now" is.
	now := time.Now()

	settings, err := useCase.settings.Load(ctx)
	if err != nil {
		return Answer{}, err
	}
	if options.EnforceDeadline && !settings.RSVPOpen(now) {
		return Answer{}, domain.NewError(domain.CodeRSVPClosed)
	}

	household, err := useCase.households.FindByID(ctx, householdID)
	if err != nil {
		return Answer{}, application.TranslateNotFound(err)
	}

	members, err := useCase.households.ListMembers(ctx, householdID)
	if err != nil {
		return Answer{}, err
	}

	answers, err := matchSubmissionToMembers(submission.Members, members)
	if err != nil {
		return Answer{}, err
	}

	updatedMembers := make([]domain.Guest, 0, len(members))
	memberChanges := make(map[int64]domain.Changes, len(members))
	for _, member := range members {
		updated, changes, err := domain.ApplyGuestAnswer(member, answers[member.ID])
		if err != nil {
			return Answer{}, MemberAgeError{MemberID: member.ID, Err: err}
		}
		updatedMembers = append(updatedMembers, updated)
		memberChanges[member.ID] = changes
	}

	// After the members are normalized, because the transport rule reads their
	// scopes and a stale one would keep seat counts the answer just gave up.
	updatedHousehold, householdChanges := domain.ApplyHouseholdAnswer(household, submission.Household, updatedMembers)

	if householdChanges.IsEmpty() && !anyMemberChanged(memberChanges) {
		return useCase.answerFor(ctx, household, now)
	}

	updatedHousehold.RSVPUpdatedAt = &now
	// Set on the first save and never moved afterwards: the pair of timestamps is what
	// F6 reads for "answered" versus "changed since we last looked". An admin save sets
	// it too — if we took the answer down the phone, the household has answered and
	// must drop off the nudge list (F3-B06).
	if updatedHousehold.RSVPSubmittedAt == nil {
		updatedHousehold.RSVPSubmittedAt = &now
	}

	if err := useCase.answers.SaveAnswer(ctx, updatedHousehold, updatedMembers); err != nil {
		return Answer{}, application.TranslateNotFound(err)
	}

	useCase.recordChanges(ctx, updatedHousehold.ID, householdChanges, memberChanges, options.ActorType, now)

	// Re-read rather than answering with the in-memory copy: the response is the
	// answer *as stored*, timestamps included, and the stored form is UTC at second
	// precision. Answering with the values we happened to hold would put a
	// nanosecond, local-zone timestamp on the wire that the next GET contradicts.
	saved, err := useCase.households.FindByID(ctx, householdID)
	if err != nil {
		return Answer{}, application.TranslateNotFound(err)
	}

	return useCase.answerFor(ctx, saved, now)
}

// Addition is a plus-one as it was stored, with the household's right to add another
// recomputed — which after a successful addition is always false.
//
// Returned together so the form can append the card and re-render the trigger from one
// response, instead of refetching the whole answer to learn what it already knows.
type Addition struct {
	Member        domain.Guest
	CanAddPlusOne bool
}

// AddPlusOne adds one adult companion to a household of one and returns them.
//
// The rule is domain.CanHouseholdAddPlusOne and it is checked inside the write
// transaction, not here: see persistence.CreateIfHouseholdAllows. The deadline is
// checked first, because a closed form is closed regardless of who is being added.
//
// Nothing but the name is taken. Kind, origin and every answer field come from
// domain.NewPlusOne, so no request body can produce a child or a pre-answered guest.
func (useCase *UseCase) AddPlusOne(ctx context.Context, householdID int64, name string, options SaveOptions) (Addition, error) {
	now := time.Now()

	if err := useCase.EnsureWritable(ctx, options); err != nil {
		return Addition{}, err
	}

	// Checked before the insert so a household id that does not exist answers 404
	// rather than surfacing a foreign-key error as a 500.
	if _, err := useCase.households.FindByID(ctx, householdID); err != nil {
		return Addition{}, application.TranslateNotFound(err)
	}

	created, err := useCase.guests.CreateIfHouseholdAllows(
		ctx, domain.NewPlusOne(householdID, name), domain.CanHouseholdAddPlusOne)
	if err != nil {
		return Addition{}, err
	}

	// The row that later answers "where did this person come from", which is the
	// question the admin delta view is built on.
	useCase.recordAudit(ctx, useCase.creationEntry(options.ActorType, householdID, created, now))

	members, err := useCase.households.ListMembers(ctx, householdID)
	if err != nil {
		return Addition{}, err
	}

	return Addition{Member: created, CanAddPlusOne: domain.CanHouseholdAddPlusOne(members) == nil}, nil
}

// RemoveMember soft-deletes a member a household added itself.
//
// A guest belonging to another household is reported as not found rather than as
// forbidden: a household must not be able to learn which ids exist by reading the
// difference between the two answers.
func (useCase *UseCase) RemoveMember(ctx context.Context, householdID, guestID int64, options SaveOptions) error {
	now := time.Now()

	if err := useCase.EnsureWritable(ctx, options); err != nil {
		return err
	}

	member, err := useCase.guests.FindByID(ctx, guestID)
	if err != nil {
		return application.TranslateNotFound(err)
	}
	if member.HouseholdID != householdID {
		return application.ErrNotFound
	}

	if err := domain.CanHouseholdRemove(member); err != nil {
		return err
	}

	if err := useCase.guests.SoftDelete(ctx, guestID, now); err != nil {
		return application.TranslateNotFound(err)
	}

	// The row that explains a headcount that went down. The name is in the payload
	// because the guest row itself stays and the audit trail is read by household.
	useCase.recordAudit(ctx, useCase.deletionEntry(options.ActorType, householdID, member, now))

	return nil
}

// creationEntry and deletionEntry build the audit entry for whoever acted, the same
// way changeEntry does for a save: an admin entry carries no actor id, because there
// is no admin row to point at.
func (useCase *UseCase) creationEntry(
	actorType domain.ActorType, householdID int64, member domain.Guest, at time.Time,
) domain.AuditEntry {
	changes := domain.CreatedChanges(map[string]any{
		"household_id": member.HouseholdID,
		"name":         member.Name,
		"kind":         string(member.Kind),
		"origin":       string(member.Origin),
	})
	if actorType == domain.ActorTypeAdmin {
		return domain.NewAdminChangeEntry(domain.AuditEntityGuest, member.ID, domain.AuditActionCreate, at, changes)
	}
	return domain.NewHouseholdChangeEntry(
		householdID, domain.AuditEntityGuest, member.ID, domain.AuditActionCreate, at, changes)
}

func (useCase *UseCase) deletionEntry(
	actorType domain.ActorType, householdID int64, member domain.Guest, at time.Time,
) domain.AuditEntry {
	changes := domain.DeletedChanges(map[string]any{
		"household_id": member.HouseholdID,
		"name":         member.Name,
		"origin":       string(member.Origin),
	})
	if actorType == domain.ActorTypeAdmin {
		return domain.NewAdminChangeEntry(domain.AuditEntityGuest, member.ID, domain.AuditActionDelete, at, changes)
	}
	return domain.NewHouseholdChangeEntry(
		householdID, domain.AuditEntityGuest, member.ID, domain.AuditActionDelete, at, changes)
}

// answerFor assembles the response value for a household that is already loaded.
//
// The members are re-read rather than reused from a caller's slice, so that what a
// save returns is what the database holds — which is the whole point of answering with
// the stored, normalized answer rather than with the submitted one.
func (useCase *UseCase) answerFor(ctx context.Context, household domain.Household, now time.Time) (Answer, error) {
	members, err := useCase.households.ListMembers(ctx, household.ID)
	if err != nil {
		return Answer{}, err
	}

	settings, err := useCase.settings.Load(ctx)
	if err != nil {
		return Answer{}, err
	}

	return Answer{
		Household:     household,
		Members:       members,
		Deadline:      settings.RSVPDeadline,
		Editable:      settings.RSVPOpen(now),
		CanAddPlusOne: domain.CanHouseholdAddPlusOne(members) == nil,
	}, nil
}

// matchSubmissionToMembers pairs each submitted answer with the member it addresses,
// and refuses anything but an exact match.
//
// A member missing from the body, a duplicate id, or an id the household does not own
// are all one failure: the body describes a different household than the one that
// exists. This is the stale-tab case — a household that added a plus-one on a phone
// and still has the form open on a laptop — and refusing is the only answer that
// cannot lose an answer somebody gave.
func matchSubmissionToMembers(submitted []MemberSubmission, members []domain.Guest) (map[int64]domain.GuestAnswer, error) {
	if len(submitted) != len(members) {
		return nil, domain.NewError(domain.CodeMemberSetMismatch)
	}

	answers := make(map[int64]domain.GuestAnswer, len(submitted))
	for _, member := range submitted {
		if _, isDuplicate := answers[member.ID]; isDuplicate {
			return nil, domain.NewError(domain.CodeMemberSetMismatch)
		}
		answers[member.ID] = member.Answer
	}

	// Counts already agree, so every member being present means the two sets are
	// equal — an id from another household would leave one of ours unmatched.
	for _, member := range members {
		if _, isAnswered := answers[member.ID]; !isAnswered {
			return nil, domain.NewError(domain.CodeMemberSetMismatch)
		}
	}
	return answers, nil
}

func anyMemberChanged(changes map[int64]domain.Changes) bool {
	for _, memberChanges := range changes {
		if !memberChanges.IsEmpty() {
			return true
		}
	}
	return false
}

// recordChanges writes one audit row per entity that actually changed: the household
// when its own fields moved, and one per member whose answer moved.
//
// One row per entity rather than one per save, because the household and each member
// are separate rows with separate ids and a single entry covering five people could
// not be read back against any of them.
func (useCase *UseCase) recordChanges(
	ctx context.Context,
	householdID int64,
	householdChanges domain.Changes,
	memberChanges map[int64]domain.Changes,
	actorType domain.ActorType,
	at time.Time,
) {
	if !householdChanges.IsEmpty() {
		useCase.recordAudit(ctx, useCase.changeEntry(
			actorType, householdID, domain.AuditEntityHousehold, householdID, at, householdChanges))
	}

	// In id order, so a hand-typed query against audit_log reads the members of one
	// save in the order the form showed them. Map iteration would shuffle them.
	for _, memberID := range sortedIDs(memberChanges) {
		useCase.recordAudit(ctx, useCase.changeEntry(
			actorType, householdID, domain.AuditEntityGuest, memberID, at, memberChanges[memberID]))
	}
}

// changeEntry builds the audit entry for whoever saved.
//
// An admin entry carries no actor id — there is no admin row to point at, and the
// entity already names the household the change was about (F3-B05).
func (useCase *UseCase) changeEntry(
	actorType domain.ActorType,
	householdID int64,
	entity string,
	entityID int64,
	at time.Time,
	changes domain.Changes,
) domain.AuditEntry {
	if actorType == domain.ActorTypeAdmin {
		return domain.NewAdminChangeEntry(entity, entityID, domain.AuditActionUpdate, at, changes)
	}
	return domain.NewHouseholdChangeEntry(householdID, entity, entityID, domain.AuditActionUpdate, at, changes)
}

// recordAudit appends an entry, and logs rather than propagates a failure. Same trade
// as Auth and Households: a broken audit table must not lose a household's answer,
// and the log line is what keeps the resulting gap from reading like a save that never
// happened.
func (useCase *UseCase) recordAudit(ctx context.Context, entry domain.AuditEntry) {
	if err := useCase.audit.Write(ctx, entry); err != nil {
		useCase.logger.Error("audit write failed",
			"action", entry.Action, "entity", entry.Entity, "entity_id", entry.EntityID, "error", err)
	}
}

func sortedIDs(changes map[int64]domain.Changes) []int64 {
	ids := make([]int64, 0, len(changes))
	for id, memberChanges := range changes {
		if memberChanges.IsEmpty() {
			continue
		}
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}
