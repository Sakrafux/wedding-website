import { useState } from "react";

import { FormField } from "@/components/FormField";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { RadioCard, RadioCardGroup } from "@/components/ui/radio-card-group";
import type { RSVPMember } from "@/lib/api/dto";
import type { Attending } from "@/lib/api/enums";
import { attendingLabels, rsvpLabels } from "@/lib/labels";

import { countScopeChanges, type RSVPDraft, sharedScope } from "./state";

/**
 * The household-level "Wir kommen zu:" bulk setter.
 *
 * Not a stored field: choosing an option writes that scope into every member card
 * below, and the cards stay individually editable afterwards. It exists so that a
 * family of four answers in one tap and then edits only the exceptions.
 *
 * Once the cards disagree it shows nothing as chosen and says so, because a selector
 * that stayed lit while the cards said something else would be a lie about what gets
 * saved. Choosing an option again then overwrites them all, which is guarded by a
 * confirmation naming how many answers would change — silent bulk overwrite is how a
 * household loses Oma's church-only answer.
 *
 * It renders nothing for a household of one: that card already carries the same
 * question, and asking it twice on one screen is the kind of duplication that makes a
 * guest wonder which of the two counts.
 */
export function ScopeSelector({
  members,
  draft,
  onSelect,
}: {
  members: RSVPMember[];
  draft: RSVPDraft;
  onSelect: (scope: Attending) => void;
}) {
  const [pendingScope, setPendingScope] = useState<Attending | null>(null);

  if (members.length < 2) {
    return null;
  }

  const selected = sharedScope(draft, members);

  function choose(scope: Attending) {
    const changed = countScopeChanges(draft, members, scope);
    if (changed === 0) {
      return;
    }

    // Confirmed only when an answer that already exists would be replaced. A household
    // filling the form in for the first time is not overwriting anything, and a dialog
    // in front of the very first tap is a dialog that teaches people to dismiss dialogs.
    const wouldOverwriteAnAnswer = members.some((member) => {
      const current = draft.members[member.id]?.attending ?? null;
      return current !== null && current !== scope;
    });

    if (wouldOverwriteAnAnswer) {
      setPendingScope(scope);
      return;
    }
    onSelect(scope);
  }

  return (
    <>
      <FormField
        id="rsvp-household-scope"
        label={rsvpLabels.householdScopeHeading}
        help={rsvpLabels.householdScopeHelp}
      >
        <RadioCardGroup
          aria-labelledby="rsvp-household-scope-label"
          value={selected ?? ""}
          onValueChange={(value) => choose(value as Attending)}
        >
          {(Object.keys(attendingLabels) as Attending[]).map((scope) => (
            <RadioCard key={scope} value={scope} label={attendingLabels[scope]} />
          ))}
        </RadioCardGroup>
        {selected === null ? <p className="text-ink-muted text-small">{rsvpLabels.householdScopeMixed}</p> : null}
      </FormField>

      <AlertDialog open={pendingScope !== null} onOpenChange={(open) => !open && setPendingScope(null)}>
        <AlertDialogContent>
          <AlertDialogTitle>{rsvpLabels.householdScopeOverwriteTitle}</AlertDialogTitle>
          <AlertDialogDescription>
            {pendingScope
              ? rsvpLabels.householdScopeOverwriteBody(
                  countScopeChanges(draft, members, pendingScope),
                  attendingLabels[pendingScope],
                )
              : null}
          </AlertDialogDescription>
          <AlertDialogFooter>
            <AlertDialogCancel>{rsvpLabels.cancel}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (pendingScope) {
                  onSelect(pendingScope);
                }
                setPendingScope(null);
              }}
            >
              {rsvpLabels.householdScopeOverwriteConfirm}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
