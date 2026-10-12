import {
  BanIcon,
  ClockIcon,
  GemIcon,
  LogInIcon,
  LogOutIcon,
  LucideIcon,
  MessageSquareDiffIcon,
  MessageSquareIcon,
  MessageSquareXIcon,
  ServerCrashIcon,
  ServerIcon,
  ShieldCheckIcon,
  ShieldMinusIcon,
  ShieldPlusIcon,
  SmilePlusIcon,
  SmileIcon,
  SquarePenIcon,
  SquarePlusIcon,
  SquareXIcon,
  TimerOffIcon,
  TimerIcon,
} from "lucide-react";
import { EventTypeScheduleCron, EventTypeWebhook } from "../types/flow.gen";

export interface DiscordEventType {
  value: string;
  label: string;
  description: string;
  icon: LucideIcon;
  category: string;
  // Shown once the event is picked, for things worth knowing up front.
  note?: string;
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
  },
  {
    value: "message_update",
    label: "Message Update",
    description: "A message is edited",
    icon: MessageSquareDiffIcon,
    category: "Messages",
  },
  {
    value: "message_delete",
    label: "Message Delete",
    description: "A message is deleted",
    icon: MessageSquareXIcon,
    category: "Messages",
  },
  {
    value: "message_reaction_add",
    label: "Message Reaction Add",
    description: "A reaction is added to a message",
    icon: SmilePlusIcon,
    category: "Reactions",
  },
  {
    value: "message_reaction_remove",
    label: "Message Reaction Remove",
    description: "A reaction is removed from a message",
    icon: SmileIcon,
    category: "Reactions",
  },
  {
    value: "guild_member_add",
    label: "Server Member Add",
    description: "A member joins the server",
    icon: LogInIcon,
    category: "Members",
    note: membersIntent,
  },
  {
    value: "guild_member_remove",
    label: "Server Member Remove",
    description: "A member leaves the server",
    icon: LogOutIcon,
    category: "Members",
    note: membersIntent,
  },
  {
    value: "guild_member_role_add",
    label: "Member Role Add",
    description: "A role is added to a member",
    icon: ShieldPlusIcon,
    category: "Members",
    note: `Runs once for every role added, with the role in {{role}}. ${membersIntent}`,
  },
  {
    value: "guild_member_role_remove",
    label: "Member Role Remove",
    description: "A role is removed from a member",
    icon: ShieldMinusIcon,
    category: "Members",
    note: `Runs once for every role removed, with the role in {{role}}. ${membersIntent}`,
  },
  {
    value: "guild_boost",
    label: "Server Boost",
    description: "A member boosts the server",
    icon: GemIcon,
    category: "Members",
    note: "Boosts are detected from the message Discord posts when someone boosts the server. The server needs a system messages channel with boost messages enabled, and your bot needs to see it.",
  },
  {
    value: "guild_ban_add",
    label: "Member Banned",
    description: "A member is banned",
    icon: BanIcon,
    category: "Moderation",
  },
  {
    value: "guild_ban_remove",
    label: "Member Unbanned",
    description: "A member is unbanned",
    icon: ShieldCheckIcon,
    category: "Moderation",
  },
  {
    value: "guild_member_timeout_add",
    label: "Member Timeout Add",
    description: "A member is timed out",
    icon: TimerIcon,
    category: "Moderation",
    note: `Also runs when a timeout is changed to end at another time. ${membersIntent}`,
  },
  {
    value: "guild_member_timeout_remove",
    label: "Member Timeout Remove",
    description: "A member's timeout is removed",
    icon: TimerOffIcon,
    category: "Moderation",
    note: `Discord doesn't say when a timeout runs out by itself. ${membersIntent}`,
  },
  {
    value: "channel_create",
    label: "Channel Create",
    description: "A channel is created",
    icon: SquarePlusIcon,
    category: "Channels",
  },
  {
    value: "channel_update",
    label: "Channel Update",
    description: "A channel's settings change",
    icon: SquarePenIcon,
    category: "Channels",
    note: "Moving a channel up or down in the list doesn't count, moving it to another category does. Threads aren't included.",
  },
  {
    value: "channel_delete",
    label: "Channel Delete",
    description: "A channel is deleted",
    icon: SquareXIcon,
    category: "Channels",
  },
  {
    value: "guild_create",
    label: "Bot Joined Server",
    description: "Your bot is added to a server",
    icon: ServerIcon,
    category: "Bot",
  },
  {
    value: "guild_delete",
    label: "Bot Left Server",
    description: "Your bot is removed from a server",
    icon: ServerCrashIcon,
    category: "Bot",
  },
];

export function getDiscordEventType(value?: string) {
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

// eventListenerListPath is the page that lists event listeners of a source.
export function eventListenerListPath(source?: string) {
  switch (source) {
    case "schedule":
      return "/apps/[appId]/events/schedules";
    case "webhook":
      return "/apps/[appId]/events/webhooks";
    default:
      return "/apps/[appId]/events";
  }
}

// eventSourceTitles names the sub pages of event listeners.
export const eventSourceTitles: Record<string, string> = {
  discord: "Discord Events",
  schedule: "Scheduled Tasks",
  webhook: "Webhooks",
};

export interface EventSourceName {
  // Lowercase, for the middle of a sentence, like "Import webhook".
  singular: string;
  // Capitalized, for titles, like "Export Webhook".
  title: string;
}

const eventSourceNames: Record<string, EventSourceName> = {
  discord: { singular: "Discord event", title: "Discord Event" },
  schedule: { singular: "scheduled task", title: "Scheduled Task" },
  webhook: { singular: "webhook", title: "Webhook" },
};

// eventSourceName is what an event listener of a source is called in the UI.
export function eventSourceName(source?: string): EventSourceName {
  return eventSourceNames[source ?? ""] ?? eventSourceNames.discord;
}

// capitalize makes a singular name start a sentence, like "Webhook deleted!".
export function capitalize(text: string) {
  return text.charAt(0).toUpperCase() + text.slice(1);
}

// eventSourceOfFlow is the source an event listener with this flow gets. The
// service derives it from the event type of the entry block the same way.
export function eventSourceOfFlow(flow?: {
  nodes?: { type?: string; data?: { event_type?: string } }[];
}): string {
  const type = flow?.nodes?.find((n) => n.type === "entry_event")?.data
    ?.event_type;
  if (type === EventTypeScheduleCron) return "schedule";
  if (type === EventTypeWebhook) return "webhook";
  return "discord";
}
