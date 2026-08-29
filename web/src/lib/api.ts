/**
 * The one place that knows where the API lives.
 *
 * The app is served under a path prefix (`/hochzeit/`), and the reverse proxy
 * strips that prefix before the request reaches Go. So the backend sees `/api/…`
 * while the browser must ask for `/hochzeit/api/…`. Vite bakes the prefix into
 * `import.meta.env.BASE_URL`, and every request goes through `apiUrl` so the literal
 * appears exactly once in the frontend — a hardcoded `/api/…` would work in dev,
 * where Vite serves at the root, and 404 in production.
 */

// BASE_URL always ends in a slash; trimming it keeps apiUrl from producing "//api".
const apiRoot = `${import.meta.env.BASE_URL.replace(/\/$/, "")}/api`;

/** apiUrl turns an API path like "/auth/login" into a URL the browser can fetch. */
export function apiUrl(path: string): string {
  return `${apiRoot}${path.startsWith("/") ? path : `/${path}`}`;
}
