import { screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ok, stubApi, unauthenticated } from "@/test/api";
import { bootstrap } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";

describe("household confirmation", () => {
  it("names the household and lists the members", async () => {
    stubApi({ "GET /api/me": ok(bootstrap()) });

    await renderApp("/willkommen");

    expect(screen.getByRole("heading")).toHaveTextContent("Willkommen, Familie Müller — seid ihr das?");
    // The member list is what actually catches a one-character-off code: two
    // households named Müller is plausible, two with the same first names is not.
    expect(screen.getByText("Anna")).toBeInTheDocument();
    expect(screen.getByText("Emil")).toBeInTheDocument();
  });

  it("renders a single-member household without a dangling list", async () => {
    stubApi({
      "GET /api/me": ok(
        bootstrap({ members: [{ id: 30, first_name: "Anna", last_name: "Müller", kind: "adult", origin: "seeded" }] }),
      ),
    });

    await renderApp("/willkommen");

    expect(screen.getAllByRole("listitem")).toHaveLength(1);
  });

  it("continues on 'Ja' and does not ask again", async () => {
    stubApi({ "GET /api/me": ok(bootstrap()) });

    const { user, router } = await renderApp("/willkommen");
    await user.click(screen.getByRole("button", { name: "Ja, das sind wir" }));

    await waitFor(() => expect(currentPath(router)).toBe("/start"));

    // A year-long session must not ask this daily, so the acknowledgement is
    // persisted and a fresh visit goes straight through.
    const returning = await renderApp("/start");
    expect(currentPath(returning.router)).toBe("/start");
  });

  // The bug this screen exists to prevent: a session left behind would walk
  // straight back into the wrong household on the next visit.
  it("logs out server-side on 'Nein'", async () => {
    const api = stubApi({ "GET /api/me": ok(bootstrap()) });
    // Logging out really does end the session, so /api/me answers 401 afterwards.
    // A stub that kept returning the household would let this test pass against an
    // app that navigated away without revoking anything — the exact bug the screen
    // exists to prevent.
    api.set("POST /api/auth/logout", () => {
      api.set("GET /api/me", unauthenticated);
      return ok();
    });

    const { user, router } = await renderApp("/willkommen");
    await user.click(screen.getByRole("button", { name: "Nein" }));

    await waitFor(() => expect(currentPath(router)).toBe("/"));

    expect(api.calls.some((call) => call.method === "POST" && call.path === "/api/auth/logout")).toBe(true);
    expect(await screen.findByRole("status")).toHaveTextContent("Kein Problem");
  });

  // Reaching content without confirming would defeat the screen entirely, so the
  // check lives in the layout guard rather than in a component.
  it("cannot be skipped by deep-linking past it", async () => {
    stubApi({ "GET /api/me": ok(bootstrap()) });

    const { router } = await renderApp("/start");

    expect(currentPath(router)).toBe("/willkommen");
  });

  it("sends an unauthenticated visitor to the login screen", async () => {
    stubApi({ "GET /api/me": unauthenticated });

    const { router } = await renderApp("/willkommen");

    expect(currentPath(router)).toBe("/");
  });
});
