import { CircleQuestionMark } from "lucide-react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { formLabels } from "@/lib/labels";

/**
 * The `?` button beside a field label, and the one or two sentences behind it.
 *
 * Every guest-facing field carries one (05-design). Behind a button rather than
 * inline under the field, because "Mitternachtssnack", "Plätze angeboten" and "Alter
 * am Hochzeitstag" all need explaining, and all of them explained at once turns the
 * RSVP form into a wall of grey text that gets skipped — including by the guest who
 * needed one of the sentences.
 *
 * `field` is the field's own label, so the accessible name says which field this
 * explains ("Hilfe zu Mitternachtssnack") instead of eight buttons all called "?".
 * The icon is small and the hit area is not: the button is 48×48 with the icon
 * centred inside it.
 */
export function HelpPopover({ field, children }: { field: string; children: React.ReactNode }) {
  return (
    <Popover>
      <PopoverTrigger
        type="button"
        aria-label={formLabels.helpFor(field)}
        className="text-ink-muted hover:text-primary focus-visible:ring-ring/50 focus-visible:border-ring -m-3 flex size-12 items-center justify-center rounded-full outline-none focus-visible:ring-[3px]"
      >
        <CircleQuestionMark className="size-5" aria-hidden="true" />
      </PopoverTrigger>
      {/* No aria-live: the popover moves focus into itself, so the text is read
          because the reader is there, not because it was announced. */}
      <PopoverContent>{children}</PopoverContent>
    </Popover>
  );
}
