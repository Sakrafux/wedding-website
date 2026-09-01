import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { apiError, ok, stubApi } from "@/test/api";
import { bootstrap, rsvpAnswer, rsvpMember } from "@/test/fixtures";
import { renderApp } from "@/test/render";

/**
 * F4-F01 and F4-F02, on the guest form.
 *
 * Every test here branches on `can_add_plus_one` from the response and never on the
 * member list: that the screen refuses to re-derive the rule is the property under
 * test, not an implementation detail of it.
 */

function confirmHousehold(householdId = 12) {
  window.localStorage.setItem("wedding.confirmed-households", JSON.stringify([householdId]));
}

/** The household we invited alone: one seeded member, and the right to add. */
function soloAnswer(overrides: Parameters<typeof rsvpAnswer>[0] = {}) {
  return rsvpAnswer({ members: [rsvpMember()], can_add_plus_one: true, ...overrides });
}

function stubRSVP(answer = soloAnswer(), extra: Record<string, Parameters<typeof stubApi>[0][string]> = {}) {
  confirmHousehold(answer.household.id);

  return stubApi({
    "GET /api/me": ok(bootstrap()),
    "GET /api/rsvp": ok(answer),
    ...extra,
  });
}

async function openForm() {
  const rendered = await renderApp("/zusagen");
  await screen.findByRole("heading", { name: "Sagt uns Bescheid", level: 1 });
  return rendered;
}

function card(name: string) {
  return within(screen.getByRole("heading", { name, level: 3 }).closest("[data-slot=card]") as HTMLElement);
}

const companion = rsvpMember({ id: 44, name: "Isabella Michelbacher", origin: "guest_added" });

describe("adding a plus-one", () => {
  it("offers the trigger and asks for exactly one field", async () => {
    stubRSVP();

    const { user } = await openForm();

    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));

    const sheet = within(await screen.findByRole("dialog"));
    expect(sheet.getByRole("textbox", { name: "Name der Begleitung" })).toBeInTheDocument();
    // No kind, no age, no meal: the server takes a name and nothing else, and a form
    // asking for more would promise something the API refuses.
    expect(sheet.queryByRole("radiogroup")).not.toBeInTheDocument();
    expect(sheet.queryByRole("spinbutton")).not.toBeInTheDocument();
  });

  it("appends the new card, closes the sheet, and switches to the explanation", async () => {
    const api = stubRSVP(soloAnswer(), {
      "POST /api/rsvp/members": ok({ member: companion, can_add_plus_one: false }),
    });

    const { user } = await openForm();

    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));
    await user.type(await screen.findByRole("textbox", { name: "Name der Begleitung" }), "Isabella Michelbacher");
    await user.click(screen.getByRole("button", { name: "Hinzufügen" }));

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(api.calls.at(-1)).toMatchObject({
      method: "POST",
      path: "/api/rsvp/members",
      body: { name: "Isabella Michelbacher" },
    });

    expect(await screen.findByRole("heading", { name: "Isabella Michelbacher", level: 3 })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Begleitung hinzufügen/ })).not.toBeInTheDocument();
    expect(screen.getByText(/Weitere Personen tragen wir gern für euch ein/)).toBeInTheDocument();
    // The asymmetry is real and confusing, so the confirmation names it.
    expect(screen.getByRole("status")).toHaveTextContent(/speichere das Formular/);
  });

  it("keeps the answers already given when the new card arrives", async () => {
    stubRSVP(soloAnswer(), { "POST /api/rsvp/members": ok({ member: companion, can_add_plus_one: false }) });

    const { user } = await openForm();
    await user.click(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ }));

    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));
    await user.type(await screen.findByRole("textbox", { name: "Name der Begleitung" }), "Isabella Michelbacher");
    await user.click(screen.getByRole("button", { name: "Hinzufügen" }));

    await screen.findByRole("heading", { name: "Isabella Michelbacher", level: 3 });
    expect(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ })).toBeChecked();
    // The new person is a new question, marked unanswered until it is answered.
    expect(card("Isabella Michelbacher").getByText("Noch keine Antwort")).toBeInTheDocument();
  });

  it("blocks the submit until the added member has been answered for", async () => {
    stubRSVP(soloAnswer(), { "POST /api/rsvp/members": ok({ member: companion, can_add_plus_one: false }) });

    const { user } = await openForm();
    await user.click(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ }));
    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));
    await user.type(await screen.findByRole("textbox", { name: "Name der Begleitung" }), "Isabella Michelbacher");
    await user.click(screen.getByRole("button", { name: "Hinzufügen" }));
    await screen.findByRole("heading", { name: "Isabella Michelbacher", level: 3 });

    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByText("Isabella Michelbacher — Antwort fehlt")).toBeInTheDocument();
  });

  /**
   * F4-F03, the regression this bug earned: `DialogContent` is a Radix portal, which
   * moves the DOM node but not the React tree, so the sheet's submit used to bubble
   * into the RSVP form and save it. The PUT carried the member list from before the
   * addition and its response overwrote the new member out of the query cache — the
   * unbidden summary, the missing card and the 409 on the second attempt were all this
   * one bug.
   *
   * Asserted as the sequence and not as a unit: a test rendering the sheet on its own
   * would have passed throughout.
   */
  it("adds the companion without saving the form, for a household that has already answered", async () => {
    const answered = soloAnswer({
      household: { ...soloAnswer().household, rsvp_submitted_at: "2026-11-04T10:00:00Z" },
      members: [rsvpMember({ attending: "both", meal_choice: "all" })],
    });
    const api = stubRSVP(answered, { "POST /api/rsvp/members": ok({ member: companion, can_add_plus_one: false }) });

    const { user } = await renderApp("/zusagen");
    // Summary-first for an answered household (F3-F09), so the form is one tap away.
    await user.click(await screen.findByRole("button", { name: "Antwort ändern" }));

    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));
    await user.type(await screen.findByRole("textbox", { name: "Name der Begleitung" }), "Isabella Michelbacher");
    await user.click(screen.getByRole("button", { name: "Hinzufügen" }));

    expect(await screen.findByRole("heading", { name: "Isabella Michelbacher", level: 3 })).toBeInTheDocument();
    expect(api.calls.filter((call) => call.method === "PUT")).toHaveLength(0);
    // Still the form, not the summary: nothing was saved, so nothing was confirmed.
    expect(screen.getByRole("button", { name: "Speichern" })).toBeInTheDocument();
  });

  // Enter inside the sheet is the same event by another route, and it went the same way.
  it("adds the companion on Enter in the name field, and still does not save the form", async () => {
    const api = stubRSVP(soloAnswer(), {
      "POST /api/rsvp/members": ok({ member: companion, can_add_plus_one: false }),
    });

    const { user } = await openForm();
    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));
    await user.type(
      await screen.findByRole("textbox", { name: "Name der Begleitung" }),
      "Isabella Michelbacher{Enter}",
    );

    expect(await screen.findByRole("heading", { name: "Isabella Michelbacher", level: 3 })).toBeInTheDocument();
    expect(api.calls.filter((call) => call.method === "PUT")).toHaveLength(0);
  });

  // Most households see this from the first render. Never a disabled button: it fails
  // contrast and reads as broken, and the answer here is yes.
  it("explains the way to add somebody to every other household", async () => {
    stubRSVP(rsvpAnswer());

    await openForm();

    expect(screen.getByText(/Weitere Personen tragen wir gern für euch ein/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "+43 650 9408100" })).toHaveAttribute("href", "tel:+436509408100");
    expect(screen.queryByRole("button", { name: /Begleitung hinzufügen/ })).not.toBeInTheDocument();
  });

  // Two tabs, one form each: the server has the last word and its sentence is shown.
  it("shows the API sentence when the addition is refused", async () => {
    stubRSVP(soloAnswer(), {
      "POST /api/rsvp/members": apiError(
        409,
        "plus_one_not_allowed",
        "Weitere Personen tragen wir gern für euch ein — ruf uns bitte kurz an: +43 650 9408100.",
      ),
    });

    const { user } = await openForm();
    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));
    await user.type(await screen.findByRole("textbox", { name: "Name der Begleitung" }), "Isabella Michelbacher");
    await user.click(screen.getByRole("button", { name: "Hinzufügen" }));

    const sheet = within(await screen.findByRole("dialog"));
    expect(await sheet.findByText(/ruf uns bitte kurz an/)).toBeInTheDocument();
  });

  it("switches the page to the read-only view when the deadline passed while it was open", async () => {
    stubRSVP(soloAnswer(), {
      "POST /api/rsvp/members": apiError(
        409,
        "rsvp_closed",
        "Die Rückmeldefrist ist vorbei. Wenn sich etwas geändert hat, ruf uns bitte kurz an.",
      ),
    });

    const { user } = await openForm();
    await user.click(screen.getByRole("button", { name: /Begleitung hinzufügen/ }));
    await user.type(await screen.findByRole("textbox", { name: "Name der Begleitung" }), "Isabella Michelbacher");
    await user.click(screen.getByRole("button", { name: "Hinzufügen" }));

    expect(await screen.findByText("Die Rückmeldefrist ist vorbei")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Speichern" })).not.toBeInTheDocument();
  });

  it("traps focus in the sheet, closes on Escape, and gives the trigger its focus back", async () => {
    stubRSVP();

    const { user } = await openForm();
    const trigger = screen.getByRole("button", { name: /Begleitung hinzufügen/ });
    await user.click(trigger);

    const sheet = await screen.findByRole("dialog");
    // Focus moves into the sheet rather than staying on the page behind it, which is
    // what makes Escape and the close button reachable from a keyboard at all.
    await waitFor(() => expect(sheet.contains(document.activeElement)).toBe(true));

    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(screen.getByRole("button", { name: /Begleitung hinzufügen/ })).toHaveFocus();
  });
});

describe("removing a member", () => {
  /** A household that used its plus-one: one seeded member and one they added. */
  function withCompanion() {
    return rsvpAnswer({ members: [rsvpMember(), companion], can_add_plus_one: false });
  }

  it("offers the control on an added member and on nobody else", async () => {
    stubRSVP(withCompanion());

    await openForm();

    expect(card("Isabella Michelbacher").getByRole("button", { name: "Isabella Michelbacher entfernen" }));
    // No disabled button and no explanation nobody asked for: a seeded member's
    // remedy is the scope control that is already on their card.
    expect(card("Anna Müller").queryByRole("button", { name: /entfernen/ })).not.toBeInTheDocument();
  });

  it("confirms by name, then removes and brings the add trigger back", async () => {
    const api = stubRSVP(withCompanion(), { "DELETE /api/rsvp/members/44": ok() });

    const { user } = await openForm();
    await user.click(card("Isabella Michelbacher").getByRole("button", { name: /entfernen/ }));

    const confirmation = within(await screen.findByRole("alertdialog"));
    expect(confirmation.getByText(/Isabella Michelbacher wird von eurer Liste genommen/)).toBeInTheDocument();

    // The refetch answers with the household as it now is, plus_one restored.
    api.set("GET /api/rsvp", ok(soloAnswer()));
    await user.click(confirmation.getByRole("button", { name: "Ja, entfernen" }));

    await waitFor(() =>
      expect(screen.queryByRole("heading", { name: "Isabella Michelbacher", level: 3 })).not.toBeInTheDocument(),
    );
    expect(api.calls.map((call) => `${call.method} ${call.path}`)).toContain("DELETE /api/rsvp/members/44");
    expect(await screen.findByRole("button", { name: /Begleitung hinzufügen/ })).toBeInTheDocument();
  });

  it("leaves the card in place and shows the API sentence when the removal fails", async () => {
    stubRSVP(withCompanion(), {
      "DELETE /api/rsvp/members/44": apiError(
        409,
        "cannot_remove_member",
        "Diese Person haben wir eingetragen. Wenn sie nicht kommt, wähl bitte «Kommt nicht» aus.",
      ),
    });

    const { user } = await openForm();
    await user.click(card("Isabella Michelbacher").getByRole("button", { name: /entfernen/ }));
    await user.click(within(await screen.findByRole("alertdialog")).getByRole("button", { name: "Ja, entfernen" }));

    expect(await screen.findByText(/Diese Person haben wir eingetragen/)).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Isabella Michelbacher", level: 3 })).toBeInTheDocument();
  });

  it("renders no remove control after the deadline", async () => {
    stubRSVP(rsvpAnswer({ members: [rsvpMember(), companion], editable: false }));

    await openForm();

    expect(screen.getByText("Die Rückmeldefrist ist vorbei")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /entfernen/ })).not.toBeInTheDocument();
  });
});
