import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Navigate, Outlet, redirect, useNavigate } from "@tanstack/react-router";

import { Button } from "@/components/ui/button";
import { adminLabels } from "@/lib/labels";
import { adminSessionQueryOptions, useLogout } from "@/lib/api/session";

/**
 * Everything behind the admin login.
 *
 * A pathless layout so /admin/login stays outside it — a guard that covered the
 * login page too would redirect anyone trying to log in, forever.
 */
export const Route = createFileRoute("/admin/_shell")({
  beforeLoad: async ({ context }) => {
    const session = await context.queryClient.ensureQueryData(adminSessionQueryOptions);

    if (!session) {
      // To the *admin* login, never the guest one. An admin whose eight hours ran
      // out mid-edit landing on "Gib den Code von deiner Einladungskarte ein"
      // would be baffling, and there is no code to give.
      throw redirect({ to: "/admin/login", search: { reason: "expired" } });
    }
  },
  component: AdminShell,
});

/**
 * Nav entries for the admin sections.
 *
 * Every one is a placeholder today. Rendering them as disabled rather than
 * omitting them is better than a 404 and keeps the remaining work visible; F5, F6,
 * F7, F8 and F9 replace them with links one at a time.
 */
const navItems = [
  adminLabels.navHouseholds,
  adminLabels.navDashboard,
  adminLabels.navSeating,
  adminLabels.navBudget,
  adminLabels.navPhotos,
];

function AdminShell() {
  const navigate = useNavigate();
  const logout = useLogout();
  // Eight hours is short enough that running out mid-edit is a normal occurrence,
  // not an edge case. Same reasoning as the guest layout: the guard covers
  // navigation, this covers sitting still.
  const { data: session } = useQuery(adminSessionQueryOptions);

  async function signOut() {
    await logout.mutateAsync();
    await navigate({ to: "/admin/login" });
  }

  if (!session) {
    return <Navigate to="/admin/login" search={{ reason: "expired" }} replace />;
  }

  return (
    // Denser than the guest side, and no serif display sizes: same tokens,
    // different register, per 05-design.
    <div className="mx-auto flex min-h-dvh max-w-5xl flex-col gap-6 px-4 py-6">
      <header className="border-line flex items-center justify-between gap-4 border-b pb-4">
        <span className="text-h3 font-body font-semibold">{adminLabels.heading}</span>
        <Button variant="outline" size="sm" onClick={() => void signOut()} disabled={logout.isPending}>
          {adminLabels.logout}
        </Button>
      </header>

      <nav aria-label={adminLabels.heading}>
        <ul className="flex flex-wrap gap-2">
          {navItems.map((item) => (
            <li key={item}>
              <span
                className="text-small text-ink-muted border-line inline-flex items-center gap-2 rounded-lg border px-3 py-2"
                aria-disabled="true"
              >
                {item}
                <span className="text-ink-muted/70">({adminLabels.comingSoon})</span>
              </span>
            </li>
          ))}
        </ul>
      </nav>

      <Outlet />
    </div>
  );
}
