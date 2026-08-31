import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useRef, useState } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NativeSelect } from "@/components/ui/native-select";
import { Textarea } from "@/components/ui/textarea";
import { fieldError } from "@/lib/api/client";
import type { AdminGuest, AdminHousehold } from "@/lib/api/dto";
import type { GuestKind, SeatingNeed } from "@/lib/api/enums";
import {
  householdQueryOptions,
  useAddMember,
  useDeleteHousehold,
  useReissueCode,
  useRemoveMember,
  useUpdateHousehold,
  useUpdateMember,
} from "@/lib/api/households";
import { formatShortDate } from "@/lib/dates";
import { guestKindLabels, guestOriginLabels, householdLabels, seatingNeedLabels } from "@/lib/labels";

export const Route = createFileRoute("/admin/_shell/haushalte/$householdId")({
  component: HouseholdDetailPage,
});

/**
 * One household, top to bottom: its own fields, its members, its code, and deleting
 * it. Entering eighty guests is the workload this page exists for, so everything that
 * task needs is here rather than spread over four screens.
 */
function HouseholdDetailPage() {
  const { householdId } = Route.useParams();
  const { data: household, isPending, error } = useQuery(householdQueryOptions(Number(householdId)));

  if (isPending) {
    return (
      <div aria-busy="true" className="bg-surface-sunken h-64 animate-pulse rounded-lg">
        <p className="sr-only">{householdLabels.loading}</p>
      </div>
    );
  }

  if (error) {
    return <p role="alert">{error.message}</p>;
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <Link to="/admin/haushalte" className="text-small underline underline-offset-4">
          {householdLabels.detailBack}
        </Link>
        <h1 className="text-h2 font-body">{household.display_name}</h1>
      </div>

      <HouseholdFields household={household} />
      <CodeSection household={household} />
      <MemberSection household={household} />
      <RSVPSummary household={household} />
      <DeleteSection household={household} />
    </div>
  );
}

/** A section wrapper, so the six blocks on this page are visually one list. */
function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="border-line flex flex-col gap-4 border-t pt-6">
      <h2 className="text-h3 font-body">{title}</h2>
      {children}
    </section>
  );
}

function HouseholdFields({ household }: { household: AdminHousehold }) {
  const update = useUpdateHousehold(household.id);

  const [displayName, setDisplayName] = useState(household.display_name);
  const [adminNote, setAdminNote] = useState(household.admin_note);
  const [seatsNeeded, setSeatsNeeded] = useState(String(household.transport_seats_needed));
  const [seatsOffered, setSeatsOffered] = useState(String(household.transport_seats_offered));
  const [hasStroller, setHasStroller] = useState(household.has_stroller);

  async function submit(event: React.FormEvent) {
    event.preventDefault();

    try {
      await update.mutateAsync({
        display_name: displayName,
        admin_note: adminNote,
        transport_seats_needed: Number(seatsNeeded),
        transport_seats_offered: Number(seatsOffered),
        has_stroller: hasStroller,
      });
    } catch {
      // Rendered per field below, from the error's `fields` map.
    }
  }

  return (
    // A fieldset whose legend doubles as the section heading: the page has several
    // "Speichern" buttons and two "Name"-shaped fields, and the group is what says
    // which household datum a control belongs to.
    <form onSubmit={submit} className="border-line flex flex-col gap-4 border-t pt-6">
      <fieldset className="flex flex-col gap-4">
        <legend className="text-h3 font-body">{householdLabels.detailDataHeading}</legend>
        <Field
          label={householdLabels.displayNameLabel}
          id="display-name"
          error={fieldError(update.error, "display_name")}
        >
          <Input id="display-name" value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
        </Field>

        <Field
          label={householdLabels.adminNoteLabel}
          id="admin-note"
          hint={householdLabels.adminNoteHint}
          error={fieldError(update.error, "admin_note")}
        >
          <Textarea id="admin-note" value={adminNote} onChange={(event) => setAdminNote(event.target.value)} />
        </Field>

        <div className="flex flex-wrap gap-4">
          <Field
            label={householdLabels.transportNeededLabel}
            id="seats-needed"
            error={fieldError(update.error, "transport_seats_needed")}
          >
            <Input
              id="seats-needed"
              type="number"
              min={0}
              max={20}
              value={seatsNeeded}
              onChange={(event) => setSeatsNeeded(event.target.value)}
            />
          </Field>

          <Field
            label={householdLabels.transportOfferedLabel}
            id="seats-offered"
            error={fieldError(update.error, "transport_seats_offered")}
          >
            <Input
              id="seats-offered"
              type="number"
              min={0}
              max={20}
              value={seatsOffered}
              onChange={(event) => setSeatsOffered(event.target.value)}
            />
          </Field>
        </div>

        <Label className="items-center gap-2">
          <input
            type="checkbox"
            className="size-5"
            checked={hasStroller}
            onChange={(event) => setHasStroller(event.target.checked)}
          />
          {householdLabels.strollerLabel}
        </Label>

        <div className="flex items-center gap-3">
          <Button type="submit" disabled={update.isPending}>
            {update.isPending ? householdLabels.saving : householdLabels.save}
          </Button>
          {update.isSuccess ? (
            <span role="status" className="text-small text-ink-muted">
              {householdLabels.saved}
            </span>
          ) : null}
        </div>
      </fieldset>
    </form>
  );
}

/**
 * The code, with a copy button and a guarded reissue.
 *
 * The displayed form is the stored form, which is also the printed form: six
 * characters, no separator. Copy copies exactly that, since what gets pasted goes
 * into the document the print shop receives.
 */
function CodeSection({ household }: { household: AdminHousehold }) {
  const reissue = useReissueCode(household.id);
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(household.code);
      setCopied(true);
    } catch {
      // No clipboard permission, or an insecure context. The code is on screen and
      // can be typed; failing loudly here would help nobody.
    }
  }

  return (
    <Section title={householdLabels.codeHeading}>
      <div className="flex flex-wrap items-center gap-3">
        <span className="text-h3 font-mono">{household.code}</span>

        <Button type="button" variant="outline" size="sm" onClick={() => void copy()}>
          {householdLabels.codeCopy}
        </Button>

        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button type="button" variant="outline" size="sm" disabled={reissue.isPending}>
              {householdLabels.codeReissue}
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogTitle>{householdLabels.codeReissueConfirmTitle}</AlertDialogTitle>
            {/* The consequence in plain German: the old code dies, a printed card
                carrying it is dead, and logged-in devices are signed out. */}
            <AlertDialogDescription>{householdLabels.codeReissueConfirmBody}</AlertDialogDescription>
            <AlertDialogFooter>
              <AlertDialogCancel>{householdLabels.cancel}</AlertDialogCancel>
              <AlertDialogAction onClick={() => reissue.mutate()}>
                {householdLabels.codeReissueConfirm}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      {copied ? (
        <p role="status" className="text-small text-ink-muted">
          {householdLabels.codeCopied}
        </p>
      ) : null}

      {/* revoked_sessions comes from the response, so the admin reads what happened
          rather than inferring it. */}
      {reissue.data ? (
        <p role="status" className="text-small">
          {householdLabels.codeReissued(reissue.data.revoked_sessions)}
        </p>
      ) : null}

      {reissue.error ? (
        <p role="alert" className="text-accent-strong text-small">
          {reissue.error.message}
        </p>
      ) : null}
    </Section>
  );
}

function MemberSection({ household }: { household: AdminHousehold }) {
  return (
    <Section title={householdLabels.membersHeading}>
      {household.members.length === 0 ? (
        <p className="text-ink-muted">{householdLabels.membersEmpty}</p>
      ) : (
        <ul className="flex flex-col gap-4">
          {household.members.map((member) => (
            <li key={member.id}>
              <MemberRow householdId={household.id} member={member} />
            </li>
          ))}
        </ul>
      )}

      <AddMemberForm household={household} />
    </Section>
  );
}

/**
 * One member, edited in place and saved on its own.
 *
 * Per row rather than through a page-wide submit: with twenty members, one validation
 * error in a single submit would discard the whole screen.
 */
function MemberRow({ householdId, member }: { householdId: number; member: AdminGuest }) {
  const update = useUpdateMember(householdId, member.id);
  const remove = useRemoveMember(householdId, member.id);

  const [name, setName] = useState(member.name);
  const [kind, setKind] = useState<GuestKind>(member.kind);
  const [age, setAge] = useState(member.age === null ? "" : String(member.age));
  const [seatingNeed, setSeatingNeed] = useState<SeatingNeed>(member.seating_need);
  const [dietaryNote, setDietaryNote] = useState(member.dietary_note);

  // The server clears the age when a child becomes an adult. The form clears
  // it visibly at the same moment, so the admin never builds a combination that will
  // be rejected and then reads about it.
  function changeKind(next: GuestKind) {
    setKind(next);
    if (next !== "child") {
      setAge("");
    }
  }

  async function submit(event: React.FormEvent) {
    event.preventDefault();

    try {
      await update.mutateAsync({
        name,
        kind,
        age: kind === "child" && age !== "" ? Number(age) : null,
        seating_need: seatingNeed,
        dietary_note: dietaryNote,
      });
    } catch {
      // Rendered per field below.
    }
  }

  return (
    <form onSubmit={submit}>
      <fieldset className="border-line flex flex-col gap-3 rounded-xl border p-4">
        {/* A named group, so a screen reader announces *whose* field this is: twenty
          members on one page otherwise means twenty controls called "Name" with
          nothing to tell them apart. The legend uses the stored name rather than the
          field's current value, so it does not change under the cursor. */}
        <legend className="text-body font-semibold">{member.name}</legend>

        {/* Only guest-added members are marked. A badge on every ordinary guest would
          drown out the one that matters. */}
        {guestOriginLabels[member.origin] ? (
          <span className="bg-accent-soft text-accent-strong text-small self-start rounded-lg px-2 py-1">
            {guestOriginLabels[member.origin]}
          </span>
        ) : null}

        <div className="flex flex-wrap gap-3">
          <Field label={householdLabels.nameLabel} id={`name-${member.id}`} error={fieldError(update.error, "name")}>
            <Input id={`name-${member.id}`} value={name} onChange={(event) => setName(event.target.value)} />
          </Field>

          <Field label={householdLabels.kindLabel} id={`kind-${member.id}`} error={fieldError(update.error, "kind")}>
            <NativeSelect
              id={`kind-${member.id}`}
              value={kind}
              onChange={(event) => changeKind(event.target.value as GuestKind)}
            >
              {Object.entries(guestKindLabels).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </NativeSelect>
          </Field>

          {kind === "child" ? (
            <Field
              label={householdLabels.ageLabel}
              id={`age-${member.id}`}
              hint={householdLabels.ageHint}
              error={fieldError(update.error, "age")}
            >
              <Input
                id={`age-${member.id}`}
                type="number"
                min={0}
                max={17}
                value={age}
                onChange={(event) => setAge(event.target.value)}
              />
            </Field>
          ) : null}

          <Field
            label={householdLabels.seatingNeedLabel}
            id={`seating-need-${member.id}`}
            error={fieldError(update.error, "seating_need")}
          >
            <NativeSelect
              id={`seating-need-${member.id}`}
              value={seatingNeed}
              onChange={(event) => setSeatingNeed(event.target.value as SeatingNeed)}
            >
              {Object.entries(seatingNeedLabels).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </NativeSelect>
          </Field>

          <Field
            label={householdLabels.dietaryNoteLabel}
            id={`dietary-note-${member.id}`}
            error={fieldError(update.error, "dietary_note")}
          >
            <Input
              id={`dietary-note-${member.id}`}
              value={dietaryNote}
              onChange={(event) => setDietaryNote(event.target.value)}
            />
          </Field>
        </div>

        <div className="flex items-center gap-3">
          <Button type="submit" size="sm" disabled={update.isPending}>
            {update.isPending ? householdLabels.saving : householdLabels.save}
          </Button>

          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button type="button" variant="outline" size="sm" disabled={remove.isPending}>
                {householdLabels.removeMember}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogTitle>{householdLabels.removeMemberConfirmTitle(member.name)}</AlertDialogTitle>
              <AlertDialogDescription>{householdLabels.removeMemberConfirmBody}</AlertDialogDescription>
              <AlertDialogFooter>
                <AlertDialogCancel>{householdLabels.cancel}</AlertDialogCancel>
                <AlertDialogAction onClick={() => remove.mutate()}>
                  {householdLabels.removeMemberConfirm}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          {update.isSuccess ? (
            <span role="status" className="text-small text-ink-muted">
              {householdLabels.saved}
            </span>
          ) : null}
        </div>
      </fieldset>
    </form>
  );
}

/**
 * The bulk-entry form: it stays open after a save and puts the cursor back in the
 * first-name field.
 *
 * Eighty guests entered four at a time is the actual workload, and a form that closed
 * after each one would triple the clicking.
 */
function AddMemberForm({ household }: { household: AdminHousehold }) {
  const add = useAddMember(household.id);
  const nameField = useRef<HTMLInputElement>(null);

  // Nothing is prefilled from the household. "Household" is a flexible term here —
  // one or more people who share an invitation — so display_name is free text like
  // "Luki & Paddi" or a single person's name, and a name derived from it would be
  // wrong as often as right.
  const [name, setName] = useState("");
  const [kind, setKind] = useState<GuestKind>("adult");
  const [age, setAge] = useState("");

  async function submit(event: React.FormEvent) {
    event.preventDefault();

    try {
      await add.mutateAsync({
        name,
        kind,
        age: kind === "child" && age !== "" ? Number(age) : null,
        seating_need: "normal",
        dietary_note: "",
      });

      setName("");
      setAge("");
      nameField.current?.focus();
    } catch {
      // Rendered per field below; the values stay so nothing has to be retyped.
    }
  }

  return (
    <form onSubmit={submit}>
      <fieldset className="bg-surface-sunken flex flex-col gap-3 rounded-xl p-4">
        {/* Named for the same reason the member rows are: the page carries several
            controls called "Name", and the group is what says which is which. */}
        <legend className="text-body font-semibold">{householdLabels.addMemberHeading}</legend>

        <div className="flex flex-wrap gap-3">
          <Field
            label={householdLabels.nameLabel}
            id="new-name"
            hint={householdLabels.nameHint}
            error={fieldError(add.error, "name")}
          >
            <Input
              id="new-name"
              ref={nameField}
              value={name}
              onChange={(event) => setName(event.target.value)}
              required
            />
          </Field>

          <Field label={householdLabels.kindLabel} id="new-kind" error={fieldError(add.error, "kind")}>
            <NativeSelect
              id="new-kind"
              value={kind}
              onChange={(event) => {
                const next = event.target.value as GuestKind;
                setKind(next);
                if (next !== "child") {
                  setAge("");
                }
              }}
            >
              {Object.entries(guestKindLabels).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </NativeSelect>
          </Field>

          {kind === "child" ? (
            <Field
              label={householdLabels.ageLabel}
              id="new-age"
              hint={householdLabels.ageHint}
              error={fieldError(add.error, "age")}
            >
              <Input
                id="new-age"
                type="number"
                min={0}
                max={17}
                value={age}
                onChange={(event) => setAge(event.target.value)}
              />
            </Field>
          ) : null}
        </div>

        <Button type="submit" size="sm" className="self-start" disabled={add.isPending}>
          {add.isPending ? householdLabels.addingMember : householdLabels.addMember}
        </Button>

        {add.error && !fieldError(add.error, "name") ? (
          <p role="alert" className="text-accent-strong text-small">
            {add.error.message}
          </p>
        ) : null}
      </fieldset>
    </form>
  );
}

/**
 * The household's answers, read-only.
 *
 * F3-F06 puts the *same* RSVP form component the guests use behind
 * /api/admin/households/{id}/rsvp and links from here. A second form here would be a
 * second field set to keep in step, which is the one thing Gate 1 exists to prevent.
 */
function RSVPSummary({ household }: { household: AdminHousehold }) {
  return (
    <Section title={householdLabels.rsvpHeading}>
      <p className="text-ink-muted">
        {household.rsvp_submitted_at
          ? householdLabels.rsvpAnsweredAt(formatShortDate(household.rsvp_submitted_at))
          : householdLabels.rsvpNotAnswered}
      </p>
      <p className="text-ink-muted text-small">{householdLabels.rsvpComingSoon}</p>
    </Section>
  );
}

/**
 * Deleting the household, behind a confirmation that names what goes with it.
 *
 * The dialog says the audit trail survives, because the fear when deleting is losing
 * the record rather than losing the row.
 */
function DeleteSection({ household }: { household: AdminHousehold }) {
  const remove = useDeleteHousehold(household.id);
  const navigate = Route.useNavigate();

  async function confirmDelete() {
    try {
      await remove.mutateAsync();
      await navigate({ to: "/admin/haushalte" });
    } catch {
      // Rendered below, and the admin stays on the page they were on.
    }
  }

  return (
    <Section title={householdLabels.deleteHeading}>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button type="button" variant="destructive" className="self-start" disabled={remove.isPending}>
            {householdLabels.delete}
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogTitle>{householdLabels.deleteConfirmTitle}</AlertDialogTitle>
          <AlertDialogDescription>
            {householdLabels.deleteConfirmBody(household.display_name, household.members.length)}
          </AlertDialogDescription>
          <AlertDialogFooter>
            <AlertDialogCancel>{householdLabels.cancel}</AlertDialogCancel>
            {/* The list is only navigated to once the delete has actually succeeded,
                so a failure leaves the admin here with the message. */}
            <AlertDialogAction
              className={buttonVariants({ variant: "destructive" })}
              onClick={() => void confirmDelete()}
            >
              {householdLabels.deleteConfirm}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {remove.error ? (
        <p role="alert" className="text-accent-strong text-small">
          {remove.error.message}
        </p>
      ) : null}
    </Section>
  );
}

/**
 * A labelled field with an optional hint and an error message.
 *
 * Every control on this page goes through it, which is what makes "every field has a
 * real label" a property of the page rather than a thing to remember: there is no way
 * to render an input here without one.
 */
function Field({
  label,
  id,
  hint,
  error,
  children,
}: {
  label: string;
  id: string;
  hint?: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-w-40 flex-col gap-1">
      <Label htmlFor={id}>{label}</Label>
      {children}
      {hint ? (
        <span id={`${id}-hint`} className="text-ink-muted text-small">
          {hint}
        </span>
      ) : null}
      {error ? (
        <span role="alert" className="text-accent-strong text-small">
          {error}
        </span>
      ) : null}
    </div>
  );
}
