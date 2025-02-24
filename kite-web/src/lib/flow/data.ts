import { Edge, Node, NodeProps as XYNodeProps } from "@xyflow/react";
import z from "zod";
import { FlowNodeData } from "../types/flow.gen";

const numericRegex = /^[0-9]+$/;
const placeholderRegex = /^\{\{[a-z0-9_.]+\}\}$/;

export interface FlowData {
  nodes: Node<NodeData>[];
  edges: Edge[];
}

export type NodeData = FlowNodeData & Record<string, unknown>;

export type NodeProps = XYNodeProps<Node<NodeData>>;

export type NodeType = Node<NodeData>;

export const auditLogReasonSchema = z.string().max(512).optional();

export const nodeBaseDataSchema = z.object({
  custom_label: z.string().optional(),
});

export const nodeEntryCommandDataSchema = nodeBaseDataSchema.extend({
  name: z
    .string()
    .max(32)
    .min(1)
    .regex(
      /^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$/,
      "Must be only lowercase alphanumeric characters and underscores, and have at most 3 words"
    ),
  description: z.string().max(100).min(1),
});

export const nodeOptionCommandArgumentDataSchema = nodeBaseDataSchema.extend({
  name: z
    .string()
    .max(32)
    .min(1)
    .regex(/^[a-z0-9_]+$/),
  description: z.string().max(100).min(1),
  command_argument_type: z
    .literal("string")
    .or(z.literal("integer"))
    .or(z.literal("boolean"))
    .or(z.literal("user"))
    .or(z.literal("channel"))
    .or(z.literal("role"))
    .or(z.literal("mentionable"))
    .or(z.literal("number"))
    .or(z.literal("attachment")),
  command_argument_required: z.boolean().optional(),
});

export const nodeOptionCommandPermissionsSchema = nodeBaseDataSchema.extend({
  command_permissions: z.string().regex(numericRegex),
});

export const nodeOptionCommandContextsSchema = nodeBaseDataSchema.extend({
  command_disabled_contexts: z
    .array(
      z
        .literal("guild")
        .or(z.literal("bot_dm"))
        .or(z.literal("private_channel"))
    )
    .optional(),
  command_disabled_integrations: z
    .array(z.literal("guild_install").or(z.literal("user_install")))
    .optional(),
});

export const nodeOptionEventFilterSchema = nodeBaseDataSchema.extend({
  event_filter_target: z.literal("message_content"),
  event_filter_expression: z.string().max(1000).min(1),
});

export const nodeEntryEventDataSchema = nodeBaseDataSchema.extend({
  event_type: z.string(),
});

export const nodeEntryComponentButtonDataSchema = nodeBaseDataSchema.extend({});

export const nodeActionResponseCreateDataSchema = nodeBaseDataSchema
  .extend({
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
    message_ephemeral: z.boolean().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionResponseEditDataSchema = nodeBaseDataSchema
  .extend({
    message_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(placeholderRegex))
      .or(z.literal("@original")),
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionResponseDeleteDataSchema = nodeBaseDataSchema.extend({
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex))
    .or(z.literal("@original")),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionResponseDeferDataSchema = nodeBaseDataSchema.extend({
  message_ephemeral: z.boolean().optional(),
});

export const nodeSuspendResponseModalDataSchema = nodeBaseDataSchema.extend({
  modal_data: z.object({
    title: z.string().max(45).min(1),
    components: z
      .array(
        z.object({
          components: z
            .array(
              z.object({
                custom_id: z.string().max(100).min(1),
                label: z.string().max(45).min(1),
                style: z.literal(1).or(z.literal(2)),
                required: z.boolean().optional(),
                min_length: z.number().optional(),
                max_length: z.number().optional(),
                value: z.string().max(4000).min(1).optional(),
                placeholder: z.string().max(4000).min(1).optional(),
              })
            )
            .min(1)
            .max(1),
        })
      )
      .min(1)
      .max(5),
  }),
});

export const nodeActionMessageCreateDataSchema = nodeBaseDataSchema
  .extend({
    channel_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(placeholderRegex)),
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionPrivateMessageCreateDataSchema = nodeBaseDataSchema
  .extend({
    user_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(placeholderRegex)),
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionMessageEditDataSchema = nodeBaseDataSchema
  .extend({
    channel_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(placeholderRegex)),
    message_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(placeholderRegex)),
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionMessageDeleteDataSchema = nodeBaseDataSchema.extend({
  channel_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberBanDataSchema = nodeBaseDataSchema.extend({
  user_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  member_ban_delete_message_duration_seconds: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex))
    .optional(),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberUnbanDataSchema = nodeBaseDataSchema.extend({
  user_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberKickDataSchema = nodeBaseDataSchema.extend({
  user_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberTimeoutDataSchema = nodeBaseDataSchema.extend({
  user_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  member_timeout_duration_seconds: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberEditDataSchema = nodeBaseDataSchema.extend({
  user_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  member_data: z.object({
    nick: z.string(),
  }),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberRoleAddDataSchema = nodeBaseDataSchema.extend({
  user_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  role_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberRoleRemoveDataSchema = nodeBaseDataSchema.extend({
  user_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  role_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionVariableSetSchema = nodeBaseDataSchema.extend({
  variable_id: z.string(),
  variable_scope: z.string().optional(),
  variable_value: z.string(),
  variable_operation: z
    .literal("overwrite")
    .or(z.literal("append"))
    .or(z.literal("prepend"))
    .or(z.literal("increment"))
    .or(z.literal("decrement")),
});

export const nodeActionVariableDeleteSchema = nodeBaseDataSchema.extend({
  variable_id: z.string(),
  variable_scope: z.string().optional(),
});

export const nodeActionVariableGetSchema = nodeBaseDataSchema.extend({
  variable_id: z.string(),
  variable_scope: z.string().optional(),
});

export const nodeActionHttpRequestDataSchema = nodeBaseDataSchema.extend({
  http_request_data: z.object({
    url: z.string().url(),
    method: z
      .literal("GET")
      .or(z.literal("POST"))
      .or(z.literal("PUT"))
      .or(z.literal("PATCH"))
      .or(z.literal("DELETE")),
  }),
});

export const nodeActionAiChatCompletionDataSchema = nodeBaseDataSchema.extend({
  ai_chat_completion_data: z.object({
    system_prompt: z.string().max(2000).optional(),
    prompt: z.string().max(2000).min(1),
    max_completion_tokens: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(placeholderRegex))
      .optional(),
  }),
});

export const nodeActionExpressionEvaluateDataSchema = nodeBaseDataSchema.extend(
  {
    expression: z.string().max(2000),
  }
);

export const nodeActionRandomGenerateDataSchema = nodeBaseDataSchema.extend({
  random_min: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
  random_max: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
});

export const nodeActionLogDataSchema = nodeBaseDataSchema.extend({
  log_level: z
    .literal("debug")
    .or(z.literal("info"))
    .or(z.literal("warn"))
    .or(z.literal("error")),
  log_message: z.string().max(2000).min(1),
});

export const nodeConditionCompareDataSchema = nodeBaseDataSchema.extend({
  condition_base_value: z.string(),
  condition_allow_multiple: z.boolean().optional(),
});

export const nodeConditionItemCompareDataSchema = nodeBaseDataSchema.extend({
  condition_item_value: z.string().optional(),
  condition_item_mode: z
    .literal("equal")
    .or(z.literal("not_equal"))
    .or(z.literal("greater_than"))
    .or(z.literal("less_than"))
    .or(z.literal("greater_than_or_equal"))
    .or(z.literal("less_than_or_equal"))
    .or(z.literal("contains"))
    .or(z.literal("has_role"))
    .or(z.literal("not_has_role"))
    .or(z.literal("has_permission"))
    .or(z.literal("not_has_permission")),
});

export const nodeControlLoopDataSchema = nodeBaseDataSchema.extend({
  loop_count: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
});

export const nodeControlSleepDataSchema = nodeBaseDataSchema.extend({
  sleep_duration_seconds: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(placeholderRegex)),
});
