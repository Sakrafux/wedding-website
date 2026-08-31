import { useState } from "react";

import { FieldError, FormField } from "@/components/FormField";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ApiError } from "@/lib/api/client";
import type { RSVPAddMemberResponse } from "@/lib/api/dto";
import { contactPhoneNumber, rsvpLabels } from "@/lib/labels";

import { memberCardId } from "./state";

/** The server's bound on a name (dto.RSVPAddMemberRequest), mirrored so the control
    cannot offer a value the server would reject. */
const maxNameLength = 160;

/**
 * The plus-one block: either the trigger that opens the sheet, or the sentence that
 * tells everybody else how to get somebody added.
 *
 * It branches on `canAdd` and on nothing else. The rule — one adult, and only for a
 * household we seeded as a single person — lives on the server, and re-deriving it
 * from the member list here is how the screen ends up refusing something the server
 * allows, in front of a guest who can see both.
 *
 * The confirmation is held here rather than inside the sheet, because a successful
 * addition flips `canAdd` to false and takes the sheet with it — a message rendered in
 * there would vanish in the same paint that earned it.
 */
export function AddPlusOne({
  canAdd,
  onAdd,
  onRSVPClosed,
}: {
  canAdd: boolean;
  onAdd: (name: string) => Promise<RSVPAddMemberResponse>;
  /** Called when the server answers `rsvp_closed`: the deadline passed while this form
      was open, and the page owes the guest the read-only view rather than a retry. */
  onRSVPClosed: () => void;
}) {
  const [added, setAdded] = useState<string | null>(null);

  return (
    <div className="flex flex-col gap-2">
      {canAdd ? (
        <AddPlusOneSheet onAdd={onAdd} onAdded={setAdded} onRSVPClosed={onRSVPClosed} />
      ) : (
        <PlusOneUnavailable />
      )}

      {/* Says both halves, because the asymmetry is real and confusing: the person is
          on the list already, and the rest of the form is still unsaved. */}
      {added ? (
        <p className="text-ink-muted text-small" role="status">
          {rsvpLabels.addPlusOneAdded(added)}
        </p>
      ) : null}
    </div>
  );
}

/**
 * What a household that may not add reads instead.
 *
 * Not a disabled button — it fails contrast and reads as broken — and not nothing: a
 * guest wondering whether they missed the control phones us, which is the call this
 * block exists to make unnecessary. Quiet on purpose: most households see it from the
 * first render, and nothing has gone wrong.
 */
function PlusOneUnavailable() {
  return (
    <p className="text-ink-muted text-small">
      {rsvpLabels.plusOneUnavailable} {rsvpLabels.plusOneUnavailableCall}{" "}
      <a className="underline" href={`tel:${contactPhoneNumber.replaceAll(" ", "")}`}>
        {contactPhoneNumber}
      </a>
    </p>
  );
}

function AddPlusOneSheet({
  onAdd,
  onAdded,
  onRSVPClosed,
}: {
  onAdd: (name: string) => Promise<RSVPAddMemberResponse>;
  onAdded: (name: string) => void;
  onRSVPClosed: () => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [name, setName] = useState("");
  const [error, setError] = useState<unknown>(null);
  const [isAdding, setIsAdding] = useState(false);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError(null);
    setIsAdding(true);

    try {
      const response = await onAdd(name.trim());
      setName("");
      setIsOpen(false);
      onAdded(response.member.name);
      // The new card carries the next question — is this person coming — so it is
      // brought on screen rather than left below the fold.
      document.getElementById(memberCardId(response.member.id))?.scrollIntoView({ block: "center" });
    } catch (failure) {
      if (failure instanceof ApiError && failure.code === "rsvp_closed") {
        setIsOpen(false);
        onRSVPClosed();
        return;
      }
      setError(failure);
    } finally {
      setIsAdding(false);
    }
  }

  // The API's own German sentence, verbatim — including the 409 a second tab produces,
  // where the server owns both the rule and the wording.
  const message = error instanceof Error ? error.message : undefined;

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button type="button" variant="outline" size="lg" className="h-14">
          {rsvpLabels.addPlusOneTrigger}
        </Button>
      </DialogTrigger>

      <DialogContent closeLabel={rsvpLabels.close}>
        <DialogTitle>{rsvpLabels.addPlusOneHeading}</DialogTitle>
        <DialogDescription>{rsvpLabels.addPlusOneBody}</DialogDescription>

        <form className="flex flex-col gap-4" onSubmit={(event) => void submit(event)} noValidate>
          <FormField id="plus-one-name" label={rsvpLabels.addPlusOneNameLabel} help={rsvpLabels.addPlusOneNameHelp}>
            <Input
              id="plus-one-name"
              value={name}
              maxLength={maxNameLength}
              autoComplete="name"
              onChange={(event) => setName(event.target.value)}
            />
          </FormField>

          {message ? <FieldError>{message}</FieldError> : null}

          <Button type="submit" size="lg" className="h-14" disabled={isAdding || name.trim() === ""}>
            {isAdding ? rsvpLabels.addPlusOneSubmitting : rsvpLabels.addPlusOneSubmit}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
