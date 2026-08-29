import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createRouter, RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

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

// basepath from BASE_URL, not a literal: the app is served under /hochzeit/, so
// every route the router builds or matches has to carry that prefix. Vite is the one
// place the prefix is configured; see vite.config.ts.
const router = createRouter({ routeTree, basepath: import.meta.env.BASE_URL });

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
