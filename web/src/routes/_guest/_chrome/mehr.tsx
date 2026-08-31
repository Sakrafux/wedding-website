import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";

import { moreEntries } from "@/components/layout/guestNavigation";
import { PageHeading, PageSections } from "@/components/layout/InfoSection";
import { Button } from "@/components/ui/button";
import { meQueryOptions, useLogout } from "@/lib/api/session";
import { moreLabels, shellLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/mehr")({
  component: MorePage,
});

/**
 * The overflow of the navigation, as a real page.
 *
 * A list of links at 48px each beats a menu that overlays the page and has to be
 * dismissed — and the entries gated by `seating_published` and `gallery_visible` are
 * **absent** rather than disabled: an unpublished seating plan is not "coming soon",
 * it is nothing a guest should be thinking about yet.
 */
function MorePage() {
  const { data: me } = useSuspenseQuery(meQueryOptions);
  const navigate = useNavigate();
  const logout = useLogout();

  async function signOut() {
    await logout.mutateAsync();
    await navigate({ to: "/" });
  }

  const entries = me ? moreEntries(me.flags) : [];

  return (
    <PageSections>
      <PageHeading title={moreLabels.heading} lead={moreLabels.intro} />

      <ul className="flex flex-col gap-2">
        {entries.map((entry) => {
          const Icon = entry.icon;
          return (
            <li key={entry.to}>
              <Link
                // Same list as the nav bars, so an entry cannot exist in one and not
                // the other. F7 and F9 add the routes the gated entries point at.
                to={entry.to as never}
                className="border-line hover:bg-primary-soft flex min-h-14 items-center gap-3 rounded-lg border px-4"
              >
                <Icon className="text-ink-muted size-5 shrink-0" aria-hidden="true" />
                <span>{entry.label}</span>
              </Link>
            </li>
          );
        })}
      </ul>

      {/* Here rather than in the bar: used once a year, and a fifth bar item that logs
          you out beside the one showing the schedule is a mis-tap waiting to happen. */}
      <Button variant="outline" size="lg" className="h-14 self-start" onClick={() => void signOut()}>
        {shellLabels.logout}
      </Button>
    </PageSections>
  );
}
