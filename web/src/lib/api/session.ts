/**
 * Session state, as data.
 *
 * `me` is the only place the frontend knows whether somebody is logged in. There is
 * no auth store, no context and no boolean kept alongside it: two sources of truth
 * for "am I logged in" is a reliable source of bugs, and the one that stays stale
 * is always the one a guard reads.
 */

import { type QueryClient, queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { ApiError, getJson, postJson } from "./api";
import type { AdminSession, BootstrapResponse } from "./dto";

export const meQueryKey = ["me"] as const;
export const adminSessionQueryKey = ["admin-session"] as const;

/**
 * A few minutes. The flags `me` carries — whether the RSVP form is open, whether
 * the seating plan is published — change a handful of times in a year, so
 * refetching them on every navigation is traffic that buys nothing.
 */
const sessionStaleTime = 5 * 60 * 1000;

/**
 * fetchMe answers "who is this" with `null` rather than by throwing.
 *
 * Not being logged in is the ordinary state of a first-time visitor, not a failure,
 * and modelling it as an error would put it in the same bucket as a dropped
 * connection — which is precisely the confusion that logs a guest out because a
 * train went into a tunnel. Everything that is not a 401 still throws.
 */
async function fetchMe(): Promise<BootstrapResponse | null> {
  try {
    return await getJson<BootstrapResponse>("/me");
  } catch (error) {
    if (error instanceof ApiError && error.isUnauthenticated) {
      return null;
    }
    throw error;
  }
}

async function fetchAdminSession(): Promise<AdminSession | null> {
  try {
    return await getJson<AdminSession>("/admin/me");
  } catch (error) {
    if (error instanceof ApiError && error.isUnauthenticated) {
      return null;
    }
    throw error;
  }
}

export const meQueryOptions = queryOptions({
  queryKey: meQueryKey,
  queryFn: fetchMe,
  staleTime: sessionStaleTime,
});

/**
 * The admin's session, kept separate from `me`.
 *
 * `/api/me` answers 401 for an admin — an admin is not a household — so a returning
 * admin can only be recognised by asking the admin side. One cookie, two questions.
 */
export const adminSessionQueryOptions = queryOptions({
  queryKey: adminSessionQueryKey,
  queryFn: fetchAdminSession,
  staleTime: sessionStaleTime,
});

/**
 * Both session queries are reset together after any login or logout, because one
 * cookie carries both subjects: logging in as the admin invalidates the household
 * session on the server, and leaving the old answer cached would have the app
 * believing in a session that no longer exists.
 */
function useSessionReset() {
  const queryClient = useQueryClient();

  return async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: meQueryKey }),
      queryClient.invalidateQueries({ queryKey: adminSessionQueryKey }),
    ]);
  };
}

/**
 * Treats a 401 from *any* query as the session being gone.
 *
 * A year-long session can still be revoked, and when that happens the app must
 * notice from whichever request found out rather than waiting for the next
 * navigation. Writing `null` into the session queries is enough: the layout guards
 * read them, so the redirect to the right login screen follows from the data.
 *
 * Returns the unsubscribe function. Exported so that main and the tests attach the
 * same handler — a test that had to reimplement it would be testing its own copy.
 */
export function watchForSessionExpiry(queryClient: QueryClient): () => void {
  return queryClient.getQueryCache().subscribe((event) => {
    if (event.type !== "updated" || event.action.type !== "error") {
      return;
    }
    const error: unknown = event.action.error;
    if (error instanceof ApiError && error.isUnauthenticated) {
      queryClient.setQueryData(meQueryKey, null);
      queryClient.setQueryData(adminSessionQueryKey, null);
    }
  });
}

export function useLogin() {
  const resetSession = useSessionReset();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (code: string) => postJson<BootstrapResponse>("/auth/login", { code }),
    onSuccess: async (bootstrap) => {
      // Seeded before the invalidation so the next screen renders from the body we
      // already have, instead of flashing a skeleton while refetching what the
      // login response just told us.
      queryClient.setQueryData(meQueryKey, bootstrap);
      await resetSession();
    },
  });
}

export function useAdminLogin() {
  const resetSession = useSessionReset();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (credentials: { user: string; password: string }) =>
      postJson<AdminSession>("/auth/admin/login", credentials),
    onSuccess: async (session) => {
      queryClient.setQueryData(adminSessionQueryKey, session);
      await resetSession();
    },
  });
}

export function useLogout() {
  const resetSession = useSessionReset();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => postJson<void>("/auth/logout"),
    onSuccess: async () => {
      queryClient.setQueryData(meQueryKey, null);
      queryClient.setQueryData(adminSessionQueryKey, null);
      await resetSession();
    },
  });
}

/**
 * Which households this browser has already confirmed, so the "is this you?" screen
 * appears after a login rather than on every visit for a year.
 *
 * Keyed by household id rather than by a per-login flag: the screen exists to catch
 * a valid code that belongs to somebody else, so what matters is whether *this
 * household* has been confirmed here. Logging in again with a different code lands
 * on a different id and asks again, which is exactly the case it is there for.
 */
const confirmedHouseholdsKey = "wedding.confirmed-households";

function readConfirmedHouseholds(): number[] {
  try {
    const stored: unknown = JSON.parse(window.localStorage.getItem(confirmedHouseholdsKey) ?? "[]");
    return Array.isArray(stored) ? stored.filter((id): id is number => typeof id === "number") : [];
  } catch {
    // Private browsing, a disabled store, or something else's data under our key.
    // Failing to "not confirmed" shows the screen once more than necessary, which
    // is the harmless direction.
    return [];
  }
}

export function isHouseholdConfirmed(householdId: number): boolean {
  return readConfirmedHouseholds().includes(householdId);
}

export function rememberHouseholdConfirmed(householdId: number): void {
  try {
    const confirmed = readConfirmedHouseholds();
    if (!confirmed.includes(householdId)) {
      window.localStorage.setItem(confirmedHouseholdsKey, JSON.stringify([...confirmed, householdId]));
    }
  } catch {
    // Storage unavailable. The confirmation screen will simply ask again next
    // time, which is a smaller cost than failing the navigation it precedes.
  }
}
