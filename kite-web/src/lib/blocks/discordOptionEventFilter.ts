import { AnyZodObject, z } from "zod";
import { BlockDefinition } from "./types";
import {
  booleanEventFilterTargets,
  equalityEventFilterTargets,
} from "../flow/eventFilterTargets";

// Yes or no filters compare against "true" or "false", roles, emojis and yes
// or no filters can only be equal or not equal, and audit log actions only one
// of or not one of.
function refineEventFilter(schema: AnyZodObject) {
  return schema
    .refine(
      (data) =>
        !booleanEventFilterTargets.includes(data.event_filter_target) ||
        data.event_filter_value === "true" ||
        data.event_filter_value === "false",
      {
        message: "Has to be true or false for this filter target",
        path: ["event_filter_value"],
      }
    )
    .refine(
      (data) =>
        !equalityEventFilterTargets.includes(data.event_filter_target) ||
        data.event_filter_mode === "equal" ||
        data.event_filter_mode === "not_equal",
      {
        message: "Has to be equal or not equal for this filter target",
        path: ["event_filter_mode"],
      }
    )
    .refine(
      (data) =>
        data.event_filter_target !== "audit_log_action" ||
        data.event_filter_mode === "one_of" ||
        data.event_filter_mode === "not_one_of",
      {
        message: "Use Is One Of or Is Not One Of for audit log actions",
        path: ["event_filter_mode"],
      }
    )
    .describe(
      "message_has_attachment, message_is_reply, message_mentions_bot and user_is_bot take the value true or false. role_id, emoji and those can only use equal or not_equal, audit_log_action only one_of or not_one_of, as the editor picks actions from a list. The targets available depend on the event type of the listener."
    );
}

export const discordOptionEventFilter: BlockDefinition = {
  type: "option_event_filter",
  id_prefix: "filter",
  title: "Event Filter",
  description:
    "Filter events based on their properties. Events that don't match don't run the flow.",
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
          "role_id",
          "emoji",
          "message_has_attachment",
          "message_is_reply",
          "message_mentions_bot",
          "user_is_bot",
          "audit_log_action",
        ])
        .describe(
          "Property of the event to filter on. role_id checks the roles of the member, emoji the emoji of a reaction. audit_log_action is the action of an audit log entry, like member_ban_add, member_kick, member_timeout_add, member_role_update or channel_create."
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
          "How the property is compared to the filter value. For role_id, equal means the member has the role. one_of and not_one_of compare with each of several values separated by commas."
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
    {
      name: "event_filter_ignore_case",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether message content is compared ignoring upper and lower case."
        ),
    },
  ],
  refine: refineEventFilter,
  run: { kind: "custom" },
};
