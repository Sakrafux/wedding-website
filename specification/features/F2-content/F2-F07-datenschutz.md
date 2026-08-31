# `F2-F07` — Datenschutz

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F05`, `F2-F06`

## Story

As a guest, I want a short page in plain German saying what you store about me, why, who sees it and how to get it changed, so that I can decide what to put in the form.

## Scope

**In:**

- `/datenschutz`, behind the household login with the other content pages.
- Plain-language sections: what we store, why, who sees it, how long, and how to reach us to change or delete it.
- The link to it from `/mehr`.

**Out:**

- An Impressum. [06-privacy-security](../../06-privacy-security.md) records why: the German imprint duty attaches to business-like telemedia, and a private wedding invitation behind a login is not that. Contact details are on `/kontakt` anyway.
- Any consent banner. There is nothing to consent to — no analytics, no CDN, no third-party fonts, no processors.

## Instructions

1. The route lives **inside** `/_guest`, with the other content pages. It was built outside the layout first, because this story and [06-privacy-security](../../06-privacy-security.md) both called for a page readable without logging in — "a privacy notice you have to authenticate to read is not a notice". **Reversed on 2026-08-31**, with the page in front of us: nobody without a code has any data described here, the guests it is about are logged in at the moment they give us that data, and the page was the only one in the product rendering without the navigation, which read as a different site rather than as a notice. The spec sentence was changed to match.
2. No link from the login screen. There is nothing there for somebody who has not typed a code, and the login screen already names a phone number for anybody who wants to ask.
3. Content is derived from [06-privacy-security](../../06-privacy-security.md), not invented here: the "What we hold" table, the retention rows, and the informal-rights paragraph. It is a translation of that document into guest German, and when the two disagree, the spec is what changes first.
4. Same register as the rest of the site: informal "du", short sentences, no legalese. [06-privacy-security](../../06-privacy-security.md) says so outright — a wall of boilerplate here would be worse than nothing, because nobody reads it and it makes the site feel like a business.
5. Say the things that are actually true and unusual, and say them plainly: everything runs on our own server, there is no third-party service involved, nothing is public or searchable, and the data is deleted after the wedding. That is a stronger statement than any list of legal bases, and it happens to be the selling point.
6. Name allergies and children's ages explicitly. They are the two fields a guest might hesitate over, and "we pass this to the caterer" is the honest reason they are asked for.
7. Rights are handled informally: ask us, we do it. Say that, with the contact route from `F2-F06` — do not describe a process that does not exist.
8. State the retention plainly: the site goes offline roughly three months after the wedding, and the data goes with it, except the archive we keep for ourselves.
9. `InfoSection` from `F2-F05` renders it. This page is prose; it does not get its own layout.
10. **Open input, tracked in [TODO.md](../../../TODO.md):** the German text itself is unwritten, and the photo-retention question is unanswered — the page cannot state a gallery lifetime until that decision is made. Write the section without a number rather than promising one we have not chosen.

## Test plan

- [ ] Component: `renderApp("/datenschutz")` renders the page with the guest navigation around it.
- [ ] Component: a visit without a session is sent to the login screen, like every other guest route.
- [ ] Component: `/mehr` links to it.
- [ ] Component: the page names allergies, children's ages, and the retention period.
- [ ] Accessibility: one `<h1>`, headings in order, prose at the capped measure, and the contact link's accessible name says where it goes.

## Done when

- [ ] A logged-in guest can read what we store about them and how to have it removed, without leaving the site's navigation.
- [ ] Checkbox ticked in `README.md`.
