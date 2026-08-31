/**
 * The API's response shapes, mirroring internal/infrastructure/web/dto.
 *
 * Field names are the wire names — snake_case — and are deliberately not converted
 * to camelCase. There is one client, built into the same binary that serves it, so
 * a renaming layer would buy nothing and cost the property that matters here: a
 * field name greps identically across the Go DTO, this file, and the JSON on the
 * wire. The story rule is "consume the contract exactly and invent no fields"; a
 * mapping layer is how a frontend starts inventing them.
 */

import type { Attending, GuestKind, GuestOrigin, MealChoice, Portion, SeatingNeed } from "./enums";

/**
 * The household as a guest may see it.
 *
 * `code` and `admin_note` are absent because the server never sends them — the
 * login code is the one secret in this application, and the note is written on the
 * assumption the household will never read it. Do not add them here in the hope
 * that they appear; their absence is enforced on the server and asserted by its
 * tests.
 */
export interface HouseholdSummary {
  id: number;
  display_name: string;
}

export interface Member {
  id: number;
  /** The whole name in one field — there is no first/last split in this app. */
  name: string;
  kind: GuestKind;
  origin: GuestOrigin;
}

/** The runtime switches: which sections exist, and whether the RSVP form takes input. */
export interface Flags {
  rsvp_open: boolean;
  seating_published: boolean;
  gallery_visible: boolean;
  uploads_open: boolean;
}

/**
 * The body of both POST /api/auth/login and GET /api/me.
 *
 * One type for both because the server sends one shape for both: the app must not
 * learn different things from "I just logged in" and "I was already logged in".
 */
export interface BootstrapResponse {
  household: HouseholdSummary;
  members: Member[];
  flags: Flags;
  /** RFC3339, UTC. Kept as a string; only the pages that display it need a Date. */
  rsvp_deadline: string;
}

/* ------------------------------------------------------------------------- *
 * RSVP
 *
 * One set of shapes for the guests' own /api/rsvp and for the admin's
 * /api/admin/households/{id}/rsvp: the server sends the same body on both, because
 * the admin answers a household's RSVP on the same form they would have used. A
 * second set of types here would be the first step towards a second form.
 * ------------------------------------------------------------------------- */

/**
 * The household's own half of its answer.
 *
 * `code`, `admin_note` and `rsvp_note_seen_at` are absent because the server never
 * sends them — the last of those because whether we have read a note is our business,
 * and a household that saw an unread marker would reasonably start chasing us.
 */
export interface RSVPHousehold {
  id: number;
  display_name: string;
  /** Church → reception only, and zero unless somebody in the household attends both. */
  transport_seats_needed: number;
  transport_seats_offered: number;
  has_stroller: boolean;
  /** The household's free-text note to us. Theirs, not `admin_note`, which is ours. */
  rsvp_note: string;
  /** RFC3339, or null while the household has not answered once. */
  rsvp_submitted_at: string | null;
  /** RFC3339 of the last change, or null for a household that never answered. */
  rsvp_updated_at: string | null;
}

/** One person with their answer. `null` in `attending` means "not answered yet". */
export interface RSVPMember {
  id: number;
  name: string;
  kind: GuestKind;
  /** Age at the wedding date, and null for an adult. */
  age: number | null;
  origin: GuestOrigin;
  attending: Attending | null;
  /**
   * Null both for "not answered" and for a guest whose scope excludes the party — the
   * server clears it, so the two are indistinguishable here on purpose. `attending`
   * is what says which case it is.
   */
  meal_choice: MealChoice | null;
  portion: Portion;
  midnight_snack: boolean;
  seating_need: SeatingNeed;
  dietary_note: string;
}

/** The body of GET and PUT /api/rsvp, and of the admin pair. */
export interface RSVPResponse {
  household: RSVPHousehold;
  members: RSVPMember[];
  /** RFC3339. Repeated here rather than read from `me`, so the form and its deadline
      always come from the same response. */
  deadline: string;
  /**
   * Whether the deadline has passed, as the server sees it — never computed from the
   * browser clock, which may be wrong by a year on a phone nobody has updated.
   *
   * It reports the *deadline*, not what this caller may do: the admin route sends
   * `false` here and still accepts a write (F3-B06).
   */
  editable: boolean;
}

/** One member's answer, as PUT sends it. */
export interface RSVPMemberRequest {
  id: number;
  attending: Attending;
  meal_choice: MealChoice | null;
  portion: Portion;
  midnight_snack: boolean;
  seating_need: SeatingNeed;
  dietary_note: string;
  age: number | null;
}

/**
 * The complete answer for a household — never a patch.
 *
 * `members` must list exactly the household's living members. A missing, duplicated
 * or foreign id answers 409 `member_set_mismatch`, which is the stale-tab case and is
 * handled by reloading rather than by merging.
 */
export interface RSVPSaveRequest {
  transport_seats_needed: number;
  transport_seats_offered: number;
  has_stroller: boolean;
  rsvp_note: string;
  members: RSVPMemberRequest[];
}

/** The body of POST /api/auth/admin/login and GET /api/admin/me. */
export interface AdminSession {
  subject_type: "admin";
}

/* ------------------------------------------------------------------------- *
 * Admin
 *
 * These shapes carry `code` and `admin_note`, which no guest-facing type may.
 * They come from endpoints under /api/admin, which the server refuses to a
 * household session — see the route table in internal/infrastructure/web/router.go.
 * ------------------------------------------------------------------------- */

/**
 * One row of the admin household list.
 *
 * `code` is the login code in printed form: the six stored characters, ungrouped,
 * exactly as they appear on the invitation card. There is no other form to render.
 */
export interface AdminHouseholdOverview {
  id: number;
  display_name: string;
  code: string;
  member_count: number;
  /** RFC3339, or null for a household that has never redeemed its code. */
  last_login_at: string | null;
  /** RFC3339, or null while the household has not answered. F3 fills it. */
  rsvp_submitted_at: string | null;
}

/** One household in full, as the detail page edits it. */
export interface AdminHousehold extends AdminHouseholdOverview {
  admin_note: string;
  transport_seats_needed: number;
  transport_seats_offered: number;
  has_stroller: boolean;
  members: AdminGuest[];
}

/**
 * One person as the admin sees them.
 *
 * The RSVP answers are absent because the endpoint does not send them: they belong
 * to F3's endpoint, addressed by household id, so that one shape owns the RSVP field
 * set. Do not add them here.
 */
export interface AdminGuest {
  id: number;
  household_id: number;
  name: string;
  kind: GuestKind;
  /** Age at the wedding date, and null for an adult. */
  age: number | null;
  origin: GuestOrigin;
  seating_need: SeatingNeed;
  dietary_note: string;
}

/** The body of POST /api/admin/households/{id}/code. */
export interface AdminCodeReissue {
  code: string;
  /**
   * How many logged-in devices the reissue signed out, so the screen can say what
   * happened rather than leaving the admin to infer it.
   */
  revoked_sessions: number;
}
