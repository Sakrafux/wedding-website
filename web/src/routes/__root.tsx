import type { QueryClient } from "@tanstack/react-query";
import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";
import { Suspense } from "react";

import { Devtools } from "@/components/Devtools";

/**
 * The router context: the query client, and nothing else.
 *
 * Route guards need it in `beforeLoad`, where hooks are not available — which is
 * the point, since a guard has to resolve before the route renders rather than
 * flashing a login screen at somebody who is already logged in.
 */
export interface RouterContext {
  queryClient: QueryClient;
}

/**
 * The root route: the shell every page renders inside.
 *
 * Deliberately almost empty — the skip link and the outlet. The guest navigation
 * lives in the `/_guest` layout, which is also where `<main id="main">` is: the skip
 * link has to land *past* the nav, and a `main` in here would put the nav inside it.
 * Every layout and every standalone page therefore renders its own `main`.
 */
export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
});

function RootLayout() {
  return (
    <>
      {/* Required by the accessibility rules: a keyboard user must be able to jump
          past the guest navigation, which renders above the `main` each layout
          provides. */}
      <a
        href="#main"
        className="focus:bg-surface sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:rounded-lg focus:px-4 focus:py-2"
      >
        Zum Inhalt springen
      </a>
      <Outlet />
      {/* Suspense because the dev build loads the panels lazily; there is nothing
          to fall back to, since in production this renders null. */}
      <Suspense fallback={null}>
        <Devtools />
      </Suspense>
    </>
  );
}
