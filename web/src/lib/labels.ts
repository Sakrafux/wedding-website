/**
 * Every German string in the application.
 *
 * Two kinds live here. First, the mapping from English enum values to German UI
 * text – the reason the file exists. Second, the screen copy: headings, buttons,
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
 * object literal – that is the whole value of the file.
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
// Type-only, and the one import here that points at a component: the transport
// direction is form state rather than an API enum, so state.ts owns it and this file
// only translates it.
import type { TransportDirection } from "@/components/rsvp/state";

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
 * not part of the Attending union – it is the absence of a value – so its labels
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
 * Help text shown under the portion options. Only `none` needs explaining – the
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

/**
 * The transport question's three answers. Not a database enum – the direction is
 * derived from the two seat counts (`components/rsvp/state.ts`) – but it is a fixed
 * set of options rendered from a `Record`, so it belongs with the other maps.
 */
export const transportDirectionLabels: Record<TransportDirection, string> = {
  none: "Wir brauchen nichts",
  needed: "Wir bräuchten Plätze",
  offered: "Wir können Plätze anbieten",
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
  /** Revealed after two failures – the point at which a person starts to feel
      stupid, which is exactly when the way out should appear. */
  fallback: "Klappt es nicht? Ruf uns an:",
  loggedOut: "Du wurdest abgemeldet. Bitte melde dich noch einmal an.",
  rejectedHousehold: "Kein Problem – bitte prüf den Code noch einmal.",
} as const;

/**
 * The confirmation screen, in both numbers – like `/start`, it greets a
 * household in the number we seeded it in (`F2-F08`).
 *
 * "Nein" is number-neutral, so it lives beside the variants rather than twice
 * inside them.
 */
export const confirmationAddressLabels: Record<
  Address,
  {
    /** Names the household, because the whole screen exists to catch a valid code
        that belongs to somebody else. */
    heading: (addressee: string) => string;
    membersIntro: string;
    confirm: string;
  }
> = {
  singular: {
    heading: (addressee) => `Willkommen, ${addressee} – bist du das?`,
    membersIntro: "So haben wir dich notiert:",
    confirm: "Ja, das bin ich",
  },
  plural: {
    heading: (addressee) => `Willkommen, ${addressee} – seid ihr das?`,
    membersIntro: "So haben wir euch notiert:",
    confirm: "Ja, das sind wir",
  },
};

export const confirmationLabels = {
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
 * The RSVP form – the screen this whole product exists for.
 *
 * Every field carries a `*Help` sentence, per the form rules in 05-design: the help
 * lives behind a `?` popover beside the label, so a guest who needs the explanation
 * gets it and the form does not become a wall of grey text for everybody else.
 *
 * The register is the same "du" as everywhere else, addressed to a household – so the
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
    "Das setzt die Antwort für alle Personen unten auf einmal. Danach könnt ihr einzelne Personen noch ändern.",
  /** Shown instead of a lit-up option once the cards disagree: a selector that stayed
      lit while the cards said something else would be a lie about what gets saved. */
  householdScopeMixed: "Ihr habt unterschiedliche Antworten – siehe unten.",
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
  /** The name is editable here, not only displayed: we typed the guest list off an
      address book, and the people on it know best how they are called (`F4-F04`). */
  memberNameLabel: "Name",
  memberNameHelp: "So steht der Name in der Liste. Bitte den vollständigen Namen angeben.",
  memberScopeLabel: (name: string) => `Wozu kommt ${name}?`,
  memberScopeHelp:
    "Kirche und Feier sind getrennt – manche kommen nur zur Trauung oder nur zum Fest. Wählt aus, was für diese Person passt.",
  mealChoiceLabel: (name: string) => `Was isst ${name}?`,
  mealChoiceHelp: "Wir geben dem Catering nur die Zahlen weiter, keine Namen. „Isst alles“ heißt: kein Sonderwunsch.",
  portionLabel: (name: string) => `Portion für ${name}`,
  portionHelp:
    "Kinderportionen sind kleiner, „Kein Essen“ ist für Babys oder wenn jemand nicht mitisst. Beides kostet uns nichts extra – sagt ruhig, was wirklich passt.",
  midnightSnackLabel: (name: string) => `Mitternachtssnack für ${name}`,
  midnightSnackHelp: "Spät am Abend gibt es noch eine kleine Stärkung. Wir müssen vorher wissen, für wie viele.",
  midnightSnackHint: "Kleine Stärkung spät am Abend",
  seatingNeedLabel: (name: string) => `Platz für ${name}`,
  seatingNeedHelp:
    "Damit planen wir die Plätze in der Kirche und am Tisch: Rollstuhl, Hochstuhl, oder ein Kind, das bei den Eltern sitzt. Der Hochstuhl gilt für die Feier – in der Kirche sitzt euer Kind normal in der Bank.",
  dietaryNoteLabel: (name: string) => `Allergien oder Unverträglichkeiten von ${name}`,
  dietaryNoteHelp:
    "Was die Küche wissen muss, zum Beispiel „Nussallergie“ oder „laktosefrei“. Kurz reicht – wir fragen nach, wenn etwas unklar ist.",
  dietaryNotePlaceholderHint: "Zum Beispiel: Nussallergie, laktosefrei",
  /** The date is in the label, not only in the help: a bare "Alter" gets answered as
      of today, and then the value drifts over the months before the wedding. */
  ageLabel: (name: string) => `Alter von ${name} am Hochzeitstag, ${weddingDateLong}`,
  ageHelp: `Gemeint ist das Alter am ${weddingDateLong}, nicht heute. Das Catering rechnet nach Alter am Tag der Feier.`,

  transportHeading: "Fahrt von der Kirche zur Feier",
  /** One question with a direction, not two counts: needing und anbieten schließen
      sich aus, und nebeneinander lasen sie sich wie dieselbe Frage zweimal (F3-F07). */
  transportDirectionLabel: "Wie sieht es bei euch mit der Fahrt aus?",
  transportDirectionHelp:
    "Gemeint ist die kurze Fahrt von der Kirche zur Feier. Wir sehen daran, ob wir mit einer Fahrt auskommen.",
  transportNeededLabel: "Wie viele Plätze braucht ihr?",
  transportOfferedLabel: "Wie viele Plätze könnt ihr anbieten?",
  transportSeatsHelp:
    "Nur eine Zahl für die Planung. Wenn sich das noch ändert, könnt ihr die Antwort jederzeit anpassen.",
  /** Said rather than silently discarded: a number that vanishes without explanation
      is exactly the kind of thing that gets phoned in about. */
  transportDropped:
    "Die Angaben zur Fahrt gelten nicht mehr, weil niemand von euch zu Kirche und Feier kommt. Wir speichern sie nicht.",
  taxiHeading: "Taxifahrt von der Feier nach St. Pölten",
  /** No direction question, unlike the shuttle above: a taxi is something we would
      order, so there is no "we can offer" side and the count starts at zero. */
  taxiSeatsLabel: "Wie viele Taxi-Plätze braucht ihr?",
  taxiSeatsHelp:
    "An dieser Zahl entscheiden wir, ob wir ein (oder mehrere) Taxis für die Fahrt von der Feier nach St. Pölten reservieren.",
  /** Same rule as the shuttle's, one trip later: the count is the household's answer,
      and the paragraph says it went rather than letting the number vanish. */
  taxiDropped:
    "Die Angaben zur Taxifahrt gelten nicht mehr, weil niemand von euch zur Feier kommt. Wir speichern sie nicht.",
  strollerLabel: "Wir bringen einen Kinderwagen mit",
  strollerHelp:
    "Ein Kinderwagen braucht Platz zum Abstellen, keinen Sitzplatz. Deshalb fragen wir das einmal für den ganzen Haushalt.",

  noteHeading: "Wollt ihr uns noch etwas sagen?",
  noteLabel: "Nachricht an uns",
  noteHelp:
    "Alles, wonach das Formular nicht fragt. Wir lesen jede Nachricht – eine Antwort hier im Formular gibt es aber nicht.",
  noteHint: "Zum Beispiel: „Wir kommen erst um 18:00“",
  noteReadPromise: "",
  noteRemaining: (remaining: number) => `Noch ${remaining} Zeichen.`,

  /* F4 – the plus-one. Phrased as the question it answers rather than as an
     administrative option: for a guest invited alone it is the most consequential
     thing on this page. Singular throughout, except the unavailable notice at the
     end: only a household we seeded alone ever sees the rest, so "du" is the
     number we invited them in (F2-F08) – every other household reads the notice
     in the plural. */
  addPlusOneTrigger: "Kommst du zu zweit? Begleitung hinzufügen",
  addPlusOneHeading: "Begleitung hinzufügen",
  addPlusOneBody: "Trag die Person ein, die mit dir kommt. Wozu sie kommt, fragen wir gleich darunter im Formular.",
  addPlusOneNameLabel: "Name der Begleitung",
  addPlusOneNameHelp:
    "Gemeint ist eine erwachsene Person, die mit dir kommt. Kinder und weitere Gäste tragen wir gern für dich ein – ruf uns dazu bitte kurz an.",
  addPlusOneSubmit: "Hinzufügen",
  addPlusOneSubmitting: "Wird hinzugefügt …",
  /** Says both things: the person is on the list, and the form still has to be
      answered and saved. The asymmetry is real, so the copy names it. */
  addPlusOneAdded: (name: string) =>
    `${name} steht jetzt auf deiner Liste. Sag uns unten noch, wozu ${name} kommt, und speichere das Formular.`,
  /** Shown in place of the trigger for every household that may not add – which is
      most of them. An offer, not a refusal: the answer is yes, we just want to hear
      the Personenzahl. */
  plusOneUnavailable: "Weitere Personen tragen wir gern für euch ein.",
  plusOneUnavailableCall: "Ruft uns dazu bitte kurz an:",

  /** "entfernen", never "löschen": it is a soft delete and the German should not
      promise otherwise. Same wording as the admin screen (F5-F02). */
  removeMember: "entfernen",
  removeMemberAccessibleName: (name: string) => `${name} entfernen`,
  removeMemberConfirmTitle: "Person entfernen?",
  removeMemberConfirmBody: (name: string) =>
    `${name} wird von deiner Liste genommen. Danach kannst du wieder jemanden hinzufügen.`,
  removeMemberConfirmAction: "Ja, entfernen",
  removeMemberFailedHeading: "Das Entfernen hat nicht geklappt",

  submit: "Speichern",
  submitting: "Wird gespeichert …",
  /** Named at the top of the form and linking to each card, so a household with one
      missing answer does not have to hunt for it. */
  missingAnswersHeading: "Bitte antwortet noch für diese Personen:",
  missingAnswerLink: (name: string) => `${name} – Antwort fehlt`,
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
  summaryTaxiNeeded: "Taxi-Plätze gesucht",
  summaryStroller: "Kinderwagen",
  summaryStrollerYes: "Ja",
  summaryNote: "Nachricht an uns",
  summaryNothing: "–",
  changeAnswer: "Antwort ändern",
  changeableUntil: (date: string) => `Ändern könnt ihr das noch bis zum ${date}.`,

  closedHeading: "Die Rückmeldefrist ist vorbei",
  closedBody: (date: string) =>
    `Am ${date} haben wir die Zahlen ans Catering weitergegeben. Das ist, was wir notiert haben – wenn sich etwas geändert hat, ruft uns bitte kurz an:`,
  closedNothingRecorded:
    "Von euch haben wir keine Rückmeldung. Die Frist ist vorbei, aber ruft uns bitte trotzdem an – wir tragen es dann ein:",
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
  deadlinePassed: "Die Frist ist abgelaufen – du kannst hier trotzdem speichern.",
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
  /** Nav entries whose pages do not exist yet render disabled – better than a 404,
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
 * Admin-facing, so the register is denser than the guest copy – but still "du" and
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
  onlyUnanswered: "Nur die ohne Antwort",
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
  createNameHint: "Nur für uns, zum Beispiel „Familie Müller“. Die Anschrift kommt danach.",
  createSubmit: "Anlegen",
  createSubmitting: "Wird angelegt …",

  detailBack: "Zurück zur Übersicht",
  /** Not "Haushalt": that is now the label of one of the fields inside this group. */
  detailDataHeading: "Haushaltsdaten",
  householdNameLabel: "Haushalt",
  householdNameHint: "Nur für uns: so steht der Haushalt in der Liste. Zum Beispiel „Familie Müller“.",
  addresseeLabel: "Anschrift",
  addresseeHint:
    "So sprechen wir sie an – das steht auf der Einladung und begrüßt sie auf der Seite. Zum Beispiel „Luki & Paddi“.",
  adminNoteLabel: "Interne Notiz (nur für uns)",
  adminNoteHint: "Der Haushalt sieht diese Notiz nie.",
  save: "Speichern",
  saving: "Wird gespeichert …",
  /** The time, not just the fact: a chip reading "Gespeichert." left us wondering
      whether it meant *these* changes, and locally the save is fast enough that the
      button's own state was a flash (F5-F05). */
  savedAt: (time: string) => `Zuletzt gespeichert um ${time}`,

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
  /** Only rendered when the household has rewritten the name we invited them under
      (`F4-F04`). "Eingeladen als", not "früher": it says which name is on the card we
      posted, which is the question this line exists to answer. */
  seededName: (name: string) => `Eingeladen als ${name}`,
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
  /** Shown here, editable only through the RSVP form: these four are the household's
      own answers, and their one writer is the endpoint that runs the RSVP rules
      (F5-B05). */
  rsvpTransportNeeded: "Plätze gesucht (Kirche → Feier)",
  rsvpTransportOffered: "Plätze angeboten (Kirche → Feier)",
  rsvpTaxiNeeded: "Taxi-Plätze gesucht (Feier → St. Pölten)",
  rsvpStroller: "Kinderwagen",
  rsvpYes: "Ja",
  rsvpNo: "Nein",
  rsvpNothing: "–",
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
    "Diese Datei enthält alle Login-Codes. Bitte lösche sie nach dem Druck wieder – hier auf dem Rechner und bei der Druckerei.",
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
 * F2 – the informational pages
 *
 * Content is hardcoded in the frontend by decision (02-features), and every German
 * string still lives here rather than in the components – including the schedule
 * entries, which are content *and* copy at the same time. The rule has no
 * exception clause, and one file is what makes the whole site proof-readable in one
 * sitting.
 *
 * Several blocks below are **placeholders**, marked as such and tracked in TODO.md:
 * the schedule is not decided yet. They are written as honest sentences ("steht noch
 * nicht fest") rather than as lorem ipsum, because the one thing worse than a page
 * that says the date is open is a page that says "Lorem ipsum" to eighty guests.
 * Dresscode, Geschenke and the bank details were placeholders until 2026-10-11, and
 * the venues and the hotel list landed 2026-10-14.
 * ------------------------------------------------------------------------- */

/**
 * The navigation. One definition, rendered as a bottom bar on a phone and a top nav
 * on a desktop – never two lists that drift apart.
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
  contact: "Kontakt",
  privacy: "Datenschutz",
  seating: "Sitzplan",
  gallery: "Galerie",
} as const;

export const moreLabels = {
  heading: "Mehr",
  intro: "Alles Weitere zur Hochzeit – und wie ihr uns erreicht.",
} as const;

/**
 * Whether a household is addressed as one person or as several.
 *
 * Chosen by how many people we **seeded**, never by how many there are now: a guest
 * invited alone who adds a companion is still the person we wrote to, and switching to
 * "ihr" the moment they name somebody reads as the site correcting them (`F2-F08`).
 */
export type Address = "singular" | "plural";

/**
 * The start page's copy, in both numbers.
 *
 * Two written-out variants rather than one sentence with a conjugation helper: German
 * does not conjugate by find-and-replace, and a `Record` makes a missing variant a
 * compile error.
 *
 * Only this page varies. The RSVP form stays plural throughout – it addresses a
 * household, and a companion may have been added to it.
 */
export const startAddressLabels: Record<
  Address,
  {
    greeting: (householdName: string) => string;
    intro: string;
    rsvpDeadline: (date: string) => string;
    rsvpAnsweredUntil: (date: string) => string;
  }
> = {
  singular: {
    /** Built around the display name rather than glued in front of it: "Luki & Paddi"
        is a valid display name, and "Liebe Luki & Paddi," would read wrong. */
    greeting: (householdName) => `Schön, dass du da bist, ${householdName}.`,
    intro: "Hier findest du alles zur Hochzeit – und hier sagst du uns, ob du kommst.",
    rsvpDeadline: (date) => `Bitte antworte bis zum ${date}.`,
    /** Once answered, the deadline stops being a request and becomes the date until
        which the answer can still change – the only thing left worth knowing. */
    rsvpAnsweredUntil: (date) => `Danke für deine Antwort! Ändern kannst du sie noch bis zum ${date}.`,
  },
  plural: {
    greeting: (householdName) => `Schön, dass ihr da seid, ${householdName}.`,
    intro: "Hier findet ihr alles zur Hochzeit – und hier sagt ihr uns, wer von euch kommt.",
    rsvpDeadline: (date) => `Bitte antwortet bis zum ${date}.`,
    rsvpAnsweredUntil: (date) => `Danke für eure Antwort! Ändern könnt ihr sie noch bis zum ${date}.`,
  },
};

export const startLabels = {
  names: "Isabella & Andreas",
  /** Days only, never a ticking clock: it is a badge, not a timer. */
  countdown: (days: number) => (days === 1 ? "Noch 1 Tag" : `Noch ${days} Tage`),
  countdownToday: "Heute ist es so weit",
  rsvpCallToAction: "Jetzt zusagen",
  rsvpCallToActionAnswered: "Antwort ändern",
} as const;

/**
 * The schedule. Times are plain strings: there is one timezone and one fixed day, and
 * parsing a time only to format it back is work that can go wrong for no gain.
 *
 * PLACEHOLDER (TODO.md): the running order is not fixed. Entries without a time are
 * rendered without one – a time on this page will be believed, so a guessed one is
 * worse than none.
 */
export const scheduleLabels = {
  heading: "Ablauf",
  intro: "",
  /** Which part of the day an entry belongs to, so a guest can see at a glance what
      their own Zusage covers. The timeline is never filtered by it: somebody who nur
      zur Kirche kommt still wants to know that a party happens. */
  church: "Kirche",
  party: "Feier",
  timeOpen: "Uhrzeit steht noch nicht fest",
  entries: [
    {
      time: "14:00 - 15:30",
      title: "Trauung",
      detail: "Die kirchliche Trauung – wo genau, steht auf der Seite Location.",
      part: "church",
    },
    {
      time: "16:00",
      title: "Empfang",
      detail:
        "Zeit zum Ankommen und für Fotos. Es gibt auch Kalte Platte als Vorspeise für den ersten Hunger. Wir machen zu diesem Zeitpunkt noch unsere Hochzeitsfotos.",
      part: "party",
    },
    {
      time: "17:00",
      title: "Unsere Ankunft",
      detail:
        "Wir würden erwarten zu dieser Zeit fertig mit unseren Hochzeitsfotos zu werden und bei der Feier Location anzukommen",
      part: "party",
    },
    {
      time: "17:30",
      title: "Abendessen",
      detail: "Gemeinsames Essen mit Suppe zum Start, gefolgt von Buffet. Was ihr esst, fragen wir im Formular.",
      part: "party",
    },
    {
      time: "18:00",
      title: "Candy Bar Eröffnung",
      detail:
        "Hier könnt ihr euch über den Abend verteilt Nachspeisen holen. Wir rechnen mit ~2 pro Person, also gerne auch mehrmals zugreifen, aber bitte auch den anderen Gästen Zeit geben, sich etwas zu holen.",
      part: "party",
    },
    {
      time: "18:30 - 22:00",
      title: "Feier",
      detail: "Musik, Tanz und Gespräche. Kein strukturiertes Programm, einfach Zeit die Feier zu genießen.",
      part: "party",
    },
    {
      time: "22:00",
      title: "Hochzeitstore",
      detail: "Wir schneiden die Hochzeitstorte an, welche anschließend komplett geschnitten und verteilt wird.",
      part: "party",
    },
    { time: "23:00", title: "Mitternachtssnack", detail: "Eine kleine Stärkung spät am Abend.", part: "party" },
    {
      time: "00:00 - 01:00",
      title: "Ausklang",
      detail: "Noch ein bisschen Zeit, die Feier ausklingen zu lassen.",
      part: "party",
    },
    {
      time: "02:00",
      title: "Sperrstunde",
      detail: "Spätestens zu diesem Zeitpunkt muss die Location zusperren.",
      part: "party",
    },
  ],
} as const;

/**
 * The two venues and how to get to them.
 *
 * PLACEHOLDER (TODO.md): names, addresses, parking and travel notes are open until
 * the venues are booked. These pages must not ship to guests with these strings in
 * them – 07-roadmap accepts a thin Ablauf page at launch and explicitly does not
 * accept an empty Location page. Splitting the page into an overview and one page per
 * venue (F2-F09) does not change that: it is now three pages waiting for the same
 * facts.
 */
export const locationLabels = {
  heading: "Location & Anreise",
  intro:
    "Getraut wird in der Kirche, gefeiert wird danach woanders. Hier findet ihr beide Orte – und auf der jeweiligen Seite alles, was ihr dazu braucht.",
  venuesHeading: "Die Orte",
  churchHeading: "Trauung in der Kirche",
  partyHeading: "Feier",
  /** One line per venue on the overview; the facts live on the venue's own page. Two
      addresses, two car parks and two arrival notes in one scroll were exactly what
      got mixed up (F2-F09). */
  churchTeaser: "Wo wir uns das Ja-Wort geben. Adresse, Parken und Anreise auf der Seite zur Trauung.",
  partyTeaser: "Wo danach gefeiert wird. Adresse, Parken, Anreise und Tipps zum Übernachten auf der Seite zur Feier.",
  churchDetailLink: "Zur Trauung",
  partyDetailLink: "Zur Feier",
  addressHeading: "Adresse",
  /** One line per entry, as every address block on these pages renders them. */
  churchAddress: ["Prandtauerkirche St. Pölten", "Prandtauerstraße 2", "3100 St. Pölten"],
  partyAddress: ["Restaurant Lukić", "Bahnhofstraße 3", "3150 Wilhelmsburg"],
  /** Short links, because they are typed by hand nowhere and outlive any embed. */
  churchMapUrl: "https://maps.app.goo.gl/KqCTTZ5ADfeC6fMv5",
  partyMapUrl: "https://maps.app.goo.gl/XTrsSH79hjR1dhJm7",
  /** The route between the two venues, on the overview where the transfer text lives. */
  transferMapUrl: "https://maps.app.goo.gl/WhqiJz9XFG92wEDy9",
  mapLink: "Auf der Karte ansehen",
  /** Said in the link text as well as by the icon: nobody in this audience should be
      surprised by a new tab. */
  externalHint: "(öffnet in einem neuen Tab)",
  arrivalHeading: "Anreise & Parken",
  /** One entry per rendered paragraph. */
  churchArrival: [
    "In der Nähe der Prandtauerkirche gibt es mehrere Parkgaragen. Die Tiefgarage Rathausplatz liegt direkt bei der Kirche, könnte am Samstag Nachmittag aber stark gefüllt sein.",
    "Falls ihr öffentlich kommt, liegt die Kirche nur wenige Gehminuten vom Bahnhof entfernt und ist somit sowohl an Busse als auch Züge gut angebunden.",
  ],
  partyArrival: [
    "Prinzipiell gibt es einen Parkplatz direkt angrenzend an den Saal. Ansonsten gibt es auch viele gratis Parkplätze entlang der Straßen im Ortsgebiet.",
    "Falls ihr öffentlich kommt, fährt die Traisentalbahn zum Bahnhof Wilhelmsburg (ca. 400 m entfernt).",
  ],
  departureHeading: "Abreise",
  /** On the party page, not the overview: it is the trip the evening ends with, and
      it is where the taxi question below points back at the RSVP. */
  partyDeparture: [
    "Für die Abreise ohne eigenes Auto bieten sich entweder Öffis oder Taxis an. Der genaue Zeitplan ist derzeit noch unklar aufgrund von Schienenersatzverkehr, aber zumindest um ca. 22:30 sollte eine späte Verbindung sein (potentiell auch später).",
    "Ansonsten würde es sich auch anbieten, gemeinsam mit dem Taxi zu fahren. Gebt deshalb bei eurer Antwort an, wie viele Plätze ihr benötigt, damit wir eine ungefähre Zahl haben.",
  ],
  /** On the party page, not the overview: somebody looking for a bed is looking for
      one near where the evening ends. */
  accommodationHeading: "Übernachtung",
  partyAccommodation: [
    "Die Übernachtungsmöglichkeiten in Wilhelmsburg sind begrenzt, aber potentiell möglich bei WALTNER Wilhelmsburg oder im Gasthof Franzl. Ein größeres Angebot an Hotels gibt es erst wieder in St. Pölten.",
    "Vermutlich ist es deshalb einfacher, das Auto in St. Pölten zu lassen und bei Bedarf dort zu übernachten.",
  ],
  /** Stays on the overview: it belongs to neither venue – it is the bit between them,
      and it is the paragraph a guest reads before answering the transport question. */
  transferHeading: "Von der Kirche zur Feier",
  transfer:
    "Zwischen Kirche und Feier liegt eine kurze Fahrt von 15–20 Minuten. Ob wir einen Shuttle organisieren, hängt davon ab, wie viele Plätze gebraucht werden – sagt uns im Antwortformular, ob ihr Plätze braucht oder welche anbieten könnt.",
  backToOverview: "Zurück zu Location & Anreise",
} as const;

/**
 * PLACEHOLDER (TODO.md): the dress code wording is unwritten. What is decided is the
 * shape – plain words and an example, never a category name: "festlich" means five
 * different things to five relatives.
 */
export const dresscodeLabels = {
  heading: "Dresscode",
  lead: "Festlich, aber nicht steif.",
  body: "Wir freuen uns über elegante Outfits – von Sommerkleid und Anzug bis hin zu Dirndl und Lederhose ist alles willkommen, was zum Feiern einlädt. Hauptsache festlich und bequem genug für eine lange Nacht. Weiße Kleider sind natürlich der Braut vorbehalten.",
} as const;

/**
 * PLACEHOLDER (TODO.md): the gift wording and the account details.
 *
 * The IBAN is published deliberately (F2-F05): a content page is compiled into the
 * bundle and the bundle is served to anyone who loads the site, so it is semi-public.
 * That was decided with the trade in view – an IBAN lets somebody send money, not take
 * it – and the placeholder below keeps the shape without publishing a real account.
 */
export const giftLabels = {
  heading: "Geschenke",
  lead: "Das größte Geschenk ist, dass ihr da seid.",
  body: "Trotzdem würden wir uns über etwas freuen. Zwischen der Hochzeit selbst, den anschließenden Flitterwochen oder den Einzug in unsere erste gemeinsame Wohnung, haben wir einige Kosten zu stemmen. Wir schätzen jegliche Beteiligung, idealerweise monetär. Ihr könnt trotzdem gerne auch andere, kleine Geschenke mitbringen, um einen persönlichen Touch beizubehalten.",
  body2:
    "Grundsätzlich wollen wir uns nicht den großen Stress machen, eine umfangreiche Geschenkeliste aufzustellen, die damit einhergehende Logistik bei und nach der Hochzeit, sowie Sorgen um die Fairness allen Gästen gegenüber. Gebt einfach, was ihr für angebracht haltet – wir können es auf jeden Fall gut gebrauchen :)",
  accountHeading: "Bankverbindung",
  accountText: "Falls ihr nicht mit Bargeld hantieren möchtet, könnt ihr auch direkt überweisen.",
  accountHolderLabel: "Kontoinhaber",
  accountHolder: "Andreas Hell",
  ibanLabel: "IBAN",
  /** Stored unspaced, which is what the copy button puts on the clipboard; the
      grouping in fours is display only. */
  iban: "AT135300006554033820",
  ibanPending: "Die Bankverbindung tragen wir hier ein, sobald sie feststeht.",
  copyIban: "IBAN kopieren",
  ibanCopied: "IBAN kopiert",
} as const;

export const contactLabels = {
  heading: "Kontakt",
  intro: "Ruft einfach an – wir sind beide erreichbar.",
  codeHelpHeading: "Wenn der Code nicht funktioniert",
  codeHelp:
    "Der Code steht auf eurer Einladungskarte, sechs Zeichen, Groß- und Kleinschreibung egal. Klappt es nicht, ruft uns an: wir sagen euch den Code am Telefon oder tragen eure Antwort gleich selbst ein.",
} as const;

/**
 * Datenschutz, in guest German.
 *
 * A translation of 06-privacy-security into "du", not a new document: when the two
 * disagree, the spec changes first. Deliberately not boilerplate – a wall of legalese
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
    "Euren Haushalt mit dem Namen, unter dem wir euch eingeladen haben, die Namen der Personen, euren Code von der Einladungskarte, und alles, was ihr im Antwortformular ausfüllt: wer kommt, wozu, was ihr esst, ob ihr Plätze im Auto braucht oder anbietet, ob ihr für die Heimfahrt Taxi-Plätze braucht, und eure Nachricht an uns.",
  sensitiveHeading: "Allergien und das Alter der Kinder",
  sensitive:
    "Beides fragen wir, weil das Catering es braucht: Allergien landen auf der Liste für die Küche, das Alter am Hochzeitstag entscheidet über Portion und Preis. Ihr müsst nichts ausfüllen, was ihr nicht wollt – dann ruft uns lieber an.",
  whyHeading: "Warum wir das speichern",
  why: "Damit wir die Hochzeit planen können: Personenzahl, Essen, Sitzplätze, Fahrten. Für nichts sonst. Wir schicken keine Werbung, es gibt keine Auswertung und keine Weitergabe an Dritte außer den Zahlen, die das Catering braucht.",
  whoHeading: "Wer das sieht",
  who: "Wir beide. Die Seite läuft auf unserem eigenen Server, ohne Google, ohne Tracking, ohne externe Schriften und ohne Cookies außer dem einen, mit dem ihr angemeldet bleibt. Die Seite steht in keiner Suchmaschine.",
  retentionHeading: "Wie lange",
  retention:
    "Die Seite geht ungefähr drei Monate nach der Hochzeit offline, und die Daten gehen mit ihr. Was wir behalten, ist eine private Kopie für uns selbst – so, wie man ein Gästebuch behält.",
  rightsHeading: "Ändern oder löschen",
  rights:
    "Ruft uns an oder schreibt uns, dann ändern oder löschen wir es. Kein Formular, kein Verfahren – bei achtzig Gästen ist ein Anruf schneller als alles andere.",
  contactLink: "Zum Kontakt",
} as const;
