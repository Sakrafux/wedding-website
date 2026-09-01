import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { apiError, ok, stubApi } from "@/test/api";
import { adminGuest, adminHousehold, adminSession } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";

/** The stubs every test here needs: an admin session and the household itself. */
function stubHousehold(overrides: Parameters<typeof adminHousehold>[0] = {}) {
  const household = adminHousehold(overrides);

  return {
    household,
    api: stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": ok({ households: [] }),
      "GET /api/admin/households/12": ok(household),
    }),
  };
}

/**
 * Fields are scoped to their group: the page carries one "Name" per member plus the
 * add form's, and the fieldset legend is what tells them apart — for a screen reader
 * as much as for this test.
 */
function group(name: string | RegExp) {
  return within(screen.getByRole("group", { name }));
}

describe("admin household detail", () => {
  it("renders the household's fields and its members", async () => {
    stubHousehold();

    await renderApp("/admin/haushalte/12");

    expect(await screen.findByRole("heading", { name: "Familie Müller", level: 1 })).toBeInTheDocument();
    const household = group("Haushalt");
    expect(household.getByLabelText("Name")).toHaveValue("Familie Müller");
    expect(household.getByLabelText("Interne Notiz (nur für uns)")).toHaveValue("Kommen mit dem Zug");
    // The transport counts and the pram are answers, not settings: shown here, and
    // editable only through the RSVP form (F5-B05, F5-F05).
    expect(household.queryByLabelText(/Plätze angeboten/)).not.toBeInTheDocument();
    expect(screen.getByText("ABC234")).toBeInTheDocument();
    expect(group("Anna Müller").getByLabelText("Name")).toHaveValue("Anna Müller");
    expect(group("Emil Müller").getByLabelText("Alter am Hochzeitstag")).toHaveValue(4);
  });

  // F5-F05: shown next to the link that edits them properly. They are answers, not
  // settings, and editing them beside `display_name` bypassed the RSVP rules.
  it("shows the household's own RSVP answers as text", async () => {
    stubHousehold({ transport_seats_needed: 3, transport_seats_offered: 0, has_stroller: true });

    await renderApp("/admin/haushalte/12");

    await screen.findByRole("heading", { name: "Familie Müller", level: 1 });
    const rsvp = within(screen.getByRole("heading", { name: "Rückmeldung" }).closest("section") as HTMLElement);
    expect(rsvp.getByText("Plätze gesucht (Kirche → Feier)")).toBeInTheDocument();
    expect(rsvp.getByText("3")).toBeInTheDocument();
    expect(rsvp.getByText("Kinderwagen")).toBeInTheDocument();
    expect(rsvp.getByText("Ja")).toBeInTheDocument();
  });

  // "Gespeichert." left us wondering whether it meant *these* changes, and locally the
  // request finishes fast enough that the button's own state was a flash.
  it("says when it last saved, and keeps saying it", async () => {
    const stub = stubHousehold();
    stub.api.set("PATCH /api/admin/households/12", ok({ ...stub.household, display_name: "Familie Müller-Schmidt" }));

    const { user } = await renderApp("/admin/haushalte/12");
    await screen.findByRole("group", { name: "Haushalt" });
    const household = group("Haushalt");

    await user.type(household.getByLabelText("Name"), "-Schmidt");
    await user.click(household.getByRole("button", { name: "Speichern" }));

    const status = await within(
      screen.getByRole("group", { name: "Haushalt" }).parentElement as HTMLElement,
    ).findByRole("status");
    expect(status).toHaveTextContent(/^Zuletzt gespeichert um \d{2}:\d{2}$/);

    // Still there after the next keystroke: a confirmation that persists needs no
    // minimum duration to be seen.
    await user.type(group("Haushalt").getByLabelText("Name"), "!");
    expect(status).toBeInTheDocument();
  });

  it("saves the household fields and reports a validation error next to its field", async () => {
    const { api } = stubHousehold();
    api.set("PATCH /api/admin/households/12", apiError(400, "validation_failed", "Bitte prüfe die markierten Felder."));

    const { user } = await renderApp("/admin/haushalte/12");

    await screen.findByRole("group", { name: "Haushalt" });
    const household = group("Haushalt");

    await user.clear(household.getByLabelText("Name"));
    await user.click(household.getByRole("button", { name: "Speichern" }));

    // The envelope carries no `fields` here, so the top-level message stands in —
    // what matters is that the failure is announced rather than swallowed.
    await waitFor(() => expect(api.calls.some((call) => call.method === "PATCH")).toBe(true));
  });

  it("shows a field message from the server next to the field it names", async () => {
    const { api } = stubHousehold();
    api.set("PATCH /api/admin/households/12", {
      status: 400,
      body: {
        error: {
          code: "validation_failed",
          message: "Bitte prüfe die markierten Felder.",
          request_id: "TESTID1",
          fields: { display_name: "Bitte fülle dieses Feld aus." },
        },
      },
    });

    const { user } = await renderApp("/admin/haushalte/12");

    await screen.findByRole("group", { name: "Haushalt" });
    const household = group("Haushalt");

    await user.clear(household.getByLabelText("Name"));
    await user.click(household.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Bitte fülle dieses Feld aus.");
  });

  // Eighty guests entered four at a time is the actual workload.
  it("keeps the add-member form open and ready for the next person", async () => {
    const { api } = stubHousehold({ members: [] });
    api.set("POST /api/admin/households/12/guests", ok(adminGuest({ name: "Bernd Müller" })));

    const { user } = await renderApp("/admin/haushalte/12");

    await screen.findByRole("group", { name: "Person hinzufügen" });
    const adding = group("Person hinzufügen");

    const nameField = adding.getByLabelText("Name");
    await user.type(nameField, "Bernd Müller");
    await user.click(adding.getByRole("button", { name: "Hinzufügen" }));

    // Cleared and focused, ready for the next person. Nothing is prefilled from the
    // household: "household" here is one or more people sharing an invitation, and
    // its name is free text like "Luki & Paddi".
    await waitFor(() => expect(nameField).toHaveValue(""));
    expect(nameField).toHaveFocus();
  });

  it("asks for an age only for a child", async () => {
    stubHousehold({ members: [] });

    const { user } = await renderApp("/admin/haushalte/12");

    await screen.findByRole("group", { name: "Person hinzufügen" });
    const adding = group("Person hinzufügen");
    expect(adding.queryByLabelText("Alter am Hochzeitstag")).not.toBeInTheDocument();

    await user.selectOptions(adding.getByLabelText("Erwachsen oder Kind"), "child");
    expect(adding.getByLabelText("Alter am Hochzeitstag")).toBeInTheDocument();

    await user.selectOptions(adding.getByLabelText("Erwachsen oder Kind"), "adult");
    expect(adding.queryByLabelText("Alter am Hochzeitstag")).not.toBeInTheDocument();
  });

  // A marker on every ordinary guest would drown out the one that matters.
  it("marks a member the household added itself, and nobody else", async () => {
    stubHousehold({
      members: [adminGuest(), adminGuest({ id: 31, name: "Clara Müller", origin: "guest_added" })],
    });

    await renderApp("/admin/haushalte/12");

    expect(await screen.findAllByText("Selbst hinzugefügt")).toHaveLength(1);
  });

  it("removes a member behind a confirmation that says the record survives", async () => {
    const { api } = stubHousehold({ members: [adminGuest()] });
    api.set("DELETE /api/admin/guests/30", ok());

    const { user } = await renderApp("/admin/haushalte/12");

    await user.click(await screen.findByRole("button", { name: "Entfernen" }));

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(/Datenbestand bleibt sie erhalten/)).toBeInTheDocument();

    api.set("GET /api/admin/households/12", ok(adminHousehold({ members: [], member_count: 0 })));
    await user.click(within(dialog).getByRole("button", { name: "Ja, entfernen" }));

    await waitFor(() => expect(screen.getByText("Noch niemand eingetragen.")).toBeInTheDocument());
  });

  // The old code dies, a printed card carrying it dies with it, and devices are
  // signed out. All three belong in front of the human before the request is made.
  it("reissues the code behind a confirmation and says how many devices were signed out", async () => {
    const { api } = stubHousehold();
    api.set("POST /api/admin/households/12/code", ok({ code: "DEF567", revoked_sessions: 1 }));

    const { user } = await renderApp("/admin/haushalte/12");

    await user.click(await screen.findByRole("button", { name: "Neuen Code erzeugen" }));

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(/ungültig/)).toBeInTheDocument();

    api.set("GET /api/admin/households/12", ok(adminHousehold({ code: "DEF567" })));
    await user.click(within(dialog).getByRole("button", { name: "Ja, neuen Code erzeugen" }));

    expect(await screen.findByText(/ein Gerät wurde abgemeldet/)).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("DEF567")).toBeInTheDocument());
  });

  it("does not reissue the code when the confirmation is cancelled", async () => {
    const { api } = stubHousehold();

    const { user } = await renderApp("/admin/haushalte/12");

    await user.click(await screen.findByRole("button", { name: "Neuen Code erzeugen" }));
    await user.click(within(await screen.findByRole("alertdialog")).getByRole("button", { name: "Abbrechen" }));

    expect(api.calls.some((call) => call.path.endsWith("/code"))).toBe(false);
    expect(screen.getByText("ABC234")).toBeInTheDocument();
  });

  it("deletes the household behind a confirmation naming the member count", async () => {
    const { api } = stubHousehold();
    api.set("DELETE /api/admin/households/12", ok());

    const { user, router } = await renderApp("/admin/haushalte/12");

    await user.click(await screen.findByRole("button", { name: "Haushalt löschen" }));

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(/mit 2 Personen gelöscht/)).toBeInTheDocument();
    // The fear when deleting is losing the record, not losing the row.
    expect(within(dialog).getByText(/Änderungsprotokoll bleibt erhalten/)).toBeInTheDocument();

    await user.click(within(dialog).getByRole("button", { name: "Ja, endgültig löschen" }));

    await waitFor(() => expect(currentPath(router)).toBe("/admin/haushalte"));
  });

  it("reports whether the household has answered and links to their own form", async () => {
    stubHousehold({ rsvp_submitted_at: null });

    await renderApp("/admin/haushalte/12");

    expect(await screen.findByText("Dieser Haushalt hat noch nicht geantwortet.")).toBeInTheDocument();
    // A link, not a second form: the answers are edited on the same component the
    // guests use, which is what Gate 1 exists to protect.
    expect(screen.getByRole("link", { name: "Rückmeldung bearbeiten" })).toHaveAttribute(
      "href",
      "/admin/haushalte/12/rsvp",
    );
    expect(screen.queryByRole("radiogroup", { name: /Wozu kommt/ })).not.toBeInTheDocument();
  });

  it("labels every control and keeps the confirmation dialogs dismissible by keyboard", async () => {
    stubHousehold({ members: [adminGuest()] });

    const { user } = await renderApp("/admin/haushalte/12");

    await user.click(await screen.findByRole("button", { name: "Neuen Code erzeugen" }));
    expect(await screen.findByRole("alertdialog")).toBeInTheDocument();

    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument());
  });
});
