import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { adminLabels } from "@/lib/labels";
import { adminSessionQueryOptions, useAdminLogin } from "@/lib/api/session";

interface AdminLoginSearch {
  /** Set when a session ran out mid-visit, so the screen can say why. */
  reason?: "expired";
}

export const Route = createFileRoute("/admin/login")({
  validateSearch: (search: Record<string, unknown>): AdminLoginSearch => ({
    reason: search.reason === "expired" ? "expired" : undefined,
  }),
  beforeLoad: async ({ context }) => {
    const session = await context.queryClient.ensureQueryData(adminSessionQueryOptions);
    if (session) {
      throw redirect({ to: "/admin" });
    }
  },
  component: AdminLoginPage,
});

/**
 * The admin login.
 *
 * Not linked from anywhere in the guest UI, and reachable only by typing the URL.
 * That is not a security control — RequireAdmin on the server is — but it keeps a
 * curious guest
 * from finding a login form and wondering what is behind it.
 */
function AdminLoginPage() {
  const { reason } = Route.useSearch();
  const navigate = useNavigate();
  const login = useAdminLogin();

  const [user, setUser] = useState("");
  const [password, setPassword] = useState("");

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    try {
      await login.mutateAsync({ user, password });
      await navigate({ to: "/admin" });
    } catch {
      // Shown from login.error below; nothing else to do with it here.
    }
  }

  return (
    <div className="mx-auto flex min-h-dvh max-w-sm flex-col justify-center gap-6 px-4 py-12">
      {/* No display serif and no decoration: the admin side shares the tokens and
          drops the guest styling, per 05-design. */}
      <h1 className="text-h2 font-body">{adminLabels.heading}</h1>

      {reason === "expired" ? (
        <p role="status" className="bg-surface-sunken text-small rounded-lg px-4 py-3">
          {adminLabels.sessionExpired}
        </p>
      ) : null}

      {/* A plain form with the standard autocomplete tokens: this is the one login
          in the product that belongs in a password manager. */}
      <form onSubmit={submit} className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <Label htmlFor="admin-user">{adminLabels.userLabel}</Label>
          <Input
            id="admin-user"
            name="username"
            value={user}
            onChange={(event) => setUser(event.target.value)}
            autoComplete="username"
            autoCapitalize="none"
            autoCorrect="off"
            spellCheck={false}
            disabled={login.isPending}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="admin-password">{adminLabels.passwordLabel}</Label>
          <Input
            id="admin-password"
            name="password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            autoComplete="current-password"
            disabled={login.isPending}
          />
        </div>

        {login.error ? (
          // The API's German sentence, verbatim. It says only "Anmeldung
          // fehlgeschlagen" — which half was wrong is deliberately not reported.
          <p role="alert" className="text-small text-danger">
            {login.error.message}
          </p>
        ) : null}

        <Button type="submit" disabled={login.isPending || user === "" || password === ""}>
          {login.isPending ? adminLabels.submitting : adminLabels.submit}
        </Button>
      </form>
    </div>
  );
}
