import {
  ClockIcon,
  GemIcon,
  LogInIcon,
  LogOutIcon,
  LucideIcon,
  MailIcon,
  MessageSquareDiffIcon,
  MessageSquareIcon,
  MessageSquareXIcon,
  ScrollTextIcon,
  ServerCrashIcon,
  ServerIcon,
  SmilePlusIcon,
  SmileIcon,
  SatelliteDishIcon,
  TimerIcon,
  WebhookIcon,
} from "lucide-react";
import { EventTypeScheduleCron, EventTypeWebhook } from "../types/flow.gen";

// EventGroup is a group of placeholders a Discord event fills in. user is the
// member in the server, user_account the user without server fields, like for
// someone who left. The *_id groups only have the ID (and the user's mention).
export type EventGroup =
  | "user"
  | "user_account"
  | "user_id"
  | "server"
  | "server_id"
  | "channel"
  | "message"
  | "message_id"
  | "emoji"
  | "audit_log";

export interface DiscordEventType {
  value: string;
  label: string;
  description: string;
  icon: LucideIcon;
  category: string;
  // The placeholder groups the event fills in, in the order of the picker.
  placeholders: EventGroup[];
  // Shown once the event is picked, for things worth knowing up front.
  note?: string;
  // What one is called, for events that can happen many times a minute in a
  // busy server, where a listener that costs credits on every one adds up.
  frequent?: string;
}

const membersIntent = "Needs the Server Members Intent.";

// discordEventTypes lists the Discord events an event listener can run for,
// grouped by category in the order of the event picker.
export const discordEventTypes: DiscordEventType[] = [
  {
    value: "message_create",
    label: "Message Create",
    description: "A message is sent",
    icon: MessageSquareIcon,
    category: "Messages",
    placeholders: ["user", "server", "channel", "message"],
    frequent: "message",
  },
  {
    value: "direct_message_create",
    label: "Direct Message Received",
    description: "A user sends your bot a direct message",
    icon: MailIcon,
    category: "Messages",
    placeholders: ["user_account", "channel", "message"],
    note: "Message Create, Update and Delete only run for messages in servers.",
  },
  {
    value: "message_update",
    label: "Message Update",
    description: "A message is edited",
    icon: MessageSquareDiffIcon,
    category: "Messages",
    placeholders: ["user", "server", "channel", "message"],
    frequent: "edited message",
  },
  {
    value: "message_delete",
    label: "Message Delete",
    description: "A message is deleted",
    icon: MessageSquareXIcon,
    category: "Messages",
    // The message and its author are only known if the bot saw the message
    // before it was deleted.
    placeholders: ["user", "server", "channel", "message"],
    frequent: "deleted message",
  },
  {
    value: "message_reaction_add",
    label: "Message Reaction Add",
    description: "A reaction is added to a message",
    icon: SmilePlusIcon,
    category: "Reactions",
    placeholders: ["user", "server", "channel", "message_id", "emoji"],
    frequent: "reaction",
  },
  {
    value: "message_reaction_remove",
    label: "Message Reaction Remove",
    description: "A reaction is removed from a message",
    icon: SmileIcon,
    category: "Reactions",
    // Discord only sends the ID of the user who removed the reaction.
    placeholders: ["user_id", "server", "channel", "message_id", "emoji"],
    frequent: "removed reaction",
  },
  {
    value: "guild_member_add",
    label: "Server Member Add",
    description: "A member joins the server",
    icon: LogInIcon,
    category: "Members",
    placeholders: ["user", "server"],
    note: membersIntent,
  },
  {
    value: "guild_member_remove",
    label: "Server Member Remove",
    description: "A member leaves the server",
    icon: LogOutIcon,
    category: "Members",
    placeholders: ["user_account", "server"],
    note: membersIntent,
  },
  {
    value: "guild_boost",
    label: "Server Boost",
    description: "A member boosts the server",
    icon: GemIcon,
    category: "Members",
    placeholders: ["user", "server", "channel", "message_id"],
    note: "Boosts are detected from the message Discord posts when someone boosts the server. The server needs a system messages channel with boost messages enabled, and your bot needs to see it.",
  },
  {
    value: "guild_audit_log_entry_create",
    label: "Audit Log Entry",
    description: "Something is logged in the server's audit log",
    icon: ScrollTextIcon,
    category: "Moderation",
    // The user is whoever did it, the channel the one the entry is about.
    placeholders: ["audit_log", "user", "server", "channel"],
    note: "Runs for bans, kicks, timeouts, role changes, channel changes and more. Your bot needs the View Audit Log permission in the server. Pick the actions with an Event Filter.",
  },
  {
    value: "guild_create",
    label: "Bot Joined Server",
    description: "Your bot is added to a server",
    icon: ServerIcon,
    category: "Bot",
    placeholders: ["server"],
  },
  {
    value: "guild_delete",
    label: "Bot Left Server",
    description: "Your bot is removed from a server",
    icon: ServerCrashIcon,
    category: "Bot",
    // The bot isn't in the server anymore, so only its ID is known.
    placeholders: ["server_id"],
  },
];

// discordEventType is the entry of an event type, if it's one Kite knows.
export function discordEventType(value?: string) {
  return discordEventTypes.find((t) => t.value === value);
}

export interface SchedulePreset {
  label: string;
  description: string;
  cron: string;
  icon: LucideIcon;
}

// schedulePresets are common schedules, so most people never have to write a
// cron expression. All of them are allowed on the free plan.
export const schedulePresets: SchedulePreset[] = [
  {
    label: "Every 5 minutes",
    description: "At :00, :05, :10 and so on",
    cron: "*/5 * * * *",
    icon: TimerIcon,
  },
  {
    label: "Every hour",
    description: "At the start of every hour",
    cron: "0 * * * *",
    icon: ClockIcon,
  },
  {
    label: "Every day",
    description: "At 00:00 UTC",
    cron: "0 0 * * *",
    icon: ClockIcon,
  },
  {
    label: "Every week",
    description: "On Mondays at 00:00 UTC",
    cron: "0 0 * * 1",
    icon: ClockIcon,
  },
];

export type EventListenerSource = "discord" | "schedule" | "webhook";

export interface EventSourceName {
  // Lowercase, for the middle of a sentence, like "Import webhook".
  singular: string;
  // Capitalized, for titles, like "Export Webhook".
  title: string;
  // The title of the source's page, like "Webhooks".
  plural: string;
  // The page that lists the event listeners of the source.
  path: string;
  icon: LucideIcon;
}

const eventSourceNames: Record<EventListenerSource, EventSourceName> = {
  discord: {
    singular: "event listener",
    title: "Event Listener",
    plural: "Event Listeners",
    path: "/apps/[appId]/events",
    icon: SatelliteDishIcon,
  },
  schedule: {
    singular: "scheduled task",
    title: "Scheduled Task",
    plural: "Scheduled Tasks",
    path: "/apps/[appId]/schedules",
    icon: ClockIcon,
  },
  webhook: {
    singular: "webhook",
    title: "Webhook",
    plural: "Webhooks",
    path: "/apps/[appId]/webhooks",
    icon: WebhookIcon,
  },
};

// eventSources lists the sources in the order of their pages in the sidebar.
export const eventSources = Object.keys(
  eventSourceNames
) as EventListenerSource[];

// eventSourceName is what an event listener of a source is called in the UI.
export function eventSourceName(source?: string): EventSourceName {
  return (
    eventSourceNames[source as EventListenerSource] ?? eventSourceNames.discord
  );
}

// eventListenerLabel names an event listener by what runs it, like "Member
// Joins" or "Webhook".
export function eventListenerLabel(source: string, type: string) {
  return (
    (source === "discord" && discordEventType(type)?.label) ||
    eventSourceName(source).title
  );
}

// capitalize makes a singular name start a sentence, like "Webhook deleted!".
export function capitalize(text: string) {
  return text.charAt(0).toUpperCase() + text.slice(1);
}

// eventSourceOfFlow is the source an event listener with this flow gets. The
// service derives it from the event type of the entry block the same way.
export function eventSourceOfFlow(flow?: {
  nodes?: { type?: string; data?: { event_type?: string } }[];
}): EventListenerSource {
  const type = flow?.nodes?.find((n) => n.type === "entry_event")?.data
    ?.event_type;
  if (type === EventTypeScheduleCron) return "schedule";
  if (type === EventTypeWebhook) return "webhook";
  return "discord";
}
