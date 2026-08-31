import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";

import { RSVPForm } from "@/components/rsvp/RSVPForm";
import { Alert, AlertTitle } from "@/components/ui/alert";
import { buttonVariants } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { adminRSVPQueryOptions, useSaveAdminRSVP } from "@/lib/api/rsvp";
import { adminRSVPLabels, householdLabels } from "@/lib/labels";

/**
 * The household's answer, entered by us — the call that comes in after the deadline.
 *
 * The **same** form component the guests use, given the admin query and the admin
 * mutation. If this page ever renders a control the guest form does not have, that
 * control belongs in the shared component or nowhere: one form, one field set, which is
 * what Gate 1 exists to protect.
 */
// The trailing underscore in the file name opts out of nesting under the household
// detail route: this is a page of its own, not a panel inside that one, and the detail
// page renders no <Outlet />.
export const Route = createFileRoute("/admin/_shell/haushalte/$householdId_/rsvp")({
  component: AdminRSVPPage,
});

function AdminRSVPPage() {
  const { householdId } = Route.useParams();
  const id = Number(householdId);
  const { data: answer, isPending, error, refetch } = useQuery(adminRSVPQueryOptions(id));
  const save = useSaveAdminRSVP(id);

  if (isPending) {
    return (
      <div aria-busy="true" className="bg-surface-sunken h-64 animate-pulse rounded-lg">
        <p className="sr-only">{householdLabels.loading}</p>
      </div>
    );
  }

  if (error) {
    // An unknown household is answered here with a sentence and a way back to the
    // list, rather than by redirecting to it: a redirect that swallows the reason
    // leaves an admin who mistyped an id wondering what just happened. What matters is
    // that this never lands on the guest login — the admin shell's guard owns that.
    const isUnknownHousehold = error instanceof ApiError && error.status === 404;

    return (
      <div className="flex flex-col items-start gap-4">
        <Alert variant="attention" role="alert">
          <AlertTitle>{isUnknownHousehold ? adminRSVPLabels.notFound : error.message}</AlertTitle>
        </Alert>
        <Link to="/admin/haushalte" className={buttonVariants({ variant: "outline" })}>
          {householdLabels.detailBack}
        </Link>
      </div>
    );
  }

  return (
    <div className="flex max-w-2xl flex-col gap-4">
      <Link to="/admin/haushalte/$householdId" params={{ householdId }} className="text-primary underline">
        {adminRSVPLabels.back}
      </Link>

      <h1 className="text-h1 font-body">{adminRSVPLabels.heading(answer.household.display_name)}</h1>
      <p className="text-ink-muted">{adminRSVPLabels.intro(answer.household.display_name)}</p>

      {/* The deadline is information here, never a lock: this page exists for the late
          call, so the override is passed explicitly and the shared default stays safe. */}
      {!answer.editable ? (
        <Alert>
          <AlertTitle>{adminRSVPLabels.deadlinePassed}</AlertTitle>
        </Alert>
      ) : null}

      <RSVPForm answer={answer} onSave={save.mutateAsync} onReload={refetch} allowEditingAfterDeadline dense />
    </div>
  );
}
