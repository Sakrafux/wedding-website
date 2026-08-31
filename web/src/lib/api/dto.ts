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

import type { GuestKind, GuestOrigin } from "./enums";

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
