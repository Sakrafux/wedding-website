import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it } from "vitest";

import { normalizeCode } from "@/lib/code";

import { CodeInput } from "./CodeInput";

/** A harness with the state the real screen holds, so typing behaves as it does there. */
function ControlledCodeInput({ error }: { error?: string }) {
  const [code, setCode] = useState("");
  return (
    <>
      <CodeInput value={code} onChange={setCode} error={error} />
      <output data-testid="submitted-value">{code}</output>
    </>
  );
}

describe("normalizeCode", () => {
  it("produces the canonical form for everything a guest might type", () => {
    // Each of these is a real path: the printed dash, a phone capitalising or not,
    // a paste out of a PDF with a non-breaking space, a word processor's en dash.
    expect(normalizeCode("abc 234")).toBe("ABC234");
    expect(normalizeCode("abc-234")).toBe("ABC234");
    expect(normalizeCode("ABC-234")).toBe("ABC234");
    expect(normalizeCode(" abc 234 ")).toBe("ABC234");
    expect(normalizeCode("abc–234")).toBe("ABC234");
  });

  it("caps at the code length, so a seventh character never silently disappears on submit", () => {
    expect(normalizeCode("ABC2345")).toBe("ABC234");
  });
});

describe("CodeInput", () => {
  it("normalises as the guest types", async () => {
    const user = userEvent.setup();
    render(<ControlledCodeInput />);

    await user.type(screen.getByLabelText(/Code/), "abc 234");

    expect(screen.getByTestId("submitted-value")).toHaveTextContent("ABC234");
  });

  it("accepts a pasted printed code", async () => {
    const user = userEvent.setup();
    render(<ControlledCodeInput />);

    await user.click(screen.getByLabelText(/Code/));
    await user.paste("abc-234");

    expect(screen.getByTestId("submitted-value")).toHaveTextContent("ABC234");
  });

  // Autocorrect mangling a six-character code is a real and infuriating failure,
  // and the attributes are the only thing standing between a guest and it.
  it("turns off every keyboard feature that would rewrite a code", () => {
    render(<ControlledCodeInput />);
    const field = screen.getByLabelText(/Code/);

    expect(field).toHaveAttribute("autocapitalize", "characters");
    expect(field).toHaveAttribute("autocorrect", "off");
    expect(field).toHaveAttribute("autocomplete", "off");
    expect(field).toHaveAttribute("spellcheck", "false");
    expect(field).toHaveAttribute("inputmode", "text");
  });

  it("associates the hint with the field, and adds the error when there is one", () => {
    const { rerender } = render(<ControlledCodeInput />);
    const field = screen.getByLabelText(/Code/);

    expect(field).toHaveAccessibleDescription(/ABC-234/);
    expect(field).not.toHaveAttribute("aria-invalid");

    rerender(<ControlledCodeInput error="Diesen Code kennen wir nicht." />);

    expect(screen.getByRole("alert")).toHaveTextContent("Diesen Code kennen wir nicht.");
    expect(screen.getByLabelText(/Code/)).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByLabelText(/Code/)).toHaveAccessibleDescription(/Diesen Code kennen wir nicht/);
  });
});
