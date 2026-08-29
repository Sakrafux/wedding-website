import { fileURLToPath } from "node:url";

import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// The public path the app is served under. Caddy routes it as
// `handle_path /hochzeit*`, which strips the prefix before the request reaches
// Go — so the Go side never sees it, but every URL the browser builds must carry it.
// Baked into the bundle at build time, which is fine because one build serves one
// deployment; it is exported through import.meta.env.BASE_URL, and everything that
// needs the prefix reads it from there rather than repeating the literal.
const basePath = "/hochzeit/";

// The router plugin must run before the React plugin: it generates the route tree
// that React's transform then compiles.
export default defineConfig({
  base: basePath,
  plugins: [tanstackRouter({ target: "react", autoCodeSplitting: true }), react(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    // The Go binary serves the API on 8080 in development; in production both come
    // from the same origin via go:embed, so no CORS anywhere.
    //
    // The rewrite mirrors Caddy's handle_path: the browser asks for
    // /hochzeit/api/…, the prefix is stripped, and Go sees /api/… in both
    // environments. Without it, dev and production would disagree about what the
    // backend receives — the exact class of bug this proxy exists to avoid.
    proxy: {
      [`${basePath}api`]: {
        target: "http://localhost:8080",
        rewrite: (path) => path.slice(basePath.length - 1),
      },
    },
  },
});
