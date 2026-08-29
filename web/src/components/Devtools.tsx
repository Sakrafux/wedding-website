import { lazy } from "react";

/**
 * The TanStack Router and Query devtool panels, mounted only in development.
 *
 * They are loaded through a dynamic `import()` rather than a top-level one guarded
 * by `import.meta.env.DEV`. A static import stays in the module graph even when the
 * branch using it is eliminated, so the panels would end up in the production
 * bundle — several hundred kilobytes shipped to guests on phones, for a tool only
 * we ever open. With the import inside `lazy`, the production build resolves this
 * to a component that renders nothing and never references either package.
 */
export const Devtools = import.meta.env.DEV
  ? lazy(async () => {
      const [router, query] = await Promise.all([
        import("@tanstack/react-router-devtools"),
        import("@tanstack/react-query-devtools"),
      ]);

      return {
        default: () => (
          <>
            <router.TanStackRouterDevtools position="bottom-right" />
            <query.ReactQueryDevtools initialIsOpen={false} />
          </>
        ),
      };
    })
  : () => null;
