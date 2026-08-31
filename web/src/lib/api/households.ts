/**
 * The admin guest list, as queries and mutations.
 *
 * Every mutation invalidates the queries it can have changed rather than patching the
 * cache by hand: this is an admin on a laptop on a good connection, and a member list
 * that reorders itself optimistically and then snaps back is worse than a spinner.
 */

import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { deleteRequest, getJson, patchJson, postJson } from "./client";
import type { AdminCodeReissue, AdminGuest, AdminHousehold, AdminHouseholdOverview } from "./dto";
import type { GuestKind, SeatingNeed } from "./enums";

export const householdsQueryKey = ["admin", "households"] as const;

/** The key of one household's detail query, so a mutation can invalidate just it. */
export function householdQueryKey(householdId: number) {
  return [...householdsQueryKey, householdId] as const;
}

export const householdsQueryOptions = queryOptions({
  queryKey: householdsQueryKey,
  queryFn: async () => {
    const body = await getJson<{ households: AdminHouseholdOverview[] }>("/admin/households");
    return body.households;
  },
});

export function householdQueryOptions(householdId: number) {
  return queryOptions({
    queryKey: householdQueryKey(householdId),
    queryFn: () => getJson<AdminHousehold>(`/admin/households/${householdId}`),
  });
}

/** The fields a household PATCH may carry. Absent means "leave alone", server-side. */
export interface HouseholdPatch {
  display_name?: string;
  admin_note?: string;
  transport_seats_needed?: number;
  transport_seats_offered?: number;
  has_stroller?: boolean;
}

/** The fields a new member needs, and the ones a member PATCH may carry. */
export interface GuestDraft {
  first_name: string;
  last_name: string;
  kind: GuestKind;
  age: number | null;
  seating_need: SeatingNeed;
  dietary_note: string;
}

export interface GuestPatch {
  first_name?: string;
  last_name?: string;
  kind?: GuestKind;
  /** `null` clears the age; omitting the key leaves it alone. The two differ. */
  age?: number | null;
  seating_need?: SeatingNeed;
  dietary_note?: string;
}

/**
 * Invalidates the list and, when given one, a single household.
 *
 * Both, always: a member added on the detail page changes that household's count in
 * the list, and a stale count is exactly the number the nudge calls are made from.
 */
function useHouseholdInvalidation() {
  const queryClient = useQueryClient();

  return async (householdId?: number) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: householdsQueryKey }),
      householdId === undefined
        ? Promise.resolve()
        : queryClient.invalidateQueries({ queryKey: householdQueryKey(householdId) }),
    ]);
  };
}

export function useCreateHousehold() {
  const invalidate = useHouseholdInvalidation();

  return useMutation({
    mutationFn: (displayName: string) => postJson<AdminHousehold>("/admin/households", { display_name: displayName }),
    onSuccess: (household) => invalidate(household.id),
  });
}

export function useUpdateHousehold(householdId: number) {
  const invalidate = useHouseholdInvalidation();

  return useMutation({
    mutationFn: (patch: HouseholdPatch) => patchJson<AdminHousehold>(`/admin/households/${householdId}`, patch),
    onSuccess: () => invalidate(householdId),
  });
}

export function useDeleteHousehold(householdId: number) {
  const invalidate = useHouseholdInvalidation();

  return useMutation({
    mutationFn: () => deleteRequest(`/admin/households/${householdId}`),
    onSuccess: () => invalidate(),
  });
}

/**
 * Reissuing a code is a POST with no body: there is nothing to choose. The response
 * says how many devices were signed out, which the screen shows afterwards.
 */
export function useReissueCode(householdId: number) {
  const invalidate = useHouseholdInvalidation();

  return useMutation({
    mutationFn: () => postJson<AdminCodeReissue>(`/admin/households/${householdId}/code`),
    onSuccess: () => invalidate(householdId),
  });
}

export function useAddMember(householdId: number) {
  const invalidate = useHouseholdInvalidation();

  return useMutation({
    mutationFn: (draft: GuestDraft) => postJson<AdminGuest>(`/admin/households/${householdId}/guests`, draft),
    onSuccess: () => invalidate(householdId),
  });
}

/**
 * Members are addressed by their own id — the household is passed only so the right
 * detail query is invalidated.
 */
export function useUpdateMember(householdId: number, memberId: number) {
  const invalidate = useHouseholdInvalidation();

  return useMutation({
    mutationFn: (patch: GuestPatch) => patchJson<AdminGuest>(`/admin/guests/${memberId}`, patch),
    onSuccess: () => invalidate(householdId),
  });
}

export function useRemoveMember(householdId: number, memberId: number) {
  const invalidate = useHouseholdInvalidation();

  return useMutation({
    mutationFn: () => deleteRequest(`/admin/guests/${memberId}`),
    onSuccess: () => invalidate(householdId),
  });
}
