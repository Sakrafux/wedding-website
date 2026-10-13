import { ExternalLink } from "lucide-react";

import { locationLabels } from "@/lib/labels";

/**
 * A link that opens a place in a map, in a new tab.
 *
 * One component because three screens link the same way — both venue pages and the
 * transfer section on the overview — and three copies of the markup would be three
 * places for the hint text to drift.
 */
export function MapLink({ url }: { url: string }) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="text-primary inline-flex items-center gap-2 underline"
    >
      {/* Marked as external in the text as well as by the icon: nobody in this
          audience should be surprised by a new tab. */}
      <ExternalLink className="size-4" aria-hidden="true" />
      <span>
        {locationLabels.mapLink} {locationLabels.externalHint}
      </span>
    </a>
  );
}
