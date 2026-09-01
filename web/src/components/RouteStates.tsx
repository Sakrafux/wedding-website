import { type ErrorComponentProps, useRouter } from "@tanstack/react-router";

import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { shellLabels } from "@/lib/labels";

/**
 * What every route shows while its guard is resolving.
 *
 * A skeleton, never a blank page and never the login screen. A flash of the login
 * screen reads as "it logged me out again" to somebody who is already logged in,
 * and that impression is expensive in exactly the audience with the least trust in
 * the thing to begin with.
 */
export function RoutePending() {
  return (
    <div className="mx-auto flex min-h-dvh max-w-md flex-col justify-center gap-4 px-4" aria-busy="true">
      <p className="sr-only">{shellLabels.loading}</p>
      <div className="bg-surface-sunken h-8 w-2/3 animate-pulse rounded-lg" />
      <div className="bg-surface-sunken h-14 w-full animate-pulse rounded-lg" />
      <div className="bg-surface-sunken h-14 w-full animate-pulse rounded-lg" />
    </div>
  );
}

/**
 * What a page inside the guest navigation shows while its own data loads.
 *
 * `RoutePending` centres a login-shaped column in the viewport, which is right at `/`
 * and wrong under a navigation bar — it reads as the whole app reloading rather than as
 * one page filling in. This one is content-shaped and sits in the page's own flow
 * (F11-04).
 */
export function SectionPending() {
  return (
    <div className="flex flex-col gap-4" aria-busy="true">
      <p className="sr-only">{shellLabels.loading}</p>
      <div className="bg-surface-sunken h-10 w-1/2 animate-pulse rounded-lg" />
      <div className="bg-surface-sunken h-32 w-full animate-pulse rounded-xl" />
      <div className="bg-surface-sunken h-32 w-full animate-pulse rounded-xl" />
    </div>
  );
}

/**
 * The error boundary for a failed route load.
 *
 * A dropped connection and a broken server are both shown here with a retry, and
 * neither logs anybody out — dropping a guest to the login screen because a train
 * went into a tunnel is a bad trade. A 401 never reaches this component: the
 * session queries answer "not logged in" as data, so the guards redirect instead.
 */
export function RouteError({ error, reset }: ErrorComponentProps) {
  const router = useRouter();
  const requestId = error instanceof ApiError ? error.requestId : undefined;

  // Both halves are needed: reset clears the boundary, and invalidate makes the
  // router run the guard again instead of re-rendering the same failed load.
  function retry() {
    reset();
    void router.invalidate();
  }

  return (
    <div className="mx-auto flex min-h-dvh max-w-md flex-col justify-center gap-4 px-4 text-center">
      <h1 className="text-h1">{shellLabels.errorHeading}</h1>
      {/* The API's own German sentence, or NetworkError's, shown verbatim. */}
      <p className="text-ink-muted">{error.message}</p>

      {requestId ? (
        <p className="text-small text-ink-muted">
          {shellLabels.requestId} <span className="font-mono">{requestId}</span>
        </p>
      ) : null}

      <Button onClick={retry} size="lg" className="h-14 w-full">
        {shellLabels.retry}
      </Button>
    </div>
  );
}
