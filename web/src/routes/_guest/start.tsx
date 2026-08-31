import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";

import { Button } from "@/components/ui/button";
import { shellLabels } from "@/lib/labels";
import { meQueryOptions, useLogout } from "@/lib/api/session";

export const Route = createFileRoute("/_guest/start")({
  component: StartPage,
});

/**
 * The authenticated landing page.
 *
 * A placeholder with the household's name on it, so that "logging in works" is
 * something a guest can see. F2-F02 replaces it with the real start page — hero,
 * greeting and countdown — and F2-F01 adds the navigation around it.
 */
function StartPage() {
  const { data: me } = useSuspenseQuery(meQueryOptions);
  const navigate = useNavigate();
  const logout = useLogout();

  async function signOut() {
    await logout.mutateAsync();
    await navigate({ to: "/" });
  }

  return (
    <div className="mx-auto flex min-h-dvh max-w-md flex-col justify-center gap-6 px-4 py-12 text-center">
      <h1 className="text-display">{shellLabels.startHeading}</h1>
      <p className="text-ink-muted">{me?.household.display_name}</p>
      <p>{shellLabels.startIntro}</p>

      <Button variant="outline" size="lg" className="h-14 w-full" onClick={() => void signOut()}>
        {shellLabels.logout}
      </Button>
    </div>
  );
}
