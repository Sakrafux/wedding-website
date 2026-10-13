import { FormField } from "@/components/FormField";
import { Stepper } from "@/components/Stepper";
import { rsvpLabels } from "@/lib/labels";

import type { RSVPDraft } from "./state";

/**
 * The household's answer for the way home: how many seats in a taxi from the
 * reception back into the city.
 *
 * A count without a direction question, unlike the shuttle section beside it
 * (F3-B07): a taxi is something *we* would order for the trip away from the
 * venue, so there is no "we can offer" side to refuse and zero needs no third
 * card — it is the honest "we need no seats".
 *
 * Shown only when somebody in the household attends the party, which is the
 * trip's scope — a church-only household never reaches the reception — and the
 * server zeroes the count in that case anyway. When the section disappears over
 * an answer that carried a count, `droppedNotice` says so rather than letting
 * the number vanish silently.
 */
export function TaxiFields({
  draft,
  visible,
  droppedNotice,
  onChange,
  fieldErrors,
}: {
  draft: RSVPDraft;
  /** True when at least one member attends the party. */
  visible: boolean;
  /** True when the section is hidden but the household had entered a count. */
  droppedNotice: boolean;
  onChange: (change: Partial<RSVPDraft>) => void;
  fieldErrors: Record<string, string>;
}) {
  if (!visible && !droppedNotice) {
    return null;
  }

  // Capped at the household's own size: nobody orders taxi seats for people who
  // are not coming, and the server's blanket bound of twenty would let a
  // household of two ask for a bus.
  const maxSeats = Object.keys(draft.members).length;

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-h2 font-display">{rsvpLabels.taxiHeading}</h2>

      {visible ? (
        <FormField
          id="rsvp-taxi-seats"
          label={rsvpLabels.taxiSeatsLabel}
          help={rsvpLabels.taxiSeatsHelp}
          error={fieldErrors.taxi_seats_needed}
        >
          <Stepper
            id="rsvp-taxi-seats"
            label={rsvpLabels.taxiSeatsLabel}
            value={draft.taxi_seats_needed}
            min={0}
            max={maxSeats}
            onChange={(seats) => onChange({ taxi_seats_needed: seats })}
          />
        </FormField>
      ) : (
        <p className="text-ink-muted text-small">{rsvpLabels.taxiDropped}</p>
      )}
    </section>
  );
}
