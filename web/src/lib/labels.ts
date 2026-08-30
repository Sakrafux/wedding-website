/**
 * Every German string in the application.
 *
 * Two kinds live here. First, the mapping from English enum values to German UI
 * text — the reason the file exists. Second, the screen copy: headings, buttons,
 * help text. No German string is written inline in a component, ever.
 *
 * That rule is what makes "enums are English everywhere" survivable: the API, the
 * database and the Go code all speak one vocabulary, and exactly one file
 * translates it. Keeping the screen copy here too means the whole of what a guest
 * can read is one file to proof-read, in a product whose text is read once, by
 * eighty people, and cannot be quietly fixed afterwards.
 *
 * Error messages are the deliberate exception and are *not* here: the API sends the
 * German sentence with the failure, and the frontend shows it verbatim. See
 * httpio's message table for why that lives on the server.
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

/* ------------------------------------------------------------------------- *
 * Screen copy
 * ------------------------------------------------------------------------- */

/**
 * Where a guest is told to turn when the code will not work.
 *
 * TODO(content): the real number, before the cards go to print. It appears on the
 * login screen after two failed attempts and is the only escape hatch a guest has
 * — a placeholder reaching the print run would strand exactly the people this
 * fallback exists for. Tracked in TODO.md.
 */
export const contactPhoneNumber = "+49 000 0000000";

export const loginLabels = {
  heading: "Willkommen",
  intro: "Gib den Code ein, der auf deiner Einladungskarte steht.",
  codeLabel: "Dein Code von der Einladungskarte",
  /** Help text under the field. Never a placeholder inside it: a placeholder is not
      a label, and it disappears at the moment it would be most useful. */
  codeHint: "Sechs Zeichen, zum Beispiel ABC-234. Groß- und Kleinschreibung ist egal.",
  submit: "Anmelden",
  submitting: "Wird geprüft …",
  /** Revealed after two failures — the point at which a person starts to feel
      stupid, which is exactly when the way out should appear. */
  fallback: "Klappt es nicht? Ruf uns an:",
  loggedOut: "Du wurdest abgemeldet. Bitte melde dich noch einmal an.",
  rejectedHousehold: "Kein Problem — bitte prüf den Code noch einmal.",
} as const;

export const confirmationLabels = {
  /** Names the household, because the whole screen exists to catch a valid code
      that belongs to somebody else. */
  heading: (householdName: string) => `Willkommen, ${householdName} — seid ihr das?`,
  membersIntro: "So haben wir euch notiert:",
  confirm: "Ja, das sind wir",
  reject: "Nein",
} as const;

export const shellLabels = {
  loading: "Einen Moment …",
  errorHeading: "Da ist etwas schiefgegangen",
  retry: "Nochmal versuchen",
  /** Shown with the id from the error envelope, so a guest on the phone can read
      out something that finds their request in the log. */
  requestId: "Fehlernummer:",
  /** The authenticated landing page until F2-F02 builds the real start page. */
  startHeading: "Ihr seid angemeldet",
  startIntro: "Die Einladung mit allen Infos kommt hier in Kürze.",
  logout: "Abmelden",
} as const;

export const adminLabels = {
  heading: "Verwaltung",
  userLabel: "Benutzername",
  passwordLabel: "Passwort",
  submit: "Anmelden",
  submitting: "Wird geprüft …",
  sessionExpired: "Sitzung abgelaufen. Bitte melde dich erneut an.",
  logout: "Abmelden",
  /** Nav entries whose pages do not exist yet render disabled — better than a 404,
      and it keeps the remaining work visible. */
  comingSoon: "Kommt später",
  navHouseholds: "Haushalte",
  navDashboard: "Dashboard",
  navSeating: "Sitzplan",
  navBudget: "Budget",
  navPhotos: "Fotos",
} as const;
