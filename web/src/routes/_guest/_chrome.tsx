import { createFileRoute, Outlet } from "@tanstack/react-router";

import { GuestNavShell } from "@/components/layout/GuestChrome";

/**
 * Every guest page that gets the navigation around it.
 *
 * A second pathless layout inside `/_guest` rather than a condition in that one: the
 * confirmation screen (`/willkommen`) must render **without** navigation — it asks one
 * question, and until it is answered the guard sends every other guest route straight
 * back to it, so a bar there would offer five links that all bounce. Expressing that
 * as "which layout is the page in" keeps it out of the render path entirely.
 */
export const Route = createFileRoute("/_guest/_chrome")({
  component: ChromeLayout,
});

function ChromeLayout() {
  return (
    <GuestNavShell>
      <Outlet />
    </GuestNavShell>
  );
}
