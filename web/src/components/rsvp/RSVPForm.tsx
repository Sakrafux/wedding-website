import { useEffect, useRef, useState } from "react";

import { FormField } from "@/components/FormField";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api/client";
import type { RSVPResponse, RSVPSaveRequest } from "@/lib/api/dto";
import type { Attending } from "@/lib/api/enums";
import { formatShortDate } from "@/lib/dates";
import { contactPhoneNumber, rsvpLabels } from "@/lib/labels";
import { cn } from "@/lib/utils";

import { RSVPMemberCard } from "./RSVPMemberCard";
import { RSVPSummary } from "./RSVPSummary";
import { ScopeSelector } from "./ScopeSelector";
import { TransportFields } from "./TransportFields";
import {
  attendsBoth,
  draftFrom,
  maxNoteLength,
  memberCardId,
  missingAnswers,
  noteCounterThreshold,
  type RSVPDraft,
  toRequest,
} from "./state";

/**
 * The RSVP form, for a household or for us on their behalf.
 *
 * **It takes its data and its save function as props and never fetches for itself.**
 * That is the constraint the whole epic is arranged around: the guest route wires it to
 * `/api/rsvp` and the admin route to `/api/admin/households/{id}/rsvp`, and a
 * `useQuery` in here is the one change that would break the admin page and force a
 * second form into existence.
 *
 * It renders no `<h1>` either — the guest page's heading and the admin page's ("you are
 * answering for Familie Müller") say different things, and two `<h1>`s on one page is a
 * heading structure a screen-reader user cannot navigate. Sections here are `<h2>`,
 * member cards `<h3>`.
 */
export function RSVPForm({
  answer,
  onSave,
  onReload,
  allowEditingAfterDeadline = false,
  dense = false,
}: {
  answer: RSVPResponse;
  onSave: (request: RSVPSaveRequest) => Promise<RSVPResponse>;
  /** Refetches the answer. Used for the stale-tab case, where merging state would be
      the dishonest option — see `member_set_mismatch` below. */
  onReload: () => Promise<unknown>;
  /**
   * The admin override. `editable: false` means "the deadline has passed", not "you may
   * not write" — the admin page exists for the late phone call and passes this
   * explicitly, so that the shared default stays the safe one (F3-B06).
   */
  allowEditingAfterDeadline?: boolean;
  /** Admin density: the same controls, without the guest page's decorative spacing. */
  dense?: boolean;
}) {
  const [draft, setDraft] = useState<RSVPDraft>(() => draftFrom(answer));
  const [submitAttempted, setSubmitAttempted] = useState(false);
  const [saved, setSaved] = useState<RSVPResponse | null>(null);
  const [saveError, setSaveError] = useState<unknown>(null);
  const [isSaving, setIsSaving] = useState(false);
  const confirmationHeading = useRef<HTMLHeadingElement>(null);

  const members = answer.members;
  const memberSetKey = members.map((member) => member.id).join(",");
  const [seededMemberSet, setSeededMemberSet] = useState(memberSetKey);

  // Re-seeding when the member set changes is the `member_set_mismatch` recovery: the
  // reload brings a different list of people, and a draft keyed by the old ids would
  // render the form without the member that was just added. Adjusting state during
  // render rather than in an effect, so the first paint after the reload is already
  // right — an effect would flash the stale list.
  if (seededMemberSet !== memberSetKey) {
    setSeededMemberSet(memberSetKey);
    setDraft(draftFrom(answer));
    setSaveError(null);
    setSubmitAttempted(false);
  }

  const missing = missingAnswers(draft, members);
  const fieldErrors = saveError instanceof ApiError ? saveError.fields : {};
  const errorCode = saveError instanceof ApiError ? saveError.code : undefined;

  // The server refused because the deadline passed while this form was open. It is a
  // race a guest can genuinely hit — filling the form in on the evening of the deadline
  // — and the honest answer is the read-only view, not a retry button.
  const closedByServer = errorCode === "rsvp_closed";
  const isReadOnly = (!answer.editable && !allowEditingAfterDeadline) || closedByServer;

  useEffect(() => {
    // Focus lands on the confirmation, so a screen-reader user is told the answer was
    // saved instead of being left at the bottom of a form that has disappeared.
    if (saved) {
      confirmationHeading.current?.focus();
    }
  }, [saved]);

  function updateMember(memberId: number, change: Partial<RSVPDraft["members"][number]>) {
    setDraft((current) => {
      const member = current.members[memberId];
      if (!member) {
        return current;
      }
      return { ...current, members: { ...current.members, [memberId]: { ...member, ...change } } };
    });
  }

  function setHouseholdScope(scope: Attending) {
    setDraft((current) => {
      const updated: RSVPDraft["members"] = {};
      for (const [id, member] of Object.entries(current.members)) {
        updated[Number(id)] = { ...member, attending: scope };
      }
      return { ...current, members: updated };
    });
  }

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setSubmitAttempted(true);
    setSaveError(null);

    // Checked here as well as on the server: the server rejects the same case, but a
    // guest who is missing one answer should not have to wait for a round trip to find
    // out which one.
    const unanswered = missingAnswers(draft, members);
    if (unanswered.length > 0) {
      const firstMissing = unanswered[0];
      if (firstMissing) {
        document.getElementById(memberCardId(firstMissing.id))?.scrollIntoView({ block: "center" });
      }
      return;
    }

    setIsSaving(true);
    try {
      setSaved(await onSave(toRequest(draft, members)));
    } catch (error) {
      // Every entered value stays on screen: losing a filled-in form to a dropped
      // connection in a village is the failure this audience will actually hit.
      setSaveError(error);
    } finally {
      setIsSaving(false);
    }
  }

  if (saved) {
    return (
      <div className="flex flex-col gap-6">
        <Alert>
          {/* The heading is the Alert's title and also the focus target, so it is
              rendered directly rather than through AlertTitle's paragraph. */}
          <h2 ref={confirmationHeading} tabIndex={-1} className="text-h2 font-display">
            {rsvpLabels.summaryHeading}
          </h2>
          <AlertDescription>
            <p>{rsvpLabels.changeableUntil(formatShortDate(answer.deadline))}</p>
          </AlertDescription>
        </Alert>

        <RSVPSummary answer={saved} />

        <Button type="button" variant="outline" size="lg" className="h-14" onClick={() => setSaved(null)}>
          {rsvpLabels.changeAnswer}
        </Button>
      </div>
    );
  }

  if (isReadOnly) {
    return <ClosedAnswer answer={answer} deadline={answer.deadline} />;
  }

  const transportVisible = attendsBoth(draft, members);
  const hadTransportSeats = answer.household.transport_seats_needed > 0 || answer.household.transport_seats_offered > 0;

  return (
    <form
      onSubmit={(event) => void submit(event)}
      className={cn("flex flex-col", dense ? "gap-6" : "gap-12")}
      noValidate
    >
      <p className="text-ink-muted">{rsvpLabels.deadlineNotice(formatShortDate(answer.deadline))}</p>
      {answer.household.rsvp_updated_at ? (
        <p className="text-ink-muted text-small">
          {rsvpLabels.lastChanged(formatShortDate(answer.household.rsvp_updated_at))}
        </p>
      ) : null}

      <ScopeSelector members={members} draft={draft} onSelect={setHouseholdScope} />

      <section className="flex flex-col gap-4">
        <h2 className="text-h2 font-display">{rsvpLabels.membersHeading}</h2>

        {/* The summary at the top links to each unanswered card, so one missing answer
            on a form with eight people is not a hunt. */}
        {submitAttempted && missing.length > 0 ? (
          <Alert variant="attention" role="alert">
            <AlertTitle>{rsvpLabels.missingAnswersHeading}</AlertTitle>
            <AlertDescription>
              <ul className="flex flex-col gap-1">
                {missing.map((member) => (
                  <li key={member.id}>
                    <a className="underline" href={`#${memberCardId(member.id)}`}>
                      {rsvpLabels.missingAnswerLink(member.name)}
                    </a>
                  </li>
                ))}
              </ul>
            </AlertDescription>
          </Alert>
        ) : null}

        {members.map((member) => {
          const memberDraft = draft.members[member.id];
          return memberDraft ? (
            <RSVPMemberCard
              key={member.id}
              member={member}
              draft={memberDraft}
              onChange={(change) => updateMember(member.id, change)}
              fieldErrors={fieldErrors}
              showMissingAnswer={submitAttempted}
            />
          ) : null;
        })}
      </section>

      <TransportFields
        draft={draft}
        visible={transportVisible}
        droppedNotice={hadTransportSeats}
        onChange={(change) => setDraft((current) => ({ ...current, ...change }))}
        fieldErrors={fieldErrors}
      />

      <section className="flex flex-col gap-4">
        <h2 className="text-h2 font-display">{rsvpLabels.noteHeading}</h2>
        <FormField id="rsvp-note" label={rsvpLabels.noteLabel} help={rsvpLabels.noteHelp} error={fieldErrors.rsvp_note}>
          <Textarea
            id="rsvp-note"
            rows={4}
            maxLength={maxNoteLength}
            value={draft.rsvp_note}
            aria-describedby="rsvp-note-hint"
            onChange={(event) => setDraft((current) => ({ ...current, rsvp_note: event.target.value }))}
          />
          <span id="rsvp-note-hint" className="text-ink-muted text-small">
            {rsvpLabels.noteHint} {rsvpLabels.noteReadPromise}
          </span>
          {/* Only near the cap: a counter on an empty field reads as a limit on what you
              are allowed to say. */}
          {maxNoteLength - draft.rsvp_note.length < noteCounterThreshold ? (
            <span className="text-ink-muted text-small" aria-live="polite">
              {rsvpLabels.noteRemaining(maxNoteLength - draft.rsvp_note.length)}
            </span>
          ) : null}
        </FormField>
      </section>

      {saveError ? <SaveFailure error={saveError} onReload={onReload} /> : null}

      <Button type="submit" size="lg" className="h-14 w-full" disabled={isSaving}>
        {isSaving ? rsvpLabels.submitting : rsvpLabels.submit}
      </Button>
    </form>
  );
}

/**
 * What went wrong on save.
 *
 * The API's own German sentence, verbatim — the server owns the wording so that every
 * screen says the same thing. `member_set_mismatch` additionally offers a reload: the
 * household's member list changed under this tab, and merging the two states is the one
 * option that can silently drop somebody's answer.
 */
function SaveFailure({ error, onReload }: { error: unknown; onReload: () => Promise<unknown> }) {
  const isMismatch = error instanceof ApiError && error.code === "member_set_mismatch";
  const message = error instanceof Error ? error.message : rsvpLabels.saveFailedHeading;

  return (
    <Alert variant="attention" role="alert">
      <AlertTitle>{rsvpLabels.saveFailedHeading}</AlertTitle>
      <AlertDescription>
        <p>{message}</p>
        {isMismatch ? (
          <Button type="button" variant="outline" className="mt-2 self-start" onClick={() => void onReload()}>
            {rsvpLabels.reload}
          </Button>
        ) : null}
      </AlertDescription>
    </Alert>
  );
}

/**
 * The page after the deadline: what we have, and what to do if it is wrong.
 *
 * Text on `surface-sunken` through the same summary component the post-save state uses,
 * never disabled inputs. A household that never answered gets a different sentence,
 * because explaining a form they cannot use would be answering a question they did not
 * ask — the useful thing to say is to call us.
 */
function ClosedAnswer({ answer, deadline }: { answer: RSVPResponse; deadline: string }) {
  const hasAnswered = answer.household.rsvp_submitted_at !== null;

  return (
    <div className="flex flex-col gap-6">
      <Alert variant="attention">
        <AlertTitle>{rsvpLabels.closedHeading}</AlertTitle>
        <AlertDescription>
          <p>{hasAnswered ? rsvpLabels.closedBody(formatShortDate(deadline)) : rsvpLabels.closedNothingRecorded}</p>
          {/* A telephone link, because the remedy is a phone call and this audience is
              holding a phone. */}
          <a className="underline" href={`tel:${contactPhoneNumber.replaceAll(" ", "")}`}>
            {contactPhoneNumber}
          </a>
        </AlertDescription>
      </Alert>

      {hasAnswered ? <RSVPSummary answer={answer} /> : null}
    </div>
  );
}
