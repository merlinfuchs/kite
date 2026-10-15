import { OpenPanel } from "@openpanel/web";

// Events of the onboarding funnel, from signing up to the first command
// running in Discord. Page views are tracked on their own.
export type AnalyticsEvent =
  | "sign_up"
  | "app_create_opened"
  | "app_create_step"
  | "app_create_failed"
  | "app_created"
  | "setup_step_done"
  | "setup_checklist_dismissed"
  | "command_created"
  | "token_updated"
  | "commands_synced";

type Properties = Record<string, string | number | boolean>;

function createOpenPanel() {
  const op = new OpenPanel({
    clientId: "b14a6614-59bb-481e-8de0-ae685dd67de1",
    apiUrl: "/ops",
    trackScreenViews: true,
    trackOutgoingLinks: true,
    // Queues the first page view until the endpoint below is set.
    disabled: true,
  });

  // The SDK posts everything to apiUrl + "/track", which blockers match. nginx
  // proxies /ops itself to OpenPanel's track endpoint.
  const fetch = op.api.fetch.bind(op.api);
  op.api.fetch = (path, data, options) =>
    fetch(path === "/track" ? "" : path, data, options);

  op.ready();
  return op;
}

// Only exists in production, so dev and the static export send nothing.
export const openPanel =
  process.env.NODE_ENV === "production" && typeof window !== "undefined"
    ? createOpenPanel()
    : null;

function track(event: AnalyticsEvent, properties?: Properties) {
  openPanel?.track(event, properties);
}

export function useTrack() {
  return track;
}
