import { AnyZodObject, z } from "zod";
import { BlockDefinition } from "./types";
import {
  booleanEventFilterTargets,
  equalityEventFilterTargets,
} from "../flow/eventFilterTargets";

// Yes or no filters compare against "true" or "false", and roles, emojis and
// yes or no filters can only be equal or not equal.
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
    .describe(
      "message_has_attachment, message_is_reply, message_mentions_bot and user_is_bot take the value true or false. role_id, emoji and those can only use equal or not_equal. The targets available depend on the event type of the listener."
    );
}

export const discordOptionEventFilter: BlockDefinition = {
  type: "option_event_filter",
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
        ])
        .describe(
          "Property of the event to filter on. role_id checks the roles of the member, emoji the emoji of a reaction."
        ),
    },
    {
      name: "event_filter_mode",
      schema: z
        .enum(["equal", "not_equal", "contains", "starts_with", "ends_with"])
        .describe(
          "How the property is compared to the filter value. For role_id, equal means the member has the role."
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
