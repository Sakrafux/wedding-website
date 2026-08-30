/**
 * Where the app is allowed to send somebody after logging in.
 */

/**
 * safeInternalPath returns a redirect target that cannot leave this site.
 *
 * The value comes out of the URL, so it is attacker-controlled: a link to
 * `/?redirect=https://example.invalid` would otherwise bounce a guest straight off
 * the site immediately after they logged in, which is the shape of every phishing
 * hand-off. Only a single-slash absolute path is allowed — `//host` is protocol
 * relative and is a different origin despite looking like a path.
 */
export function safeInternalPath(candidate: string | undefined, fallback: string): string {
  if (candidate === undefined || !candidate.startsWith("/") || candidate.startsWith("//")) {
    return fallback;
  }
  return candidate;
}
