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
          "message_poll_vote_add",
          "message_poll_vote_remove",
          "message_poll_end",
          "cron",
          "webhook",
        ])
        .describe(
          "Discord event that triggers the flow, cron for a scheduled flow, or webhook for a flow that runs when a request is sent to its webhook URL. guild_audit_log_entry_create runs for every audit log entry, so pick the actions with an Event Filter on audit_log_action. Its placeholders are {{audit_log.action}}, {{audit_log.reason}}, {{audit_log.target}} and more, and {{user}} is whoever did it. message_poll_vote_add and message_poll_vote_remove run for each answer a member votes for or takes back, with {{poll.question}}, {{poll.answer}} and {{poll.answer_id}}. message_poll_end runs when a poll ends, with {{poll.question}}, {{poll.winner}}, {{poll.winner_votes}} and {{poll.total_votes}}, and {{user}} is the author of the poll. {{message.id}} is the poll for all three."
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
