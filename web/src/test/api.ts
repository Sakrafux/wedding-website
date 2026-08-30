import { vi } from "vitest";

/**
 * A stand-in for the API, keyed by "METHOD /path".
 *
 * The component tests drive the real router, the real query client and the real
 * fetch wrapper; only the network is replaced. Anything below that — the session
 * queries, the guards, the error mapping — is the thing under test, and a mock of
 * it would be a mock of the answer.
 *
 * The API itself is covered by the Go integration suite against a real database, so
 * nothing here is trying to be a second implementation of the server.
 */
export interface StubResponse {
  status: number;
  body?: unknown;
}

/**
 * A handler may return a promise, which is how a test holds a request open and
 * observes the in-flight state of the screen that made it.
 */
export type StubHandler = StubResponse | ((body: unknown) => StubResponse | Promise<StubResponse>);

/** A success body, with the 204-no-body case the logout endpoint uses. */
export function ok(body?: unknown): StubResponse {
  return body === undefined ? { status: 204 } : { status: 200, body };
}

/** The error envelope, exactly as httpio writes it. */
export function apiError(status: number, code: string, message: string): StubResponse {
  return { status, body: { error: { code, message, request_id: "TESTID1" } } };
}

/** 401 with the message the server really sends for an unknown code. */
export const unknownCode = apiError(
  401,
  "unknown_login_code",
  "Diesen Code kennen wir nicht. Schau bitte noch mal auf deine Karte — Groß- und Kleinschreibung ist egal.",
);

export const unauthenticated = apiError(401, "unauthenticated", "Bitte melde dich an.");

export interface ApiStub {
  /** Every request the app made, in order, for asserting on what was called. */
  calls: { method: string; path: string; body: unknown }[];
  /** Replaces a handler mid-test, for the "session expires while you are here" cases. */
  set: (route: string, handler: StubHandler) => void;
}

export function stubApi(handlers: Record<string, StubHandler>): ApiStub {
  const routes = { ...handlers };
  const calls: ApiStub["calls"] = [];

  vi.stubGlobal("fetch", (input: string, init?: RequestInit) => {
    const method = init?.method ?? "GET";
    const path = String(input);
    const body: unknown = typeof init?.body === "string" ? JSON.parse(init.body) : undefined;

    calls.push({ method, path, body });

    const handler = routes[`${method} ${path}`];
    if (!handler) {
      // Better than a silent 404: an unexpected request is nearly always a test
      // that set up the wrong route, and the message names it.
      throw new Error(`no stub for ${method} ${path}`);
    }

    return Promise.resolve(typeof handler === "function" ? handler(body) : handler).then(
      (response) =>
        new Response(response.status === 204 ? null : JSON.stringify(response.body), {
          status: response.status,
          headers: { "Content-Type": "application/json" },
        }),
    );
  });

  return {
    calls,
    set: (route, handler) => {
      routes[route] = handler;
    },
  };
}
