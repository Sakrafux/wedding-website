import { createFileRoute, Link } from "@tanstack/react-router";

import { InfoSection, PageHeading, PageSections } from "@/components/layout/InfoSection";
import { privacyLabels } from "@/lib/labels";

/**
 * What we store, why, who sees it, and how to have it changed.
 *
 * Behind the household login like every other page, reached from `/mehr`.
 *
 * It sat outside the guest layout until 2026-08-31, on the grounds that a notice you
 * have to authenticate to read is not a notice. Reversed deliberately: the site is one
 * invitation to eighty known guests, nobody without a code has any data described
 * here, and a page rendering without the navigation looked like a different site. The
 * people this notice is about are logged in at the moment they give us the data.
 *
 * The text is a translation of 06-privacy-security into guest German, not a new
 * document — when the two disagree, the spec changes first. It says the things that
 * are actually true and unusual (our own server, no third parties, nothing public,
 * deleted afterwards), which is a stronger statement than any list of legal bases.
 */
export const Route = createFileRoute("/_guest/_chrome/datenschutz")({
  component: PrivacyPage,
});

function PrivacyPage() {
  return (
    <PageSections>
      <PageHeading title={privacyLabels.heading} lead={privacyLabels.lead} />

      <InfoSection title={privacyLabels.storedHeading}>
        <p>{privacyLabels.stored}</p>
      </InfoSection>

      {/* Named explicitly, because they are the two fields a guest might hesitate over,
          and the caterer is the honest reason they are asked for. */}
      <InfoSection title={privacyLabels.sensitiveHeading}>
        <p>{privacyLabels.sensitive}</p>
      </InfoSection>

      <InfoSection title={privacyLabels.whyHeading}>
        <p>{privacyLabels.why}</p>
      </InfoSection>

      <InfoSection title={privacyLabels.whoHeading}>
        <p>{privacyLabels.who}</p>
      </InfoSection>

      <InfoSection title={privacyLabels.retentionHeading}>
        <p>{privacyLabels.retention}</p>
      </InfoSection>

      {/* Rights are handled informally: ask us, we do it. Describing a process that
          does not exist would be worse than saying what actually happens. */}
      <InfoSection title={privacyLabels.rightsHeading}>
        <p>{privacyLabels.rights}</p>
        <Link to="/kontakt" className="text-primary underline">
          {privacyLabels.contactLink}
        </Link>
      </InfoSection>
    </PageSections>
  );
}
