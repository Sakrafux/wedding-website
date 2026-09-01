/**
 * The router, configured once.
 *
 * `main.tsx` and the test harness both build one from here rather than each writing
 * out the options: they differ only in the history they run on, and a test router with
 * its own copy of these settings is a test that keeps passing against a configuration
 * the application no longer has (F11-06).
 */

import type { QueryClient } from "@tanstack/react-query";
import { createRouter, type RouterHistory } from "@tanstack/react-router";

import { RouteError, SectionPending } from "@/components/RouteStates";
import { routeTree } from "@/routeTree.gen";

/**
 * How long a transition may take before anything is drawn for it.
 *
 * Under this, nothing renders at all — which is the whole point: with route chunks
 * code-split (`autoCodeSplitting` in `vite.config.ts`), a first visit to a page fetches
 * a sub-kilobyte file, and drawing a skeleton for that reads as a flash. Above it, the
 * wait is real and a pending state is the honest answer.
 *
 * It does not contradict the reasoning that had this at 0 (F11-04): that was about the
 * cold load, where the guard waits on the network and 150 ms is invisible.
 */
const pendingThresholdMs = 150;

export function createAppRouter(queryClient: QueryClient, history?: RouterHistory) {
  return createRouter({
    routeTree,
    context: { queryClient },
    history,
    // Content-shaped by default, because most routes render inside a navigation. The
    // three that are genuinely full-screen — `_guest`'s guard, the login screen and the
    // admin login — name `RoutePending` themselves. Pending components resolve per
    // route with a fallback to this, and never from a parent layout, so a default is
    // the only way to cover a subtree.
    defaultPendingComponent: SectionPending,
    defaultErrorComponent: RouteError,
    defaultPendingMs: pendingThresholdMs,
    // Every rendered `<Link>` preloads its route: the guest navigation links to every
    // page from every page, so the whole site is warm shortly after first paint and a
    // tap fetches nothing. "intent" was the alternative and still races a fast tap.
    defaultPreload: "render",
  });
}
