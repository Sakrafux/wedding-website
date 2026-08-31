import { createFileRoute } from "@tanstack/react-router";

import { PageHeading, PageSections } from "@/components/layout/InfoSection";
import { dresscodeLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/dresscode")({
  component: DresscodePage,
});

/**
 * What to wear, in plain words.
 *
 * No image: the page answers a question in a sentence, and a decorative photo would
 * push the answer below the fold.
 *
 * PLACEHOLDER: the wording is unwritten (TODO.md). What is decided is its shape — an
 * example rather than a category name, because "festlich" means five different things
 * to five relatives.
 */
function DresscodePage() {
  return (
    <PageSections>
      <PageHeading title={dresscodeLabels.heading} lead={dresscodeLabels.lead} />
      <p className="max-w-prose">{dresscodeLabels.body}</p>
    </PageSections>
  );
}
