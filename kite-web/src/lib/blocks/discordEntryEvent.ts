import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordEntryEvent: BlockDefinition = {
  type: "entry_event",
  id_prefix: "event",
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
          "guild_audit_log_entry_create",
          "auto_moderation_action_execution",
          "cron",
          "webhook",
        ])
        .describe(
          "Discord event that triggers the flow, cron for a scheduled flow, or webhook for a flow that runs when a request is sent to its webhook URL. guild_audit_log_entry_create runs for every audit log entry, so pick the actions with an Event Filter on audit_log_action. Its placeholders are {{audit_log.action}}, {{audit_log.reason}}, {{audit_log.target}} and more, and {{user}} is whoever did it. auto_moderation_action_execution runs once for each action AutoMod takes, so pick them with an Event Filter on automod_action. Its placeholders are {{automod.action}}, {{automod.rule_name}}, {{automod.trigger}}, {{automod.content}}, {{automod.matched_keyword}} and more, and {{user}} is whoever wrote what AutoMod acted on."
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
