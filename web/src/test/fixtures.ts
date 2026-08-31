import type { AdminSession, BootstrapResponse } from "@/lib/api/dto";

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
