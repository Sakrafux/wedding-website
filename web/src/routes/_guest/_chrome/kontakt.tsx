import { createFileRoute } from "@tanstack/react-router";

import { InfoSection, PageHeading, PageSections } from "@/components/layout/InfoSection";
import { contactLabels, contacts } from "@/lib/labels";

export const Route = createFileRoute("/_guest/_chrome/kontakt")({
  component: ContactPage,
});

/**
 * How to reach a human.
 *
 * No contact form: there is no mail path in this application and no SMTP dependency,
 * and a form that silently goes nowhere is worse than a phone number. No Impressum
 * either — 06-privacy-security records why a private, login-gated wedding invitation
 * is not "geschäftsmäßig" telemedia, and adding one would put a home address on a page
 * that does not need it.
 *
 * Both numbers, each labelled with whose it is: a guest ringing about a lost code and
 * a guest ringing about an allergy want different people.
 */
function ContactPage() {
  return (
    <PageSections>
      <PageHeading title={contactLabels.heading} lead={contactLabels.intro} />

      <ul className="flex flex-col gap-4">
        {contacts.map((contact) => (
          <li key={contact.phone} className="flex flex-col gap-1">
            <span>{contact.name}</span>
            {/* A tel: link: on a phone, the fallback for a guest who cannot log in
                should be one tap rather than a transcription. */}
            <a href={`tel:${contact.phone.replaceAll(" ", "")}`} className="text-primary text-h3 underline">
              {contact.phone}
            </a>
          </li>
        ))}
      </ul>

      {/* Repeated here rather than linked: the guest reading this may have got to it
          from somebody else's phone, because their own code did not work. */}
      <InfoSection title={contactLabels.codeHelpHeading}>
        <p>{contactLabels.codeHelp}</p>
      </InfoSection>
    </PageSections>
  );
}
