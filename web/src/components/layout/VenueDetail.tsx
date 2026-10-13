import { Link } from "@tanstack/react-router";

import { locationLabels } from "@/lib/labels";

import { InfoSection, PageHeading, PageSections } from "./InfoSection";
import { MapLink } from "./MapLink";

/**
 * One venue's page: the address, a link that opens it in a map, and how to get there.
 *
 * Both venue pages render through this component — two places with the same set of
 * facts, and two hand-written pages are two places for a postcode to be wrong. What
 * differs goes in `children`: only the party page carries Abreise and Übernachtung
 * (F2-F09).
 *
 * The map is a **plain external link**, never an embedded frame: an iframe is a third
 * party in the critical path, it needs a `frame-src` hole in the CSP (E0-07), and it
 * hands every guest's IP to a mapping company — for a link that does the same job.
 * Reaffirmed 2026-10-14, when the venue pages got their real addresses and maps.
 */
export function VenueDetail({
  title,
  address,
  mapUrl,
  arrival,
  children,
}: {
  title: string;
  /** One string per rendered line. */
  address: readonly string[];
  mapUrl: string;
  /** One string per rendered paragraph. */
  arrival: readonly string[];
  children?: React.ReactNode;
}) {
  return (
    <PageSections>
      <PageHeading title={title} />

      <InfoSection id="adresse" title={locationLabels.addressHeading}>
        <address className="text-ink-muted not-italic">
          {address.map((line) => (
            <span key={line} className="block">
              {line}
            </span>
          ))}
        </address>

        <MapLink url={mapUrl} />
      </InfoSection>

      <InfoSection id="anreise" title={locationLabels.arrivalHeading}>
        {arrival.map((paragraph) => (
          <p key={paragraph}>{paragraph}</p>
        ))}
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
