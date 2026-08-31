import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";

import { InfoSection, PageHeading, PageSections } from "@/components/layout/InfoSection";
import { Button } from "@/components/ui/button";
import { giftLabels } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/geschenke")({
  component: GiftsPage,
});

/** The IBAN placeholder, which is not a real account yet (TODO.md). */
const ibanIsPending = giftLabels.iban.startsWith("AT0000");

/**
 * What we would like, and how to send it.
 *
 * The bank details are published deliberately: everything on a content page is
 * compiled into the JavaScript bundle, and the bundle is served to anyone who loads
 * the site — the session gate covers `/api`, not the SPA. An IBAN here is therefore
 * semi-public. Accepted on 2026-08-31 (F2-F05): an IBAN lets somebody send money, not
 * take it, the site is unindexed (E0-07), and "IBAN auf Anfrage" costs every guest a
 * phone call to do the thing we asked them to do. What would not be acceptable is
 * publishing it while believing it sits behind the login — hence this comment.
 */
function GiftsPage() {
  return (
    <PageSections>
      <PageHeading title={giftLabels.heading} lead={giftLabels.lead} />
      <p className="max-w-prose">{giftLabels.body}</p>

      <InfoSection title={giftLabels.accountHeading}>
        {ibanIsPending ? <p>{giftLabels.ibanPending}</p> : <BankDetails />}
      </InfoSection>
    </PageSections>
  );
}

/**
 * The account, with a copy button.
 *
 * Nobody should transcribe an IBAN from a phone screen by hand, so the grouped,
 * tabular rendering is display only and the clipboard gets the unspaced value — a
 * pasted IBAN with spaces in it is one a banking app rejects.
 */
function BankDetails() {
  const [isCopied, setIsCopied] = useState(false);

  async function copy() {
    await navigator.clipboard.writeText(giftLabels.iban);
    setIsCopied(true);
  }

  return (
    <div className="flex flex-col gap-2">
      <p>
        <span className="text-ink-muted text-small block">{giftLabels.accountHolderLabel}</span>
        {giftLabels.accountHolder}
      </p>

      <p>
        <span className="text-ink-muted text-small block">{giftLabels.ibanLabel}</span>
        <span className="tabular-nums">{groupIban(giftLabels.iban)}</span>
      </p>

      <Button type="button" variant="outline" className="self-start" onClick={() => void copy()}>
        {giftLabels.copyIban}
      </Button>

      {/* A live region rather than a label swap: the confirmation has to be announced,
          and a button whose name changes under a screen reader is a different button. */}
      <span role="status" className="text-ink-muted text-small">
        {isCopied ? giftLabels.ibanCopied : ""}
      </span>
    </div>
  );
}

/** Groups an IBAN in fours, the way it is printed. Display only. */
function groupIban(iban: string): string {
  return iban.replace(/(.{4})/g, "$1 ").trim();
}
