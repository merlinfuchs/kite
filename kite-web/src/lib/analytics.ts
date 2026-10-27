import env from "@/lib/env/client";

// Page views are tracked on their own.
export type AnalyticsEvent =
  // The onboarding funnel, from signing up to the first command running in
  // Discord.
  | "sign_up"
  | "app_create_opened"
  | "app_create_step"
  | "app_create_failed"
  | "app_create_portal_opened"
  | "app_created"
  | "setup_step_done"
  | "setup_checklist_dismissed"
  | "command_created"
  | "token_updated"
  | "commands_synced"
  // The flow editor.
  | "block_added"
  | "blocks_pasted"
  | "note_added"
  | "flow_ai_opened"
  | "flow_ai_prompt_sent"
  | "flow_ai_prompt_answered"
  | "flow_ai_prompt_failed"
  | "flow_ai_build_clicked"
  | "flow_ai_changes_kept"
  | "flow_ai_changes_undone"
  | "flow_ai_limit_reached"
  | "command_saved"
  | "event_listener_saved"
  | "flow_save_failed"
  // Everything else in an app.
  | "event_listener_created"
  | "message_created"
  | "message_saved"
  | "message_sent"
  | "template_added"
  | "template_updated"
  | "template_removed"
  | "variable_created"
  | "secret_created"
  | "api_token_created"
  | "plugin_enabled"
  | "integration_enabled"
  | "collaborator_added"
  | "flow_exported"
  | "bundle_imported"
  | "app_deleted"
  // Premium.
  | "upsell_shown"
  | "upsell_clicked"
  | "checkout_started"
  | "plan_switched";

type Properties = Record<string, string | number | boolean>;

async function createOpenPanel(clientId: string) {
  const { OpenPanel } = await import("@openpanel/web");
  const op = new OpenPanel({
    clientId,
    apiUrl: "/ops",
    trackScreenViews: true,
    trackOutgoingLinks: true,
  });

  // The SDK posts everything to apiUrl + "/track", which blockers match. nginx
  // proxies /ops itself to OpenPanel's track endpoint.
  const fetch = op.api.fetch.bind(op.api);
  op.api.fetch = (path, data, options) =>
    fetch(path === "/track" ? "" : path, data, options);

  return op;
}

// Loaded after the page, since the SDK bundles rrweb for session replays.
export const openPanel =
  env.NEXT_PUBLIC_OPENPANEL_CLIENT_ID && typeof window !== "undefined"
    ? createOpenPanel(env.NEXT_PUBLIC_OPENPANEL_CLIENT_ID)
    : null;

export function track(event: AnalyticsEvent, properties?: Properties) {
  openPanel?.then((op) => op.track(event, properties));
}

// Events are grouped by the app they happen in, so features can be counted by
// apps and not only by users. The user is linked to the app once identified.
export function setAppGroup(appId: string | undefined, identified: boolean) {
  openPanel?.then((op) => {
    // setGroup adds to the groups of every following event, which must only
    // be the current app.
    op.groups = appId ? [appId] : [];
    if (appId && identified) op.setGroup(appId);
  });
}

export function updateAppGroup(
  app: { id: string; name: string },
  properties: Properties
) {
  openPanel?.then((op) =>
    op.upsertGroup({ id: app.id, type: "app", name: app.name, properties })
  );
}
