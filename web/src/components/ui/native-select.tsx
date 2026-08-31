import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * A plain `<select>`, styled like Input.
 *
 * Not shadcn's `Select`, and not by accident: the native control gets the platform's
 * own picker — the iOS wheel, Android's list — which is both familiar and reliably
 * accessible. shadcn's version is a custom listbox, which is the right choice when a
 * design demands it and the wrong one for an admin form with four fixed options.
 */
function NativeSelect({ className, ...props }: React.ComponentProps<"select">) {
  return (
    <select
      data-slot="native-select"
      className={cn(
        "border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:border-destructive h-9 w-full rounded-lg border bg-transparent px-3 text-base shadow-xs outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        className,
      )}
      {...props}
    />
  );
}

export { NativeSelect };
