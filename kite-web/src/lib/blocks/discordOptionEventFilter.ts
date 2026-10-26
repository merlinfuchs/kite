import { z } from "zod";
import { BlockDefinition, SettingsSchema } from "./types";
import {
  booleanEventFilterTargets,
  getEventFilterModes,
} from "../flow/eventFilterTargets";

// Yes or no filters have no value, every other filter needs one, and some
// targets only take some modes, see getEventFilterModes.
function refineEventFilter(schema: SettingsSchema) {
  return schema
    .refine(
      (data) =>
        booleanEventFilterTargets.includes(data.event_filter_target) ||
        !!data.event_filter_value,
      {
        message: "Required",
        path: ["event_filter_value"],
      }
    )
    .superRefine((data, ctx) => {
      const modes = getEventFilterModes(data.event_filter_target);
      if (modes && !modes.includes(data.event_filter_mode)) {
        ctx.addIssue({
          code: "custom",
          message: `Has to be ${modes.join(" or ")} for this filter target`,
          path: ["event_filter_mode"],
        });
      }
    })
    .describe(
      "message_has_attachment, message_mentions_bot and user_is_bot have no value, equal means yes and not_equal no, so not_equal on user_is_bot ignores bots. They, role_id and emoji only take equal or not_equal, audit_log_action only one_of or not_one_of, as the editor picks actions from a list. The targets available depend on the event type of the listener."
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
        .optional()
        .describe(
          "Value to compare against. This is fixed text, the only placeholders it can use are the flow's parameters, like {{params.staff_role}}. Yes or no targets have none."
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
