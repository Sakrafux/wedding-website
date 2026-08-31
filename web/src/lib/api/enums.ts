/**
 * The enum values of the API, as string unions.
 *
 * Every enum in this app is English in the database, the API, Go and TypeScript;
 * German exists only as display labels in labels.ts. These unions are the single
 * declaration of the allowed values on the frontend — DTO types reference them
 * rather than restating a string literal list that can drift.
 */

/**
 * Attendance and its scope in one value, mirroring `guest.attending`. `null` means
 * unanswered. The states are combined on purpose: a guest who is not coming cannot
 * also have a venue scope, and that contradiction should be unrepresentable.
 */
export type Attending = "no" | "church_only" | "party_only" | "both";

/** Attendance scopes that include the party — the ones that gate catering. */
export const partyScopes = ["party_only", "both"] as const satisfies Attending[];

/** Attendance scopes that include the church service. */
export const churchScopes = ["church_only", "both"] as const satisfies Attending[];

export type MealChoice = "all" | "vegetarian" | "vegan";

/** Portion size, orthogonal to MealChoice. `none` covers infants and anyone not eating. */
export type Portion = "none" | "kids" | "full";

/** What the guest needs at the seat itself. Applies to adults too, not only children. */
export type SeatingNeed = "normal" | "with_parent" | "high_chair" | "wheelchair";

export type GuestKind = "adult" | "child";

/** Whether we seeded the guest from the invite list or the household added them. */
export type GuestOrigin = "seeded" | "guest_added";

export type BudgetItemStatus = "planned" | "booked" | "partially_paid" | "paid" | "cancelled";

export type AuditAction = "create" | "update" | "delete" | "login" | "login_failed";
