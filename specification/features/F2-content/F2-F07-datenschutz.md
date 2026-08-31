# `F2-F07` — Datenschutz

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F05`, `F2-F06`

## Story

As a guest, I want a short page in plain German saying what you store about me, why, who sees it and how to get it changed, so that I can decide what to put in the form.

## Scope

**In:**

- `/datenschutz`, reachable **without logging in**.
- Plain-language sections: what we store, why, who sees it, how long, and how to reach us to change or delete it.
- Links to it from `/mehr` and from the login screen.

**Out:**

- An Impressum. [06-privacy-security](../../06-privacy-security.md) records why: the German imprint duty attaches to business-like telemedia, and a private wedding invitation behind a login is not that. Contact details are on `/kontakt` anyway.
- Any consent banner. There is nothing to consent to — no analytics, no CDN, no third-party fonts, no processors.

## Instructions

1. The route lives **outside** the `/_guest` layout, next to the login screen, because [06-privacy-security](../../06-privacy-security.md) requires it to be readable without logging in. A privacy notice you have to authenticate to read is not a notice.
2. Add the link to the login screen too, small, below the form. Somebody who has not typed their code yet is exactly the person who might want to know what happens to it.
3. Content is derived from [06-privacy-security](../../06-privacy-security.md), not invented here: the "What we hold" table, the retention rows, and the informal-rights paragraph. It is a translation of that document into guest German, and when the two disagree, the spec is what changes first.
4. Same register as the rest of the site: informal "du", short sentences, no legalese. [06-privacy-security](../../06-privacy-security.md) says so outright — a wall of boilerplate here would be worse than nothing, because nobody reads it and it makes the site feel like a business.
5. Say the things that are actually true and unusual, and say them plainly: everything runs on our own server, there is no third-party service involved, nothing is public or searchable, and the data is deleted after the wedding. That is a stronger statement than any list of legal bases, and it happens to be the selling point.
6. Name allergies and children's ages explicitly. They are the two fields a guest might hesitate over, and "we pass this to the caterer" is the honest reason they are asked for.
7. Rights are handled informally: ask us, we do it. Say that, with the contact route from `F2-F06` — do not describe a process that does not exist.
8. State the retention plainly: the site goes offline roughly three months after the wedding, and the data goes with it, except the archive we keep for ourselves.
9. `InfoSection` from `F2-F05` renders it. This page is prose; it does not get its own layout.
10. **Open input, tracked in [TODO.md](../../../TODO.md):** the German text itself is unwritten, and the photo-retention question is unanswered — the page cannot state a gallery lifetime until that decision is made. Write the section without a number rather than promising one we have not chosen.

## Test plan

- [ ] Component: `renderApp("/datenschutz")` renders the page **without** a session, and the guard does not redirect to the login screen.
- [ ] Component: the login screen links to it, and so does `/mehr`.
- [ ] Component: the page names allergies, children's ages, and the retention period.
- [ ] Accessibility: one `<h1>`, headings in order, prose at the capped measure, and the contact link's accessible name says where it goes.

## Done when

- [ ] A guest who has not logged in can read what we store about them and how to have it removed.
- [ ] Checkbox ticked in `README.md`.
