import { useOpenPanel } from "@openpanel/nextjs";
import { useCallback } from "react";

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

// Does nothing outside of production, where the OpenPanel script isn't loaded.
export function useTrack() {
  const { track } = useOpenPanel();
  return useCallback(
    (event: AnalyticsEvent, properties?: Properties) =>
      track(event, properties),
    [track]
  );
}
