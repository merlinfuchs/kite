import generated from "./eventFilterTargets.gen.json";

// What event filters can check for each event type, generated from
// eventFilterTargets in the service with `go test ./pkg/flow -update`. A target
// that isn't listed doesn't exist for the event, and the service rejects it on
// save.
export const eventFilterTargets: Record<string, string[]> = generated.events;

// Yes or no filters have no value, equal means yes and not_equal no.
export const booleanEventFilterTargets: string[] = generated.boolean;

const eventFilterModes: Record<string, string[]> = generated.modes;

export const eventFilterTargetLabels: Record<string, string> = {
  message_content: "Message Content",
  user_id: "User ID",
  guild_id: "Guild ID",
  channel_id: "Channel ID",
  message_id: "Message ID",
  role_id: "Member Role",
  emoji: "Reaction Emoji",
  message_has_attachment: "Message Has Attachment",
  message_mentions_bot: "Message Mentions Bot",
  user_is_bot: "User Is Bot",
  audit_log_action: "Audit Log Action",
};

// getEventFilterModes returns the modes a target can use, undefined if it can
// use all.
export function getEventFilterModes(target?: string): string[] | undefined {
  return target ? eventFilterModes[target] : undefined;
}

// getEventFilterTargets returns the targets available for an event type, or
// every target if the filter isn't connected to an event yet.
export function getEventFilterTargets(eventType?: string) {
  if (!eventType) return Object.keys(eventFilterTargetLabels);
  return eventFilterTargets[eventType] ?? [];
}
