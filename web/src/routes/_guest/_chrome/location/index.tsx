import { createFileRoute, Link } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";

import { InfoSection, PageHeading, PageSections } from "@/components/layout/InfoSection";
import { locationLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/location/")({
  component: LocationOverviewPage,
});

/**
 * Where the two places are, in one line each, and the trip between them.
 *
 * An overview and two detail pages rather than one long page (F2-F09): a guest looking
 * for the church address had to read past the party's car park to find it, and the two
 * sets of facts are exactly similar enough to be mixed up. The transfer section stays
 * here, because it belongs to neither venue.
 *
 * PLACEHOLDER: the venues are not booked (TODO.md). None of these three pages may go
 * out with the invitations in this state — 07-roadmap accepts a thin Ablauf page at
 * launch and explicitly does not accept an empty Location page.
 */
function LocationOverviewPage() {
  return (
    <PageSections>
      <PageHeading title={locationLabels.heading} lead={locationLabels.intro} />

      <InfoSection id="orte" title={locationLabels.venuesHeading}>
        <VenueTeaser
          title={locationLabels.churchHeading}
          teaser={locationLabels.churchTeaser}
          linkLabel={locationLabels.churchDetailLink}
          to="/location/kirche"
        />
        <VenueTeaser
          title={locationLabels.partyHeading}
          teaser={locationLabels.partyTeaser}
          linkLabel={locationLabels.partyDetailLink}
          to="/location/feier"
        />
      </InfoSection>

      <InfoSection id="fahrt" title={locationLabels.transferHeading}>
        <p>{locationLabels.transfer}</p>
      </InfoSection>
    </PageSections>
  );
}

/**
 * One venue in one line, and the way to its page.
 *
 * The whole card is not the link: the link is a named target with its own words, which
 * is what a screen reader reads out of a list of two and what a thumb aims at.
 */
function VenueTeaser({
  title,
  teaser,
  linkLabel,
  to,
}: {
  title: string;
  teaser: string;
  linkLabel: string;
  to: "/location/kirche" | "/location/feier";
}) {
  return (
    <div className="border-line flex flex-col gap-2 rounded-xl border p-4">
      <h3 className="text-h3 font-body">{title}</h3>
      <p className="text-ink-muted">{teaser}</p>
      <Link to={to} className="text-primary inline-flex min-h-12 items-center gap-1 underline underline-offset-4">
        {linkLabel}
        <ChevronRight className="size-4" aria-hidden="true" />
      </Link>
    </div>
  );
}
