/**
 * The single way this app talks to its API.
 *
 * Every failure arrives as one of two errors, and the distinction is the point:
 * ApiError means the server answered and said no, NetworkError means it never
 * answered at all. Treating the second as the first is how a guest on a train gets
 * logged out because they went through a tunnel.
 */

/** The error envelope every failing endpoint returns. */
interface ErrorEnvelope {
  error: {
    code: string;
    message: string;
    request_id: string;
    fields?: Record<string, string>;
  };
}

/**
 * ApiError is a response the server produced deliberately.
 *
 * `message` is the German sentence the API chose, and is shown to the guest
 * verbatim. The frontend does not re-map it: the server owns the wording so that
 * every screen says the same thing and a wording fix is one deploy, not two.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId: string;
  readonly fields: Record<string, string>;

  constructor(status: number, envelope: ErrorEnvelope["error"]) {
    super(envelope.message);
    this.name = "ApiError";
    this.status = status;
    this.code = envelope.code;
    this.requestId = envelope.request_id;
    this.fields = envelope.fields ?? {};
  }

  /** True when the session is missing, expired or of the wrong kind. */
  get isUnauthenticated(): boolean {
    return this.status === 401;
  }
}

/**
 * NetworkError is the request never completing: offline, DNS, a dropped
 * connection, or a response that is not the JSON this API always sends.
 *
 * Its message is German and safe to show, because there is no envelope to take one
 * from — this is the one case where the frontend has to phrase the failure itself.
 */
export class NetworkError extends Error {
  constructor(cause?: unknown) {
    super("Keine Verbindung zum Server. Bitte prüf deine Internetverbindung.");
    this.name = "NetworkError";
    this.cause = cause;
  }
}

/** JSON bodies only: the API requires the content type, as a CSRF control. */
const jsonHeaders = { "Content-Type": "application/json" };

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`/api${path}`, {
      // Same origin in production and through the dev proxy alike, so the cookie
      // rides along without any cross-origin credential handling.
      credentials: "same-origin",
      ...init,
    });
  } catch (cause) {
    throw new NetworkError(cause);
  }

  // 204 is the whole answer for logout; asking for a body would fail on an empty
  // one.
  if (response.status === 204) {
    return undefined as T;
  }

  let body: unknown;
  try {
    body = await response.json();
  } catch (cause) {
    // A non-JSON body from this API means something in front of it answered — a
    // proxy error page, a captive portal. That is a connection problem, not a
    // decision the application made.
    throw new NetworkError(cause);
  }

  if (!response.ok) {
    throw new ApiError(response.status, (body as ErrorEnvelope).error);
  }
  return body as T;
}

export function getJson<T>(path: string): Promise<T> {
  return request<T>(path, { method: "GET" });
}

export function postJson<T>(path: string, payload?: unknown): Promise<T> {
  return request<T>(path, {
    method: "POST",
    headers: jsonHeaders,
    // A body of "null" is still valid JSON and still carries the content type,
    // which is what the server's CSRF check looks for on a bodyless POST.
    body: JSON.stringify(payload ?? null),
  });
}

/**
 * PATCH sends a partial update: only the fields in `payload` are touched.
 *
 * The server rejects an unknown field outright, so a typo in a payload key is an
 * error rather than an answer silently dropped.
 */
export function patchJson<T>(path: string, payload: unknown): Promise<T> {
  return request<T>(path, {
    method: "PATCH",
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  });
}

/** DELETE answers 204 with no body, which `request` maps to undefined. */
export function deleteRequest(path: string): Promise<void> {
  return request<void>(path, { method: "DELETE" });
}

/**
 * fieldError picks one field's message out of a failed request.
 *
 * The server reports rejected inputs keyed by the JSON field name, so a form renders
 * each message next to its own control instead of one sentence at the top — which is
 * the difference between "prüfe die markierten Felder" being useful and being noise.
 * Anything that is not an ApiError has no fields, and answers undefined.
 */
export function fieldError(error: unknown, field: string): string | undefined {
  return error instanceof ApiError ? error.fields[field] : undefined;
}
