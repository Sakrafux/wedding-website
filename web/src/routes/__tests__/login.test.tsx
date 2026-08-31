import { screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { apiError, ok, type StubResponse, stubApi, unauthenticated, unknownCode } from "@/test/api";
import { bootstrap } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";

describe("login screen", () => {
  it("logs a household in and moves on to the confirmation", async () => {
    const api = stubApi({
      "GET /api/me": unauthenticated,
      "POST /api/auth/login": ok(bootstrap()),
    });

    const { user, router } = await renderApp("/");

    await user.type(screen.getByLabelText(/Code/), "abc-234");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));

    await waitFor(() => expect(currentPath(router)).toBe("/willkommen"));

    // The normalised code is what reaches the wire, not what was typed.
    const login = api.calls.find((call) => call.path === "/api/auth/login");
    expect(login?.body).toEqual({ code: "ABC234" });
  });

  it("shows the API's own German message under the field", async () => {
    stubApi({ "GET /api/me": unauthenticated, "POST /api/auth/login": unknownCode });

    const { user } = await renderApp("/");

    await user.type(screen.getByLabelText(/Code/), "ZZZZZZ");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));

    // Verbatim: the server owns the wording so every screen says the same thing
    // and a fix is one deploy rather than two.
    expect(await screen.findByRole("alert")).toHaveTextContent("Diesen Code kennen wir nicht.");
  });

  it("reveals the phone number after the second failure", async () => {
    stubApi({ "GET /api/me": unauthenticated, "POST /api/auth/login": unknownCode });

    const { user } = await renderApp("/");
    const field = screen.getByLabelText(/Code/);
    const submit = screen.getByRole("button", { name: "Anmelden" });

    await user.type(field, "ZZZZZZ");
    await user.click(submit);
    await screen.findByRole("alert");
    expect(screen.queryByText(/Ruf uns an/)).not.toBeInTheDocument();

    await user.clear(field);
    await user.type(field, "YYYYYY");
    await user.click(submit);

    // Two failures is where a person stops assuming they mistyped and starts
    // assuming they are the problem, which is when the way out should appear.
    expect(await screen.findByText(/Ruf uns an/)).toBeInTheDocument();
  });

  // A dropped connection is not the guest failing at anything, so it must not
  // count towards the fallback and must not read as a rejection.
  it("does not count a rate limit as a network problem, or the other way round", async () => {
    stubApi({
      "GET /api/me": unauthenticated,
      "POST /api/auth/login": apiError(429, "rate_limited", "Zu viele Versuche. Bitte warte ein paar Minuten."),
    });

    const { user } = await renderApp("/");

    await user.type(screen.getByLabelText(/Code/), "ZZZZZZ");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Zu viele Versuche.");
  });

  // A silent second press on a slow connection is how a guest submits twice, and
  // the changed label matters as much as the disabled state: a spinner on its own
  // reads as "broken" to somebody already unsure of the app.
  it("disables the button and says so while the request is in flight", async () => {
    let release: (response: StubResponse) => void = () => {};

    stubApi({
      "GET /api/me": unauthenticated,
      "POST /api/auth/login": () =>
        new Promise<StubResponse>((resolve) => {
          release = resolve;
        }),
    });

    const { user } = await renderApp("/");
    await user.type(screen.getByLabelText(/Code/), "ABC234");

    const submit = screen.getByRole("button", { name: "Anmelden" });
    expect(submit).toBeEnabled();

    await user.click(submit);

    const pendingButton = await screen.findByRole("button", { name: "Wird geprüft …" });
    expect(pendingButton).toBeDisabled();

    release(ok(bootstrap()));
  });

  it("is disabled before anything has been typed", async () => {
    stubApi({ "GET /api/me": unauthenticated });

    await renderApp("/");

    expect(screen.getByRole("button", { name: "Anmelden" })).toBeDisabled();
  });

  // A year-long session should take a returning guest straight to the content,
  // never past the screen asking for a code they no longer need.
  it("sends an already-logged-in household onward", async () => {
    stubApi({ "GET /api/me": ok(bootstrap()) });

    const { router } = await renderApp("/");

    await waitFor(() => expect(currentPath(router)).toBe("/willkommen"));
  });
});
