import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

import { rsvpQueryOptions } from "@/lib/api/rsvp";
import { navLabels } from "@/lib/labels";
import { cn } from "@/lib/utils";

import { barEntries, type NavEntry } from "./guestNavigation";

/**
 * The chrome with its entries worked out: what every page a logged-in household sees
 * renders inside.
 *
 * Split from the layout route that renders it so the entry list and the query behind
 * it stay one thing, testable without a router match.
 */
export function GuestNavShell({ children }: { children: React.ReactNode }) {
  // The answer entry reads "Antwort ändern" once the household has answered, from the
  // query F3 already owns rather than from a new field on `me`. While it loads the
  // entry reads "Antwort" — the safe half of the pair, since it invites an answer
  // rather than implying one was given.
  const { data: answer } = useQuery(rsvpQueryOptions);

  return <GuestChrome entries={barEntries(answer?.household.rsvp_submitted_at != null)}>{children}</GuestChrome>;
}

/**
 * The guest chrome: the same navigation on every page a logged-in household sees.
 *
 * A fixed bottom bar on a phone with icon **and** visible label, a horizontal top nav
 * from `sm` upwards. Both render one list (`guestNavigation.ts`) — never two that
 * drift apart — and neither is a hamburger: a hamburger hides exactly the links an
 * unconfident guest is hunting for (05-design).
 *
 * It lives in the `/_guest` layout rather than in the root, because the root also
 * renders the login screen, the admin area and `/datenschutz`. A nav offering "Ablauf"
 * to somebody who has not logged in is both wrong and a small leak of what is behind
 * the door.
 */
export function GuestChrome({ entries, children }: { entries: NavEntry[]; children: React.ReactNode }) {
  return (
    <>
      <nav aria-label={navLabels.guestNav} className="border-line bg-surface hidden border-b sm:block">
        <ul className="mx-auto flex max-w-2xl items-center gap-2 px-6 py-2">
          {entries.map((entry) => (
            <li key={entry.to}>
              <NavLink entry={entry} className="min-h-12 gap-2 px-3 py-2" />
            </li>
          ))}
        </ul>
      </nav>

      {/* The bar is fixed, so the page needs padding of its own height underneath:
          content that ends beneath a fixed bar is content nobody scrolls to. */}
      <main id="main" className="mx-auto w-full max-w-2xl px-4 pt-8 pb-32 sm:px-6 sm:pb-16">
        {children}
      </main>

      <nav
        aria-label={navLabels.guestNav}
        className="border-line bg-surface fixed inset-x-0 bottom-0 border-t sm:hidden"
        // The last item must not sit under a home indicator.
        style={{ paddingBottom: "env(safe-area-inset-bottom)" }}
      >
        <ul className="mx-auto flex max-w-2xl items-stretch justify-between gap-2 px-2 py-1">
          {entries.map((entry) => (
            <li key={entry.to} className="flex-1">
              <NavLink entry={entry} className="min-h-12 w-full flex-col gap-1 px-1 py-2" />
            </li>
          ))}
        </ul>
      </nav>
    </>
  );
}

/**
 * One entry, in either bar.
 *
 * The active entry carries `aria-current="page"`, the `primary` tint **and** a bold
 * label: colour is never the only signal (05-design), and sage against cream is
 * exactly the axis red-green colour blindness sits on.
 */
function NavLink({ entry, className }: { entry: NavEntry; className: string }) {
  const Icon = entry.icon;

  return (
    <Link
      // The flag-gated entries point at routes F7 and F9 will add. They are filtered
      // out until then, so the cast is about the type of a list that outlives this
      // story rather than about a link a guest can reach.
      to={entry.to as never}
      className={cn(
        "text-ink-muted text-small focus-visible:ring-ring flex items-center justify-center rounded-lg text-center focus-visible:ring-2 focus-visible:outline-none",
        className,
      )}
      activeProps={{ className: "bg-primary-soft text-primary-hover font-semibold", "aria-current": "page" }}
    >
      <Icon className="size-5 shrink-0" aria-hidden="true" />
      <span>{entry.label}</span>
    </Link>
  );
}
