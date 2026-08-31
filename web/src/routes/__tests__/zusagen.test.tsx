import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { apiError, ok, stubApi } from "@/test/api";
import { bootstrap, rsvpAnswer, rsvpMember } from "@/test/fixtures";
import { renderApp } from "@/test/render";

/** The confirmation screen is F1's; a test about the RSVP form starts past it. */
function confirmHousehold(householdId = 12) {
  window.localStorage.setItem("wedding.confirmed-households", JSON.stringify([householdId]));
}

function stubRSVP(overrides: Parameters<typeof rsvpAnswer>[0] = {}) {
  const answer = rsvpAnswer(overrides);
  confirmHousehold(answer.household.id);

  return {
    answer,
    api: stubApi({
      "GET /api/me": ok(bootstrap()),
      "GET /api/rsvp": ok(answer),
    }),
  };
}

/** One member's card, so an assertion about Anna cannot pass because of Emil. */
function card(name: string) {
  return within(screen.getByRole("heading", { name, level: 3 }).closest("[data-slot=card]") as HTMLElement);
}

async function openForm() {
  const rendered = await renderApp("/zusagen");
  await screen.findByRole("heading", { name: "Sagt uns Bescheid", level: 1 });
  return rendered;
}

describe("the RSVP form", () => {
  it("renders every member from the response, in order, with no answer marked as an error", async () => {
    stubRSVP();

    await openForm();

    const names = screen.getAllByRole("heading", { level: 3 }).map((heading) => heading.textContent);
    expect(names).toEqual(["Anna Müller", "Emil Müller"]);

    // A form that opens red at a household who has not answered yet reads as broken.
    // The unanswered marker is a statement of fact until a submit is attempted.
    expect(screen.getAllByText("Noch keine Antwort")).toHaveLength(2);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("sets every member from the household selector in one tap", async () => {
    stubRSVP();

    const { user } = await openForm();

    await user.click(within(screen.getByRole("radiogroup", { name: /Wir kommen zu/ })).getByText("Kirche und Feier"));

    for (const name of ["Anna Müller", "Emil Müller"]) {
      expect(card(name).getByRole("radio", { name: /Kirche und Feier/ })).toBeChecked();
    }
  });

  // A selector that stayed lit while the cards disagreed would be a lie about what
  // gets saved.
  it("clears the household selection once one member differs", async () => {
    stubRSVP();

    const { user } = await openForm();
    const selector = within(screen.getByRole("radiogroup", { name: /Wir kommen zu/ }));

    await user.click(selector.getByText("Kirche und Feier"));
    await user.click(card("Anna Müller").getByRole("radio", { name: /Nur zur Kirche/ }));

    expect(selector.getByRole("radio", { name: /Kirche und Feier/ })).not.toBeChecked();
    expect(screen.getByText(/unterschiedliche Antworten/)).toBeInTheDocument();
  });

  // Silent bulk overwrite is how a household loses Oma's church-only answer.
  it("asks before overwriting answers from the household selector, and names the count", async () => {
    stubRSVP();

    const { user } = await openForm();
    const selector = within(screen.getByRole("radiogroup", { name: /Wir kommen zu/ }));

    await user.click(card("Anna Müller").getByRole("radio", { name: /Nur zur Kirche/ }));
    await user.click(selector.getByText("Kirche und Feier"));

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(/2 Personen/)).toBeInTheDocument();

    await user.click(within(dialog).getByRole("button", { name: "Ja, für alle setzen" }));

    expect(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ })).toBeChecked();
  });

  // One member, one scope control: the bulk selector would be the same question asked
  // twice on one screen.
  it("hides the household selector for a household of one", async () => {
    stubRSVP({ members: [rsvpMember()] });

    await openForm();

    expect(screen.queryByRole("radiogroup", { name: /Wir kommen zu/ })).not.toBeInTheDocument();
    expect(screen.getByRole("radiogroup", { name: /Wozu kommt Anna Müller/ })).toBeInTheDocument();
  });

  it("reveals the catering fields only for a scope that covers the party", async () => {
    stubRSVP({ members: [rsvpMember()] });

    const { user } = await openForm();
    const anna = card("Anna Müller");

    expect(anna.queryByRole("radiogroup", { name: /Was isst Anna Müller/ })).not.toBeInTheDocument();

    await user.click(anna.getByRole("radio", { name: /Kirche und Feier/ }));
    expect(anna.getByRole("radiogroup", { name: /Was isst Anna Müller/ })).toBeInTheDocument();
    expect(anna.getByRole("radiogroup", { name: /Portion für Anna Müller/ })).toBeInTheDocument();
    expect(anna.getByRole("checkbox", { name: /Mitternachtssnack/ })).toBeInTheDocument();

    // Not disabled — absent. A disabled control fails contrast and reads as broken.
    await user.click(anna.getByRole("radio", { name: /Nur zur Kirche/ }));
    expect(anna.queryByRole("radiogroup", { name: /Was isst Anna Müller/ })).not.toBeInTheDocument();

    // The one place the scope gate deliberately does not apply.
    expect(anna.getByRole("radiogroup", { name: /Platz für Anna Müller/ })).toBeInTheDocument();
    expect(anna.getByRole("textbox", { name: /Allergien oder Unverträglichkeiten/ })).toBeInTheDocument();

    await user.click(anna.getByRole("radio", { name: "Kommt nicht" }));
    expect(anna.queryByRole("radiogroup", { name: /Platz für Anna Müller/ })).not.toBeInTheDocument();
    expect(anna.queryByRole("textbox", { name: /Allergien oder Unverträglichkeiten/ })).not.toBeInTheDocument();
  });

  it("asks a child's age as the age on the wedding day, and only for children", async () => {
    stubRSVP();

    const { user } = await openForm();

    await user.click(card("Emil Müller").getByRole("radio", { name: /Kirche und Feier/ }));
    expect(
      card("Emil Müller").getByRole("spinbutton", { name: /Alter von Emil Müller am Hochzeitstag, 17\. Juli 2027/ }),
    ).toHaveValue(4);

    await user.click(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ }));
    expect(card("Anna Müller").queryByRole("spinbutton", { name: /Alter von/ })).not.toBeInTheDocument();
  });

  it("shows the transport section only when somebody attends both halves of the day", async () => {
    stubRSVP({ members: [rsvpMember()] });

    const { user } = await openForm();
    const anna = card("Anna Müller");

    await user.click(anna.getByRole("radio", { name: /Nur zur Feier/ }));
    expect(screen.queryByText("Plätze gesucht")).not.toBeInTheDocument();

    await user.click(anna.getByRole("radio", { name: /Kirche und Feier/ }));
    expect(screen.getByText("Plätze gesucht")).toBeInTheDocument();
  });

  // A number that vanishes without explanation is exactly the thing that gets phoned
  // in about.
  it("explains that the transport answers no longer apply when the section disappears", async () => {
    stubRSVP({
      members: [rsvpMember({ attending: "both" })],
      household: { ...rsvpAnswer().household, transport_seats_needed: 2 },
    });

    const { user } = await openForm();

    await user.click(card("Anna Müller").getByRole("radio", { name: /Nur zur Kirche/ }));

    expect(screen.getByText(/Angaben zur Fahrt gelten nicht mehr/)).toBeInTheDocument();
  });

  it("steps the transport seats without going below zero or above the cap", async () => {
    stubRSVP({ members: [rsvpMember({ attending: "both" })] });

    const { user } = await openForm();

    const decrease = screen.getByRole("button", { name: "Einen weniger: Plätze gesucht" });
    const increase = screen.getByRole("button", { name: "Einen mehr: Plätze gesucht" });
    expect(decrease).toBeDisabled();

    await user.click(increase);
    expect(screen.getByRole("button", { name: "Einen weniger: Plätze gesucht" })).toBeEnabled();
  });

  it("names the field in every help button and closes the popover with Escape", async () => {
    stubRSVP({ members: [rsvpMember()] });

    const { user } = await openForm();

    const help = screen.getByRole("button", { name: "Hilfe zu Wozu kommt Anna Müller?" });
    await user.click(help);

    expect(await screen.findByText(/Kirche und Feier sind getrennt/)).toBeInTheDocument();

    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByText(/Kirche und Feier sind getrennt/)).not.toBeInTheDocument());
    expect(help).toHaveFocus();
  });

  it("submits exactly the form state and shows the saved answer instead of a toast", async () => {
    const { api } = stubRSVP({ members: [rsvpMember()] });
    const saved = rsvpAnswer({
      members: [rsvpMember({ attending: "both", meal_choice: "vegan", midnight_snack: true })],
      household: {
        ...rsvpAnswer().household,
        rsvp_submitted_at: "2026-11-09T19:04:00Z",
        rsvp_updated_at: "2026-11-09T19:04:00Z",
      },
    });
    api.set("PUT /api/rsvp", ok(saved));

    const { user } = await openForm();
    const anna = card("Anna Müller");

    await user.click(anna.getByRole("radio", { name: /Kirche und Feier/ }));
    await user.click(anna.getByRole("radio", { name: "Vegan" }));
    await user.click(anna.getByLabelText(/Kleine Stärkung/));
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    const confirmation = await screen.findByRole("heading", { name: "Danke, wir haben es notiert" });
    expect(confirmation).toBeInTheDocument();
    // Focus moves to the confirmation, so a screen-reader user is told the answer was
    // saved rather than being left at the bottom of a form that has disappeared.
    expect(confirmation).toHaveFocus();

    const request = api.calls.find((call) => call.method === "PUT")?.body as Record<string, unknown>;
    expect(request.members).toEqual([
      {
        id: 30,
        attending: "both",
        meal_choice: "vegan",
        portion: "full",
        midnight_snack: true,
        seating_need: "normal",
        dietary_note: "",
        age: null,
      },
    ]);

    // The recap is of the response, not of the form state, so a value the server
    // normalized away is visibly absent.
    expect(screen.getByText("Vegan")).toBeInTheDocument();
  });

  it("recaps what the server stored rather than what was typed", async () => {
    const { api } = stubRSVP({ members: [rsvpMember()] });
    api.set(
      "PUT /api/rsvp",
      // The scope excludes the party, so the server cleared the meal choice.
      ok(rsvpAnswer({ members: [rsvpMember({ attending: "church_only", meal_choice: null })] })),
    );

    const { user } = await openForm();

    await user.click(card("Anna Müller").getByRole("radio", { name: /Nur zur Feier/ }));
    await user.click(card("Anna Müller").getByRole("radio", { name: "Vegan" }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    await screen.findByRole("heading", { name: "Danke, wir haben es notiert" });
    expect(screen.getByText("Nur zur Kirche")).toBeInTheDocument();
    expect(screen.queryByText("Vegan")).not.toBeInTheDocument();
  });

  it("blocks a submit with a missing answer and links to the member it belongs to", async () => {
    const { api } = stubRSVP();

    const { user } = await openForm();

    await user.click(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(screen.getByText("Bitte antworte noch für diese Personen:")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Emil Müller — Antwort fehlt" })).toBeInTheDocument();
    expect(api.calls.some((call) => call.method === "PUT")).toBe(false);
  });

  it("renders a field error from the API on the right member's control", async () => {
    const { api } = stubRSVP();
    api.set("PUT /api/rsvp", {
      status: 400,
      body: {
        error: {
          code: "validation_failed",
          message: "Bitte prüfe die markierten Felder.",
          request_id: "TESTID1",
          fields: { "members.31.age": "Bitte gib ein Alter zwischen 0 und 17 Jahren an." },
        },
      },
    });

    const { user } = await openForm();

    await user.click(within(screen.getByRole("radiogroup", { name: /Wir kommen zu/ })).getByText("Kirche und Feier"));
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByText("Bitte gib ein Alter zwischen 0 und 17 Jahren an.")).toBeInTheDocument();
    expect(card("Emil Müller").getByText("Bitte gib ein Alter zwischen 0 und 17 Jahren an.")).toBeInTheDocument();
  });

  it("offers a reload when the member list changed under the form", async () => {
    const { api } = stubRSVP({ members: [rsvpMember()] });
    api.set(
      "PUT /api/rsvp",
      apiError(409, "member_set_mismatch", "Die Liste der Personen hat sich geändert. Bitte lade die Seite neu."),
    );

    const { user } = await openForm();

    await user.click(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByText(/Liste der Personen hat sich geändert/)).toBeInTheDocument();

    // Merging state would be the option that can silently drop somebody's answer.
    api.set("GET /api/rsvp", ok(rsvpAnswer({ members: [rsvpMember(), rsvpMember({ id: 32, name: "Neue Person" })] })));
    await user.click(screen.getByRole("button", { name: "Neu laden" }));

    expect(await screen.findByRole("heading", { name: "Neue Person", level: 3 })).toBeInTheDocument();
  });

  it("switches to the read-only view when the deadline passed while the form was open", async () => {
    const { api } = stubRSVP({ members: [rsvpMember({ attending: "both" })] });
    api.set(
      "PUT /api/rsvp",
      apiError(
        409,
        "rsvp_closed",
        "Die Rückmeldefrist ist vorbei. Wenn sich etwas geändert hat, ruf uns bitte kurz an.",
      ),
    );

    const { user } = await openForm();
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByText("Die Rückmeldefrist ist vorbei")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Speichern" })).not.toBeInTheDocument();
  });

  // Losing a filled-in form to a dropped connection in a village is the failure this
  // audience will actually hit.
  it("keeps every entered value when the connection drops", async () => {
    const { api } = stubRSVP({ members: [rsvpMember()] });
    api.set("PUT /api/rsvp", () => {
      throw new Error("offline");
    });

    const { user } = await openForm();

    await user.click(card("Anna Müller").getByRole("radio", { name: /Kirche und Feier/ }));
    await user.click(card("Anna Müller").getByRole("radio", { name: "Vegan" }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByText(/Keine Verbindung zum Server/)).toBeInTheDocument();
    expect(card("Anna Müller").getByRole("radio", { name: "Vegan" })).toBeChecked();
  });

  it("renders the saved answer as text after the deadline, with the phone number", async () => {
    stubRSVP({
      editable: false,
      members: [rsvpMember({ attending: "both", meal_choice: "vegetarian" })],
      household: { ...rsvpAnswer().household, rsvp_submitted_at: "2026-11-09T19:04:00Z" },
    });

    await openForm();

    expect(screen.getByText("Die Rückmeldefrist ist vorbei")).toBeInTheDocument();
    expect(screen.getByText("Vegetarisch")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "+43 650 9408100" })).toBeInTheDocument();
    // Text on surface-sunken, never inputs: disabled text fails contrast and reads as
    // broken.
    expect(screen.queryByRole("radio")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Speichern" })).not.toBeInTheDocument();
  });

  it("tells a household that never answered to call us, rather than explaining a closed form", async () => {
    stubRSVP({ editable: false });

    await openForm();

    expect(screen.getByText(/keine Rückmeldung/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "+43 650 9408100" })).toBeInTheDocument();
  });

  // The browser clock is not the authority: a phone with a wrong date must not be shown
  // a live form that every save rejects, nor a closed one it should be able to use.
  it("renders the form whenever the server says it is editable", async () => {
    stubRSVP({ editable: true, deadline: "2020-01-01T00:00:00Z" });

    await openForm();

    expect(screen.getByRole("button", { name: "Speichern" })).toBeInTheDocument();
  });

  it("renders the shared error state when the answer cannot be loaded", async () => {
    confirmHousehold();
    stubApi({
      "GET /api/me": ok(bootstrap()),
      "GET /api/rsvp": apiError(
        500,
        "internal_error",
        "Da ist etwas schiefgegangen. Bitte versuche es später noch einmal.",
      ),
    });

    await renderApp("/zusagen");

    expect(await screen.findByRole("heading", { name: "Da ist etwas schiefgegangen" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Speichern" })).not.toBeInTheDocument();
  });
});
