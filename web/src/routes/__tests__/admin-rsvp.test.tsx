import { screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { apiError, ok, stubApi, unauthenticated } from "@/test/api";
import { adminHousehold, adminSession, rsvpAnswer, rsvpMember } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";

function stubAdminRSVP(overrides: Parameters<typeof rsvpAnswer>[0] = {}) {
  const answer = rsvpAnswer(overrides);

  return {
    answer,
    api: stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": ok({ households: [] }),
      "GET /api/admin/households/12": ok(adminHousehold()),
      "GET /api/admin/households/12/rsvp": ok(answer),
    }),
  };
}

function card(name: string) {
  return within(screen.getByRole("heading", { name, level: 3 }).closest("[data-slot=card]") as HTMLElement);
}

describe("the admin RSVP page", () => {
  // An admin with two tabs open must not write Familie Müller's answer into Familie
  // Schmidt, so the household is named on the page and not only in the URL.
  it("names the household it is answering for and renders the guests' own form", async () => {
    stubAdminRSVP();

    await renderApp("/admin/haushalte/12/rsvp");

    expect(
      await screen.findByRole("heading", { name: "Rückmeldung für Familie Müller", level: 1 }),
    ).toBeInTheDocument();
    expect(screen.getByText("Du beantwortest dieses Formular für Familie Müller.")).toBeInTheDocument();
    // The same controls as the guest form, by construction: it is the same component.
    expect(screen.getByRole("radiogroup", { name: /Wozu kommt Anna Müller/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Speichern" })).toBeInTheDocument();
  });

  it("posts to the admin endpoint with the household id from the path", async () => {
    const { api } = stubAdminRSVP({ members: [rsvpMember()] });
    api.set("PUT /api/admin/households/12/rsvp", ok(rsvpAnswer({ members: [rsvpMember({ attending: "both" })] })));

    const { user } = await renderApp("/admin/haushalte/12/rsvp");

    await user.click((await screen.findAllByRole("radio", { name: /Kirche und Feier/ }))[0] as HTMLElement);
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByRole("heading", { name: "Danke, wir haben es notiert" })).toBeInTheDocument();
    expect(api.calls.some((call) => call.method === "PUT" && call.path === "/api/admin/households/12/rsvp")).toBe(true);
  });

  // The page exists for the late phone call: `editable: false` is information here,
  // never a lock — while the guest form renders read-only on the same response.
  it("still renders inputs after the deadline, and says so", async () => {
    stubAdminRSVP({ editable: false, members: [rsvpMember({ attending: "both" })] });

    await renderApp("/admin/haushalte/12/rsvp");

    expect(
      await screen.findByText("Die Frist ist abgelaufen — du kannst hier trotzdem speichern."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Speichern" })).toBeInTheDocument();
    expect(screen.queryByText("Die Rückmeldefrist ist vorbei")).not.toBeInTheDocument();
  });

  it("stays in the admin area for an unknown household", async () => {
    stubApi({
      "GET /api/admin/me": ok(adminSession),
      "GET /api/admin/households": ok({ households: [] }),
      "GET /api/admin/households/4711/rsvp": apiError(404, "not_found", "Diese Adresse gibt es nicht."),
    });

    const { router } = await renderApp("/admin/haushalte/4711/rsvp");

    expect(await screen.findByText("Diesen Haushalt gibt es nicht.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Zurück zur Übersicht" })).toBeInTheDocument();
    expect(currentPath(router)).toBe("/admin/haushalte/4711/rsvp");
  });

  it("sends an expired admin session to the admin login, not the guest one", async () => {
    stubApi({ "GET /api/admin/me": unauthenticated });

    const { router } = await renderApp("/admin/haushalte/12/rsvp");

    await waitFor(() => expect(currentPath(router)).toBe("/admin/login"));
  });

  it("shows a member's field error on that member's card, exactly as the guest form does", async () => {
    const { api } = stubAdminRSVP();
    api.set("PUT /api/admin/households/12/rsvp", {
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

    const { user } = await renderApp("/admin/haushalte/12/rsvp");

    await user.click(
      within(await screen.findByRole("radiogroup", { name: /Wir kommen zu/ })).getByText("Kirche und Feier"),
    );
    await user.click(screen.getByRole("button", { name: "Speichern" }));

    expect(await screen.findByText("Bitte gib ein Alter zwischen 0 und 17 Jahren an.")).toBeInTheDocument();
    expect(card("Emil Müller").getByText("Bitte gib ein Alter zwischen 0 und 17 Jahren an.")).toBeInTheDocument();
  });
});
