import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { watchForSessionExpiry } from "@/lib/api/session";
import { createAppRouter } from "@/lib/routing/router";

/**
 * Renders the whole application at a given path.
 *
 * The real route tree, the real guards, the real query client and — since F11-06 —
 * literally the same router factory main.tsx uses. A test that assembled a smaller
 * router would prove the smaller router works, and every bug worth catching here
 * lives in the guards.
 */
export async function renderApp(initialPath = "/") {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } },
  });
  watchForSessionExpiry(queryClient);

  // The application's own router, differing only in its history: a test that built
  // its own would keep passing against options the app no longer has (F11-06).
  const router = createAppRouter(queryClient, createMemoryHistory({ initialEntries: [initialPath] }));

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
