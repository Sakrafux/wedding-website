import * as React from "react";
import { Dialog as DialogPrimitive } from "radix-ui";
import { X } from "lucide-react";

import { cn } from "@/lib/utils";

/**
 * shadcn/ui dialog, copied in — Radix's `Dialog`, not `AlertDialog`.
 *
 * The distinction is the one alert-dialog.tsx states from the other side: this one is
 * dismissible by clicking beside it, which is right for a form somebody opened by
 * choice and wrong for a destructive confirmation.
 *
 * `DialogContent` is a **sheet on a phone and a dialog on a desktop**: it sits on the
 * bottom edge, full width, up to `sm`, and centres itself above that. One component
 * rather than two, because the behaviour that is hard to get right — focus trapping,
 * Escape, returning focus to the trigger, `aria-modal` — is Radix's either way, and a
 * second component would be a second place for it to be got wrong.
 */
function Dialog(props: React.ComponentProps<typeof DialogPrimitive.Root>) {
  return <DialogPrimitive.Root data-slot="dialog" {...props} />;
}

function DialogTrigger(props: React.ComponentProps<typeof DialogPrimitive.Trigger>) {
  return <DialogPrimitive.Trigger data-slot="dialog-trigger" {...props} />;
}

function DialogOverlay({ className, ...props }: React.ComponentProps<typeof DialogPrimitive.Overlay>) {
  return (
    <DialogPrimitive.Overlay
      data-slot="dialog-overlay"
      className={cn("fixed inset-0 z-50 bg-black/40", className)}
      {...props}
    />
  );
}

function DialogContent({
  className,
  children,
  closeLabel,
  ...props
}: React.ComponentProps<typeof DialogPrimitive.Content> & {
  /** The close button's accessible name. German, from labels.ts, like every other
      piece of visible or announced text. */
  closeLabel: string;
}) {
  return (
    <DialogPrimitive.Portal>
      <DialogOverlay />
      <DialogPrimitive.Content
        data-slot="dialog-content"
        className={cn(
          "bg-surface shadow-card fixed inset-x-0 bottom-0 z-50 flex flex-col gap-4 rounded-t-xl border p-6",
          "sm:top-1/2 sm:bottom-auto sm:left-1/2 sm:w-[calc(100%-2rem)] sm:max-w-md sm:-translate-x-1/2 sm:-translate-y-1/2 sm:rounded-xl",
          className,
        )}
        {...props}
      >
        {children}
        {/* A visible close target as well as Escape: a thumb on a phone has no Escape
            key, and dismissing by tapping the overlay is not discoverable. */}
        <DialogPrimitive.Close
          className="text-ink-muted focus-visible:ring-ring absolute top-4 right-4 flex size-11 items-center justify-center rounded-md focus-visible:ring-2 focus-visible:outline-none"
          aria-label={closeLabel}
        >
          <X className="size-5" aria-hidden="true" />
        </DialogPrimitive.Close>
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  );
}

function DialogTitle({ className, ...props }: React.ComponentProps<typeof DialogPrimitive.Title>) {
  return <DialogPrimitive.Title data-slot="dialog-title" className={cn("text-h3 pr-12", className)} {...props} />;
}

function DialogDescription({ className, ...props }: React.ComponentProps<typeof DialogPrimitive.Description>) {
  return (
    <DialogPrimitive.Description
      data-slot="dialog-description"
      className={cn("text-ink-muted text-small", className)}
      {...props}
    />
  );
}

export { Dialog, DialogTrigger, DialogContent, DialogTitle, DialogDescription };
