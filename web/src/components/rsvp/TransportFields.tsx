import { FormField } from "@/components/FormField";
import { Stepper } from "@/components/Stepper";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { rsvpLabels } from "@/lib/labels";

import { maxTransportSeats, type RSVPDraft } from "./state";

/**
 * The household's transport answers, and the pram.
 *
 * The seat counts are shown only when somebody in the household attends both halves of
 * the day: the trip is church → reception, nobody else makes it, and the server zeroes
 * the counts in that case anyway. Hiding them keeps two questions off the screen of
 * every household that is coming to only one half.
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
  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-h2 font-display">{rsvpLabels.transportHeading}</h2>

      {visible ? (
        <>
          <FormField
            id="rsvp-transport-needed"
            label={rsvpLabels.transportNeededLabel}
            help={rsvpLabels.transportNeededHelp}
            error={fieldErrors.transport_seats_needed}
          >
            <Stepper
              id="rsvp-transport-needed"
              label={rsvpLabels.transportNeededLabel}
              value={draft.transport_seats_needed}
              max={maxTransportSeats}
              onChange={(value) => onChange({ transport_seats_needed: value })}
            />
          </FormField>

          <FormField
            id="rsvp-transport-offered"
            label={rsvpLabels.transportOfferedLabel}
            help={rsvpLabels.transportOfferedHelp}
            error={fieldErrors.transport_seats_offered}
          >
            <Stepper
              id="rsvp-transport-offered"
              label={rsvpLabels.transportOfferedLabel}
              value={draft.transport_seats_offered}
              max={maxTransportSeats}
              onChange={(value) => onChange({ transport_seats_offered: value })}
            />
          </FormField>
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
