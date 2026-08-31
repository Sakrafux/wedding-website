import * as React from "react";
import { RadioGroup as RadioGroupPrimitive } from "radix-ui";

import { cn } from "@/lib/utils";

/**
 * A radio group rendered as large tappable cards.
 *
 * Cards rather than a `<select>` for every guest-facing choice in the RSVP form: a
 * native select on Android hides three options behind a modal, and every option here
 * is a decision worth seeing at once. Each card is its own 48px-plus target, which is
 * the rule that matters most for this audience — most failures will be mis-taps.
 *
 * Radix supplies the arrow-key roving focus and the single tab stop, so a keyboard
 * user moves through the *group* rather than through every option in the form.
 */
function RadioCardGroup({ className, ...props }: React.ComponentProps<typeof RadioGroupPrimitive.Root>) {
  return (
    <RadioGroupPrimitive.Root
      data-slot="radio-card-group"
      className={cn("flex flex-col gap-2", className)}
      {...props}
    />
  );
}

/**
 * One card. `hint` is the inline explanation under the label — used only where an
 * option needs disambiguating from its neighbour (`portion: none`), never as a
 * substitute for the field's help popover.
 */
function RadioCard({
  value,
  label,
  hint,
  className,
  ...props
}: Omit<React.ComponentProps<typeof RadioGroupPrimitive.Item>, "children"> & {
  label: string;
  hint?: string | null;
}) {
  return (
    <RadioGroupPrimitive.Item
      data-slot="radio-card"
      value={value}
      className={cn(
        "border-line bg-surface hover:border-primary focus-visible:ring-ring/50 focus-visible:border-ring data-[state=checked]:border-primary data-[state=checked]:bg-primary-soft flex min-h-12 w-full cursor-pointer items-start gap-3 rounded-lg border px-4 py-3 text-left outline-none focus-visible:ring-[3px]",
        className,
      )}
      {...props}
    >
      {/* A drawn circle rather than the platform radio: the card carries the state
          visually, and the dot is what keeps "selected" from being colour alone. */}
      <span
        aria-hidden="true"
        className="border-line bg-surface data-[state=checked]:border-primary mt-1 flex size-5 shrink-0 items-center justify-center rounded-full border"
      >
        <RadioGroupPrimitive.Indicator className="bg-primary size-2.5 rounded-full" />
      </span>
      <span className="flex flex-col gap-0.5">
        <span className="font-medium">{label}</span>
        {hint ? <span className="text-ink-muted text-small">{hint}</span> : null}
      </span>
    </RadioGroupPrimitive.Item>
  );
}

export { RadioCard, RadioCardGroup };
