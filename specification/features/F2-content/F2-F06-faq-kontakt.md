# `F2-F06` — FAQ, Kontakt

**Epic:** F2 — Informational content · **Layer:** frontend · **Depends on:** `F2-F05`

## Story

As a guest, I want the questions other people have already asked answered on one page, and a way to reach you when mine is not there, so that I am never stuck.

## Scope

**In:**

- `/faq` and `/kontakt`, both built on `InfoSection` from `F2-F05`.
- The FAQ entries as data: question, answer, and where relevant a link to the page that has the detail.
- Contact: our names and the phone number, plus what to do when the login code does not work.

**Out:**

- A contact form. There is no mail path in this application and no SMTP dependency ([01-vision-scope](../../01-vision-scope.md)); a form that silently goes nowhere is worse than a phone number.
- The Datenschutz page → `F2-F07`.

## Instructions

1. FAQ entries are an array of question/answer pairs in the module, rendered by one component — same reasoning as the timeline in `F2-F03`.
2. **All answers are expanded. No accordion.** Rejected deliberately: a collapsed answer hides the text from the browser's own find-in-page, costs a tap for every question, and this audience reads a page rather than operating a widget. Ten questions of three lines each is a scroll, not a problem.
3. Each question is a real heading, so the list is navigable by heading with a screen reader and linkable by anchor. An FAQ answer that duplicates a content page links to it instead of restating it — a copy of the dress code here is a second copy to keep in step.
4. Seed the list from what we actually get asked: children, arriving late, parking, what to wear, the gift question, how the code works, what happens if two households share a car.
5. Kontakt names both of us and gives the phone number. No Impressum — [06-privacy-security](../../06-privacy-security.md) records why a private, login-gated wedding invitation is not "geschäftsmäßig" telemedia, and adding one anyway would put a home address on a page that does not need it.
6. **Two numbers, both on this page**, from `labels.ts`: **Andreas Hell — +43 650 9408100** and **Isabella Michelbacher — +43 677 63668655**, each labelled with whose it is — a guest ringing about a lost code and a guest ringing about a dietary question want different people. `contactPhoneNumber` stays the single constant the login fallback uses (`F11-02`) and is Andreas' number; this page renders both from the same file. No number is written inline in a component.
7. Render it as a `tel:` link. On a phone, the fallback for a guest who cannot log in should be one tap, not a transcription.
8. Say on the Kontakt page what to do when the code does not work, because the guest reading it may have got here from somebody else's phone. Repeat the phone number there rather than linking to it.
9. **Open input, tracked in [TODO.md](../../../TODO.md):** the final FAQ list, which is partly a guess until the invitations go out and the first questions actually arrive. The phone numbers and names are decided (above).

## Test plan

- [ ] Component: `renderApp("/faq")` renders every entry, with all answers visible without interaction.
- [ ] Component: an FAQ answer that links to a content page navigates there.
- [ ] Component: `renderApp("/kontakt")` shows the phone number as a `tel:` link, and it is the same value the login screen uses.
- [ ] Component: both numbers render, each with whose it is, as `tel:` links.
- [ ] Accessibility: FAQ questions are headings in order, and every link's accessible name says where it goes.

## Done when

- [ ] A guest with a question that is not on the site can reach a human in one tap.
- [ ] Checkbox ticked in `README.md`.
