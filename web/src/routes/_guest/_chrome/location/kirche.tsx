import { createFileRoute } from "@tanstack/react-router";

import { VenueDetail } from "@/components/layout/VenueDetail";
import { locationLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/location/kirche")({
  component: ChurchPage,
});

/**
 * The church: address, map, parking, how to get there.
 *
 * No accommodation section — hotels belong beside the place the evening ends, which is
 * the party page (F2-F09).
 *
 * PLACEHOLDER: the venue is not booked (TODO.md).
 */
function ChurchPage() {
  return <VenueDetail title={locationLabels.churchHeading} />;
}
