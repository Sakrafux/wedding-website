import { screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { getJson } from "@/lib/api";
import { apiError, ok, stubApi, unauthenticated } from "@/test/api";
import { adminSession, bootstrap } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";

const invalidCredentials = apiError(401, "invalid_credentials", "Anmeldung fehlgeschlagen.");

describe("admin login and shell", () => {
  it("logs in and lands on the shell", async () => {
    const api = stubApi({ "GET /api/admin/me": unauthenticated });
    api.set("POST /api/auth/admin/login", () => {
      api.set("GET /api/admin/me", ok(adminSession));
      return ok(adminSession);
    });

    const { user, router } = await renderApp("/admin/login");

    await user.type(screen.getByLabelText("Benutzername"), "admin");
    await user.type(screen.getByLabelText("Passwort"), "correct horse battery staple");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));

    await waitFor(() => expect(currentPath(router)).toBe("/admin"));
    expect(await screen.findByRole("navigation")).toBeInTheDocument();
  });

  it("shows the API's German error and says nothing about which half was wrong", async () => {
    stubApi({
      "GET /api/admin/me": unauthenticated,
      "POST /api/auth/admin/login": invalidCredentials,
    });

    const { user } = await renderApp("/admin/login");

    await user.type(screen.getByLabelText("Benutzername"), "admin");
    await user.type(screen.getByLabelText("Passwort"), "wrong");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Anmeldung fehlgeschlagen.");
  });

  // This is the one login in the product that belongs in a password manager, and
  // the tokens are the whole of what makes that work.
  it("carries the autocomplete tokens a password manager needs", async () => {
    stubApi({ "GET /api/admin/me": unauthenticated });

    await renderApp("/admin/login");

    expect(screen.getByLabelText("Benutzername")).toHaveAttribute("autocomplete", "username");
    expect(screen.getByLabelText("Passwort")).toHaveAttribute("autocomplete", "current-password");
    expect(screen.getByLabelText("Passwort")).toHaveAttribute("type", "password");
  });

  // Eight hours means running out mid-edit is routine. Landing on "Gib den Code
  // von deiner Einladungskarte ein" would be baffling: there is no such card.
  it("returns to the admin login on expiry, never the guest one", async () => {
    const api = stubApi({ "GET /api/admin/me": ok(adminSession) });

    const { router, queryClient } = await renderApp("/admin");
    expect(currentPath(router)).toBe("/admin");

    api.set("GET /api/admin/households", unauthenticated);
    await queryClient
      .fetchQuery({ queryKey: ["probe"], queryFn: () => getJson("/admin/households") })
      .catch(() => undefined);

    await waitFor(() => expect(currentPath(router)).toBe("/admin/login"));
    expect(await screen.findByRole("status")).toHaveTextContent("Sitzung abgelaufen.");
  });

  it("logs out back to the admin login", async () => {
    const api = stubApi({ "GET /api/admin/me": ok(adminSession) });
    api.set("POST /api/auth/logout", () => {
      api.set("GET /api/admin/me", unauthenticated);
      return ok();
    });

    const { user, router } = await renderApp("/admin");

    await user.click(screen.getByRole("button", { name: "Abmelden" }));

    await waitFor(() => expect(currentPath(router)).toBe("/admin/login"));
    expect(api.calls.some((call) => call.path === "/api/auth/logout")).toBe(true);
  });

  // Better than a 404, and it keeps the remaining work visible.
  it("renders the not-yet-built sections as disabled placeholders", async () => {
    stubApi({ "GET /api/admin/me": ok(adminSession) });

    await renderApp("/admin");

    for (const section of ["Haushalte", "Dashboard", "Sitzplan", "Budget", "Fotos"]) {
      expect(screen.getByText(section)).toBeInTheDocument();
    }
    expect(screen.queryAllByRole("link", { name: /Haushalte/ })).toHaveLength(0);
  });

  it("sends an already-signed-in admin past the login screen", async () => {
    stubApi({ "GET /api/admin/me": ok(adminSession) });

    const { router } = await renderApp("/admin/login");

    expect(currentPath(router)).toBe("/admin");
  });

  // The guest side and the admin side share one cookie but not one identity, and
  // the admin area must not be reachable by holding a household session.
  it("refuses a household session", async () => {
    stubApi({
      "GET /api/me": ok(bootstrap()),
      "GET /api/admin/me": unauthenticated,
    });

    const { router } = await renderApp("/admin");

    expect(currentPath(router)).toBe("/admin/login");
  });
});
