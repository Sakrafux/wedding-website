import { Minus, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { formLabels } from "@/lib/labels";

/**
 * A number chosen with a `−` and a `+` button.
 *
 * Not a number input: this audience mis-taps far more often than it mistypes, and the
 * spinner arrows on mobile Safari are a two-pixel target. The buttons are 48×48, and
 * each one's accessible name names the field ("Ein Platz mehr: Plätze gesucht") rather
 * than being a bare "+", which is unreadable out of context.
 *
 * The value is rendered as text with `aria-live="polite"`, so a change is announced
 * without moving focus.
 */
export function Stepper({
  id,
  label,
  value,
  min = 0,
  max,
  onChange,
}: {
  id: string;
  /** The field's label, used only in the buttons' accessible names. */
  label: string;
  value: number;
  min?: number;
  max: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="flex items-center gap-4" id={id}>
      <Button
        type="button"
        variant="outline"
        className="size-12 rounded-full p-0"
        aria-label={formLabels.stepperDecrease(label)}
        disabled={value <= min}
        onClick={() => onChange(Math.max(min, value - 1))}
      >
        <Minus className="size-5" aria-hidden="true" />
      </Button>

      <output aria-live="polite" className="text-h3 min-w-8 text-center tabular-nums">
        {value}
      </output>

      <Button
        type="button"
        variant="outline"
        className="size-12 rounded-full p-0"
        aria-label={formLabels.stepperIncrease(label)}
        disabled={value >= max}
        onClick={() => onChange(Math.min(max, value + 1))}
      >
        <Plus className="size-5" aria-hidden="true" />
      </Button>
    </div>
  );
}
