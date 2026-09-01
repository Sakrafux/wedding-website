import { FormField } from "@/components/FormField";
import { Stepper } from "@/components/Stepper";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { RadioCard, RadioCardGroup } from "@/components/ui/radio-card-group";
import { rsvpLabels, transportDirectionLabels } from "@/lib/labels";

import {
  maxTransportSeats,
  minChosenTransportSeats,
  type RSVPDraft,
  type TransportDirection,
  transportDirection,
  withTransportDirection,
} from "./state";

/**
 * The household's transport answer, and the pram.
 *
 * **One question with a direction**, not two counts: needing seats and offering them
 * are mutually exclusive, the server refuses a household that claims both (`F3-B07`),
 * and as two steppers side by side they read as the same question asked twice. The
 * direction is derived from the counts rather than stored beside them — `needed > 0`
 * *is* "we need seats".
 *
 * The whole section is shown only when somebody in the household attends both halves of
 * the day: the trip is church → reception, nobody else makes it, and the server zeroes
 * the counts in that case anyway.
 *
 * When they *were* filled in and the section then disappears, `droppedNotice` says so
 * rather than letting the numbers vanish silently — a number that disappears without
 * explanation is exactly the kind of thing that gets phoned in about.
 */
export function TransportFields({
  draft,
  visible,
  droppedNotice,
  onChange,
  fieldErrors,
}: {
  draft: RSVPDraft;
  /** True when at least one member attends both. */
  visible: boolean;
  /** True when the section is hidden but the household had entered a count. */
  droppedNotice: boolean;
  onChange: (change: Partial<RSVPDraft>) => void;
  fieldErrors: Record<string, string>;
}) {
  const direction = transportDirection(draft);

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-h2 font-display">{rsvpLabels.transportHeading}</h2>

      {visible ? (
        <>
          <FormField
            id="rsvp-transport-direction"
            label={rsvpLabels.transportDirectionLabel}
            help={rsvpLabels.transportDirectionHelp}
            // The server keys its refusal to both counts, because it cannot know which
            // of the two the form is showing (F3-B07). Either message belongs here.
            error={fieldErrors.transport_seats_needed ?? fieldErrors.transport_seats_offered}
          >
            <RadioCardGroup
              aria-labelledby="rsvp-transport-direction-label"
              value={direction}
              onValueChange={(value) => onChange(withTransportDirection(draft, value as TransportDirection))}
            >
              {(Object.keys(transportDirectionLabels) as TransportDirection[]).map((option) => (
                <RadioCard key={option} value={option} label={transportDirectionLabels[option]} />
              ))}
            </RadioCardGroup>
          </FormField>

          {direction === "none" ? null : <SeatCount direction={direction} draft={draft} onChange={onChange} />}
        </>
      ) : (
        droppedNotice && <p className="text-ink-muted text-small">{rsvpLabels.transportDropped}</p>
      )}

      {/* A household fact rather than a person's: a pram is parked, not sat on, and
          nobody brings two. */}
      <FormField id="rsvp-stroller" label={rsvpLabels.strollerLabel} help={rsvpLabels.strollerHelp}>
        <div className="flex min-h-12 items-center gap-3">
          <Checkbox
            id="rsvp-stroller"
            checked={draft.has_stroller}
            onCheckedChange={(checked) => onChange({ has_stroller: checked === true })}
          />
          <Label htmlFor="rsvp-stroller" className="text-ink-muted text-small">
            {rsvpLabels.strollerLabel}
          </Label>
        </div>
      </FormField>
    </section>
  );
}

/**
 * How many seats, once a direction is chosen.
 *
 * One stepper for both directions, labelled by the chosen one: two steppers with one
 * hidden is what this control replaced. The minimum is one, not zero — going back to
 * nothing is what the third card above is for, and it says so in words.
 */
function SeatCount({
  direction,
  draft,
  onChange,
}: {
  direction: Exclude<TransportDirection, "none">;
  draft: RSVPDraft;
  onChange: (change: Partial<RSVPDraft>) => void;
}) {
  const isNeeded = direction === "needed";
  const label = isNeeded ? rsvpLabels.transportNeededLabel : rsvpLabels.transportOfferedLabel;
  const value = isNeeded ? draft.transport_seats_needed : draft.transport_seats_offered;

  return (
    <FormField id="rsvp-transport-seats" label={label} help={rsvpLabels.transportSeatsHelp}>
      <Stepper
        id="rsvp-transport-seats"
        label={label}
        value={value}
        min={minChosenTransportSeats}
        max={maxTransportSeats}
        onChange={(seats) =>
          onChange(isNeeded ? { transport_seats_needed: seats } : { transport_seats_offered: seats })
        }
      />
    </FormField>
  );
}
