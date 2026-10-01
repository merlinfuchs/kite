import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordOptionEventFilter: BlockDefinition = {
  type: "option_event_filter",
  title: "Event Filter",
  description: "Filter events based on their properties.",
  icon: "filter",
  category: "Events",
  component: "option",
  requires: ["discord"],
  custom_label: false,
  fields: [
    {
      name: "event_filter_target",
      schema: z
        .enum(["message_content", "user_id", "guild_id", "channel_id"])
        .describe("Property of the event to filter on."),
    },
    {
      name: "event_filter_mode",
      schema: z
        .enum(["equal", "not_equal", "contains", "starts_with", "ends_with"])
        .describe("How the property is compared to the filter value."),
    },
    {
      name: "event_filter_value",
      schema: z
        .string()
        .max(1000)
        .min(1)
        .describe(
          "Value to compare against. This is fixed text, placeholders aren't supported."
        ),
    },
  ],
  run: { kind: "custom" },
};
