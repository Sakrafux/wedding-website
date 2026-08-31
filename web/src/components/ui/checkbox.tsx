import * as React from "react";
import { Checkbox as CheckboxPrimitive } from "radix-ui";
import { Check } from "lucide-react";

import { cn } from "@/lib/utils";

/**
 * shadcn/ui checkbox, copied in.
 *
 * The box itself is 24px but sits in a 48px row (see the call sites), because the
 * touch-target rule is about what a thumb can hit and not about what the design draws.
 */
function Checkbox({ className, ...props }: React.ComponentProps<typeof CheckboxPrimitive.Root>) {
  return (
    <CheckboxPrimitive.Root
      data-slot="checkbox"
      className={cn(
        "border-line data-[state=checked]:bg-primary data-[state=checked]:border-primary data-[state=checked]:text-paper focus-visible:ring-ring/50 focus-visible:border-ring peer bg-surface size-6 shrink-0 rounded-md border outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    >
      <CheckboxPrimitive.Indicator data-slot="checkbox-indicator" className="flex items-center justify-center">
        <Check className="size-4" aria-hidden="true" />
      </CheckboxPrimitive.Indicator>
    </CheckboxPrimitive.Root>
  );
}

export { Checkbox };
