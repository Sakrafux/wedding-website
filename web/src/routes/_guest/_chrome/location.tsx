import { createFileRoute } from "@tanstack/react-router";
import { ExternalLink } from "lucide-react";

import { InfoSection, PageHeading, PageSections } from "@/components/layout/InfoSection";
import { locationLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/location")({
  component: LocationPage,
});

/**
 * Both venues, how to get there, and where to sleep — on one page.
 *
 * One page and not three: a guest asking "where is it" wants the address, the parking
 * and the train in the same scroll. The sections carry `id` anchors so the nav or an
 * FAQ answer can link straight at one.
 *
 * PLACEHOLDER: the venues are not booked (TODO.md). This page must not go out with the
 * invitations in this state — 07-roadmap accepts a thin Ablauf page at launch and
 * explicitly does not accept an empty Location page.
 */
function LocationPage() {
  return (
    <PageSections>
      <PageHeading title={locationLabels.heading} lead={locationLabels.intro} />

      <InfoSection id="orte" title={locationLabels.venuesHeading}>
        <VenueBlock title={locationLabels.churchHeading} />
        <VenueBlock title={locationLabels.partyHeading} />
      </InfoSection>

      <InfoSection id="anreise" title={locationLabels.arrivalHeading}>
        <p>{locationLabels.arrivalOpen}</p>
      </InfoSection>

      <InfoSection id="fahrt" title={locationLabels.transferHeading}>
        <p>{locationLabels.transfer}</p>
      </InfoSection>

      <InfoSection id="uebernachtung" title={locationLabels.accommodationHeading}>
        <p>{locationLabels.accommodationOpen}</p>
      </InfoSection>
    </PageSections>
  );
}

/**
 * One venue: the address, and a link that opens it in a map.
 *
 * Both venues render through this component — two places with the same set of facts,
 * and two hand-written blocks are two places for a postcode to be wrong.
 *
 * The map is a **plain external link**, never an embedded frame: an iframe is a third
 * party in the critical path, it needs a `frame-src` hole in the CSP (E0-07), and it
 * hands every guest's IP to a mapping company — for a link that does the same job.
 *
 * `mapUrl` is optional only while the venues are unbooked; the address block ships
 * with the address (TODO.md).
 */
function VenueBlock({ title, address, mapUrl }: { title: string; address?: string[]; mapUrl?: string }) {
  return (
    <div className="border-line flex flex-col gap-2 rounded-xl border p-4">
      <h3 className="text-h3 font-body">{title}</h3>

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
    </div>
  );
}
