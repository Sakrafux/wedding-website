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
    expect(household.getByLabelText("Plätze angeboten (Kirche → Feier)")).toHaveValue(4);
    expect(screen.getByText("ABC234")).toBeInTheDocument();
    expect(group("Anna Müller").getByLabelText("Name")).toHaveValue("Anna Müller");
    expect(group("Emil Müller").getByLabelText("Alter am Hochzeitstag")).toHaveValue(4);
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

  it("shows a read-only RSVP summary until F3-F06 exists", async () => {
    stubHousehold({ rsvp_submitted_at: null });

    await renderApp("/admin/haushalte/12");

    expect(await screen.findByText("Dieser Haushalt hat noch nicht geantwortet.")).toBeInTheDocument();
    // No second RSVP form here: F3-F06 renders the guests' own form against the
    // admin endpoint, and a second field set is what Gate 1 exists to prevent.
    expect(screen.queryByLabelText("Kommt")).not.toBeInTheDocument();
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
