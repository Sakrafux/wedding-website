import { createFileRoute } from "@tanstack/react-router";

import { adminLabels } from "@/lib/labels";

export const Route = createFileRoute("/admin/_shell/")({
  component: AdminHome,
});

/**
 * The admin landing page: proof that the login and the shell work, and nothing
 * more. F6-F01 puts the dashboard here.
 */
function AdminHome() {
  return <p className="text-ink-muted">{adminLabels.dashboardPlaceholder}</p>;
}
