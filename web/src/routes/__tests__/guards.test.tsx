import { screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { getJson } from "@/lib/api/client";
import { ok, stubApi, unauthenticated } from "@/test/api";
import { adminSession, bootstrap, rsvpAnswer } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";

/** Marks the household as already confirmed, so a test can reach the content. */
function confirmHousehold(householdId = 12) {
  window.localStorage.setItem("wedding.confirmed-households", JSON.stringify([householdId]));
}

describe("route guards", () => {
  it("sends an unauthenticated visit to a guarded route to the login screen", async () => {
    stubApi({ "GET /api/me": unauthenticated });

    const { router } = await renderApp("/start");

    expect(currentPath(router)).toBe("/");
  });

  // Losing the destination would land every guest on the same page regardless of
  // the link they followed, which gets worse the more pages F2 adds.
  it("carries the intended path through the login and back", async () => {
    const api = stubApi({ "GET /api/me": unauthenticated });
    confirmHousehold();

    const { user, router } = await renderApp("/start");
    expect(router.state.location.search).toMatchObject({ redirect: "/start" });

    api.set("POST /api/auth/login", () => {
      api.set("GET /api/me", ok(bootstrap()));
      return ok(bootstrap());
    });

    await user.type(screen.getByLabelText(/Code/), "ABC234");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));

    await waitFor(() => expect(currentPath(router)).toBe("/start"));
  });

  // The value comes out of the URL, so it is attacker-controlled: bouncing a guest
  // off-site straight after they logged in is the shape of every phishing hand-off.
  it("refuses to redirect anywhere but this site", async () => {
    const api = stubApi({ "GET /api/me": unauthenticated });
    confirmHousehold();

    const { user, router } = await renderApp("/?redirect=https://example.invalid/steal");

    api.set("POST /api/auth/login", () => {
      api.set("GET /api/me", ok(bootstrap()));
      return ok(bootstrap());
    });

    await user.type(screen.getByLabelText(/Code/), "ABC234");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));

    await waitFor(() => expect(currentPath(router)).toBe("/start"));
  });

  // A year-long session can still be revoked, and the app must notice from
  // whichever request found out rather than showing a white screen.
  it("drops to the login screen when the session is revoked mid-visit", async () => {
    // The guest chrome asks for the RSVP answer to word its "Antwort" entry, so every
    // test that reaches a guest page stubs it (F2-F01).
    const api = stubApi({ "GET /api/me": ok(bootstrap()), "GET /api/rsvp": ok(rsvpAnswer()) });
    confirmHousehold();

    const { router, queryClient } = await renderApp("/start");
    expect(currentPath(router)).toBe("/start");

    // A background query discovers the revocation, exactly as a real one would —
    // the 401 does not come from /api/me, and the app still has to react to it.
    api.set("GET /api/rsvp", unauthenticated);
    await queryClient.fetchQuery({ queryKey: ["probe"], queryFn: () => getJson("/rsvp") }).catch(() => undefined);

    await waitFor(() => expect(currentPath(router)).toBe("/"));
    expect(await screen.findByRole("status")).toHaveTextContent("Du wurdest abgemeldet.");
  });

  // Dropping a guest to the login screen because a train went into a tunnel is a
  // bad trade: they would type their code again for nothing, and it would fail.
  it("shows a retry for a network failure rather than logging anybody out", async () => {
    stubApi({
      "GET /api/me": () => {
        throw new TypeError("Failed to fetch");
      },
    });

    await renderApp("/start");

    expect(await screen.findByRole("heading")).toHaveTextContent("Da ist etwas schiefgegangen");
    expect(screen.getByText(/Keine Verbindung/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Nochmal versuchen" })).toBeInTheDocument();
    expect(screen.queryByLabelText(/Code/)).not.toBeInTheDocument();
  });

  it("keeps a household session out of the admin area", async () => {
    stubApi({
      "GET /api/me": ok(bootstrap()),
      "GET /api/admin/me": unauthenticated,
    });
    confirmHousehold();

    const { router } = await renderApp("/admin");

    // To the admin login, never the guest one: there is no card with an admin
    // password on it.
    expect(currentPath(router)).toBe("/admin/login");
  });

  it("lets an admin session through to the admin area", async () => {
    stubApi({
      "GET /api/me": unauthenticated,
      "GET /api/admin/me": ok(adminSession),
    });

    const { router } = await renderApp("/admin");

    expect(currentPath(router)).toBe("/admin");
  });
});
