import { fileURLToPath } from "node:url";

import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// The router plugin must run before the React plugin: it generates the route tree
// that React's transform then compiles.
export default defineConfig({
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
    // The site has its own subdomain and is served at the root, so the dev server
    // forwards /api unchanged — dev and production agree on what the backend
    // receives without a rewrite in between.
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
