import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createRouter, RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { RouteError, RoutePending } from "@/components/RouteStates";
import { watchForSessionExpiry } from "@/lib/api/session";

import "./index.css";
import { routeTree } from "./routeTree.gen";

/**
 * One QueryClient for the app.
 *
 * `retry: false` because every failing request here is either a real server error
 * or an expired session, and silently retrying both delays the message the guest
 * needs to see. `refetchOnWindowFocus: false` for the same reason in reverse: a
 * half-filled RSVP form must not be overwritten because somebody switched tabs.
 */
const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, refetchOnWindowFocus: false },
  },
});

// Never unsubscribed: it lives as long as the page does.
watchForSessionExpiry(queryClient);

const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPendingComponent: RoutePending,
  defaultErrorComponent: RouteError,
  // Shown from the first millisecond rather than after a delay: the guards await
  // the session query, and an unstyled gap before the skeleton reads as a broken
  // page on a slow connection.
  defaultPendingMs: 0,
});

// Gives useNavigate, Link and the rest their typed route paths. Without it every
// route string is just a string.
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("missing #root element in index.html");
}

createRoot(rootElement).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
);
