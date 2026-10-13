import { createFileRoute } from "@tanstack/react-router";

import { InfoSection } from "@/components/layout/InfoSection";
import { VenueDetail } from "@/components/layout/VenueDetail";
import { locationLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/location/feier")({
  component: PartyPage,
});

/**
 * The party: address, map, parking — and how to get home and where to sleep, which are
 * this page's because the evening ends here (F2-F09).
 */
function PartyPage() {
  return (
    <VenueDetail
      title={locationLabels.partyHeading}
      address={locationLabels.partyAddress}
      mapUrl={locationLabels.partyMapUrl}
      arrival={locationLabels.partyArrival}
    >
      <InfoSection id="abreise" title={locationLabels.departureHeading}>
        {locationLabels.partyDeparture.map((paragraph) => (
          <p key={paragraph}>{paragraph}</p>
        ))}
      </InfoSection>

      <InfoSection id="uebernachtung" title={locationLabels.accommodationHeading}>
        {locationLabels.partyAccommodation.map((paragraph) => (
          <p key={paragraph}>{paragraph}</p>
        ))}
      </InfoSection>
    </VenueDetail>
  );
}
