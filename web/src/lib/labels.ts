/**
 * The one and only mapping from English enum values to German UI text.
 *
 * No German string is written inline in a component. That rule is what makes
 * "enums are English everywhere" survivable: the API, the database and the Go code
 * all speak one vocabulary, and exactly one file translates it.
 *
 * Every map is typed `Record<TheUnion, ...>`, so adding a variant to enums.ts and
 * forgetting it here is a compile error rather than a raw English word appearing on
 * a guest's screen. Do not loosen these annotations to `Partial` or to a plain
 * object literal — that is the whole value of the file.
 *
 * Copy register: informal "du" throughout, short sentences, no jargon. See
 * specification/05-design.md.
 */

import type {
  Attending,
  AuditAction,
  BudgetItemStatus,
  GuestKind,
  GuestOrigin,
  MealChoice,
  Portion,
  SeatingNeed,
} from "./enums";

export const attendingLabels: Record<Attending, string> = {
  no: "Kommt nicht",
  church_only: "Nur zur Kirche",
  party_only: "Nur zur Feier",
  both: "Kirche und Feier",
};

/** Short forms for admin tables, where the full phrases wrap every column. */
export const attendingShortLabels: Record<Attending, string> = {
  no: "Nein",
  church_only: "Kirche",
  party_only: "Feier",
  both: "Beides",
};

/**
 * `guest.attending` is NULL until the household answers. The unanswered state is
 * not part of the Attending union — it is the absence of a value — so its labels
 * live beside the maps instead of inside them.
 */
export const attendingUnansweredLabel = "Noch keine Antwort";
export const attendingUnansweredShortLabel = "Offen";

export const mealChoiceLabels: Record<MealChoice, string> = {
  all: "Isst alles",
  vegetarian: "Vegetarisch",
  vegan: "Vegan",
};

export const portionLabels: Record<Portion, string> = {
  none: "Kein Essen",
  kids: "Kinderportion",
  full: "Normale Portion",
};

/**
 * Help text shown under the portion options. Only `none` needs explaining — the
 * other two are self-evident, and a hint under every option is noise that makes the
 * one that matters easy to skip.
 */
export const portionHelpLabels: Record<Portion, string | null> = {
  none: "Für Babys oder wenn jemand nicht mitisst",
  kids: null,
  full: null,
};

export const seatingNeedLabels: Record<SeatingNeed, string> = {
  normal: "Normaler Platz",
  with_parent: "Sitzt bei den Eltern (kein eigener Platz)",
  high_chair: "Hochstuhl",
  wheelchair: "Platz für Rollstuhl",
};

export const guestKindLabels: Record<GuestKind, string> = {
  adult: "Erwachsene:r",
  child: "Kind",
};

/**
 * Admin-only. `seeded` is the normal case and deliberately carries no label: a
 * marker on every ordinary guest would drown out the one on the guest a household
 * added itself.
 */
export const guestOriginLabels: Record<GuestOrigin, string | null> = {
  seeded: null,
  guest_added: "Selbst hinzugefügt",
};

/** Admin-only. Budget data never reaches a guest response. */
export const budgetItemStatusLabels: Record<BudgetItemStatus, string> = {
  planned: "Geplant",
  booked: "Gebucht",
  partially_paid: "Teilweise bezahlt",
  paid: "Bezahlt",
  cancelled: "Storniert",
};

/** Admin-only. The audit log is never shown to guests. */
export const auditActionLabels: Record<AuditAction, string> = {
  create: "Angelegt",
  update: "Geändert",
  delete: "Gelöscht",
  login: "Anmeldung",
  login_failed: "Fehlgeschlagene Anmeldung",
};
