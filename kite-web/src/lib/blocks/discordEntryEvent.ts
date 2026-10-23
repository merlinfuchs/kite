import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordEntryEvent: BlockDefinition = {
  type: "entry_event",
  title: "Listen for Event",
  description:
    "Listens for an event to trigger the flow. Drop different actions here!",
  icon: "satellite-dish",
  contexts: ["event_discord", "event_schedule", "event_webhook"],
  fixed: true,
  component: "entry_event",
  requires: ["discord"],
  custom_label: false,
  fields: [
    {
      name: "event_type",
      schema: z
        .enum([
          "message_create",
          "message_update",
          "message_delete",
          "guild_member_add",
          "guild_member_remove",
          "message_reaction_add",
          "message_reaction_remove",
          "guild_create",
          "guild_delete",
          "direct_message_create",
          "guild_boost",
          "guild_ban_add",
          "guild_ban_remove",
          "channel_create",
          "channel_update",
          "channel_delete",
          "guild_member_role_add",
          "guild_member_role_remove",
          "guild_member_timeout_add",
          "guild_member_timeout_remove",
          "cron",
          "webhook",
        ])
        .describe(
          "Discord event that triggers the flow, cron for a scheduled flow, or webhook for a flow that runs when a request is sent to its webhook URL. guild_member_role_add and guild_member_role_remove run once per role, with the role in {{role}}. channel_update has the channel from before in {{event.before}} and what changed in {{event.changes}}."
        ),
    },
    {
      name: "event_schedule_cron",
      schema: z
        .string()
        .max(100)
        .optional()
        .describe(
          "Cron expression in UTC for scheduled flows, e.g. */5 * * * * for every five minutes."
        ),
    },
    {
      name: "description",
      schema: z
        .string()
        .max(100)
        .min(1)
        .describe("Description of what the event listener does."),
    },
  ],
  run: { kind: "custom" },
};
