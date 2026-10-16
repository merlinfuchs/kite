import env from "@/lib/env/client";

// Events of the onboarding funnel, from signing up to the first command
// running in Discord. Page views are tracked on their own.
export type AnalyticsEvent =
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
  | "commands_synced";

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

function track(event: AnalyticsEvent, properties?: Properties) {
  openPanel?.then((op) => op.track(event, properties));
}

export function useTrack() {
  return track;
}
