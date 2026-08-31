import { screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ok, stubApi, unauthenticated } from "@/test/api";
import { bootstrap, rsvpAnswer } from "@/test/fixtures";
import { currentPath, renderApp } from "@/test/render";
import { weddingDate } from "@/lib/wedding";

/** F2: the guest chrome and the informational pages. */

function confirmHousehold(householdId = 12) {
  window.localStorage.setItem("wedding.confirmed-households", JSON.stringify([householdId]));
}

/** A logged-in household, with the two queries every guest page needs. */
function stubGuest(overrides: { me?: ReturnType<typeof bootstrap>; answer?: ReturnType<typeof rsvpAnswer> } = {}) {
  confirmHousehold();
  return stubApi({
    "GET /api/me": ok(overrides.me ?? bootstrap()),
    "GET /api/rsvp": ok(overrides.answer ?? rsvpAnswer()),
  });
}

/** A household that has already answered, which is what changes the answer entry. */
function answeredHousehold() {
  const answer = rsvpAnswer();
  return rsvpAnswer({ household: { ...answer.household, rsvp_submitted_at: "2026-12-01T10:00:00Z" } });
}

/** The bottom bar and the top nav render the same entries, so a query by role finds
    each label twice. This is the count that says both are there. */
function navEntries(label: string | RegExp) {
  return screen.getAllByRole("link", { name: label });
}

describe("the guest navigation", () => {
  it("shows every entry, with a visible label, on every guest page", async () => {
    stubGuest();

    await renderApp("/start");

    for (const label of ["Start", "Ablauf", "Location", "Antwort", "Mehr"]) {
      // Two of each: the top nav and the bottom bar, from one definition.
      expect(navEntries(label)).toHaveLength(2);
    }
  });

  it("marks the page you are on", async () => {
    stubGuest();

    const { user } = await renderApp("/start");
    await user.click(navEntries("Ablauf")[0] as HTMLElement);

    expect(await screen.findByRole("heading", { name: "Ablauf", level: 1 })).toBeInTheDocument();
    for (const entry of navEntries("Ablauf")) {
      expect(entry).toHaveAttribute("aria-current", "page");
    }
  });

  it("words the answer entry as an invitation while the household has not answered", async () => {
    stubGuest();

    await renderApp("/start");

    expect(navEntries("Antwort")).toHaveLength(2);
    expect(screen.queryByRole("link", { name: "Antwort ändern" })).not.toBeInTheDocument();
  });

  it("words the answer entry as a change once they have", async () => {
    stubGuest({ answer: answeredHousehold() });

    await renderApp("/start");

    // Awaited: the entry is worded from the RSVP query, which the chrome asks for as
    // it renders rather than in a loader — the page must not wait on it.
    // Three: both nav entries and the start page's own call to action, which follows
    // the same state.
    expect(await screen.findAllByRole("link", { name: "Antwort ändern" })).toHaveLength(3);
  });

  // Absent, never disabled: an unpublished seating plan is not "coming soon", it is
  // nothing a guest should be thinking about yet.
  it("leaves the flag-gated entries out of /mehr entirely", async () => {
    stubGuest();

    await renderApp("/mehr");

    for (const label of ["Dresscode", "Geschenke", "Häufige Fragen", "Kontakt", "Datenschutz"]) {
      expect(screen.getByRole("link", { name: label })).toBeInTheDocument();
    }
    expect(screen.queryByText("Sitzplan")).not.toBeInTheDocument();
    expect(screen.queryByText("Galerie")).not.toBeInTheDocument();
  });

  it("renders the seating and gallery entries once their flags are on", async () => {
    stubGuest({
      me: bootstrap({
        flags: { rsvp_open: true, seating_published: true, gallery_visible: true, uploads_open: false },
      }),
    });

    await renderApp("/mehr");

    expect(screen.getByRole("link", { name: "Sitzplan" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Galerie" })).toBeInTheDocument();
  });

  it("carries the logout on /mehr rather than in the bar", async () => {
    stubGuest();

    await renderApp("/mehr");

    expect(screen.getByRole("button", { name: "Abmelden" })).toBeInTheDocument();
  });

  // A nav offering "Ablauf" to somebody who has not logged in is both wrong and a
  // small leak of what is behind the door.
  it("renders no guest nav on the login screen", async () => {
    stubApi({ "GET /api/me": unauthenticated });

    await renderApp("/");

    expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
  });

  // The confirmation screen asks one question, and until it is answered every other
  // guest route redirects straight back to it — a nav bar there is five links that
  // all bounce.
  it("renders no guest nav on the confirmation screen", async () => {
    stubApi({ "GET /api/me": ok(bootstrap()), "GET /api/rsvp": ok(rsvpAnswer()) });

    await renderApp("/willkommen");

    expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
  });

  it("keeps the skip link pointing past the navigation", async () => {
    stubGuest();

    await renderApp("/start");

    expect(screen.getByRole("link", { name: "Zum Inhalt springen" })).toHaveAttribute("href", "#main");
    // The nav must not be inside what the skip link skips to.
    const main = document.querySelector("main");
    expect(main?.querySelector("nav")).toBeNull();
  });
});

describe("the start page", () => {
  it("greets the household by its display name", async () => {
    stubGuest();

    await renderApp("/start");

    expect(screen.getByText(/Familie Müller/)).toBeInTheDocument();
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
  });

  /** Renders the start page with the clock set to a fixed local moment. */
  async function renderOn(dayOffset: number, hour: number) {
    vi.useFakeTimers();
    vi.setSystemTime(
      new Date(weddingDate.getFullYear(), weddingDate.getMonth(), weddingDate.getDate() + dayOffset, hour),
    );
    // Real timers again before rendering: the fake clock is what the countdown reads,
    // and userEvent and the query client both need timers that actually run.
    const now = new Date();
    vi.useRealTimers();
    vi.setSystemTime(now);
    vi.useFakeTimers({ shouldAdvanceTime: true });

    await renderApp("/start");
  }

  it("counts the days remaining before the wedding", async () => {
    stubGuest();

    await renderOn(-3, 8);

    expect(screen.getByText("Noch 3 Tage")).toBeInTheDocument();
  });

  // A guest opening the site at 23:50 must be told the same number as the person
  // beside them at 00:10 — the count is taken against local midnight.
  it("does not change within the same local day", async () => {
    stubGuest();

    await renderOn(-3, 23);

    expect(screen.getByText("Noch 3 Tage")).toBeInTheDocument();
  });

  it("says heute on the day itself", async () => {
    stubGuest();

    await renderOn(0, 9);

    expect(screen.getByText("Heute ist es so weit")).toBeInTheDocument();
  });

  // A badge reading "-3 Tage" in August 2027 is the kind of thing nobody tests and
  // everybody sees.
  it("disappears after the wedding", async () => {
    stubGuest();

    await renderOn(1, 9);

    expect(screen.queryByText(/Noch \d+ Tage/)).not.toBeInTheDocument();
    expect(screen.queryByText("Heute ist es so weit")).not.toBeInTheDocument();
  });

  it("points at the RSVP form", async () => {
    stubGuest();

    const { user, router } = await renderApp("/start");
    await user.click(screen.getByRole("link", { name: "Jetzt zusagen" }));

    expect(currentPath(router)).toBe("/zusagen");
  });

  it("offers to change the answer once the household has given one", async () => {
    stubGuest({ answer: answeredHousehold() });

    await renderApp("/start");

    // Three: the two nav entries and the page's own call to action.
    expect(await screen.findAllByRole("link", { name: "Antwort ändern" })).toHaveLength(3);
  });

  it("carries the hero photo as decoration, not as content", async () => {
    stubGuest();

    await renderApp("/start");

    // alt="" — the image sits beside text that already says who and when.
    const image = document.querySelector("img");
    expect(image).toHaveAttribute("alt", "");
    expect(image).not.toHaveAttribute("loading", "lazy");
  });
});

describe("the content pages", () => {
  it("renders the schedule in order, and says so when a time is not fixed", async () => {
    stubGuest();

    await renderApp("/ablauf");

    const entries = screen.getAllByRole("listitem").filter((item) => item.querySelector("h2"));
    expect(entries.map((item) => item.querySelector("h2")?.textContent)).toEqual([
      "Trauung",
      "Empfang",
      "Abendessen",
      "Feier",
    ]);
    expect(screen.getAllByText("Uhrzeit steht noch nicht fest").length).toBe(entries.length);
    // The church entry is marked, and the party entries are not hidden from a guest
    // who is only coming to the church.
    expect(within(entries[0] as HTMLElement).getByText("Kirche")).toBeInTheDocument();
    expect(within(entries[1] as HTMLElement).getByText("Feier")).toBeInTheDocument();
  });

  it("renders both venues and every section anchor on the location page", async () => {
    stubGuest();

    await renderApp("/location");

    expect(screen.getByRole("heading", { name: "Kirche", level: 3 })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Feier", level: 3 })).toBeInTheDocument();
    for (const anchor of ["orte", "anreise", "fahrt", "uebernachtung"]) {
      expect(document.getElementById(anchor)).not.toBeNull();
    }
  });

  it("renders the dress code and the gift pages", async () => {
    stubGuest();

    await renderApp("/dresscode");
    expect(screen.getByRole("heading", { name: "Dresscode", level: 1 })).toBeInTheDocument();

    await renderApp("/geschenke");
    expect(screen.getByRole("heading", { name: "Geschenke", level: 1 })).toBeInTheDocument();
    // The account is a placeholder, so the page says so rather than showing a copy
    // button for an IBAN that is not ours.
    expect(screen.getByText(/Bankverbindung tragen wir hier ein/)).toBeInTheDocument();
  });

  it("shows every FAQ answer without interaction", async () => {
    stubGuest();

    await renderApp("/faq");

    expect(screen.getByRole("heading", { name: "Sind Kinder eingeladen?", level: 2 })).toBeInTheDocument();
    expect(screen.getByText(/Wer allein eingeladen ist/)).toBeInTheDocument();
    // No accordion: nothing here is a disclosure control.
    expect(screen.queryByRole("button", { name: /Kinder/ })).not.toBeInTheDocument();
  });

  it("links an FAQ answer to the page that holds the detail", async () => {
    stubGuest();

    const { user, router } = await renderApp("/faq");
    await user.click(screen.getByRole("link", { name: "Zum Dresscode" }));

    expect(currentPath(router)).toBe("/dresscode");
  });

  it("lists both of us on the contact page, each as a tel: link", async () => {
    stubGuest();

    await renderApp("/kontakt");

    expect(screen.getByText("Andreas Hell")).toBeInTheDocument();
    expect(screen.getByText("Isabella Michelbacher")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "+43 650 9408100" })).toHaveAttribute("href", "tel:+436509408100");
    expect(screen.getByRole("link", { name: "+43 677 63668655" })).toHaveAttribute(
      "href",
      "tel:+436776366865 5".replace(" ", ""),
    );
  });
});

describe("the privacy page", () => {
  // Behind the login like every other page: nobody without a code has any data
  // described here, and the version without navigation looked like a different site.
  it("renders with the guest navigation", async () => {
    stubGuest();

    await renderApp("/datenschutz");

    expect(screen.getByRole("heading", { name: "Datenschutz", level: 1 })).toBeInTheDocument();
    expect(navEntries("Start")).toHaveLength(2);
  });

  it("sends a visit without a session to the login screen", async () => {
    stubApi({ "GET /api/me": unauthenticated });

    const { router } = await renderApp("/datenschutz");

    expect(currentPath(router)).toBe("/");
  });

  it("names the two fields a guest might hesitate over, and what happens afterwards", async () => {
    stubGuest();

    await renderApp("/datenschutz");

    expect(screen.getByText(/Allergien landen auf der Liste für die Küche/)).toBeInTheDocument();
    expect(screen.getByText(/drei Monate nach der Hochzeit offline/)).toBeInTheDocument();
  });

  it("is reachable from /mehr", async () => {
    stubGuest();

    const { user, router } = await renderApp("/mehr");
    await user.click(screen.getByRole("link", { name: "Datenschutz" }));

    expect(currentPath(router)).toBe("/datenschutz");
  });
});
