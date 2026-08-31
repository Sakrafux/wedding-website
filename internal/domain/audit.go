package domain

import "time"

// ActorType is who performed an audited action. Wider than SubjectType because the
// application itself acts: a failed login has no actor to name, and attributing it
// to a household would be a guess.
type ActorType string

const (
	ActorTypeHousehold ActorType = "household"
	ActorTypeAdmin     ActorType = "admin"
	ActorTypeSystem    ActorType = "system"
)

// AuditAction is what happened. The values are the ones audit_log's CHECK
// constraint allows; adding one is a migration, which is the intended friction.
type AuditAction string

const (
	AuditActionCreate      AuditAction = "create"
	AuditActionUpdate      AuditAction = "update"
	AuditActionDelete      AuditAction = "delete"
	AuditActionLogin       AuditAction = "login"
	AuditActionLoginFailed AuditAction = "login_failed"
)

// Audited entity names — what an action was done to.
//
// Each value is spelled exactly like the SQLite table it names, so that a query
// typed by hand months from now (`WHERE entity = 'household' AND entity_id = 12`)
// needs no lookup table to read and hits the (entity, entity_id) index built for it.
// The rule matters when F6 adds 'budget_item': the name of the table is the name of
// the entity, always.
const (
	AuditEntityHousehold = "household"
	AuditEntityGuest     = "guest"
	AuditEntityAdmin     = "admin"
)

// AuditEntityNone is the entity id for an event that concerns no particular row: a
// failed login, where the household is by definition unknown, or an admin login,
// where there is no admin table to point at.
//
// Zero rather than NULL because audit_log.entity_id is NOT NULL, and it is NOT NULL
// because every other event in the system does concern a row. Treating the one
// exception as a sentinel is cheaper than making the column nullable and then
// checking for NULL in every later query.
const AuditEntityNone int64 = 0

// AuditEntry is one append-only record of something having happened.
//
// It answers two separate questions, which is why there are two pairs of columns
// rather than one:
//
//   - the actor — *who did it*: ActorType plus ActorID.
//   - the entity — *what it was done to*: Entity plus EntityID.
//
// They coincide for a household login and diverge everywhere else, which is the
// whole reason for keeping both:
//
//	event                 actor             entity
//	household logs in     household / 12    household / 12
//	admin logs in         admin / nil       admin / 0
//	failed login          system / nil      household or admin / 0   (which door was tried)
//	admin edits a guest   admin / nil       guest / 47
//
// Before and After hold the changed fields only, never whole rows: the log is a
// history, not a second copy of the database, and a full snapshot of a guest row
// would duplicate personal data into a table nobody ever deletes from.
type AuditEntry struct {
	At        time.Time
	ActorType ActorType
	// ActorID is the household id, or nil for the admin and for the system — both
	// of which have no row to reference.
	ActorID  *int64
	Entity   string
	EntityID int64
	Action   AuditAction
	Before   map[string]any
	After    map[string]any
}

// NewHouseholdLoginEntry records a household redeeming its code.
//
// The entity is the household rather than the session: sessions are identified by
// a token hash, which is text and could not go in entity_id anyway, and the useful
// question later is "what has this household done", which this makes one indexed
// query. The session is not the interesting noun; the household is.
func NewHouseholdLoginEntry(householdID int64, at time.Time, userAgent, ip string) AuditEntry {
	return AuditEntry{
		At:        at,
		ActorType: ActorTypeHousehold,
		ActorID:   &householdID,
		Entity:    AuditEntityHousehold,
		EntityID:  householdID,
		Action:    AuditActionLogin,
		After:     connectionDetails(userAgent, ip),
	}
}

// NewAdminLoginEntry records the admin logging in. No actor id and no entity id:
// there is exactly one admin and no row anywhere that describes them.
func NewAdminLoginEntry(at time.Time, userAgent, ip string) AuditEntry {
	return AuditEntry{
		At:        at,
		ActorType: ActorTypeAdmin,
		Entity:    AuditEntityAdmin,
		EntityID:  AuditEntityNone,
		Action:    AuditActionLogin,
		After:     connectionDetails(userAgent, ip),
	}
}

// NewLoginFailureEntry records a rejected login attempt.
//
// The actor is the system, because nobody has been identified — that is what
// "failed" means here. The entity names which door was tried, so that a run of
// attempts against the admin login is visible as such rather than buried among
// mistyped guest codes.
func NewLoginFailureEntry(entity string, at time.Time, userAgent, ip string) AuditEntry {
	return AuditEntry{
		At:        at,
		ActorType: ActorTypeSystem,
		Entity:    entity,
		EntityID:  AuditEntityNone,
		Action:    AuditActionLoginFailed,
		After:     connectionDetails(userAgent, ip),
	}
}

// connectionDetails is what an auth event records about the caller.
//
// The submitted code or password is deliberately absent, and must stay absent. A
// log of near-misses is a partial key list, and a log of typos will eventually
// contain some other household's real code — typed by a guest holding the wrong
// card. "Log the input so we can debug it" is exactly the well-meaning change that
// would turn this table into the thing the login code is protected from.
func connectionDetails(userAgent, ip string) map[string]any {
	return map[string]any{"ip": ip, "user_agent": userAgent}
}

// NewAdminChangeEntry records the admin creating, updating or deleting a row.
//
// One constructor for all three actions and both entities, because the shape is the
// same and the only thing that differs is which of Before and After is populated —
// which Changes already decides. A per-endpoint constructor would be six functions
// that must not drift.
//
// The actor has no id: there is exactly one admin and no row anywhere describing
// them, so ActorType is the whole answer to who did this.
//
// **Never pass a login code in changes.** Neither the old value nor the new one:
// audit_log is append-only and nobody ever deletes from it, so a code recorded here
// would turn the table into a second, permanent copy of the key list. A code change
// is recorded as the fact that it happened — see the code_changed flag the
// regeneration path passes.
func NewAdminChangeEntry(entity string, entityID int64, action AuditAction, at time.Time, changes Changes) AuditEntry {
	return AuditEntry{
		At:        at,
		ActorType: ActorTypeAdmin,
		Entity:    entity,
		EntityID:  entityID,
		Action:    action,
		Before:    changes.Before,
		After:     changes.After,
	}
}
