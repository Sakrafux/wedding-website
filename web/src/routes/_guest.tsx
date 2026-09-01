import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Navigate, Outlet, redirect } from "@tanstack/react-router";

import type { BootstrapResponse } from "@/lib/api/dto";
import { RoutePending } from "@/components/RouteStates";
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
  beforeLoad: ({ context, location }) => {
    // Cache first, and **synchronously** when it answers. An async guard puts the
    // router into its pending state for at least a tick even when the value is
    // already in hand, and with defaultPendingMs: 0 that rendered a full-screen
    // skeleton on every navigation between two hardcoded content pages (F11-04).
    //
    // `undefined` is "never fetched"; a cached `null` is "not logged in", which is an
    // answer and must not send us back to the network.
    const cached = context.queryClient.getQueryData(meQueryOptions.queryKey);
    if (cached !== undefined) {
      guardHousehold(cached, location.pathname, location.href);
      return;
    }

    // Only the cold load waits. Awaited rather than resolved during render, so a
    // caller who turns out not to be logged in never sees a flash of guest content.
    return context.queryClient.ensureQueryData(meQueryOptions).then((me) => {
      guardHousehold(me, location.pathname, location.href);
    });
  },
  // Full-screen, unlike the content routes below it: when this guard is pending the
  // app is booting and there is no navigation to sit inside yet (F11-06).
  pendingComponent: RoutePending,
  component: GuestLayout,
});

/**
 * The two redirects every guest route inherits, in one place so that both the cached
 * and the fetched path apply exactly the same rules.
 *
 * Throws rather than returns: a redirect from `beforeLoad` is a thrown value in
 * TanStack Router, and returning one here would guard nothing at all.
 */
function guardHousehold(me: BootstrapResponse | null, pathname: string, href: string) {
  if (!me) {
    // The intended path travels along, so logging in lands the guest where they were
    // going rather than on a generic start page.
    throw redirect({ to: "/", search: { redirect: href } });
  }

  // Shown once per household, not once per visit: a year-long session must not ask
  // this daily. The check lives here rather than in a component so that a deep link
  // cannot skip it.
  if (!isHouseholdConfirmed(me.household.id) && pathname !== confirmationPath) {
    throw redirect({ to: confirmationPath });
  }
}

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
