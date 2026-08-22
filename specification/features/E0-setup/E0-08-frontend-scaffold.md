# `E0-08` — Frontend scaffold and design tokens

**Epic:** E0 — Project setup · **Layer:** frontend · **Depends on:** `E0-01`

## Story

As a developer, I want the React app scaffolded with the design tokens already in place, so that no component is ever written against ad-hoc colours and spacing that have to be unpicked later.

## Scope

**In:**

- Vite + React + TypeScript in `web/`, strict mode on.
- Tailwind configured with the tokens from [05-design](../../05-design.md).
- shadcn/ui initialised; `Button`, `Input`, `Label`, `Card` pulled in as the first components.
- Self-hosted fonts.
- `web/src/lib/labels.ts` with the German enum label map.
- TanStack Router and TanStack Query installed and wired to an empty root route.

**Out:**

- `go:embed` and the SPA fallback → `E0-09`.
- Any real screen → the feature epics.

## Instructions

1. **pnpm is the package manager.** Enable it with `corepack enable` and pin the exact version in `package.json` via `packageManager`, so the Docker build stage and your machine resolve `pnpm-lock.yaml` identically. Commit the lockfile. Never mix in an `npm install` — a stray `package-lock.json` alongside a `pnpm-lock.yaml` is a slow, confusing failure.
2. `tsconfig` with `strict: true`, `noUncheckedIndexedAccess: true`. Turning these on later means fixing a hundred call sites at the worst possible time.
3. Tailwind theme extension carries the token names verbatim — `paper`, `ink`, `ink-muted`, `line`, `primary`, `primary-hover`, `primary-soft`, `accent`, `accent-strong`, `accent-soft`, `success`, `warning`, `danger`. **Components reference token names, never hex values.**
4. Base font size 18px. Set it on `html` in rem terms so browser zoom and OS font scaling still work.
5. Download Cormorant Garamond and Source Serif 4 as `.woff2`, subset to Latin, into `web/src/assets/fonts/`, and declare them with `@font-face` and `font-display: swap`. **No Google Fonts link tag** — it is a third-party request and the CSP forbids it.
6. `labels.ts` transcribes the full German label map from the design document as `const` objects keyed by enum type, typed so that a missing enum value is a compile error. That last part is the whole value: adding an enum variant should not silently render as a raw English string.
7. Global focus-visible style: 2px `primary` outline, 2px offset. Set it once, globally, so no component can forget it.
8. `.gitignore` for `web/node_modules` and `web/dist`. `pnpm-lock.yaml` is committed.

## Test plan

- [ ] `pnpm build` produces a `dist/` with no type errors.
- [ ] Type check: removing a key from a label map fails compilation.
- [ ] Manual: fonts load from the local origin — no network request to a third party in the Network tab.
- [ ] Manual: a tabbed button shows the focus ring.

## Done when

- [ ] A placeholder page renders in the project's typography and colours, not Vite's defaults.
- [ ] Checkbox ticked in `README.md`.
