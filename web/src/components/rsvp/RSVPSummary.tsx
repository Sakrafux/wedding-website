import type { RSVPMember, RSVPResponse } from "@/lib/api/dto";
import { formatShortDate } from "@/lib/dates";
import {
  attendingLabels,
  attendingUnansweredLabel,
  mealChoiceLabels,
  portionLabels,
  rsvpLabels,
  seatingNeedLabels,
} from "@/lib/labels";

import { coversParty } from "./state";

/**
 * A recap of exactly what is stored, in German labels rather than enum values.
 *
 * Rendered from the **response**, never from the form state, so a value the server
 * normalized away is visibly absent — a church-only guest shows no meal, which is the
 * honest recap of what we will actually order.
 *
 * Two screens use this one component: the post-save confirmation and the read-only view
 * after the deadline (F3-F05). Text on `surface-sunken`, not disabled inputs — disabled
 * text fails contrast and reads as broken.
 */
export function RSVPSummary({ answer }: { answer: RSVPResponse }) {
  return (
    <div className="bg-surface-sunken flex flex-col gap-6 rounded-xl p-4">
      <section className="flex flex-col gap-3">
        <h3 className="text-h3 font-body">{rsvpLabels.summaryMembersHeading}</h3>
        <ul className="flex flex-col gap-4">
          {answer.members.map((member) => (
            <li key={member.id} className="flex flex-col gap-1">
              <p className="font-medium">{member.name}</p>
              <p>{member.attending ? attendingLabels[member.attending] : attendingUnansweredLabel}</p>
              <MemberDetails member={member} />
            </li>
          ))}
        </ul>
      </section>

      <section className="flex flex-col gap-3">
        <h3 className="text-h3 font-body">{rsvpLabels.summaryHouseholdHeading}</h3>
        <dl className="flex flex-col gap-1">
          {/* The counts are shown only when they are stored: zero seats on a household
              that never made the trip is a number nobody entered. */}
          {answer.household.transport_seats_needed > 0 ? (
            <SummaryRow label={rsvpLabels.summaryTransportNeeded} value={answer.household.transport_seats_needed} />
          ) : null}
          {answer.household.transport_seats_offered > 0 ? (
            <SummaryRow label={rsvpLabels.summaryTransportOffered} value={answer.household.transport_seats_offered} />
          ) : null}
          {answer.household.has_stroller ? (
            <SummaryRow label={rsvpLabels.summaryStroller} value={rsvpLabels.summaryStrollerYes} />
          ) : null}
          {answer.household.rsvp_note ? (
            <SummaryRow label={rsvpLabels.summaryNote} value={answer.household.rsvp_note} />
          ) : null}
        </dl>
      </section>

      {answer.household.rsvp_updated_at ? (
        <p className="text-ink-muted text-small">
          {rsvpLabels.summarySavedAt(formatShortDate(answer.household.rsvp_updated_at))}
        </p>
      ) : null}
    </div>
  );
}

/** The catering half of one member's recap, which exists only for a party guest. */
function MemberDetails({ member }: { member: RSVPMember }) {
  const isComing = member.attending !== null && member.attending !== "no";

  return (
    <dl className="text-ink-muted text-small flex flex-col gap-0.5">
      {coversParty(member.attending) ? (
        <>
          <SummaryRow
            label={rsvpLabels.summaryMeal}
            value={member.meal_choice ? mealChoiceLabels[member.meal_choice] : rsvpLabels.summaryNothing}
          />
          <SummaryRow label={rsvpLabels.summaryPortion} value={portionLabels[member.portion]} />
          <SummaryRow
            label={rsvpLabels.summarySnack}
            value={member.midnight_snack ? rsvpLabels.summarySnackYes : rsvpLabels.summarySnackNo}
          />
        </>
      ) : null}

      {/* Not gated by the party: a wheelchair space is needed in the pew too. */}
      {isComing && member.seating_need !== "normal" ? (
        <SummaryRow label={rsvpLabels.summarySeatingNeed} value={seatingNeedLabels[member.seating_need]} />
      ) : null}
      {isComing && member.dietary_note ? (
        <SummaryRow label={rsvpLabels.summaryDietaryNote} value={member.dietary_note} />
      ) : null}
    </dl>
  );
}

function SummaryRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex flex-wrap gap-x-2">
      <dt className="after:content-[':']">{label}</dt>
      <dd>{value}</dd>
    </div>
  );
}
