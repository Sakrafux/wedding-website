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
  /** The authenticated landing page until F2-F02 builds the real start page. */
  startHeading: "Ihr seid angemeldet",
  startIntro: "Die Einladung mit allen Infos kommt hier in Kürze.",
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
  ageLabel: (name: string) => `Alter von ${name} am Hochzeitstag, 17. Juli 2027`,
  ageHelp: "Gemeint ist das Alter am 17. Juli 2027, nicht heute. Das Catering rechnet nach Alter am Tag der Feier.",

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
  ageHint: "Nur bei Kindern, und gemeint ist das Alter am 17.07.2027.",
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
