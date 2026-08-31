import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { apiError, ok, stubApi, unauthenticated } from "@/test/api";
import { adminHousehold, adminHouseholdOverview, adminSession } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";

/** The list endpoint's body, which is a wrapper around the rows. */
function householdList(...households: ReturnType<typeof adminHouseholdOverview>[]) {
  return ok({ households });
}

describe("admin household list", () => {
  it("renders a row per household, with the code in printed form", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": householdList(
        adminHouseholdOverview(),
        adminHouseholdOverview({ id: 13, display_name: "Familie Albrecht", code: "DEF567", member_count: 3 }),
      ),
    });

    await renderApp("/admin/haushalte");

    const table = await screen.findByRole("table");
    expect(within(table).getByRole("link", { name: "Familie Müller" })).toBeInTheDocument();
    // Six characters, ungrouped, exactly as printed on the card.
    expect(within(table).getByText("ABC234")).toBeInTheDocument();
    expect(within(table).getByText("DEF567")).toBeInTheDocument();
    expect(screen.getByText("2 Haushalte, davon 0 nie angemeldet")).toBeInTheDocument();
  });

  // The column that answers "did they even see the invitation?". A blank cell reads
  // as "not loaded", which is why the empty case carries a word of its own.
  it("marks a household that has never logged in and dates one that has", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": householdList(
        adminHouseholdOverview({ last_login_at: "2026-11-03T18:22:00Z" }),
        adminHouseholdOverview({ id: 13, display_name: "Familie Albrecht", last_login_at: null }),
      ),
    });

    await renderApp("/admin/haushalte");

    expect(await screen.findByText("03.11.2026", { exact: false })).toBeInTheDocument();
    expect(screen.getByText("Nie angemeldet")).toBeInTheDocument();
    expect(screen.getByText("2 Haushalte, davon 1 nie angemeldet")).toBeInTheDocument();
  });

  it("narrows the rows by name, case-insensitively", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": householdList(
        adminHouseholdOverview(),
        adminHouseholdOverview({ id: 13, display_name: "Familie Albrecht" }),
      ),
    });

    const { user } = await renderApp("/admin/haushalte");
    await screen.findByRole("table");

    await user.type(screen.getByLabelText("Haushalt suchen"), "müll");

    expect(screen.getByRole("link", { name: "Familie Müller" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Familie Albrecht" })).not.toBeInTheDocument();
  });

  it("filters down to the households that never logged in", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": householdList(
        adminHouseholdOverview({ last_login_at: "2026-11-03T18:22:00Z" }),
        adminHouseholdOverview({ id: 13, display_name: "Familie Albrecht", last_login_at: null }),
      ),
    });

    const { user } = await renderApp("/admin/haushalte");
    await screen.findByRole("table");

    await user.click(screen.getByLabelText("Nur die, die sich nie angemeldet haben"));

    expect(screen.getByRole("link", { name: "Familie Albrecht" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Familie Müller" })).not.toBeInTheDocument();
  });

  // Creating and then hunting for the row you just created is the flow to avoid.
  it("creates a household and goes straight to its detail page", async () => {
    const api = stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": householdList(),
      "POST /api/admin/households": ok(adminHousehold({ id: 77, display_name: "Familie Neu", members: [] })),
      "GET /api/admin/households/77": ok(adminHousehold({ id: 77, display_name: "Familie Neu", members: [] })),
    });

    const { user, router } = await renderApp("/admin/haushalte");

    await user.type(await screen.findByLabelText("Name des Haushalts"), "Familie Neu");
    await user.click(screen.getByRole("button", { name: "Anlegen" }));

    await waitFor(() => expect(currentPath(router)).toBe("/admin/haushalte/77"));
    expect(api.calls).toContainEqual({
      method: "POST",
      path: "/api/admin/households",
      body: { display_name: "Familie Neu" },
    });
  });

  it("shows the API's German message when creating fails", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": householdList(),
      "POST /api/admin/households": apiError(400, "validation_failed", "Bitte prüfe die markierten Felder."),
    });

    const { user } = await renderApp("/admin/haushalte");

    await user.type(await screen.findByLabelText("Name des Haushalts"), "x");
    await user.click(screen.getByRole("button", { name: "Anlegen" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Bitte prüfe die markierten Felder.");
  });

  // An admin whose eight hours ran out mid-edit must not land on "Gib den Code von
  // deiner Einladungskarte ein" — there is no such card.
  it("sends a 401 from the list to the admin login, not the guest one", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": unauthenticated,
    });

    const { router } = await renderApp("/admin/haushalte");

    await waitFor(() => expect(currentPath(router)).toBe("/admin/login"));
  });

  it("is a real table with an accessible name and column headers", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": householdList(adminHouseholdOverview()),
    });

    await renderApp("/admin/haushalte");

    const table = await screen.findByRole("table", {
      name: "Alle Haushalte mit Code, Personenzahl und Anmeldestatus",
    });
    const headers = within(table).getAllByRole("columnheader");
    expect(headers.map((header) => header.textContent)).toEqual([
      "Haushalt",
      "Code",
      "Personen",
      "Letzte Anmeldung",
      "RSVP",
    ]);
  });
});
