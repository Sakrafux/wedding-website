import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { CodeInput } from "@/components/CodeInput";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { normalizeCode } from "@/lib/code";
import { contactPhoneNumber, loginLabels } from "@/lib/labels";
import { safeInternalPath } from "@/lib/routing/navigation";
import { meQueryOptions, useLogin } from "@/lib/api/session";

/** Where a guest lands when they had no particular destination in mind. */
const defaultDestination = "/start";

/**
 * Failed attempts before the phone number appears. Two, because that is where a
 * person stops assuming they mistyped and starts assuming they are the problem.
 */
const failuresBeforeFallback = 2;

interface LoginSearch {
  /** The guarded path the guest was heading for before being sent here. */
  redirect?: string;
  /** Set when a session ended mid-visit, so the screen can say why. */
  reason?: "expired" | "rejected";
}

export const Route = createFileRoute("/")({
  validateSearch: (search: Record<string, unknown>): LoginSearch => ({
    redirect: typeof search.redirect === "string" ? search.redirect : undefined,
    reason: search.reason === "expired" || search.reason === "rejected" ? search.reason : undefined,
  }),
  beforeLoad: async ({ context, search }) => {
    // Somebody already logged in has no business on the login screen: a year-long
    // session should take them straight to the content.
    const me = await context.queryClient.ensureQueryData(meQueryOptions);
    if (me) {
      throw redirect({ href: safeInternalPath(search.redirect, defaultDestination) });
    }
  },
  component: LoginPage,
});

function LoginPage() {
  const { redirect: intendedPath, reason } = Route.useSearch();
  const navigate = useNavigate();
  const login = useLogin();

  const [code, setCode] = useState("");
  const [failures, setFailures] = useState(0);

  // The API owns the wording of every failure, so it is shown verbatim. A
  // NetworkError carries a German sentence of its own for the one case the server
  // never answered.
  const errorMessage = login.error?.message;

  const noticeMessage =
    reason === "expired" ? loginLabels.loggedOut : reason === "rejected" ? loginLabels.rejectedHousehold : undefined;

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    try {
      // Normalized here rather than in the field: what the guest sees keeps the
      // dash they typed, what the API is asked about never has one.
      await login.mutateAsync(normalizeCode(code));
      await navigate({ href: safeInternalPath(intendedPath, defaultDestination) });
    } catch (error) {
      // Only a rejection by the server counts towards revealing the phone number;
      // a dropped connection is not the guest failing at anything.
      if (error instanceof ApiError) {
        setFailures((previous) => previous + 1);
      }
    }
  }

  return (
    <div className="mx-auto flex min-h-dvh max-w-md flex-col justify-center gap-8 px-4 py-12">
      <header className="flex flex-col gap-3 text-center">
        <h1 className="text-display">{loginLabels.heading}</h1>
        <p className="text-ink-muted">{loginLabels.intro}</p>
      </header>

      {noticeMessage ? (
        <p role="status" className="bg-surface-sunken text-small rounded-lg px-4 py-3 text-center">
          {noticeMessage}
        </p>
      ) : null}

      {/* A real form, so the on-screen keyboard offers "go" and Enter submits. */}
      <form onSubmit={submit} className="flex flex-col gap-6" noValidate>
        <CodeInput value={code} onChange={setCode} error={errorMessage} disabled={login.isPending} />

        {/* Disabled while in flight, because a silent second press on a slow
            connection is how a guest ends up submitting twice. The label changes
            too: a spinner alone reads as "broken" to someone unsure of the app. */}
        <Button
          type="submit"
          size="lg"
          className="h-14 w-full text-lg"
          disabled={login.isPending || normalizeCode(code).length === 0}
        >
          {login.isPending ? loginLabels.submitting : loginLabels.submit}
        </Button>
      </form>

      {failures >= failuresBeforeFallback ? (
        <p className="text-small text-ink-muted text-center">
          {loginLabels.fallback} {/* A tel: link, because the person reading this is holding a phone. */}
          <a href={`tel:${contactPhoneNumber.replace(/\s/g, "")}`} className="text-primary underline">
            {contactPhoneNumber}
          </a>
        </p>
      ) : null}
    </div>
  );
}
