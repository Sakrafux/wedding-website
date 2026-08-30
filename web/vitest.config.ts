import { fileURLToPath } from "node:url";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

/**
 * A config of its own rather than a `test` block inside vite.config.ts.
 *
 * The Vite config carries the TanStack Router plugin, which generates the route
 * tree by scanning src/routes and writing routeTree.gen.ts. Running that during a
 * test run would rewrite a checked-in file as a side effect of `pnpm test`, so the
 * test setup deliberately loads only the React transform and the path alias.
 */
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
    // Component tests only. The integration suite lives in Go and drives the real
    // API; duplicating that here against a mocked fetch would test the mock.
    include: ["src/**/*.test.{ts,tsx}"],
  },
});
