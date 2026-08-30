import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { normalizeCode } from "@/lib/code";
import { loginLabels } from "@/lib/labels";

interface CodeInputProps {
  value: string;
  onChange: (code: string) => void;
  /** The German message from the API, shown under the field. Empty when there is none. */
  error?: string;
  disabled?: boolean;
}

/**
 * The login code field: one large input, not six boxes.
 *
 * Segmented inputs look tidy and break everything that matters here — pasting a
 * code out of a message, screen readers announcing one character at a time, and
 * on-screen keyboards that fight the focus jumps between boxes.
 */
export function CodeInput({ value, onChange, error, disabled }: CodeInputProps) {
  const hintId = "code-hint";
  const errorId = "code-error";

  return (
    <div className="flex flex-col gap-2">
      <Label htmlFor="code" className="text-h3">
        {loginLabels.codeLabel}
      </Label>

      <Input
        id="code"
        name="code"
        value={value}
        onChange={(event) => onChange(normalizeCode(event.target.value))}
        disabled={disabled}
        // A code is letters and digits, so the plain text keyboard is right:
        // inputmode="numeric" would hide the letters, and "email" or "url" bring
        // an @ key nobody needs.
        inputMode="text"
        autoCapitalize="characters"
        autoCorrect="off"
        autoComplete="off"
        spellCheck={false}
        // No autofocus: on a phone it opens the keyboard over the page before the
        // guest has read what is being asked of them.
        //
        // No maxLength either, deliberately: the attribute counts raw characters
        // and would truncate a pasted ABC-234 before the dash is stripped. See
        // normalizeCode.
        aria-describedby={error ? `${hintId} ${errorId}` : hintId}
        aria-invalid={error ? true : undefined}
        // Larger than the body scale and widely spaced: this is the one field in
        // the product that is transcribed character by character off a card.
        className="h-14 text-center font-mono text-[1.6rem] tracking-[0.3em]"
      />

      <p id={hintId} className="text-small text-ink-muted">
        {loginLabels.codeHint}
      </p>

      {error ? (
        // role="alert" so the message is announced when it appears, rather than
        // only being found by someone who goes looking for it.
        <p id={errorId} role="alert" className="text-small text-danger flex items-start gap-2">
          <span aria-hidden="true">⚠</span>
          <span>{error}</span>
        </p>
      ) : null}
    </div>
  );
}
