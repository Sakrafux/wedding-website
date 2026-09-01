import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { watchForSessionExpiry } from "@/lib/api/session";
import { createAppRouter } from "@/lib/routing/router";

import "./index.css";

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

// Every option lives in createAppRouter, so the test harness runs the real
// configuration rather than its own copy of it (F11-06).
const router = createAppRouter(queryClient);

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
