import { createFileRoute } from "@tanstack/react-router";

import { PageHeading } from "@/components/layout/InfoSection";
import { scheduleLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/ablauf")({
  component: SchedulePage,
});

/**
 * The day, in order.
 *
 * One vertical timeline on every screen size: a two-sided alternating timeline is
 * decorative and halves the line length on the device most guests use. The entries are
 * data in `labels.ts` rendered by one component — twelve copies of the same
 * three-element block is where a typo hides.
 */
function SchedulePage() {
  return (
    <div className="flex flex-col gap-8">
      <PageHeading title={scheduleLabels.heading} lead={scheduleLabels.intro} />

      <ol className="flex flex-col gap-6">
        {scheduleLabels.entries.map((entry) => (
          <li key={entry.title} className="border-line flex flex-col gap-1 border-l-2 pl-4">
            <div className="flex flex-wrap items-baseline gap-3">
              {/* Tabular figures so the times line up down the column; an entry whose
                  time is not fixed says so rather than showing an empty gap or a
                  stray dash — a time on this page will be believed. */}
              <span className="text-ink-muted text-small tabular-nums">{entry.time ?? scheduleLabels.timeOpen}</span>
              <h2 className="text-h3 font-body">{entry.title}</h2>
              <PartMarker part={entry.part} />
            </div>
            <p className="text-ink-muted max-w-prose">{entry.detail}</p>
          </li>
        ))}
      </ol>
    </div>
  );
}

/**
 * Which half of the day an entry belongs to, so a guest can see at a glance what their
 * own Zusage covers.
 *
 * The timeline is deliberately **not** filtered by the household's answer: somebody
 * coming only to the church still wants to know that a party happens, and hiding it
 * would read as a mistake rather than as tact.
 */
function PartMarker({ part }: { part: string }) {
  const label = part === "church" ? scheduleLabels.church : scheduleLabels.party;

  return <span className="border-line text-ink-muted text-small rounded-lg border px-2 py-0.5">{label}</span>;
}
