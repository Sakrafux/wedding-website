import type {
  AdminGuest,
  AdminHousehold,
  AdminHouseholdOverview,
  AdminSession,
  BootstrapResponse,
} from "@/lib/api/dto";

/**
 * The bootstrap body, shaped exactly as the Go DTO sends it.
 *
 * Overridable field by field, because a fixture that demands every field makes a
 * test unreadable and hides which part it is actually about.
 */
export function bootstrap(overrides: Partial<BootstrapResponse> = {}): BootstrapResponse {
  return {
    household: { id: 12, display_name: "Familie Müller" },
    members: [
      { id: 30, first_name: "Anna", last_name: "Müller", kind: "adult", origin: "seeded" },
      { id: 31, first_name: "Emil", last_name: "Müller", kind: "child", origin: "seeded" },
    ],
    flags: { rsvp_open: true, seating_published: false, gallery_visible: false, uploads_open: false },
    rsvp_deadline: "2027-05-17T21:59:59Z",
    ...overrides,
  };
}

export const adminSession: AdminSession = { subject_type: "admin" };

/**
 * A household as the admin endpoints send it, `code` and `admin_note` included —
 * which is exactly what makes these responses admin-only.
 */
export function adminHousehold(overrides: Partial<AdminHousehold> = {}): AdminHousehold {
  return {
    id: 12,
    display_name: "Familie Müller",
    code: "ABC234",
    member_count: 2,
    last_login_at: "2026-11-03T18:22:00Z",
    rsvp_submitted_at: null,
    admin_note: "Kommen mit dem Zug",
    transport_seats_needed: 0,
    transport_seats_offered: 4,
    has_stroller: false,
    members: [adminGuest(), adminGuest({ id: 31, first_name: "Emil", kind: "child", age: 4 })],
    ...overrides,
  };
}

export function adminGuest(overrides: Partial<AdminGuest> = {}): AdminGuest {
  return {
    id: 30,
    household_id: 12,
    first_name: "Anna",
    last_name: "Müller",
    kind: "adult",
    age: null,
    origin: "seeded",
    seating_need: "normal",
    dietary_note: "",
    ...overrides,
  };
}

/**
 * The list row, projected from the full household so the two fixtures cannot disagree
 * about a shared field.
 */
export function adminHouseholdOverview(overrides: Partial<AdminHouseholdOverview> = {}): AdminHouseholdOverview {
  const household = adminHousehold();

  return {
    id: household.id,
    display_name: household.display_name,
    code: household.code,
    member_count: household.member_count,
    last_login_at: household.last_login_at,
    rsvp_submitted_at: household.rsvp_submitted_at,
    ...overrides,
  };
}
