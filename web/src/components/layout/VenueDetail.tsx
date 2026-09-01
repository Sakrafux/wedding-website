import { Link } from "@tanstack/react-router";
import { ExternalLink } from "lucide-react";

import { locationLabels } from "@/lib/labels";

import { InfoSection, PageHeading, PageSections } from "./InfoSection";

/**
 * One venue's page: the address, a link that opens it in a map, and how to get there.
 *
 * Both venue pages render through this component — two places with the same set of
 * facts, and two hand-written pages are two places for a postcode to be wrong. What
 * differs goes in `children`: only the party page carries Übernachtung (F2-F09).
 *
 * The map is a **plain external link**, never an embedded frame: an iframe is a third
 * party in the critical path, it needs a `frame-src` hole in the CSP (E0-07), and it
 * hands every guest's IP to a mapping company — for a link that does the same job.
 *
 * `address` and `mapUrl` are optional only while the venues are unbooked; both pages
 * ship with the address (TODO.md).
 */
export function VenueDetail({
  title,
  address,
  mapUrl,
  children,
}: {
  title: string;
  address?: string[];
  mapUrl?: string;
  children?: React.ReactNode;
}) {
  return (
    <PageSections>
      <PageHeading title={title} />

      <InfoSection id="adresse" title={locationLabels.addressHeading}>
        {address ? (
          <address className="text-ink-muted not-italic">
            {address.map((line) => (
              <span key={line} className="block">
                {line}
              </span>
            ))}
          </address>
        ) : (
          <p className="text-ink-muted">{locationLabels.addressOpen}</p>
        )}

        {mapUrl ? (
          // Marked as external in the text as well as by the icon: nobody in this
          // audience should be surprised by a new tab.
          <a
            href={mapUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary inline-flex items-center gap-2 underline"
          >
            <ExternalLink className="size-4" aria-hidden="true" />
            <span>
              {locationLabels.mapLink} {locationLabels.externalHint}
            </span>
          </a>
        ) : null}
      </InfoSection>

      <InfoSection id="anreise" title={locationLabels.arrivalHeading}>
        <p>{locationLabels.arrivalOpen}</p>
      </InfoSection>

      {children}

      {/* The way back, in words: these pages are reached from the overview and not from
          the bottom bar, so the bar cannot bring a guest back to where they were. */}
      <Link to="/location" className="text-primary min-h-12 underline underline-offset-4">
        {locationLabels.backToOverview}
      </Link>
    </PageSections>
  );
}
