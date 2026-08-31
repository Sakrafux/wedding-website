# Frontend

React + TypeScript + Vite. The built `dist/` is embedded into the Go binary with `go:embed`, so there is one artifact and a frontend/backend version skew is impossible.

Setup and how to run the two dev servers together: [root README, "Local development"](../README.md#local-development). This file covers only what is specific to `web/`.

## Dependencies

Few and boring, same rule as the backend. Every entry below has to justify itself; anything that can be a hundred lines of our own code instead of a dependency, is.

### Runtime

| Package | Why |
|---|---|
| `react`, `react-dom` | The framework. |
| `@tanstack/react-router` | Type-safe routing. Route paths and params are checked at compile time, so a renamed route is a build error rather than a 404 a guest finds. File-based: `src/routes/` is the route table. |
| `@tanstack/react-query` | Server state: caching, in-flight dedupe, invalidation after a mutation. Chosen so that "refetch the RSVP after saving it" is one line and not a hand-rolled store. |
| `radix-ui` | Unstyled accessible primitives underneath the shadcn components — focus trapping, ARIA wiring, keyboard handling. This is the part that is genuinely hard to get right, and the accessibility requirements are not negotiable. |
| `lucide-react` | Icon set. Tree-shaken per icon. Needed because colour is never the only signal — every state marker carries an icon too. |
| `class-variance-authority` | Declares component variants (`Button` primary/secondary/ghost/danger) as data instead of nested ternaries. Required by shadcn's generated components. |
| `clsx`, `tailwind-merge` | The `cn()` helper: conditional class names, with a later Tailwind class correctly beating an earlier one of the same kind. Also required by shadcn. |

### Build and tooling

| Package | Why |
|---|---|
| `vite`, `@vitejs/plugin-react` | Dev server and production bundler. |
| `typescript` | `strict` and `noUncheckedIndexedAccess`, on from the first commit. |
| `tailwindcss`, `@tailwindcss/vite` | v4, configured in CSS via `@theme` — there is no `tailwind.config.js`. The design tokens live in `src/index.css`. |
| `@tanstack/router-plugin` | Generates `src/routeTree.gen.ts` from `src/routes/`. |
| `@tanstack/react-router-devtools`, `@tanstack/react-query-devtools` | The route-match and query-cache inspectors. Mounted from `src/components/Devtools.tsx` through a dynamic `import()`, so they exist in `pnpm dev` and are absent from the production bundle — verify with `grep -i devtools dist/assets/*.js` after a build. |
| `oxlint` | Linter. See below. |
| `prettier`, `prettier-plugin-tailwindcss` | Formatter; the plugin sorts Tailwind classes into a canonical order, which kills the class-order half of every review diff. |
| `vitest` | Test runner. Vite's own, so the tests resolve `@/` imports and transform TSX through the same pipeline the app is built with — a second bundler config is a second thing to keep in step. |
| `jsdom` | The DOM the component tests run against. |
| `@testing-library/react`, `@testing-library/user-event` | Renders components and drives them the way a guest does — by label text and button name, through real typing and clicking. Queries by accessible name, so a test that passes is also evidence the control is reachable by a screen reader. |
| `@testing-library/jest-dom` | The DOM matchers (`toBeDisabled`, `toHaveAccessibleDescription`). Registered in `src/test/setup.ts`. |
| `@types/*` | Type definitions. |

**No shadcn dependency.** `pnpm dlx shadcn@latest add <component>` copies component source into `src/components/ui/`. We own those files and edit them freely; the trade-off is that upstream fixes have to be pulled in by re-running `add --overwrite`.

**pnpm only.** Never `npm install` here: a `package-lock.json` beside `pnpm-lock.yaml` resolves to a different tree in the Docker build than on your machine, and the symptom surfaces as a runtime error days later. The version is pinned in `package.json` via `packageManager`; `corepack enable` makes your shell honour it.

**Versions are pinned by `pnpm-lock.yaml`, not by `package.json`.** The lockfile fixes every transitive package to an exact version and integrity hash, so the `^` ranges are only consulted on a deliberate `pnpm update`. The Docker build must install with `--frozen-lockfile` so that a drifting tree fails the build instead of shipping. pnpm 10 also refuses to run install scripts unless a package is listed in `onlyBuiltDependencies` — none are, and adding one deserves a look at what it runs.

## Linting and formatting

`oxlint` for lint, `prettier` for format. **No eslint.** oxlint ports the rules that matter here and runs in milliseconds; the one thing it cannot do is type-aware rules, and `strict` plus `noUncheckedIndexedAccess` in `tsc` already cover most of that ground for an app this size. Running both linters means two configs that disagree at the edges. If a floating-promise bug ever costs us an evening, that is the moment to add eslint — not before.

`react/only-export-components` is off for `src/routes/` and `src/components/ui/`: a route file must export its `Route` object beside its component, and a shadcn primitive exports its cva variants. The warning is unactionable in both, and a lint run with permanent warnings is a lint run nobody reads. `.oxlintrc.json` is JSON and cannot carry that note itself.

Prettier settings worth knowing: **double quotes** (JSX attributes are double-quoted whatever you pick, so this leaves one quote character per file), **semicolons**, **2-space indent** (JSX plus Tailwind nests four or five levels deep before any content, and every `shadcn add` arrives at 2, so 4 would mean reformatting each new component), and **120 columns**.

120 is chosen to sit alongside the Go side: `gofmt` enforces no width at all, but `golangci-lint`'s `lll` defaults to 120 and this repo's Go code measures p99 = 119. Both languages overrun it in the same one place — an unbreakable string literal, a long Tailwind `className` here and a long SQL string there — and that is accepted rather than worked around. Indentation cannot be kept in step the same way: Go indents with tabs, so its width is a property of your editor, not of the file.

Markdown is excluded from prettier so the repo keeps one Markdown convention: no manual line wrapping, tables written by hand.

## Layout

| Path | What |
|---|---|
| `src/routes/` | File-based TanStack Router routes. `__root.tsx` is the shell. |
| `src/routeTree.gen.ts` | Generated by the router plugin. **Committed**, because `pnpm build` type checks before Vite runs and would otherwise not find it. Never edited by hand. |
| `src/components/ui/` | shadcn/ui primitives, copied in. |
| `src/lib/api/` | Everything that knows the HTTP contract: `client.ts`, `dto.ts`, `enums.ts`, `session.ts`. Grouped rather than loose in `lib/`, which is not a dumping ground — a new file belongs in a named subdirectory or it does not belong in `lib/`. |
| `src/lib/api/client.ts` | The single fetch wrapper. Every failure arrives as `ApiError` (the server said no) or `NetworkError` (it never answered); nothing else talks to `fetch`. |
| `src/lib/api/dto.ts`, `src/lib/api/enums.ts` | The response shapes and the API's enum values as string unions. English, always. |
| `src/lib/api/session.ts` | `me` and the admin session as queries. The only place the app knows whether somebody is logged in. |
| `src/lib/routing/` | Route helpers that are not routes — `navigation.ts` sanitises a `?redirect=` target. |
| `src/lib/labels.ts` | The only place a German string is written — enum labels and screen copy alike. |
| `src/lib/code.ts` | The login code on the client: what the field displays, and what gets sent. |
| `src/lib/utils.ts` | `cn` only, and it stays at this path: `components.json` aliases it and every `shadcn add` writes `@/lib/utils` into the component it generates. |
| `**/__tests__/` | Component and route tests, one directory per level, so a source directory lists only source. Matched by vitest's `src/**/*.test.{ts,tsx}` and ignored by the router plugin, which skips `*.test.*` — a test file under `src/routes/` never becomes a route. |
| `src/test/` | The component-test harness: a fetch stub keyed by `"METHOD /path"`, and `renderApp` which mounts the real router with the real guards. |
| `src/index.css` | Design tokens from [05-design](../specification/05-design.md), plus the global base layer. |
| `src/assets/fonts/` | Self-hosted `.woff2`. No Google Fonts request at runtime — the CSP forbids it. |
| `scripts/write-dist-gitkeep.mjs` | Runs at the end of `pnpm build`. Vite empties `dist/`, and `dist/.gitkeep` has to survive: the Go binary embeds `web/dist`, and an embed with no matching files fails the backend build on a clean checkout. |

`@/` resolves to `src/`, declared twice on purpose: in `vite.config.ts` for the bundler and in `tsconfig` `paths` for the type checker. They are separate resolvers and must be kept in step.

`vitest.config.ts` is separate from `vite.config.ts` for a similar reason in reverse: the Vite config carries the router plugin, which rewrites the checked-in `routeTree.gen.ts` as it runs. A test run must not have that side effect, so the test config loads only the React transform and the alias.

## Testing

`pnpm test` runs the component tests. They mount the **real** route tree, the real query client and the real guards, and replace only `fetch` — everything below that is the thing under test, and a mock of it would be a mock of the answer.

The API itself is covered by the Go integration suite in `tests/integration`, which drives the real HTTP surface against a real SQLite file. Nothing here tries to be a second implementation of the server: if a test needs to know what an endpoint returns, the fixture is copied from that endpoint's contract.

Queries go through accessible names — `getByLabelText`, `getByRole("button", { name })` — so a passing test is also evidence the control can be found by somebody using a screen reader. Do not reach for `data-testid` when a label would do.

## Rules that are easy to break

- Components use **token names** (`bg-paper`, `text-ink-muted`), never hex values.
- No German string inline in a component. It goes in `labels.ts`.
- No `dangerouslySetInnerHTML`.
- Nothing below 16px ships, and every interactive element keeps a visible focus ring.
