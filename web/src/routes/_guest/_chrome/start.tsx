import { useQuery, useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";

import heroPhotoJpg from "@/assets/hero.jpg";
import heroPhotoWebp from "@/assets/hero.webp";
import { Button } from "@/components/ui/button";
import { rsvpQueryOptions } from "@/lib/api/rsvp";
import { meQueryOptions } from "@/lib/api/session";
import { formatShortDate } from "@/lib/dates";
import { startLabels } from "@/lib/labels";
import { daysUntilWedding, weddingDateLong } from "@/lib/wedding";

export const Route = createFileRoute("/_guest/_chrome/start")({
  component: StartPage,
});

/**
 * The first screen after logging in: the photo, our names, the date, how long it is
 * until the wedding, and the one thing we need the household to do.
 *
 * One primary action. A start page with six equal buttons is a menu with a photo on
 * it, and the nav already carries the secondary links.
 */
function StartPage() {
  const { data: me } = useSuspenseQuery(meQueryOptions);
  // Not a suspense query and not a loader: the start page is what a guest lands on,
  // and it must render the photo, the date and the countdown even if the RSVP request
  // is slow or fails. Until it answers, the call to action reads "Jetzt zusagen" —
  // the safe half of the pair.
  const { data: answer } = useQuery(rsvpQueryOptions);
  const hasAnswered = answer?.household.rsvp_submitted_at != null;

  // The layout's guard has already redirected a caller with no session; this is the
  // type narrowing, not a second check.
  if (!me) {
    return null;
  }

  return (
    <div className="flex flex-col gap-8">
      <HeroSection />

      <div className="flex flex-col gap-3">
        <p className="text-h3">{startLabels.greeting(me.household.display_name)}</p>
        <p className="text-ink-muted">{startLabels.intro}</p>
      </div>

      <div className="flex flex-col gap-3">
        <Button asChild size="lg" className="h-14 w-full">
          <Link to="/zusagen">{hasAnswered ? startLabels.rsvpCallToActionAnswered : startLabels.rsvpCallToAction}</Link>
        </Button>

        {/* The deadline is a written-out date beside the answer link, never a second
            countdown: two counters compete for the same attention and the urgent one
            loses to the bigger number. */}
        <p className="text-ink-muted text-small">
          {hasAnswered ? `${startLabels.rsvpAnswered} ` : ""}
          {/* From `me`, which every guest page already has: moving the setting moves
              this sentence, and the page does not wait on the RSVP request to say it. */}
          {startLabels.rsvpDeadline(formatShortDate(me.rsvp_deadline))}
        </p>
      </div>
    </div>
  );
}

/**
 * The hero: photo, names, date.
 *
 * No venue town — the venues have their own page, and a start page that names a town
 * invites a guest to plan travel from the one line with no address on it.
 *
 * The photo is `alt=""`: it is decoration beside text that already says who and when,
 * and a description of our own engagement photo helps nobody. It is not lazy-loaded
 * either — it is the first thing above the fold, and `loading="lazy"` there is a
 * visible delay for a guest on a train.
 */
function HeroSection() {
  return (
    <section className="relative overflow-hidden rounded-xl">
      <picture>
        <source srcSet={heroPhotoWebp} type="image/webp" />
        <img
          src={heroPhotoJpg}
          alt=""
          // Explicit dimensions, so the text below does not jump when the image lands.
          width={1200}
          height={1500}
          className="aspect-[4/5] w-full object-cover sm:aspect-[16/9]"
        />
      </picture>

      {/* A warm scrim rather than a tint on the image itself: the display text has to
          keep its contrast whatever the final photo turns out to look like. */}
      <div className="absolute inset-0 flex flex-col items-center justify-end gap-2 bg-gradient-to-t from-[rgb(58_50_38_/_0.72)] to-transparent p-6 text-center text-[color:var(--color-paper)]">
        <h1 className="text-display font-display">{startLabels.names}</h1>
        <p className="text-small">{weddingDateLong}</p>
        <CountdownBadge />
      </div>
    </section>
  );
}

/**
 * Days until the wedding — three states, all of them written out: the count, "heute"
 * on the day, and nothing at all afterwards. A badge reading "-3 Tage" in August 2027
 * is the kind of thing nobody tests and everybody sees.
 *
 * Days, never seconds: a live counter is motion for nothing (05-design), and it is
 * computed against local midnight so two guests in the same room are told the same
 * number either side of midnight.
 */
function CountdownBadge() {
  const days = daysUntilWedding();

  if (days < 0) {
    return null;
  }

  return (
    <p className="bg-surface/90 text-ink text-small rounded-lg px-3 py-1">
      {days === 0 ? startLabels.countdownToday : startLabels.countdown(days)}
    </p>
  );
}
