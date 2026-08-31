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
} from "./api/enums";
import { weddingDateLong, weddingDateShort } from "./wedding";

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
 * One number, not both: somebody whose code has just failed twice needs a person to
 * ring, not a choice. Both numbers are listed on the Kontakt page (F2-F06).
 */
export const contactPhoneNumber = "+43 650 9408100";

/**
 * The Kontakt page's list, and the source for any copy naming a specific person.
 * Written out here rather than in a component, like every other German string.
 */
export const contacts = [
  { name: "Andreas Hell", phone: contactPhoneNumber },
  { name: "Isabella Michelbacher", phone: "+43 677 63668655" },
] as const;

export const loginLabels = {
  heading: "Willkommen",
  intro: "Gib den Code ein, der auf deiner Einladungskarte steht.",
  codeLabel: "Dein Code von der Einladungskarte",
  /** Help text under the field. Never a placeholder inside it: a placeholder is not
      a label, and it disappears at the moment it would be most useful. */
  codeHint: "Sechs Zeichen, genau wie auf der Karte. Groß- und Kleinschreibung ist egal.",
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
  /** Lives on /mehr rather than in the bottom bar: it is used once a year, and a
      fifth bar item that logs you out beside the one showing the schedule is a
      mis-tap waiting to happen. */
  logout: "Abmelden",
} as const;

/**
 * Copy shared by every form control, rather than by one screen.
 *
 * The accessible names in here are the reason they are functions: a help button called
 * "?" and a stepper button called "+" are unreadable out of context, and eight of each
 * on one page are indistinguishable. Naming the field is what makes them usable.
 */
export const formLabels = {
  helpFor: (field: string) => `Hilfe zu ${field}`,
  stepperDecrease: (field: string) => `Einen weniger: ${field}`,
  stepperIncrease: (field: string) => `Einen mehr: ${field}`,
} as const;

/**
 * The RSVP form — the screen this whole product exists for.
 *
 * Every field carries a `*Help` sentence, per the form rules in 05-design: the help
 * lives behind a `?` popover beside the label, so a guest who needs the explanation
 * gets it and the form does not become a wall of grey text for everybody else.
 *
 * The register is the same "du" as everywhere else, addressed to a household — so the
 * questions are plural where they concern the group ("Wozu kommt ihr?") and singular
 * where they concern one person ("Was isst Anna?").
 */
export const rsvpLabels = {
  heading: "Sagt uns Bescheid",
  intro: "Bitte sagt uns, wer von euch kommt. Ändern könnt ihr die Antwort jederzeit bis zum Stichtag.",
  deadlineNotice: (date: string) => `Bitte antwortet bis zum ${date}.`,
  lastChanged: (date: string) => `Zuletzt geändert am ${date}.`,

  householdScopeHeading: "Wir kommen zu:",
  householdScopeHelp:
    "Das setzt die Antwort für alle Personen unten auf einmal. Danach kannst du einzelne Personen noch ändern.",
  /** Shown instead of a lit-up option once the cards disagree: a selector that stayed
      lit while the cards said something else would be a lie about what gets saved. */
  householdScopeMixed: "Ihr habt unterschiedliche Antworten — siehe unten.",
  householdScopeOverwriteTitle: "Antworten überschreiben?",
  householdScopeOverwriteBody: (changed: number, scope: string) =>
    changed === 1
      ? `Damit wird die Antwort einer Person auf „${scope}“ geändert.`
      : `Damit werden die Antworten von ${changed} Personen auf „${scope}“ geändert.`,
  householdScopeOverwriteConfirm: "Ja, für alle setzen",
  cancel: "Abbrechen",
  close: "Schließen",

  membersHeading: "Wer kommt?",
  /** A statement of fact in the muted ink, not an error: a form that opens red at a
      household who has not answered yet reads as broken. */
  memberUnanswered: "Noch keine Antwort",
  memberScopeLabel: (name: string) => `Wozu kommt ${name}?`,
  memberScopeHelp:
    "Kirche und Feier sind getrennt — manche kommen nur zur Trauung oder nur zum Fest. Wähl aus, was für diese Person passt.",
  mealChoiceLabel: (name: string) => `Was isst ${name}?`,
  mealChoiceHelp: "Wir geben dem Catering nur die Zahlen weiter, keine Namen. „Isst alles“ heißt: kein Sonderwunsch.",
  portionLabel: (name: string) => `Portion für ${name}`,
  portionHelp:
    "Kinderportionen sind kleiner, „Kein Essen“ ist für Babys oder wenn jemand nicht mitisst. Beides kostet uns nichts extra — sag ruhig, was wirklich passt.",
  midnightSnackLabel: (name: string) => `Mitternachtssnack für ${name}`,
  midnightSnackHelp:
    "Spät am Abend gibt es noch eine kleine Stärkung. Wir müssen vorher wissen, für wie viele — wer schon im Bett ist, braucht keinen.",
  midnightSnackHint: "Kleine Stärkung spät am Abend",
  seatingNeedLabel: (name: string) => `Platz für ${name}`,
  seatingNeedHelp:
    "Damit planen wir die Plätze in der Kirche und am Tisch: Rollstuhl, Hochstuhl, oder ein Kind, das bei den Eltern sitzt.",
  dietaryNoteLabel: (name: string) => `Allergien oder Unverträglichkeiten von ${name}`,
  dietaryNoteHelp:
    "Was die Küche wissen muss, zum Beispiel „Nussallergie“ oder „laktosefrei“. Kurz reicht — wir fragen nach, wenn etwas unklar ist.",
  dietaryNotePlaceholderHint: "Zum Beispiel: Nussallergie, laktosefrei",
  /** The date is in the label, not only in the help: a bare "Alter" gets answered as
      of today, and then the value drifts over the months before the wedding. */
  ageLabel: (name: string) => `Alter von ${name} am Hochzeitstag, ${weddingDateLong}`,
  ageHelp: `Gemeint ist das Alter am ${weddingDateLong}, nicht heute. Das Catering rechnet nach Alter am Tag der Feier.`,

  transportHeading: "Fahrt von der Kirche zur Feier",
  transportNeededLabel: "Plätze gesucht",
  transportNeededHelp:
    "Wie viele von euch bräuchten eine Mitfahrgelegenheit von der Kirche zur Feier? Wir sehen daran, ob sich ein Shuttle lohnt — wir vermitteln keine Fahrgemeinschaften.",
  transportOfferedLabel: "Plätze angeboten",
  transportOfferedHelp:
    "Wie viele freie Plätze hättet ihr im Auto von der Kirche zur Feier? Auch das ist nur für die Planung; wer mitfährt, klärt sich vor Ort.",
  /** Said rather than silently discarded: a number that vanishes without explanation
      is exactly the kind of thing that gets phoned in about. */
  transportDropped:
    "Die Angaben zur Fahrt gelten nicht mehr, weil niemand von euch zu Kirche und Feier kommt. Wir speichern sie nicht.",
  strollerLabel: "Wir bringen einen Kinderwagen mit",
  strollerHelp:
    "Ein Kinderwagen braucht Platz zum Abstellen, keinen Sitzplatz. Deshalb fragen wir das einmal für den ganzen Haushalt.",

  noteHeading: "Willst du uns noch etwas sagen?",
  noteLabel: "Nachricht an uns",
  noteHelp:
    "Alles, wonach das Formular nicht fragt. Wir lesen jede Nachricht — eine Antwort hier im Formular gibt es aber nicht.",
  noteHint: "Zum Beispiel: „Wir kommen erst nach der Zeremonie“ oder „Oma braucht einen Platz nah am Ausgang“.",
  noteReadPromise: "Wir lesen das.",
  noteRemaining: (remaining: number) => `Noch ${remaining} Zeichen.`,

  /* F4 — the plus-one. Phrased as the question it answers rather than as an
     administrative option: for a guest invited alone it is the most consequential
     thing on this page. */
  addPlusOneTrigger: "Kommst du zu zweit? Begleitung hinzufügen",
  addPlusOneHeading: "Begleitung hinzufügen",
  addPlusOneBody: "Trag die Person ein, die mit dir kommt. Wozu sie kommt, fragen wir gleich darunter im Formular.",
  addPlusOneNameLabel: "Name der Begleitung",
  addPlusOneNameHelp:
    "Gemeint ist eine erwachsene Person, die mit dir kommt. Kinder und weitere Gäste tragen wir gern für euch ein — ruf uns dazu bitte kurz an.",
  addPlusOneSubmit: "Hinzufügen",
  addPlusOneSubmitting: "Wird hinzugefügt …",
  /** Says both things: the person is on the list, and the form still has to be
      answered and saved. The asymmetry is real, so the copy names it. */
  addPlusOneAdded: (name: string) =>
    `${name} steht jetzt auf eurer Liste. Sag uns unten noch, wozu ${name} kommt, und speichere das Formular.`,
  /** Shown in place of the trigger for every household that may not add — which is
      most of them. An offer, not a refusal: the answer is yes, we just want to hear
      the Personenzahl. */
  plusOneUnavailable: "Weitere Personen tragen wir gern für euch ein.",
  plusOneUnavailableCall: "Ruf uns dazu bitte kurz an:",

  /** "entfernen", never "löschen": it is a soft delete and the German should not
      promise otherwise. Same wording as the admin screen (F5-F02). */
  removeMember: "entfernen",
  removeMemberAccessibleName: (name: string) => `${name} entfernen`,
  removeMemberConfirmTitle: "Person entfernen?",
  removeMemberConfirmBody: (name: string) =>
    `${name} wird von eurer Liste genommen. Danach könnt ihr wieder jemanden hinzufügen.`,
  removeMemberConfirmAction: "Ja, entfernen",
  removeMemberFailedHeading: "Das Entfernen hat nicht geklappt",

  submit: "Speichern",
  submitting: "Wird gespeichert …",
  /** Named at the top of the form and linking to each card, so a household with one
      missing answer does not have to hunt for it. */
  missingAnswersHeading: "Bitte antworte noch für diese Personen:",
  missingAnswerLink: (name: string) => `${name} — Antwort fehlt`,
  saveFailedHeading: "Das Speichern hat nicht geklappt",
  reload: "Neu laden",

  summaryHeading: "Danke, wir haben es notiert",
  summarySavedAt: (date: string) => `Gespeichert am ${date}.`,
  summaryHouseholdHeading: "Euer Haushalt",
  summaryMembersHeading: "Personen",
  summaryMeal: "Essen",
  summaryPortion: "Portion",
  summarySnack: "Mitternachtssnack",
  summarySnackYes: "Ja",
  summarySnackNo: "Nein",
  summarySeatingNeed: "Platz",
  summaryDietaryNote: "Allergien",
  summaryTransportNeeded: "Plätze gesucht",
  summaryTransportOffered: "Plätze angeboten",
  summaryStroller: "Kinderwagen",
  summaryStrollerYes: "Ja",
  summaryNote: "Nachricht an uns",
  summaryNothing: "—",
  changeAnswer: "Antwort ändern",
  changeableUntil: (date: string) => `Ändern kannst du das noch bis zum ${date}.`,

  closedHeading: "Die Rückmeldefrist ist vorbei",
  closedBody: (date: string) =>
    `Am ${date} haben wir die Zahlen ans Catering weitergegeben. Das ist, was wir notiert haben — wenn sich etwas geändert hat, ruf uns bitte kurz an:`,
  closedNothingRecorded:
    "Von euch haben wir keine Rückmeldung. Die Frist ist vorbei, aber ruf uns bitte trotzdem an — wir tragen es dann ein:",
} as const;

/**
 * The admin's view of the same form. Its own block, because these sentences are
 * addressed to us and not to a household.
 */
export const adminRSVPLabels = {
  heading: (householdName: string) => `Rückmeldung für ${householdName}`,
  /** An admin with two tabs open must not write Familie Müller's answer into Familie
      Schmidt: the household's name is on the page, in the heading, not only in the URL. */
  intro: (householdName: string) => `Du beantwortest dieses Formular für ${householdName}.`,
  deadlinePassed: "Die Frist ist abgelaufen — du kannst hier trotzdem speichern.",
  back: "Zurück zum Haushalt",
  notFound: "Diesen Haushalt gibt es nicht.",
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
  /** The admin landing page until F6-F01 puts the dashboard there. Admin-facing, so
      it may say what it is. */
  dashboardPlaceholder: "Das Dashboard kommt mit F6. Bis dahin geht es über Haushalte.",
} as const;

/**
 * The admin guest list: the household table and the detail page.
 *
 * Admin-facing, so the register is denser than the guest copy — but still "du" and
 * still German, because there is exactly one admin and no reason to switch voice.
 */
export const householdLabels = {
  heading: "Haushalte",
  /** The table needs an accessible name; a caption is the one that also reads on
      screen, so it doubles as the summary line under the table. */
  tableCaption: "Alle Haushalte mit Code, Personenzahl und Anmeldestatus",
  columnHousehold: "Haushalt",
  columnCode: "Code",
  columnMembers: "Personen",
  columnLastLogin: "Letzte Anmeldung",
  columnRSVP: "RSVP",
  searchLabel: "Haushalt suchen",
  onlyNeverLoggedIn: "Nur die, die sich nie angemeldet haben",
  /** Not a stat tile: two numbers do not need a component, and the tiles are F6's. */
  summary: (households: number, neverLoggedIn: number) =>
    `${households} Haushalte, davon ${neverLoggedIn} nie angemeldet`,
  /** Never a blank cell: a blank reads as "not loaded", and this column is the answer
      to "haben sie es überhaupt gesehen?". */
  neverLoggedIn: "Nie angemeldet",
  rsvpAnswered: "Beantwortet",
  rsvpOpen: "Offen",
  empty: "Noch keine Haushalte angelegt.",
  noMatches: "Kein Haushalt passt zur Suche.",
  loading: "Haushalte werden geladen …",

  createHeading: "Haushalt anlegen",
  createNameLabel: "Name des Haushalts",
  createNameHint: "Zum Beispiel „Familie Müller“ oder „Anna und Bernd“.",
  createSubmit: "Anlegen",
  createSubmitting: "Wird angelegt …",

  detailBack: "Zurück zur Übersicht",
  detailDataHeading: "Haushalt",
  displayNameLabel: "Name",
  adminNoteLabel: "Interne Notiz (nur für uns)",
  adminNoteHint: "Der Haushalt sieht diese Notiz nie.",
  transportNeededLabel: "Plätze gesucht (Kirche → Feier)",
  transportOfferedLabel: "Plätze angeboten (Kirche → Feier)",
  strollerLabel: "Bringt einen Kinderwagen",
  save: "Speichern",
  saving: "Wird gespeichert …",
  saved: "Gespeichert.",

  codeHeading: "Login-Code",
  codeCopy: "Code kopieren",
  codeCopied: "Code kopiert.",
  codeReissue: "Neuen Code erzeugen",
  codeReissueConfirmTitle: "Neuen Code erzeugen?",
  codeReissueConfirmBody:
    "Der alte Code funktioniert danach nicht mehr. Eine bereits gedruckte Karte mit diesem Code ist damit ungültig, und angemeldete Geräte werden abgemeldet.",
  codeReissueConfirm: "Ja, neuen Code erzeugen",
  codeReissued: (revokedSessions: number) =>
    revokedSessions === 0
      ? "Neuer Code erzeugt. Der alte Code funktioniert nicht mehr."
      : `Neuer Code erzeugt. Der alte Code funktioniert nicht mehr, ${revokedSessions === 1 ? "ein Gerät wurde" : `${revokedSessions} Geräte wurden`} abgemeldet.`,

  membersHeading: "Personen",
  membersEmpty: "Noch niemand eingetragen.",
  addMemberHeading: "Person hinzufügen",
  /** One field, not two. What every list, place card and caterer sheet needs is the
      whole name, and asking for halves would make a guest decide which part of
      "Oma Erika" is the surname. */
  nameLabel: "Name",
  nameHint: "Vor- und Nachname, so wie die Person genannt werden möchte.",
  kindLabel: "Erwachsen oder Kind",
  ageLabel: "Alter am Hochzeitstag",
  ageHint: `Nur bei Kindern, und gemeint ist das Alter am ${weddingDateShort}.`,
  seatingNeedLabel: "Platz",
  dietaryNoteLabel: "Allergien und Unverträglichkeiten",
  addMember: "Hinzufügen",
  addingMember: "Wird hinzugefügt …",
  /** "entfernen", never "löschen": it is a soft delete, the person stays in the
      record, and the German should not promise otherwise. */
  removeMember: "Entfernen",
  removeMemberConfirmTitle: (name: string) => `${name} entfernen?`,
  removeMemberConfirmBody:
    "Die Person wird aus der Liste entfernt. Im Datenbestand bleibt sie erhalten, damit frühere Zählungen nachvollziehbar bleiben.",
  removeMemberConfirm: "Ja, entfernen",

  rsvpHeading: "Rückmeldung",
  rsvpNotAnswered: "Dieser Haushalt hat noch nicht geantwortet.",
  rsvpAnsweredAt: (date: string) => `Beantwortet am ${date}.`,
  /** A link, not a second form: the same component the guests use, addressed by id. */
  rsvpOpenForm: "Rückmeldung bearbeiten",

  exportHeading: "Downloads",
  /** Named for what they are for, not for their filename: nobody downloading these
      thinks in filenames. */
  exportCodes: "Codes für die Druckerei",
  /** Shown next to the link, so a truncated or empty file is obvious before it
      reaches the printer rather than after. */
  exportCodesCount: (households: number) => `${households} Haushalte, je eine Zeile`,
  /** The app cannot enforce E-OPS-07, and a warning nobody sees is no warning. */
  exportCodesWarning:
    "Diese Datei enthält alle Login-Codes. Bitte lösche sie nach dem Druck wieder — hier auf dem Rechner und bei der Druckerei.",
  exportGuests: "Gästeliste (Rohdaten)",
  exportGuestsWarning:
    "Vollständiger Auszug aus der Datenbank, entfernte Personen inklusive. Also keine Zählgrundlage: zum Zählen zuerst die entfernten Zeilen herausfiltern.",

  deleteHeading: "Haushalt löschen",
  delete: "Haushalt löschen",
  deleteConfirmTitle: "Haushalt löschen?",
  deleteConfirmBody: (name: string, members: number) =>
    `„${name}“ wird mit ${members === 1 ? "einer Person" : `${members} Personen`} gelöscht, dazu ihre Rückmeldungen und Sitzplätze. Der Eintrag im Änderungsprotokoll bleibt erhalten.`,
  deleteConfirm: "Ja, endgültig löschen",
  cancel: "Abbrechen",
} as const;

/* ------------------------------------------------------------------------- *
 * F2 — the informational pages
 *
 * Content is hardcoded in the frontend by decision (02-features), and every German
 * string still lives here rather than in the components — including the schedule
 * entries and the FAQ, which are content *and* copy at the same time. The rule has no
 * exception clause, and one file is what makes the whole site proof-readable in one
 * sitting.
 *
 * Several blocks below are **placeholders**, marked as such and tracked in TODO.md:
 * the venues, the schedule, the dress code, the gift wording and the account details
 * are not decided yet. They are written as honest sentences ("steht noch nicht fest")
 * rather than as lorem ipsum, because the one thing worse than a page that says the
 * date is open is a page that says "Lorem ipsum" to eighty guests.
 * ------------------------------------------------------------------------- */

/**
 * The navigation. One definition, rendered as a bottom bar on a phone and a top nav
 * on a desktop — never two lists that drift apart.
 */
export const navLabels = {
  guestNav: "Seiten",
  start: "Start",
  schedule: "Ablauf",
  location: "Location",
  /** The RSVP entry, in its two states. Primary until the household has answered,
      because it is the one thing we actually need them to do. */
  rsvp: "Antwort",
  rsvpAnswered: "Antwort ändern",
  more: "Mehr",
  /** The overflow page's own entries. */
  dresscode: "Dresscode",
  gifts: "Geschenke",
  faq: "Häufige Fragen",
  contact: "Kontakt",
  privacy: "Datenschutz",
  seating: "Sitzplan",
  gallery: "Galerie",
} as const;

export const moreLabels = {
  heading: "Mehr",
  intro: "Alles Weitere zur Hochzeit — und wie du uns erreichst.",
} as const;

export const startLabels = {
  /** Built around the display name rather than glued in front of it: "Luki & Paddi"
      is a valid display name, and "Liebe Luki & Paddi," would read wrong. */
  greeting: (householdName: string) => `Schön, dass ihr da seid, ${householdName}.`,
  intro: "Hier findet ihr alles zur Hochzeit — und hier sagt ihr uns, wer von euch kommt.",
  names: "Isabella & Andreas",
  /** Days only, never a ticking clock: it is a badge, not a timer. */
  countdown: (days: number) => (days === 1 ? "Noch 1 Tag" : `Noch ${days} Tage`),
  countdownToday: "Heute ist es so weit",
  rsvpCallToAction: "Jetzt zusagen",
  rsvpCallToActionAnswered: "Antwort ändern",
  /** Beside the answer link, spelled out — the deadline is a sentence, not a second
      countdown competing with the big number above it. */
  rsvpDeadline: (date: string) => `Bitte antwortet bis zum ${date}.`,
  rsvpAnswered: "Ihr habt uns schon geantwortet. Danke!",
} as const;

/**
 * The schedule. Times are plain strings: there is one timezone and one fixed day, and
 * parsing a time only to format it back is work that can go wrong for no gain.
 *
 * PLACEHOLDER (TODO.md): the running order is not fixed. Entries without a time are
 * rendered without one — a time on this page will be believed, so a guessed one is
 * worse than none.
 */
export const scheduleLabels = {
  heading: "Ablauf",
  intro: "So ist der Tag geplant. Sobald die Zeiten feststehen, tragen wir sie hier ein.",
  /** Which part of the day an entry belongs to, so a guest can see at a glance what
      their own Zusage covers. The timeline is never filtered by it: somebody who nur
      zur Kirche kommt still wants to know that a party happens. */
  church: "Kirche",
  party: "Feier",
  timeOpen: "Uhrzeit steht noch nicht fest",
  entries: [
    {
      time: null,
      title: "Trauung",
      detail: "Die kirchliche Trauung — wo genau, steht auf der Seite Location.",
      part: "church",
    },
    { time: null, title: "Empfang", detail: "Sekt, Zeit zum Ankommen und für Fotos.", part: "party" },
    {
      time: null,
      title: "Abendessen",
      detail: "Gemeinsames Essen. Was ihr esst, fragen wir im Formular.",
      part: "party",
    },
    { time: null, title: "Feier", detail: "Musik, Tanz und später am Abend eine kleine Stärkung.", part: "party" },
  ],
} as const;

/**
 * The two venues and how to get to them.
 *
 * PLACEHOLDER (TODO.md): names, addresses, parking and travel notes are open until
 * the venues are booked. This page must not ship to guests with these strings in it —
 * 07-roadmap accepts a thin Ablauf page at launch and explicitly does not accept an
 * empty Location page.
 */
export const locationLabels = {
  heading: "Location & Anreise",
  intro: "Getraut wird in der Kirche, gefeiert wird danach. Beide Adressen stehen hier, sobald sie feststehen.",
  venuesHeading: "Die Orte",
  churchHeading: "Kirche",
  partyHeading: "Feier",
  addressOpen: "Die Adresse steht noch nicht fest. Sie kommt hier auf die Seite, sobald wir gebucht haben.",
  mapLink: "Auf der Karte ansehen",
  /** Said in the link text as well as by the icon: nobody in this audience should be
      surprised by a new tab. */
  externalHint: "(öffnet in einem neuen Tab)",
  arrivalHeading: "Anreise & Parken",
  arrivalOpen:
    "Wie ihr am besten hinkommt und wo ihr parken könnt, schreiben wir hier auf, sobald die Orte feststehen.",
  transferHeading: "Von der Kirche zur Feier",
  transfer:
    "Zwischen Kirche und Feier liegt eine kurze Fahrt. Ob wir einen Shuttle organisieren, hängt davon ab, wie viele Plätze gebraucht werden — sag uns im Antwortformular, ob ihr Plätze braucht oder welche anbieten könnt.",
  accommodationHeading: "Übernachtung",
  accommodationOpen:
    "Eine Liste mit Hotels und Pensionen in der Nähe kommt hier auf die Seite. Zimmer reserviert haben wir nicht.",
} as const;

/**
 * PLACEHOLDER (TODO.md): the dress code wording is unwritten. What is decided is the
 * shape — plain words and an example, never a category name: "festlich" means five
 * different things to five relatives.
 */
export const dresscodeLabels = {
  heading: "Dresscode",
  lead: "Zieht an, worin ihr euch wohlfühlt — festlich, aber nicht steif.",
  body: "Wir schreiben hier noch genauer auf, was wir selbst anziehen, damit ihr eine Vorstellung habt. Wenn der Boden oder das Wetter eine Rolle für die Schuhwahl spielen, sagen wir das an dieser Stelle dazu.",
} as const;

/**
 * PLACEHOLDER (TODO.md): the gift wording and the account details.
 *
 * The IBAN is published deliberately (F2-F05): a content page is compiled into the
 * bundle and the bundle is served to anyone who loads the site, so it is semi-public.
 * That was decided with the trade in view — an IBAN lets somebody send money, not take
 * it — and the placeholder below keeps the shape without publishing a real account.
 */
export const giftLabels = {
  heading: "Geschenke",
  lead: "Das größte Geschenk ist, dass ihr da seid.",
  body: "Wenn ihr uns trotzdem etwas mitbringen möchtet, freuen wir uns über einen Beitrag zu unserer Hochzeitsreise. Die genauen Worte dazu schreiben wir noch — und niemand muss etwas mitbringen.",
  accountHeading: "Bankverbindung",
  accountHolderLabel: "Kontoinhaber",
  accountHolder: "Noch nicht eingetragen",
  ibanLabel: "IBAN",
  /** Stored unspaced, which is what the copy button puts on the clipboard; the
      grouping in fours is display only. */
  iban: "AT000000000000000000",
  ibanPending: "Die Bankverbindung tragen wir hier ein, sobald sie feststeht.",
  copyIban: "IBAN kopieren",
  ibanCopied: "IBAN kopiert",
} as const;

/**
 * The FAQ. Answers are always expanded — no accordion: a collapsed answer hides the
 * text from find-in-page and costs a tap per question.
 *
 * Seeded from what we expect to be asked; the real list follows the first phone calls
 * after send-out (TODO.md). An answer that a content page already gives links there
 * rather than restating it, so there is one copy to keep in step.
 */
export const faqLabels = {
  heading: "Häufige Fragen",
  intro: "Was uns schon gefragt wurde. Ist deine Frage nicht dabei, ruf uns einfach an.",
  entries: [
    {
      question: "Sind Kinder eingeladen?",
      answer:
        "Ja. Sag uns im Antwortformular Bescheid, wer mitkommt und wie alt die Kinder am Hochzeitstag sind — daran hängen Essen und Sitzplatz. Wenn wir ein Kind noch nicht auf eurer Liste haben, ruf uns bitte kurz an, dann tragen wir es ein.",
      link: null,
    },
    {
      question: "Darf ich jemanden mitbringen?",
      answer:
        "Wer allein eingeladen ist, kann im Antwortformular eine Begleitung eintragen. Für alle weiteren Personen ruft uns bitte an — wir tragen sie gern ein, wollen die Anzahl aber vorher wissen.",
      link: null,
    },
    {
      question: "Was ziehe ich an?",
      answer: "Festlich, aber ohne Zwang. Die Einzelheiten stehen auf der Seite Dresscode.",
      link: { to: "/dresscode", label: "Zum Dresscode" },
    },
    {
      question: "Wo ist die Hochzeit, und wo kann ich parken?",
      answer: "Beide Adressen, Parken und Anreise stehen auf der Seite Location & Anreise.",
      link: { to: "/location", label: "Zu Location & Anreise" },
    },
    {
      question: "Wie komme ich von der Kirche zur Feier?",
      answer:
        "Es ist eine kurze Fahrt. Im Antwortformular fragen wir, wie viele Plätze ihr braucht und wie viele ihr anbieten könnt — daran sehen wir, ob sich ein Shuttle lohnt. Wer mit wem fährt, klären wir nicht über die Seite.",
      link: null,
    },
    {
      question: "Was, wenn wir später kommen oder früher gehen müssen?",
      answer:
        "Kein Problem. Sag uns im Antwortformular, ob ihr zur Kirche, zur Feier oder zu beidem kommt, und schreib uns den Rest in die Nachricht am Ende des Formulars.",
      link: null,
    },
    {
      question: "Was schenken wir euch?",
      answer:
        "Am liebsten gar nichts außer eurer Zeit. Wenn ihr doch etwas möchtet, steht das auf der Seite Geschenke.",
      link: { to: "/geschenke", label: "Zu den Geschenken" },
    },
    {
      question: "Mein Code funktioniert nicht — was jetzt?",
      answer:
        "Groß- und Kleinschreibung ist egal, Leerzeichen und Bindestriche auch. Klappt es trotzdem nicht, ruf uns an — wir sagen dir den Code oder tragen deine Antwort direkt für dich ein.",
      link: { to: "/kontakt", label: "Zum Kontakt" },
    },
    {
      question: "Kann ich meine Antwort noch ändern?",
      answer:
        "Ja, bis zum Stichtag jederzeit. Danach geht es nicht mehr über die Seite — dann ruf uns bitte an, wir ändern es für dich.",
      link: null,
    },
  ],
} as const;

export const contactLabels = {
  heading: "Kontakt",
  intro: "Ruf einfach an — wir sind beide erreichbar.",
  codeHelpHeading: "Wenn der Code nicht funktioniert",
  codeHelp:
    "Der Code steht auf eurer Einladungskarte, sechs Zeichen, Groß- und Kleinschreibung egal. Klappt es nicht, ruf uns an: wir sagen dir den Code am Telefon oder tragen eure Antwort gleich selbst ein.",
} as const;

/**
 * Datenschutz, in guest German.
 *
 * A translation of 06-privacy-security into "du", not a new document: when the two
 * disagree, the spec changes first. Deliberately not boilerplate — a wall of legalese
 * would be worse than nothing here, because nobody reads it and it makes a wedding
 * invitation feel like a business.
 *
 * Wants a proof-read before send-out (TODO.md). The photo-retention sentence is
 * missing on purpose: the gallery lifetime is undecided, and a number we have not
 * chosen must not be promised.
 */
export const privacyLabels = {
  heading: "Datenschutz",
  lead: "Kurz und ohne Juristendeutsch: Das hier läuft auf unserem eigenen Server, und außer uns bekommt niemand eure Daten zu sehen.",
  storedHeading: "Was wir speichern",
  stored:
    "Euren Haushalt mit dem Namen, unter dem wir euch eingeladen haben, die Namen der Personen, euren Code von der Einladungskarte, und alles, was ihr im Antwortformular ausfüllt: wer kommt, wozu, was ihr esst, ob ihr Plätze im Auto braucht oder anbietet, und eure Nachricht an uns.",
  sensitiveHeading: "Allergien und das Alter der Kinder",
  sensitive:
    "Beides fragen wir, weil das Catering es braucht: Allergien landen auf der Liste für die Küche, das Alter am Hochzeitstag entscheidet über Portion und Preis. Ihr müsst nichts ausfüllen, was ihr nicht wollt — dann ruft uns lieber an.",
  whyHeading: "Warum wir das speichern",
  why: "Damit wir die Hochzeit planen können: Personenzahl, Essen, Sitzplätze, Fahrten. Für nichts sonst. Wir schicken keine Werbung, es gibt keine Auswertung und keine Weitergabe an Dritte außer den Zahlen, die das Catering braucht.",
  whoHeading: "Wer das sieht",
  who: "Wir beide. Die Seite läuft auf unserem eigenen Server, ohne Google, ohne Tracking, ohne externe Schriften und ohne Cookies außer dem einen, mit dem ihr angemeldet bleibt. Die Seite steht in keiner Suchmaschine.",
  retentionHeading: "Wie lange",
  retention:
    "Die Seite geht ungefähr drei Monate nach der Hochzeit offline, und die Daten gehen mit ihr. Was wir behalten, ist eine private Kopie für uns selbst — so, wie man ein Gästebuch behält.",
  rightsHeading: "Ändern oder löschen",
  rights:
    "Ruf uns an oder schreib uns, dann ändern oder löschen wir es. Kein Formular, kein Verfahren — bei achtzig Gästen ist ein Anruf schneller als alles andere.",
  contactLink: "Zum Kontakt",
} as const;
