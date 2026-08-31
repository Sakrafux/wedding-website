import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import { RSVPForm } from "@/components/rsvp/RSVPForm";
import { rsvpQueryOptions, useAddPlusOne, useRemoveMember, useSaveRSVP } from "@/lib/api/rsvp";
import { rsvpLabels } from "@/lib/labels";

/**
 * The route is `/zusagen`, not `/rsvp`: "RSVP" is jargon the copy rules forbid, and
 * this URL gets read out on the telephone as often as it gets clicked.
 *
 * The loader is what puts the load on the router, so a failure renders the shared
 * error state and a slow load the shared skeleton — rather than a blank form that
 * fills in later.
 */
export const Route = createFileRoute("/_guest/_chrome/zusagen")({
  loader: ({ context }) => context.queryClient.ensureQueryData(rsvpQueryOptions),
  component: RSVPPage,
});

function RSVPPage() {
  const { data: answer, refetch } = useSuspenseQuery(rsvpQueryOptions);
  const save = useSaveRSVP();
  const addPlusOne = useAddPlusOne();
  const removeMember = useRemoveMember();

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-8 px-4 py-12">
      <h1 className="text-display font-display">{rsvpLabels.heading}</h1>
      <p>{rsvpLabels.intro}</p>

      {/* The form fetches nothing itself: the query and the mutation are handed in, so
          the admin route can hand it different ones (F3-F06). */}
      <RSVPForm
        answer={answer}
        onSave={save.mutateAsync}
        onReload={refetch}
        onAddMember={(name) => addPlusOne.mutateAsync({ name })}
        onRemoveMember={removeMember.mutateAsync}
      />
    </div>
  );
}
