import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { CircleAlert } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { householdsQueryOptions, useCreateHousehold } from "@/lib/api/households";
import type { AdminHouseholdOverview } from "@/lib/api/dto";
import { formatRelativeDays, formatShortDate } from "@/lib/dates";
import { householdLabels } from "@/lib/labels";

export const Route = createFileRoute("/admin/_shell/haushalte/")({
  component: HouseholdListPage,
});

/**
 * Every household on one screen: who exists, what their code is, and whether they
 * have ever logged in.
 *
 * The data is fetched with useQuery rather than in a route loader on purpose. A 401
 * from here has to reach the session query — that is what makes the admin shell
 * redirect to the *admin* login — and a loader that threw would land on the route
 * error boundary instead, showing "etwas ist schiefgegangen" to somebody whose
 * session has merely run out.
 */
function HouseholdListPage() {
  const { data: households, isPending, error } = useQuery(householdsQueryOptions);

  const [search, setSearch] = useState("");
  const [onlyNeverLoggedIn, setOnlyNeverLoggedIn] = useState(false);
  const [onlyUnanswered, setOnlyUnanswered] = useState(false);

  if (isPending) {
    return (
      <div aria-busy="true" className="flex flex-col gap-3">
        <p className="sr-only">{householdLabels.loading}</p>
        <div className="bg-surface-sunken h-10 animate-pulse rounded-lg" />
        <div className="bg-surface-sunken h-40 animate-pulse rounded-lg" />
      </div>
    );
  }

  if (error) {
    // The API's own German sentence, or NetworkError's, shown verbatim — the server
    // owns the wording so every screen says the same thing.
    return <p role="alert">{error.message}</p>;
  }

  const neverLoggedIn = households.filter((household) => !household.last_login_at).length;

  // Filtered in the browser: sixty rows are already in memory, and a round trip per
  // keystroke would be latency for nothing. The two filters are an AND — two filters
  // that quietly ORed would produce a list nobody can explain.
  const visible = households.filter((household) => {
    if (onlyNeverLoggedIn && household.last_login_at) {
      return false;
    }
    if (onlyUnanswered && household.rsvp_submitted_at) {
      return false;
    }
    return household.display_name.toLocaleLowerCase("de-DE").includes(search.toLocaleLowerCase("de-DE"));
  });

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-h2 font-body">{householdLabels.heading}</h1>

      <p className="text-ink-muted text-small">{householdLabels.summary(households.length, neverLoggedIn)}</p>

      {/* Above the table, because the table is the only element on this page that
          grows: everything above it keeps its position for the life of the project
          (F5-F04). */}
      <CreateHouseholdForm />

      {/* Directly above the table they filter, not up in the action strip: "what I do"
          and "what I see" are different rows. */}
      <div className="flex flex-wrap items-end gap-4">
        <div className="flex min-w-56 flex-col gap-2">
          <Label htmlFor="household-search">{householdLabels.searchLabel}</Label>
          <Input
            id="household-search"
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            autoComplete="off"
          />
        </div>

        <Label className="min-h-12 items-center gap-2">
          <input
            type="checkbox"
            checked={onlyNeverLoggedIn}
            onChange={(event) => setOnlyNeverLoggedIn(event.target.checked)}
            className="size-5"
          />
          {householdLabels.onlyNeverLoggedIn}
        </Label>

        <Label className="min-h-12 items-center gap-2">
          <input
            type="checkbox"
            checked={onlyUnanswered}
            onChange={(event) => setOnlyUnanswered(event.target.checked)}
            className="size-5"
          />
          {householdLabels.onlyUnanswered}
        </Label>
      </div>

      {visible.length === 0 ? (
        <p className="text-ink-muted">{households.length === 0 ? householdLabels.empty : householdLabels.noMatches}</p>
      ) : (
        <Table>
          <TableCaption>{householdLabels.tableCaption}</TableCaption>
          {/* Sticky, with the surface colour set: the table is the tail of the page
              now, and a transparent sticky header shows rows sliding under the words. */}
          <TableHeader className="bg-surface sticky top-0 z-10">
            <TableRow>
              <TableHead scope="col">{householdLabels.columnHousehold}</TableHead>
              <TableHead scope="col">{householdLabels.columnCode}</TableHead>
              <TableHead scope="col">{householdLabels.columnMembers}</TableHead>
              <TableHead scope="col">{householdLabels.columnLastLogin}</TableHead>
              <TableHead scope="col">{householdLabels.columnRSVP}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {visible.map((household) => (
              <HouseholdRow key={household.id} household={household} />
            ))}
          </TableBody>
        </Table>
      )}

      <ExportLinks households={households.length} />
    </div>
  );
}

function HouseholdRow({ household }: { household: AdminHouseholdOverview }) {
  return (
    <TableRow>
      <TableCell>
        <Link
          to="/admin/haushalte/$householdId"
          params={{ householdId: String(household.id) }}
          className="underline underline-offset-4"
        >
          {household.display_name}
        </Link>
      </TableCell>
      {/* Monospaced: the only use of a code on this screen is being read aloud or
          compared against a printed card, character by character. */}
      <TableCell className="font-mono">{household.code}</TableCell>
      <TableCell>{household.member_count}</TableCell>
      <TableCell>
        <LastLoginCell lastLoginAt={household.last_login_at} />
      </TableCell>
      <TableCell>{household.rsvp_submitted_at ? householdLabels.rsvpAnswered : householdLabels.rsvpOpen}</TableCell>
    </TableRow>
  );
}

/**
 * The column that answers "have they even seen it?", which is what drives the nudge
 * calls before send-out.
 *
 * A household that never logged in gets a marker with an icon and a word, never a
 * blank cell and never colour alone: a blank reads as "not loaded", and colour as the
 * only signal fails the accessibility rule in 05-design.
 */
function LastLoginCell({ lastLoginAt }: { lastLoginAt: string | null }) {
  if (!lastLoginAt) {
    return (
      <span className="text-accent-strong inline-flex items-center gap-1 whitespace-nowrap">
        <CircleAlert aria-hidden="true" className="size-4" />
        {householdLabels.neverLoggedIn}
      </span>
    );
  }

  const relative = formatRelativeDays(lastLoginAt);

  return (
    <span className="whitespace-nowrap">
      {formatShortDate(lastLoginAt)}
      {/* The absolute date is what gets read out on the phone; the relative hint is
          what makes the column scannable. */}
      {relative ? <span className="text-ink-muted"> ({relative})</span> : null}
    </span>
  );
}

/**
 * Creating a household asks for the name and nothing else, then goes straight to its
 * detail page — where the code is now visible and the members can be entered.
 * Creating and then hunting for the row you just created is the flow to avoid.
 *
 * An inline one-row form rather than a dialog: this is the control used sixty times in
 * a single seeding sitting (E-OPS-01), and hiding it behind a tap buys nothing.
 */
function CreateHouseholdForm() {
  const navigate = useNavigate();
  const create = useCreateHousehold();
  const [displayName, setDisplayName] = useState("");

  async function submit(event: React.FormEvent) {
    event.preventDefault();

    try {
      const household = await create.mutateAsync(displayName);
      await navigate({ to: "/admin/haushalte/$householdId", params: { householdId: String(household.id) } });
    } catch {
      // Shown from create.error below.
    }
  }

  return (
    <form onSubmit={submit} className="border-line flex flex-col gap-3 border-t pt-6">
      <h2 className="text-h3 font-body">{householdLabels.createHeading}</h2>

      <div className="flex flex-wrap items-end gap-3">
        <div className="flex min-w-64 flex-col gap-2">
          <Label htmlFor="new-household-name">{householdLabels.createNameLabel}</Label>
          <Input
            id="new-household-name"
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            aria-describedby="new-household-name-hint"
            aria-invalid={Boolean(create.error)}
            disabled={create.isPending}
            required
          />
        </div>

        <Button type="submit" disabled={create.isPending}>
          {create.isPending ? householdLabels.createSubmitting : householdLabels.createSubmit}
        </Button>
      </div>

      <p id="new-household-name-hint" className="text-ink-muted text-small">
        {householdLabels.createNameHint}
      </p>

      {create.error ? (
        <p role="alert" className="text-accent-strong text-small">
          {create.error.message}
        </p>
      ) : null}
    </form>
  );
}

/**
 * The two CSV downloads, on this page rather than on an exports page of their own — a
 * separate page for two links is a page nobody would find.
 *
 * Plain `<a download>` and no fetch-and-blob dance: the cookie rides along, the
 * browser handles the save dialog, and there is no in-memory copy of the code list
 * sitting in a tab all afternoon. No progress UI either — sixty rows are
 * instantaneous, and a spinner would be a lie about the work involved.
 */
function ExportLinks({ households }: { households: number }) {
  return (
    <section className="border-line flex flex-col gap-4 border-t pt-6">
      <h2 className="text-h3 font-body">{householdLabels.exportHeading}</h2>

      <div className="flex flex-col gap-1">
        <a href="/api/admin/export/codes.csv" download className="self-start underline underline-offset-4">
          {householdLabels.exportCodes}
        </a>
        <span className="text-ink-muted text-small">{householdLabels.exportCodesCount(households)}</span>
        {/* The most sensitive artefact this application produces. The app cannot
            delete it again for anybody, so the sentence sits where the download
            happens. */}
        <p className="text-accent-strong text-small">{householdLabels.exportCodesWarning}</p>
      </div>

      <div className="flex flex-col gap-1">
        <a href="/api/admin/export/guests.csv" download className="self-start underline underline-offset-4">
          {householdLabels.exportGuests}
        </a>
        <p className="text-ink-muted text-small">{householdLabels.exportGuestsWarning}</p>
      </div>
    </section>
  );
}
