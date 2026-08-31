import { createFileRoute, Link } from "@tanstack/react-router";

import { PageHeading, PageSections } from "@/components/layout/InfoSection";
import { faqLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/faq")({
  component: FAQPage,
});

/**
 * The questions we already get asked.
 *
 * **All answers are expanded — no accordion.** A collapsed answer hides its text from
 * the browser's own find-in-page, costs a tap per question, and this audience reads a
 * page rather than operating a widget. Every question is a real heading, so the list
 * is navigable by heading and linkable by anchor.
 *
 * An answer whose detail lives on a content page links there instead of restating it:
 * a copy of the dress code here would be a second copy to keep in step.
 */
function FAQPage() {
  return (
    <PageSections>
      <PageHeading title={faqLabels.heading} lead={faqLabels.intro} />

      <div className="flex flex-col gap-8">
        {faqLabels.entries.map((entry) => (
          <section key={entry.question} className="flex max-w-prose flex-col gap-2">
            <h2 className="text-h3 font-body">{entry.question}</h2>
            <p>{entry.answer}</p>
            {entry.link ? (
              <Link to={entry.link.to} className="text-primary underline">
                {entry.link.label}
              </Link>
            ) : null}
          </section>
        ))}
      </div>
    </PageSections>
  );
}
