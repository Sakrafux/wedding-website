import { createRootRoute, Outlet } from "@tanstack/react-router";
import { Suspense } from "react";

import { Devtools } from "@/components/Devtools";

/**
 * The root route: the shell every page renders inside.
 *
 * It is deliberately almost empty. Navigation, the household session guard and the
 * error boundary arrive with F1-F03, which is the story that knows what a logged-in
 * shell looks like; putting a placeholder nav here now would be a layout that gets
 * thrown away.
 */
export const Route = createRootRoute({
  component: RootLayout,
});

function RootLayout() {
  return (
    <>
      {/* Required by the accessibility rules: a keyboard user must be able to jump
          past the navigation that F1-F03 adds above this outlet. */}
      <a
        href="#main"
        className="focus:bg-surface sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:rounded-lg focus:px-4 focus:py-2"
      >
        Zum Inhalt springen
      </a>
      <main id="main">
        <Outlet />
      </main>
      {/* Suspense because the dev build loads the panels lazily; there is nothing
          to fall back to, since in production this renders null. */}
      <Suspense fallback={null}>
        <Devtools />
      </Suspense>
    </>
  );
}
