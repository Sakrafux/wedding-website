import { useState } from "react";

import { FieldError } from "@/components/FormField";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { rsvpLabels } from "@/lib/labels";

/**
 * Takes a member the household added itself back off the list (F4-F02).
 *
 * Rendered only on `guest_added` cards. A seeded member gets no control and no
 * explanation nobody asked for: the way to say that person is not coming is the scope
 * control already on their card, which is what the server's refusal says too.
 *
 * Confirmed, and the confirmation names the person: one tap from a list is how the
 * wrong child gets removed on a phone. Not optimistic either — a card that vanishes
 * and comes back is worse than a brief pending state, and this is a rare action.
 */
export function RemoveMemberButton({ name, onRemove }: { name: string; onRemove: () => Promise<unknown> }) {
  const [isConfirming, setIsConfirming] = useState(false);
  const [isRemoving, setIsRemoving] = useState(false);
  const [error, setError] = useState<unknown>(null);

  async function remove() {
    setError(null);
    setIsRemoving(true);
    try {
      await onRemove();
    } catch (failure) {
      // Unreachable from this UI for `cannot_remove_member` — the control is not
      // rendered for a seeded member — but handled anyway: a stale tab may be looking
      // at somebody whose origin we changed in admin. The API's sentence, verbatim,
      // shown on the card the household is still looking at.
      setError(failure);
    } finally {
      setIsRemoving(false);
      setIsConfirming(false);
    }
  }

  const message = error instanceof Error ? error.message : undefined;

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        // The accessible name carries the person, because "entfernen" repeated once per
        // card tells a screen-reader user nothing about which one they are on.
        aria-label={rsvpLabels.removeMemberAccessibleName(name)}
        disabled={isRemoving}
        onClick={() => setIsConfirming(true)}
      >
        {rsvpLabels.removeMember}
      </Button>

      {message ? <FieldError>{message}</FieldError> : null}

      <AlertDialog open={isConfirming} onOpenChange={(open) => !open && setIsConfirming(false)}>
        <AlertDialogContent>
          <AlertDialogTitle>{rsvpLabels.removeMemberConfirmTitle}</AlertDialogTitle>
          <AlertDialogDescription>{rsvpLabels.removeMemberConfirmBody(name)}</AlertDialogDescription>
          <AlertDialogFooter>
            <AlertDialogCancel>{rsvpLabels.cancel}</AlertDialogCancel>
            {/* A plain Button rather than AlertDialogAction: that one closes on click,
                which would drop the dialog before the request has answered and leave
                the pending state with nowhere to show. */}
            <Button type="button" variant="destructive" disabled={isRemoving} onClick={() => void remove()}>
              {rsvpLabels.removeMemberConfirmAction}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
