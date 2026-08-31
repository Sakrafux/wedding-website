import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Navigate, Outlet, redirect } from "@tanstack/react-router";

import { isHouseholdConfirmed, meQueryOptions } from "@/lib/api/session";

/** The route that asks a freshly logged-in household whether the code was theirs. */
export const confirmationPath = "/willkommen";

/**
 * Everything a logged-in household may see.
 *
 * A pathless layout, so the guard is written once and every page underneath
 * inherits it. A per-page check is how one page eventually ships without one.
 */
export const Route = createFileRoute("/_guest")({
  beforeLoad: async ({ context, location }) => {
    // Awaited, so the redirect happens before anything renders. A guard that
    // resolved during render would flash the guest's own content at somebody who
    // turns out not to be logged in.
    const me = await context.queryClient.ensureQueryData(meQueryOptions);

    if (!me) {
      // The intended path travels along, so logging in lands the guest where they
      // were going rather than on a generic start page.
      throw redirect({ to: "/", search: { redirect: location.href } });
    }

    // Shown once per household, not once per visit: a year-long session must not
    // ask this daily. The check lives here rather than in a component so that a
    // deep link cannot skip it.
    if (!isHouseholdConfirmed(me.household.id) && location.pathname !== confirmationPath) {
      throw redirect({ to: confirmationPath });
    }
  },
  component: GuestLayout,
});

/**
 * Renders the page, and leaves the moment the session stops existing.
 *
 * The guard above runs on navigation, which is not enough on its own: a session can
 * be revoked while somebody is sitting on a page, and the 401 that discovers it
 * arrives from whatever query happened to be running. watchForSessionExpiry writes
 * `null` into the session query, and this is what reacts to it — without it the app
 * would keep rendering content it can no longer fetch until the next navigation.
 */
function GuestLayout() {
  const { data: me } = useQuery(meQueryOptions);

  if (!me) {
    return <Navigate to="/" search={{ reason: "expired" }} replace />;
  }
  return <Outlet />;
}
