/**
 * The page shape every informational page is built from: a heading, an optional lead
 * sentence, and prose capped at the readable measure.
 *
 * It exists so `F2-F05`'s two pages, the FAQ, Kontakt and Datenschutz are one layout
 * decision rather than five. Prose is `max-w-prose` — about 66 characters — because a
 * line that runs the full width of a desktop is a line people lose their place in.
 */
export function PageHeading({ title, lead }: { title: string; lead?: string }) {
  return (
    <header className="flex flex-col gap-3">
      <h1 className="text-h1 font-display">{title}</h1>
      {lead ? <p className="text-ink-muted max-w-prose">{lead}</p> : null}
    </header>
  );
}

/** One titled block of prose inside a page. `id` is what an FAQ answer or the nav
    links straight at. */
export function InfoSection({ id, title, children }: { id?: string; title: string; children: React.ReactNode }) {
  return (
    <section id={id} className="flex flex-col gap-3">
      <h2 className="text-h2 font-display">{title}</h2>
      <div className="flex max-w-prose flex-col gap-3">{children}</div>
    </section>
  );
}

/** The vertical rhythm between sections, in one place: 48px on mobile, 64px up. */
export function PageSections({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-col gap-12 sm:gap-16">{children}</div>;
}
