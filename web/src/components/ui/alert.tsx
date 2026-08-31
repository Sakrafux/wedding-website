import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const alertVariants = cva("flex flex-col gap-1 rounded-xl border p-4 text-body", {
  variants: {
    variant: {
      /** Something to know. `accent` is the attention colour and never an action. */
      info: "border-line bg-surface-sunken text-ink",
      /** Something that went wrong or has closed. Never the only signal — the copy
          says what happened, per the colour rule in 05-design. */
      attention: "border-accent bg-accent-soft text-ink",
    },
  },
  defaultVariants: { variant: "info" },
});

/**
 * A block of text that explains a state.
 *
 * `role="status"` by default so a screen reader announces it when it appears, which
 * is the case it exists for: the deadline notice and the "your answer was saved"
 * confirmation both appear after an action rather than on load.
 */
function Alert({
  className,
  variant,
  role = "status",
  ...props
}: React.ComponentProps<"div"> & VariantProps<typeof alertVariants>) {
  return <div data-slot="alert" role={role} className={cn(alertVariants({ variant, className }))} {...props} />;
}

function AlertTitle({ className, ...props }: React.ComponentProps<"p">) {
  return <p data-slot="alert-title" className={cn("font-medium", className)} {...props} />;
}

function AlertDescription({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div data-slot="alert-description" className={cn("text-ink-muted flex flex-col gap-1", className)} {...props} />
  );
}

export { Alert, AlertDescription, AlertTitle };
