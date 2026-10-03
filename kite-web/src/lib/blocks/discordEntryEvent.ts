import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordEntryEvent: BlockDefinition = {
  type: "entry_event",
  title: "Listen for Event",
  description:
    "Listens for an event to trigger the flow. Drop different actions here!",
  icon: "satellite-dish",
  contexts: ["event_discord", "event_schedule"],
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
          "guild_create",
          "guild_delete",
          "voice_state_update",
          "cron",
        ])
        .describe(
          "Discord event that triggers the flow, or cron for a scheduled flow."
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
