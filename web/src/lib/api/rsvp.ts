/**
 * The RSVP answer, as queries and mutations.
 *
 * Two pairs of them, for the same server-side use case: the household's own
 * `/api/rsvp` and the admin's `/api/admin/households/{id}/rsvp`. They are separate
 * query keys because they are separate caches — the admin has many households open
 * over an evening and the guest has one — but they return the identical shape, which
 * is what lets one form component render either (F3-F06).
 */

import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { deleteRequest, getJson, postJson, putJson } from "./client";
import type { RSVPAddMemberRequest, RSVPAddMemberResponse, RSVPResponse, RSVPSaveRequest } from "./dto";
import { householdQueryKey, householdsQueryKey } from "./households";
import { meQueryKey } from "./session";

export const rsvpQueryKey = ["rsvp"] as const;

export function adminRSVPQueryKey(householdId: number) {
  return ["admin", "rsvp", householdId] as const;
}

export const rsvpQueryOptions = queryOptions({
  queryKey: rsvpQueryKey,
  queryFn: () => getJson<RSVPResponse>("/rsvp"),
});

export function adminRSVPQueryOptions(householdId: number) {
  return queryOptions({
    queryKey: adminRSVPQueryKey(householdId),
    queryFn: () => getJson<RSVPResponse>(`/admin/households/${householdId}/rsvp`),
  });
}

/**
 * Saves the household's own answer.
 *
 * The response is written straight into the query cache: it *is* the stored answer,
 * normalized, so refetching it would ask the server to repeat itself. `me` is
 * invalidated as well, because `rsvp_submitted_at` decides whether the navigation says
 * "Zusagen" or "Antwort ändern", and a bar still saying "Zusagen" after a successful
 * answer is a bar that gets pressed again.
 */
export function useSaveRSVP() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: RSVPSaveRequest) => putJson<RSVPResponse>("/rsvp", request),
    onSuccess: async (answer) => {
      queryClient.setQueryData(rsvpQueryKey, answer);
      await queryClient.invalidateQueries({ queryKey: meQueryKey });
    },
  });
}

/**
 * Saves a household's answer as the admin — the call we take on the phone.
 *
 * Invalidates the admin household queries too: the list and the detail page both show
 * `rsvp_submitted_at`, and that is exactly the column that just changed.
 */
export function useSaveAdminRSVP(householdId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: RSVPSaveRequest) => putJson<RSVPResponse>(`/admin/households/${householdId}/rsvp`, request),
    onSuccess: async (answer) => {
      queryClient.setQueryData(adminRSVPQueryKey(householdId), answer);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: householdsQueryKey }),
        queryClient.invalidateQueries({ queryKey: householdQueryKey(householdId) }),
      ]);
    },
  });
}

/**
 * Adds the household's plus-one.
 *
 * The response carries the created member and the recomputed flag, so the cache is
 * patched rather than refetched: the form the guest is filling in stays exactly as it
 * is, with one more card at the end. A refetch here would be a round trip for data we
 * were just handed — and, worse, would race with unsaved answers on screen.
 */
export function useAddPlusOne() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: RSVPAddMemberRequest) => postJson<RSVPAddMemberResponse>("/rsvp/members", request),
    onSuccess: (added) => {
      queryClient.setQueryData<RSVPResponse>(rsvpQueryKey, (answer) =>
        answer
          ? { ...answer, members: [...answer.members, added.member], can_add_plus_one: added.can_add_plus_one }
          : answer,
      );
    },
  });
}

/**
 * Removes a member the household added itself.
 *
 * The endpoint answers 204, so the flag comes back from a refetch: whether the
 * household may now add somebody else is the server's answer, and assuming `true` here
 * would be this file re-deriving the rule.
 */
export function useRemoveMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (memberId: number) => deleteRequest(`/rsvp/members/${memberId}`),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: rsvpQueryKey });
    },
  });
}
