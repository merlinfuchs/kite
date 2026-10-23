import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordOptionEventFilter: BlockDefinition = {
  type: "option_event_filter",
  id_prefix: "filter",
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
        .enum([
          "message_content",
          "user_id",
          "guild_id",
          "channel_id",
          "message_id",
          "audit_log_action",
        ])
        .describe(
          "Property of the event to filter on. audit_log_action is the action of an audit log entry, like member_ban_add, member_kick, member_timeout_add, member_role_update or channel_create."
        ),
    },
    {
      name: "event_filter_mode",
      schema: z
        .enum([
          "equal",
          "not_equal",
          "contains",
          "starts_with",
          "ends_with",
          "one_of",
          "not_one_of",
        ])
        .describe(
          "How the property is compared to the filter value. one_of and not_one_of compare with each of several values separated by commas."
        ),
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
