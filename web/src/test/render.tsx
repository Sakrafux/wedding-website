import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, createRouter, RouterProvider } from "@tanstack/react-router";
import { render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { RouteError, RoutePending } from "@/components/RouteStates";
import { watchForSessionExpiry } from "@/lib/api/session";
import { routeTree } from "@/routeTree.gen";

/**
 * Renders the whole application at a given path.
 *
 * The real route tree, the real guards and the real query client — wired the same
 * way main.tsx wires them. A test that assembled a smaller router would prove the
 * smaller router works, and every bug worth catching here lives in the guards.
 */
export async function renderApp(initialPath = "/") {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } },
  });
  watchForSessionExpiry(queryClient);

  const router = createRouter({
    routeTree,
    context: { queryClient },
    history: createMemoryHistory({ initialEntries: [initialPath] }),
    defaultPendingComponent: RoutePending,
    defaultErrorComponent: RouteError,
    defaultPendingMs: 0,
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router as never} />
    </QueryClientProvider>,
  );

  // The guards await the session query, so nothing meaningful is on screen until
  // the first load settles.
  await router.load();

  return { router, queryClient, user: userEvent.setup() };
}

/** Where the router currently is, for asserting on a redirect. */
export function currentPath(router: { state: { location: { pathname: string } } }): string {
  return router.state.location.pathname;
}
