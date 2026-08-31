import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { confirmationLabels } from "@/lib/labels";
import { meQueryOptions, rememberHouseholdConfirmed, useLogout } from "@/lib/api/session";

export const Route = createFileRoute("/_guest/willkommen")({
  component: ConfirmationPage,
});

/**
 * "Is this you?", asked once, immediately after logging in.
 *
 * This screen exists for exactly one failure: two valid codes differing by one
 * character. The alphabet makes that unlikely; this makes it harmless — the
 * alternative is discovering it after somebody has answered the RSVP on another
 * household's behalf.
 */
function ConfirmationPage() {
  // Guaranteed present: the layout's guard has already resolved the session, and
  // redirects when there is none.
  const { data: me } = useSuspenseQuery(meQueryOptions);
  const navigate = useNavigate();
  const logout = useLogout();

  if (!me) {
    return null;
  }

  function confirm() {
    rememberHouseholdConfirmed(me!.household.id);
    void navigate({ to: "/start" });
  }

  async function reject() {
    // Logging out server-side, not merely navigating away. A session left behind
    // is precisely the bug this screen exists to prevent — the next visit would
    // walk straight back into the wrong household.
    await logout.mutateAsync();
    await navigate({ to: "/", search: { reason: "rejected" } });
  }

  return (
    <div className="mx-auto flex min-h-dvh max-w-md flex-col justify-center gap-8 px-4 py-12">
      <header className="flex flex-col gap-3 text-center">
        <h1 className="text-h1">{confirmationLabels.heading(me.household.display_name)}</h1>
      </header>

      {/* The member list is what actually catches the error: two households with
          adjacent codes will almost never have the same first names, whereas
          "Familie Müller" twice is entirely plausible. */}
      <section className="bg-surface border-line shadow-card flex flex-col gap-2 rounded-xl border p-6">
        <p className="text-small text-ink-muted">{confirmationLabels.membersIntro}</p>
        <ul className="flex flex-col gap-1">
          {me.members.map((member) => (
            <li key={member.id} className="text-h3">
              {member.first_name}
            </li>
          ))}
        </ul>
      </section>

      {/* Two actions of equal weight. No third option and no "remind me later":
          every extra choice here is a chance to pick neither and carry on as
          somebody else. */}
      <div className="flex flex-col gap-3">
        <Button size="lg" className="h-14 w-full text-lg" onClick={confirm} disabled={logout.isPending}>
          {confirmationLabels.confirm}
        </Button>
        <Button
          size="lg"
          variant="outline"
          className="h-14 w-full text-lg"
          onClick={() => void reject()}
          disabled={logout.isPending}
        >
          {confirmationLabels.reject}
        </Button>
      </div>
    </div>
  );
}
