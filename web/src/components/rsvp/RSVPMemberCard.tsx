import { FormField } from "@/components/FormField";
import { Card } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioCard, RadioCardGroup } from "@/components/ui/radio-card-group";
import type { RSVPMember } from "@/lib/api/dto";
import type { Attending, MealChoice, Portion, SeatingNeed } from "@/lib/api/enums";
import {
  attendingLabels,
  mealChoiceLabels,
  portionHelpLabels,
  portionLabels,
  rsvpLabels,
  seatingNeedLabels,
} from "@/lib/labels";

import {
  attendsAnything,
  coversParty,
  maxDietaryNoteLength,
  memberCardId,
  memberFieldKey,
  type MemberDraft,
} from "./state";

/** The age range for a child, matching the server's rule (domain.ResolveAge). */
const maxChildAge = 17;

/**
 * One person's answer.
 *
 * The catering fields are **revealed**, not disabled, once the scope covers the party.
 * A disabled control fails contrast and reads as broken; an absent question reads as a
 * question that was not asked, which is exactly what it is. Scope `no` collapses
 * everything but the name and the scope, so a declining household is finished in a few
 * taps.
 *
 * Seating need and allergies stay visible for `church_only` as well. That looks like an
 * inconsistency and is not: a wheelchair space is needed in the pew, and an allergy
 * matters wherever somebody eats — the scope gate is about food, not about the person.
 */
export function RSVPMemberCard({
  member,
  draft,
  onChange,
  fieldErrors,
  showMissingAnswer,
}: {
  member: RSVPMember;
  draft: MemberDraft;
  onChange: (change: Partial<MemberDraft>) => void;
  /** The API's field errors for this member, already keyed by `members.<id>.<field>`. */
  fieldErrors: Record<string, string>;
  /**
   * Only true after a submit attempt. Nothing on this card is marked as an error
   * before then: a form that opens red at a household who has not answered yet reads
   * as broken, and it is the opposite of the register the copy is written in.
   */
  showMissingAnswer: boolean;
}) {
  const scopeFieldId = `${memberCardId(member.id)}-attending`;
  const isMissing = showMissingAnswer && draft.attending === null;

  function errorFor(field: keyof MemberDraft): string | undefined {
    return fieldErrors[memberFieldKey(member.id, field)];
  }

  return (
    <Card id={memberCardId(member.id)} className="gap-4 px-4 py-4" data-testid={memberCardId(member.id)}>
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h3 className="text-h3 font-body">{member.name}</h3>
        {draft.attending === null ? (
          <p className={isMissing ? "text-danger text-small" : "text-ink-muted text-small"}>
            {rsvpLabels.memberUnanswered}
          </p>
        ) : null}
      </div>

      <FormField
        id={scopeFieldId}
        label={rsvpLabels.memberScopeLabel(member.name)}
        help={rsvpLabels.memberScopeHelp}
        error={errorFor("attending") ?? (isMissing ? rsvpLabels.memberUnanswered : undefined)}
      >
        <RadioCardGroup
          aria-labelledby={`${scopeFieldId}-label`}
          aria-invalid={isMissing || errorFor("attending") !== undefined}
          value={draft.attending ?? ""}
          onValueChange={(value) => onChange({ attending: value as Attending })}
        >
          {(Object.keys(attendingLabels) as Attending[]).map((scope) => (
            <RadioCard key={scope} value={scope} label={attendingLabels[scope]} />
          ))}
        </RadioCardGroup>
      </FormField>

      {/* The revealed block sits below the control that caused it, so the page does not
          jump under a thumb. */}
      {coversParty(draft.attending) ? (
        <div className="flex flex-col gap-4">
          <FormField
            id={`${scopeFieldId}-meal`}
            label={rsvpLabels.mealChoiceLabel(member.name)}
            help={rsvpLabels.mealChoiceHelp}
            error={errorFor("meal_choice")}
          >
            <RadioCardGroup
              aria-labelledby={`${scopeFieldId}-meal-label`}
              value={draft.meal_choice ?? ""}
              onValueChange={(value) => onChange({ meal_choice: value as MealChoice })}
            >
              {(Object.keys(mealChoiceLabels) as MealChoice[]).map((choice) => (
                <RadioCard key={choice} value={choice} label={mealChoiceLabels[choice]} />
              ))}
            </RadioCardGroup>
          </FormField>

          <FormField
            id={`${scopeFieldId}-portion`}
            label={rsvpLabels.portionLabel(member.name)}
            help={rsvpLabels.portionHelp}
            error={errorFor("portion")}
          >
            <RadioCardGroup
              aria-labelledby={`${scopeFieldId}-portion-label`}
              value={draft.portion}
              onValueChange={(value) => onChange({ portion: value as Portion })}
            >
              {(Object.keys(portionLabels) as Portion[]).map((portion) => (
                <RadioCard
                  key={portion}
                  value={portion}
                  label={portionLabels[portion]}
                  // `none` keeps its inline hint: it disambiguates two options rather
                  // than explaining the field, which is what the popover is for.
                  hint={portionHelpLabels[portion]}
                />
              ))}
            </RadioCardGroup>
          </FormField>

          <FormField
            id={`${scopeFieldId}-snack`}
            label={rsvpLabels.midnightSnackLabel(member.name)}
            help={rsvpLabels.midnightSnackHelp}
            error={errorFor("midnight_snack")}
          >
            <div className="flex min-h-12 items-center gap-3">
              <Checkbox
                id={`${scopeFieldId}-snack`}
                checked={draft.midnight_snack}
                onCheckedChange={(checked) => onChange({ midnight_snack: checked === true })}
              />
              <Label htmlFor={`${scopeFieldId}-snack`} className="text-ink-muted text-small">
                {rsvpLabels.midnightSnackHint}
              </Label>
            </div>
          </FormField>
        </div>
      ) : null}

      {attendsAnything(draft.attending) ? (
        <div className="flex flex-col gap-4">
          <FormField
            id={`${scopeFieldId}-seating`}
            label={rsvpLabels.seatingNeedLabel(member.name)}
            help={rsvpLabels.seatingNeedHelp}
            error={errorFor("seating_need")}
          >
            <RadioCardGroup
              aria-labelledby={`${scopeFieldId}-seating-label`}
              value={draft.seating_need}
              onValueChange={(value) => onChange({ seating_need: value as SeatingNeed })}
            >
              {(Object.keys(seatingNeedLabels) as SeatingNeed[]).map((need) => (
                <RadioCard key={need} value={need} label={seatingNeedLabels[need]} />
              ))}
            </RadioCardGroup>
          </FormField>

          <FormField
            id={`${scopeFieldId}-dietary`}
            label={rsvpLabels.dietaryNoteLabel(member.name)}
            help={rsvpLabels.dietaryNoteHelp}
            error={errorFor("dietary_note")}
          >
            <Input
              id={`${scopeFieldId}-dietary`}
              value={draft.dietary_note}
              maxLength={maxDietaryNoteLength}
              aria-describedby={`${scopeFieldId}-dietary-hint`}
              onChange={(event) => onChange({ dietary_note: event.target.value })}
            />
            <span id={`${scopeFieldId}-dietary-hint`} className="text-ink-muted text-small">
              {rsvpLabels.dietaryNotePlaceholderHint}
            </span>
          </FormField>

          {/* Children only, and `kind` is not editable here: a household that needs to
              change it phones us and we fix it in F5-F02. */}
          {member.kind === "child" ? (
            <FormField
              id={`${scopeFieldId}-age`}
              label={rsvpLabels.ageLabel(member.name)}
              help={rsvpLabels.ageHelp}
              error={errorFor("age")}
            >
              <Input
                id={`${scopeFieldId}-age`}
                type="number"
                inputMode="numeric"
                min={0}
                max={maxChildAge}
                className="max-w-24"
                value={draft.age ?? ""}
                onChange={(event) => onChange({ age: event.target.value === "" ? null : Number(event.target.value) })}
              />
            </FormField>
          ) : null}
        </div>
      ) : null}
    </Card>
  );
}
