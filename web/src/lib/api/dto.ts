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

import type { GuestKind, GuestOrigin, SeatingNeed } from "./enums";

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
  first_name: string;
  last_name: string;
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
  first_name: string;
  last_name: string;
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
