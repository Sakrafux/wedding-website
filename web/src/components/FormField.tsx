import { CircleAlert } from "lucide-react";

import { HelpPopover } from "@/components/HelpPopover";
import { Label } from "@/components/ui/label";

/**
 * One labelled field: its label, its help popover, the control, and its error.
 *
 * Every guest-facing field goes through here, which is what keeps four rules from
 * 05-design in one place rather than repeated per control: the label is visible and
 * never a placeholder, the help button names its field, the error sits under the
 * control in `danger` with an icon, and the control is wired to both by id.
 *
 * `labelledControl` is for a control that is not a single form element — a radio group
 * or a stepper. Those get `aria-labelledby` pointing at this label instead of an
 * `htmlFor` relationship, so the caller renders them inside `children` and passes the
 * ids it was given.
 */
export function FormField({
  id,
  label,
  help,
  error,
  children,
}: {
  id: string;
  label: string;
  /** The help sentences. Required by the design rules for every guest-facing field. */
  help: React.ReactNode;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-1">
        <Label htmlFor={id} id={`${id}-label`}>
          {label}
        </Label>
        <HelpPopover field={label}>{help}</HelpPopover>
      </div>
      {children}
      {error ? <FieldError id={`${id}-error`}>{error}</FieldError> : null}
    </div>
  );
}

/**
 * A rejected input's message.
 *
 * `role="alert"` so it is announced when it appears after a submit, and an icon
 * beside it because colour is never the only signal.
 */
export function FieldError({ id, children }: { id?: string; children: React.ReactNode }) {
  return (
    <p id={id} role="alert" className="text-danger text-small flex items-start gap-2">
      <CircleAlert className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
      <span>{children}</span>
    </p>
  );
}
