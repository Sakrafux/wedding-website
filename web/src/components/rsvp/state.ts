/**
 * The RSVP form's state, and the pure functions over it.
 *
 * One object keyed by member id, mirroring the request shape of `PUT /api/rsvp`, so
 * that submitting is a serialisation rather than a gathering exercise — and so that a
 * field error keyed `members.<id>.<field>` is a lookup rather than a search.
 *
 * Everything here is pure. The rules that decide what gets *stored* live on the server
 * (domain.NormalizeGuestAnswer); what lives here is only what decides what gets
 * *shown*, and the two are deliberately not the same thing: the form keeps a meal
 * choice in state while somebody flips a scope back and forth, and the server drops it.
 */

import type { RSVPMember, RSVPResponse, RSVPSaveRequest } from "@/lib/api/dto";
import type { Attending, MealChoice, Portion, SeatingNeed } from "@/lib/api/enums";

/** The bounds the server enforces, mirrored so the controls cannot offer a rejected value. */
export const maxTransportSeats = 20;
export const maxNoteLength = 2000;
export const maxDietaryNoteLength = 500;

/** The point at which the note's character counter appears. Not before: a counter on
    an empty field reads as a limit on what you are allowed to say. */
export const noteCounterThreshold = 200;

/**
 * The transport answer as one question: nothing, we need seats, or we can offer them.
 *
 * Derived from the two counts and never stored beside them — `needed > 0` *is* "we
 * need". A third piece of state would be a third thing to keep in step with a request
 * body that has two numbers in it (F3-F07).
 */
export type TransportDirection = "none" | "needed" | "offered";

/** The count a chosen direction starts at. Not zero: a household that just said it
    needs seats, looking at a zero, has been told its answer does not count. */
const firstTransportSeat = 1;

/** One member's answer as the form holds it. `null` in `attending` is "not answered". */
export interface MemberDraft {
  id: number;
  attending: Attending | null;
  meal_choice: MealChoice | null;
  portion: Portion;
  midnight_snack: boolean;
  seating_need: SeatingNeed;
  dietary_note: string;
  age: number | null;
}

export interface RSVPDraft {
  transport_seats_needed: number;
  transport_seats_offered: number;
  has_stroller: boolean;
  rsvp_note: string;
  /** Keyed by member id. The order comes from the response, never from this object. */
  members: Record<number, MemberDraft>;
}

/** draftFrom seeds the form from the server's answer, unanswered members included. */
export function draftFrom(answer: RSVPResponse): RSVPDraft {
  const members: Record<number, MemberDraft> = {};
  for (const member of answer.members) {
    members[member.id] = withMealDefault({
      id: member.id,
      attending: member.attending,
      meal_choice: member.meal_choice,
      portion: member.portion,
      midnight_snack: member.midnight_snack,
      seating_need: member.seating_need,
      dietary_note: member.dietary_note,
      age: member.age,
    });
  }

  return {
    transport_seats_needed: answer.household.transport_seats_needed,
    transport_seats_offered: answer.household.transport_seats_offered,
    has_stroller: answer.household.has_stroller,
    rsvp_note: answer.household.rsvp_note,
    members,
  };
}

/**
 * reseedDraft takes a fresh answer and keeps what the household has typed.
 *
 * The member list is the server's — a member who was added or removed appears or
 * disappears — while every answer already on screen survives. Re-seeding wholesale
 * would be simpler and wrong: adding a plus-one changes the member set (F4-F01), and a
 * household that lost a half-filled form to the act of naming their companion would
 * have no reason to trust the next attempt either.
 */
export function reseedDraft(current: RSVPDraft, answer: RSVPResponse): RSVPDraft {
  const seeded = draftFrom(answer);
  const members: Record<number, MemberDraft> = {};

  for (const member of answer.members) {
    const typed = current.members[member.id];
    const fresh = seeded.members[member.id];
    if (typed) {
      members[member.id] = typed;
    } else if (fresh) {
      members[member.id] = fresh;
    }
  }

  return { ...current, members };
}

/**
 * toRequest serialises the draft in the order the response listed the members.
 *
 * The order is the response's because the body has to name exactly the household's
 * living members, and iterating the object's keys would silently drop a member the
 * draft never learned about.
 *
 * A member with no scope cannot be serialised — the field is required on the wire —
 * so this is only ever called once `missingAnswers` is empty.
 */
export function toRequest(draft: RSVPDraft, members: RSVPMember[]): RSVPSaveRequest {
  return {
    transport_seats_needed: draft.transport_seats_needed,
    transport_seats_offered: draft.transport_seats_offered,
    has_stroller: draft.has_stroller,
    rsvp_note: draft.rsvp_note,
    members: members.map((member) => {
      const memberDraft = draft.members[member.id];
      return {
        id: member.id,
        // Defaulted only to satisfy the type: submit is blocked while any scope is
        // missing, and a request built here always has one.
        attending: memberDraft?.attending ?? "no",
        meal_choice: memberDraft?.meal_choice ?? null,
        portion: memberDraft?.portion ?? "full",
        midnight_snack: memberDraft?.midnight_snack ?? false,
        seating_need: memberDraft?.seating_need ?? "normal",
        dietary_note: memberDraft?.dietary_note ?? "",
        age: memberDraft?.age ?? null,
      };
    }),
  };
}

/**
 * withMealDefault pre-answers the meal with `all` once a scope covers the party.
 *
 * Applied both when the draft is seeded and when a scope changes, so the default
 * reaches the request body rather than only the radio group — a control defaulted in
 * its `value` prop submits whatever the state still says, which is null.
 *
 * The accepted cost: we can no longer tell "said: eats everything" from "did not look
 * at the field". The caterer numbers are unaffected — a plate is ordered either way —
 * and three quarters of the guest list would have tapped this option.
 */
export function withMealDefault(member: MemberDraft): MemberDraft {
  if (!coversParty(member.attending) || member.meal_choice !== null) {
    return member;
  }
  return { ...member, meal_choice: "all" };
}

/** transportDirection reads the household's transport answer out of the two counts. */
export function transportDirection(draft: RSVPDraft): TransportDirection {
  if (draft.transport_seats_needed > 0) {
    return "needed";
  }
  if (draft.transport_seats_offered > 0) {
    return "offered";
  }
  return "none";
}

/**
 * withTransportDirection is the change a direction card makes: one count set, the
 * other zeroed.
 *
 * Zeroing the other one is the whole point — a household that both needs and offers
 * seats is refused by the server (F3-B07), and this is what makes that state
 * unreachable from the form rather than merely unsaved.
 */
export function withTransportDirection(draft: RSVPDraft, direction: TransportDirection): Partial<RSVPDraft> {
  switch (direction) {
    case "needed":
      return {
        transport_seats_needed: Math.max(draft.transport_seats_needed, firstTransportSeat),
        transport_seats_offered: 0,
      };
    case "offered":
      return {
        transport_seats_needed: 0,
        transport_seats_offered: Math.max(draft.transport_seats_offered, firstTransportSeat),
      };
    case "none":
      return { transport_seats_needed: 0, transport_seats_offered: 0 };
  }
}

/** The lowest count a chosen direction may hold. Going back to zero is what the
    "we need nothing" card is for, and it says so in words. */
export const minChosenTransportSeats = firstTransportSeat;

/** Members the household has not answered for yet, in the order they are shown. */
export function missingAnswers(draft: RSVPDraft, members: RSVPMember[]): RSVPMember[] {
  return members.filter((member) => !draft.members[member.id]?.attending);
}

/**
 * The scope the whole household shares, or null when they differ.
 *
 * Null is what makes the bulk selector show nothing as chosen: a selector that stayed
 * lit while the cards below disagreed would be a lie about what gets saved. An
 * unanswered member counts as a difference, so a half-filled form does not look
 * uniform.
 */
export function sharedScope(draft: RSVPDraft, members: RSVPMember[]): Attending | null {
  const scopes = members.map((member) => draft.members[member.id]?.attending ?? null);
  const first = scopes[0] ?? null;

  return first !== null && scopes.every((scope) => scope === first) ? first : null;
}

/** How many members a bulk scope change would actually change. Zero means no
    confirmation is needed, because nothing would be overwritten. */
export function countScopeChanges(draft: RSVPDraft, members: RSVPMember[], scope: Attending): number {
  return members.filter((member) => (draft.members[member.id]?.attending ?? null) !== scope).length;
}

/**
 * Whether anybody in the household attends both halves of the day.
 *
 * The transport section exists only for that case: the trip is church → reception, and
 * the server zeroes the counts otherwise (domain.NormalizeHouseholdAnswer).
 */
export function attendsBoth(draft: RSVPDraft, members: RSVPMember[]): boolean {
  return members.some((member) => draft.members[member.id]?.attending === "both");
}

/** Whether the scope covers the party, which is what gates the catering fields. */
export function coversParty(scope: Attending | null): boolean {
  return scope === "party_only" || scope === "both";
}

/** Whether the guest is coming at all — the gate for seating need and allergies, which
    apply in the pew as much as at the table. */
export function attendsAnything(scope: Attending | null): boolean {
  return scope !== null && scope !== "no";
}

/** The field error key the API uses for one member's field: `members.31.attending`. */
export function memberFieldKey(memberId: number, field: keyof MemberDraft): string {
  return `members.${memberId}.${field}`;
}

/** The DOM id of a member's card, so the error summary at the top can link to it. */
export function memberCardId(memberId: number): string {
  return `rsvp-member-${memberId}`;
}
