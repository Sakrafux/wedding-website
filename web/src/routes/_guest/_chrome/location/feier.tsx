import { createFileRoute } from "@tanstack/react-router";

import { InfoSection } from "@/components/layout/InfoSection";
import { VenueDetail } from "@/components/layout/VenueDetail";
import { locationLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/location/feier")({
  component: PartyPage,
});

/**
 * The party: address, map, parking, how to get there — and where to sleep.
 *
 * PLACEHOLDER: the venue is not booked (TODO.md).
 */
function PartyPage() {
  return (
    <VenueDetail title={locationLabels.partyHeading}>
      <InfoSection id="uebernachtung" title={locationLabels.accommodationHeading}>
        <p>{locationLabels.accommodationOpen}</p>
      </InfoSection>
    </VenueDetail>
  );
}
