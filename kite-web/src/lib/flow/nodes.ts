import { Edge, Node, XYPosition } from "@xyflow/react";
import { humanId } from "human-id";
import { useMemo } from "react";
import { ZodSchema } from "zod";
import { getUniqueId } from "../utils";
import {
  nodeActionAiChatCompletionDataSchema,
  nodeActionAiWebSearchCompletionDataSchema,
  nodeActionChannelCreateDataSchema,
  nodeActionChannelDeleteDataSchema,
  nodeActionChannelEditDataSchema,
  nodeActionChannelGetDataSchema,
  nodeActionExpressionEvaluateDataSchema,
  nodeActionForumPostCreateDataSchema,
  nodeActionGuildGetDataSchema,
  nodeActionHttpRequestDataSchema,
  nodeActionLogDataSchema,
  nodeActionMemberBanDataSchema,
  nodeActionMemberEditDataSchema,
  nodeActionMemberGetDataSchema,
  nodeActionMemberKickDataSchema,
  nodeActionMemberRoleAddDataSchema,
  nodeActionMemberRoleRemoveDataSchema,
  nodeActionMemberTimeoutDataSchema,
  nodeActionMemberUnbanDataSchema,
  nodeActionMessageCreateDataSchema,
  nodeActionMessageDeleteDataSchema,
  nodeActionMessageEditDataSchema,
  nodeActionMessageGetDataSchema,
  nodeActionMessageReactionCreateDataSchema,
  nodeActionMessageReactionDeleteDataSchema,
  nodeActionPrivateMessageCreateDataSchema,
  nodeActionRandomGenerateDataSchema,
  nodeActionResponseCreateDataSchema,
  nodeActionResponseDeferDataSchema,
  nodeActionResponseDeleteDataSchema,
  nodeActionResponseEditDataSchema,
  nodeActionRobloxUserGetDataSchema,
  nodeActionRoleGetDataSchema,
  nodeActionThreadCreateDataSchema,
  nodeActionThreadMemberAddDataSchema,
  nodeActionThreadMemberRemoveDataSchema,
  nodeActionUserGetDataSchema,
  nodeActionVariableDeleteSchema,
  nodeActionVariableGetSchema,
  nodeActionVariableSetSchema,
  nodeConditionCompareDataSchema,
  nodeConditionItemCompareDataSchema,
  nodeControlLoopDataSchema,
  nodeControlSleepDataSchema,
  NodeData,
  nodeEntryCommandDataSchema,
  nodeEntryComponentButtonDataSchema,
  nodeEntryEventDataSchema,
  nodeOptionCommandArgumentDataSchema,
  nodeOptionCommandContextsSchema,
  nodeOptionCommandPermissionsSchema,
  nodeOptionEventFilterSchema,
  nodeSuspendResponseModalDataSchema,
} from "./dataSchema";
import {
  nodeActionChannelGetResultSchema,
  nodeActionGuildGetResultSchema,
  nodeActionMemberGetResultSchema,
  nodeActionMessageCreateResultSchema,
  nodeActionMessageEditResultSchema,
  nodeActionMessageGetResultSchema,
  nodeActionPrivateMessageCreateResultSchema,
  nodeActionResponseCreateResultSchema,
  nodeActionResponseEditResultSchema,
  nodeActionRobloxUserGetResultSchema,
  nodeActionRoleGetResultSchema,
} from "./resultSchema";

export const primaryColor = "#3B82F6";

export const actionColor = "#3b82f6";
export const entryColor = "#eab308";
export const errorColor = "#ef4444";
export const controlColor = "#22c55e";
export const optionColor = "#8b5cf6";
export const suspendColor = "#d946ef";

export interface NodeValues {
  color: string;
  icon: string;
  defaultTitle: string;
  defaultDescription: string;
  dataSchema?: ZodSchema;
  dataFields: string[];
  resultSchema?: ZodSchema;
  ownsChildren?: boolean;
  fixed?: boolean;
  creditsCost?: number | ((data: NodeData) => number);
}

export const nodeTypes: Record<string, NodeValues> = {
  entry_command: {
    color: entryColor,
    icon: "square-slash",
    defaultTitle: "Command",
    defaultDescription:
      "Command entry. Drop different actions and options here!",
    dataSchema: nodeEntryCommandDataSchema,
    dataFields: ["name", "description"],
    fixed: true,
  },
  entry_event: {
    color: entryColor,
    icon: "satellite-dish",
    defaultTitle: "Listen for Event",
    defaultDescription:
      "Listens for an event to trigger the flow. Drop different actions here!",
    dataSchema: nodeEntryEventDataSchema,
    dataFields: ["event_type", "description"],
    fixed: true,
  },
  entry_component_button: {
    color: entryColor,
    icon: "mouse-pointer-click",
    defaultTitle: "Button",
    defaultDescription:
      "This gets triggered when a user clicks the button. Drop different actions here!",
    dataSchema: nodeEntryComponentButtonDataSchema,
    dataFields: [],
    fixed: true,
  },
  action_response_create: {
    color: actionColor,
    icon: "message-circle-reply",
    defaultTitle: "Create response message",
    defaultDescription: "Bot replies to the interaction with a message",
    dataSchema: nodeActionResponseCreateDataSchema,
    resultSchema: nodeActionResponseCreateResultSchema,
    dataFields: [
      "message_template_id",
      "message_data",
      "message_ephemeral",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_response_edit: {
    color: actionColor,
    icon: "pen",
    defaultTitle: "Edit response message",
    defaultDescription: "Bot edits an existing interaction response message",
    dataSchema: nodeActionResponseEditDataSchema,
    resultSchema: nodeActionResponseEditResultSchema,
    dataFields: [
      "response_target",
      "message_template_id",
      "message_data",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_response_delete: {
    color: actionColor,
    icon: "message-circle-x",
    defaultTitle: "Delete response message",
    defaultDescription: "Bot deletes an existing interaction response message",
    dataSchema: nodeActionResponseDeleteDataSchema,
    dataFields: ["response_target", "custom_label"],
    creditsCost: 1,
  },
  action_response_defer: {
    color: actionColor,
    icon: "message-circle-question",
    defaultTitle: "Defer response",
    defaultDescription:
      "Bot defers the response to the interaction to give time for further processing",
    dataSchema: nodeActionResponseDeferDataSchema,
    dataFields: ["message_ephemeral", "custom_label"],
    creditsCost: 1,
  },
  action_message_create: {
    color: actionColor,
    icon: "message-circle-plus",
    defaultTitle: "Create channel message",
    defaultDescription: "Bot sends a message to a channel",
    dataSchema: nodeActionMessageCreateDataSchema,
    resultSchema: nodeActionMessageCreateResultSchema,
    dataFields: [
      "channel_target",
      "message_template_id",
      "message_data",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_message_edit: {
    color: actionColor,
    icon: "pen",
    defaultTitle: "Edit channel message",
    defaultDescription: "Bot edits an existing message in a channel",
    dataSchema: nodeActionMessageEditDataSchema,
    resultSchema: nodeActionMessageEditResultSchema,
    dataFields: [
      "channel_target",
      "message_target",
      "message_template_id",
      "message_data",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_private_message_create: {
    color: actionColor,
    icon: "message-circle-plus",
    defaultTitle: "Send direct message",
    defaultDescription:
      "Bot sends a private message to a user if the user allows it",
    dataSchema: nodeActionPrivateMessageCreateDataSchema,
    resultSchema: nodeActionPrivateMessageCreateResultSchema,
    dataFields: [
      "user_target",
      "message_data",
      "message_template_id",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_message_delete: {
    color: actionColor,
    icon: "message-circle-x",
    defaultTitle: "Delete channel message",
    defaultDescription: "Bot deletes an existing message in a channel",
    dataSchema: nodeActionMessageDeleteDataSchema,
    dataFields: [
      "channel_target",
      "message_target",
      "audit_log_reason",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_message_reaction_create: {
    color: actionColor,
    icon: "smile-plus",
    defaultTitle: "Create message reaction",
    defaultDescription: "Bot adds a reaction to a message",
    dataSchema: nodeActionMessageReactionCreateDataSchema,
    dataFields: [
      "channel_target",
      "message_target",
      "emoji_data",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_message_reaction_delete: {
    color: actionColor,
    icon: "frown",
    defaultTitle: "Delete message reaction",
    defaultDescription: "Bot deletes a reaction from a message",
    dataSchema: nodeActionMessageReactionDeleteDataSchema,
    dataFields: [
      "channel_target",
      "message_target",
      "emoji_data",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_member_ban: {
    color: actionColor,
    icon: "user-round-x",
    defaultTitle: "Ban member",
    defaultDescription: "Ban a member from the server",
    dataSchema: nodeActionMemberBanDataSchema,
    dataFields: [
      "user_target",
      "member_ban_delete_message_duration_seconds",
      "audit_log_reason",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_member_unban: {
    color: actionColor,
    icon: "user-round-check",
    defaultTitle: "Unban member",
    defaultDescription: "Unban a member from the server",
    dataSchema: nodeActionMemberUnbanDataSchema,
    dataFields: ["user_target", "audit_log_reason", "custom_label"],
    creditsCost: 1,
  },
  action_member_kick: {
    color: actionColor,
    icon: "user-round-minus",
    defaultTitle: "Kick member",
    defaultDescription: "Kick a member from the server",
    dataSchema: nodeActionMemberKickDataSchema,
    dataFields: ["user_target", "audit_log_reason", "custom_label"],
    creditsCost: 1,
  },
  action_member_timeout: {
    color: actionColor,
    icon: "message-circle-off",
    defaultTitle: "Timeout member",
    defaultDescription: "Timeout a member in the server",
    dataSchema: nodeActionMemberTimeoutDataSchema,
    dataFields: [
      "user_target",
      "member_timeout_duration_seconds",
      "audit_log_reason",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_member_edit: {
    color: actionColor,
    icon: "user-round-pen",
    defaultTitle: "Edit member nickname",
    defaultDescription: "Edit a member in the server",
    dataSchema: nodeActionMemberEditDataSchema,
    dataFields: [
      "user_target",
      "member_nick",
      "audit_log_reason",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_member_role_add: {
    color: actionColor,
    icon: "bookmark-plus",
    defaultTitle: "Add role to member",
    defaultDescription: "Add a role to a member",
    dataSchema: nodeActionMemberRoleAddDataSchema,
    dataFields: [
      "user_target",
      "role_target",
      "audit_log_reason",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_member_role_remove: {
    color: actionColor,
    icon: "bookmark-minus",
    defaultTitle: "Remove role from member",
    defaultDescription: "Remove a role from a member",
    dataSchema: nodeActionMemberRoleRemoveDataSchema,
    dataFields: [
      "user_target",
      "role_target",
      "audit_log_reason",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_member_get: {
    color: actionColor,
    icon: "user-round-search",
    defaultTitle: "Get member",
    defaultDescription: "Get a member by ID",
    dataSchema: nodeActionMemberGetDataSchema,
    resultSchema: nodeActionMemberGetResultSchema,
    dataFields: [
      "guild_target",
      "user_target",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_user_get: {
    color: actionColor,
    icon: "user-round-search",
    defaultTitle: "Get user",
    defaultDescription: "Get a user by ID",
    dataSchema: nodeActionUserGetDataSchema,
    dataFields: ["user_target", "temporary_name", "custom_label"],
    creditsCost: 1,
  },
  action_channel_get: {
    color: actionColor,
    icon: "folder-search",
    defaultTitle: "Get channel",
    defaultDescription: "Get a channel by ID",
    dataSchema: nodeActionChannelGetDataSchema,
    resultSchema: nodeActionChannelGetResultSchema,
    dataFields: ["channel_target", "temporary_name", "custom_label"],
    creditsCost: 1,
  },
  action_channel_create: {
    color: actionColor,
    icon: "folder-plus",
    defaultTitle: "Create channel",
    defaultDescription: "Create a channel",
    dataSchema: nodeActionChannelCreateDataSchema,
    dataFields: [
      "guild_target",
      "channel_data",
      "audit_log_reason",
      "temporary_name",
      "custom_label",
    ],
  },
  action_channel_edit: {
    color: actionColor,
    icon: "folder-pen",
    defaultTitle: "Edit channel",
    defaultDescription: "Edit a channel or thread",
    dataSchema: nodeActionChannelEditDataSchema,
    dataFields: [
      "channel_target",
      "channel_data",
      "audit_log_reason",
      "temporary_name",
      "custom_label",
    ],
  },
  action_channel_delete: {
    color: actionColor,
    icon: "folder-x",
    defaultTitle: "Delete channel",
    defaultDescription: "Delete a channel or thread ",
    dataSchema: nodeActionChannelDeleteDataSchema,
    dataFields: ["channel_target", "audit_log_reason", "custom_label"],
  },
  action_thread_create: {
    color: actionColor,
    icon: "message-circle-plus",
    defaultTitle: "Create thread",
    defaultDescription: "Create a thread",
    dataSchema: nodeActionThreadCreateDataSchema,
    dataFields: [
      "thread_data",
      "audit_log_reason",
      "temporary_name",
      "custom_label",
    ],
  },
  action_forum_post_create: {
    color: actionColor,
    icon: "message-circle-plus",
    defaultTitle: "Create forum post",
    defaultDescription: "Create a forum post",
    dataSchema: nodeActionForumPostCreateDataSchema,
    dataFields: [
      "channel_target",
      "channel_data",
      "audit_log_reason",
      "temporary_name",
      "custom_label",
    ],
  },

  action_thread_member_add: {
    color: actionColor,
    icon: "user-plus",
    defaultTitle: "Add member to thread",
    defaultDescription: "Add a member to a thread",
    dataSchema: nodeActionThreadMemberAddDataSchema,
    dataFields: [
      "channel_target",
      "user_target",
      "audit_log_reason",
      "custom_label",
    ],
  },
  action_thread_member_remove: {
    color: actionColor,
    icon: "user-minus",
    defaultTitle: "Remove member from thread",
    defaultDescription: "Remove a member from a thread",
    dataSchema: nodeActionThreadMemberRemoveDataSchema,
    dataFields: [
      "channel_target",
      "user_target",
      "audit_log_reason",
      "custom_label",
    ],
  },
  action_role_get: {
    color: actionColor,
    icon: "bookmark",
    defaultTitle: "Get role",
    defaultDescription: "Get a role by ID",
    dataSchema: nodeActionRoleGetDataSchema,
    resultSchema: nodeActionRoleGetResultSchema,
    dataFields: [
      "guild_target",
      "role_target",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_guild_get: {
    color: actionColor,
    icon: "server",
    defaultTitle: "Get server",
    defaultDescription: "Get a server / guild by ID",
    dataSchema: nodeActionGuildGetDataSchema,
    resultSchema: nodeActionGuildGetResultSchema,
    dataFields: ["guild_target", "temporary_name", "custom_label"],
    creditsCost: 1,
  },
  action_message_get: {
    color: actionColor,
    icon: "mail-search",
    defaultTitle: "Get channel message",
    defaultDescription: "Get a message from a channel",
    dataSchema: nodeActionMessageGetDataSchema,
    resultSchema: nodeActionMessageGetResultSchema,
    dataFields: ["message_target", "temporary_name", "custom_label"],
    creditsCost: 1,
  },
  action_roblox_user_get: {
    color: actionColor,
    icon: "gamepad",
    defaultTitle: "Get Roblox User",
    defaultDescription: "Get a Roblox user by ID or username",
    dataSchema: nodeActionRobloxUserGetDataSchema,
    resultSchema: nodeActionRobloxUserGetResultSchema,
    dataFields: [
      "roblox_user_target",
      "roblox_lookup_mode",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_variable_set: {
    color: actionColor,
    icon: "variable",
    defaultTitle: "Set stored variable",
    defaultDescription: "Set the value of a stored variable",
    dataSchema: nodeActionVariableSetSchema,
    dataFields: [
      "variable_id",
      "variable_scope",
      "variable_operation",
      "variable_value",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_variable_delete: {
    color: actionColor,
    icon: "variable",
    defaultTitle: "Delete stored variable",
    defaultDescription: "Delete the value of a stored variable",
    dataSchema: nodeActionVariableDeleteSchema,
    dataFields: ["variable_id", "variable_scope", "custom_label"],
    creditsCost: 1,
  },
  action_variable_get: {
    color: actionColor,
    icon: "variable",
    defaultTitle: "Get stored variable",
    defaultDescription: "Get the value of a stored variable",
    dataSchema: nodeActionVariableGetSchema,
    dataFields: [
      "variable_id",
      "variable_scope",
      "temporary_name",
      "custom_label",
    ],
    creditsCost: 1,
  },
  action_http_request: {
    color: actionColor,
    icon: "webhook",
    defaultTitle: "Send API Request",
    defaultDescription: "Send an API request to an external server",
    dataSchema: nodeActionHttpRequestDataSchema,
    dataFields: ["http_request_data", "temporary_name", "custom_label"],
    creditsCost: 3,
  },
  action_ai_chat_completion: {
    color: actionColor,
    icon: "brain-circuit",
    defaultTitle: "Ask AI",
    defaultDescription:
      "Ask artificial intelligence a question or let it respond to a prompt",
    dataSchema: nodeActionAiChatCompletionDataSchema,
    dataFields: ["ai_chat_completion_data", "temporary_name", "custom_label"],
    creditsCost: (data) => {
      const model = data.ai_chat_completion_data?.model;
      switch (model) {
        case "gpt-4.1":
          return 100;
        case "gpt-4.1-mini":
          return 20;
        default:
          return 5;
      }
    },
  },
  action_ai_web_search: {
    color: actionColor,
    icon: "search",
    defaultTitle: "Search the Web",
    defaultDescription: "Search the web for the latest information using AI",
    dataSchema: nodeActionAiWebSearchCompletionDataSchema,
    dataFields: ["ai_web_search_data", "temporary_name", "custom_label"],
    creditsCost: (data) => {
      const model = data.ai_chat_completion_data?.model;
      switch (model) {
        case "gpt-4.1":
          return 500;
        case "gpt-4.1-mini":
          return 100;
        default:
          return 25;
      }
    },
  },
  action_expression_evaluate: {
    color: actionColor,
    icon: "calculator",
    defaultTitle: "Calculate Value",
    defaultDescription:
      "Evaluate math or other logical expressions and use the result later",
    dataSchema: nodeActionExpressionEvaluateDataSchema,
    dataFields: ["expression", "temporary_name", "custom_label"],
    creditsCost: 1,
  },
  action_random_generate: {
    color: actionColor,
    icon: "dices",
    defaultTitle: "Generate Random Number",
    defaultDescription: "Generate a random number in a range",
    dataSchema: nodeActionRandomGenerateDataSchema,
    dataFields: ["random_min", "random_max", "temporary_name", "custom_label"],
    creditsCost: 1,
  },
  action_log: {
    color: actionColor,
    icon: "scroll-text",
    defaultTitle: "Log Message",
    defaultDescription:
      "Log some text which is only visible in the application logs",
    dataSchema: nodeActionLogDataSchema,
    dataFields: ["log_level", "log_message", "custom_label"],
    creditsCost: 1,
  },
  control_condition_compare: {
    color: controlColor,
    icon: "arrow-left-right",
    defaultTitle: "Comparison Condition",
    defaultDescription:
      "Run actions based on the difference between two values.",
    dataSchema: nodeConditionCompareDataSchema,
    dataFields: [
      "condition_compare_base_value",
      "condition_allow_multiple",
      "custom_label",
    ],
    ownsChildren: true,
  },
  control_condition_item_compare: {
    color: controlColor,
    icon: "circle-help",
    defaultTitle: "Match Condition",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the two values are equal.",
    dataFields: ["condition_item_compare_mode", "condition_item_compare_value"],
  },
  control_condition_user: {
    color: controlColor,
    icon: "user-search",
    defaultTitle: "User Condition",
    defaultDescription: "Run actions based on a user.",
    dataSchema: nodeConditionCompareDataSchema,
    dataFields: [
      "condition_user_base_value",
      "condition_allow_multiple",
      "custom_label",
    ],
    ownsChildren: true,
  },
  control_condition_item_user: {
    color: controlColor,
    icon: "circle-help",
    defaultTitle: "Match User",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the user meets the criteria.",
    dataFields: ["condition_item_user_mode", "condition_item_user_value"],
  },
  control_condition_channel: {
    color: controlColor,
    icon: "folder-search",
    defaultTitle: "Channel Condition",
    defaultDescription: "Run actions based on a channel.",
    dataSchema: nodeConditionCompareDataSchema,
    dataFields: [
      "condition_channel_base_value",
      "condition_allow_multiple",
      "custom_label",
    ],
    ownsChildren: true,
  },
  control_condition_item_channel: {
    color: controlColor,
    icon: "circle-help",
    defaultTitle: "Match Channel",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the channel meets the criteria.",
    dataFields: ["condition_item_channel_mode", "condition_item_channel_value"],
  },
  control_condition_role: {
    color: controlColor,
    icon: "bookmark",
    defaultTitle: "Role Condition",
    defaultDescription: "Run actions based on a role.",
    dataSchema: nodeConditionCompareDataSchema,
    dataFields: [
      "condition_role_base_value",
      "condition_allow_multiple",
      "custom_label",
    ],
    ownsChildren: true,
  },
  control_condition_item_role: {
    color: controlColor,
    icon: "circle-help",
    defaultTitle: "Match Role",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the role meets the criteria.",
    dataFields: ["condition_item_role_mode", "condition_item_role_value"],
  },
  control_condition_item_else: {
    color: errorColor,
    icon: "circle-x",
    defaultTitle: "Else",
    defaultDescription: "Run actions if no other conditions are met.",
    dataFields: [],
    fixed: true,
  },
  control_error_handler: {
    color: errorColor,
    icon: "circle-alert",
    defaultTitle: "Handle Errors",
    defaultDescription:
      "Handle errors that occur in the flow after this block.",
    dataFields: ["temporary_name", "custom_label"],
    ownsChildren: true,
  },
  control_loop: {
    color: controlColor,
    icon: "repeat-2",
    defaultTitle: "Run a loop",
    dataSchema: nodeControlLoopDataSchema,
    defaultDescription: "Run a set of actions multiple times.",
    dataFields: ["loop_count", "custom_label"],
    ownsChildren: true,
  },
  control_loop_each: {
    color: controlColor,
    icon: "repeat-2",
    defaultTitle: "Each loop iteration",
    defaultDescription: "Run actions for each iteration of the loop.",
    dataFields: [],
    fixed: true,
  },
  control_loop_end: {
    color: controlColor,
    icon: "corner-down-right",
    defaultTitle: "After loop",
    defaultDescription: "Run actions after the loop has finished.",
    dataFields: [],
    fixed: true,
  },
  control_loop_exit: {
    color: controlColor,
    icon: "log-out",
    defaultTitle: "Exit loop",
    defaultDescription: "Exit out of the loop.",
    dataFields: [],
  },
  control_sleep: {
    color: controlColor,
    icon: "timer",
    defaultTitle: "Wait",
    defaultDescription: "Pause the flow for a set amount of time.",
    dataSchema: nodeControlSleepDataSchema,
    dataFields: ["sleep_duration_seconds"],
  },
  option_command_argument: {
    color: optionColor,
    icon: "text-cursor-input",
    defaultTitle: "Command Argument",
    defaultDescription: "Argument for a command.",
    dataSchema: nodeOptionCommandArgumentDataSchema,
    dataFields: [
      "name",
      "description",
      "command_argument_type",
      "command_argument_required",
      "command_argument_min_value",
      "command_argument_max_value",
      "command_argument_max_length",
      "command_argument_choices",
    ],
  },
  option_command_permissions: {
    color: optionColor,
    icon: "shield-check",
    defaultTitle: "Command Permissions",
    defaultDescription:
      "Make the command only available to users with the specified permissions.",
    dataSchema: nodeOptionCommandPermissionsSchema,
    dataFields: ["command_permissions"],
  },
  option_command_contexts: {
    color: optionColor,
    icon: "map-pin",
    defaultTitle: "Command Contexts",
    defaultDescription:
      "Define where your command should be available. By default, it will be available everywhere.",
    dataSchema: nodeOptionCommandContextsSchema,
    dataFields: ["command_contexts", "command_integrations"],
  },
  option_event_filter: {
    color: optionColor,
    icon: "filter",
    defaultTitle: "Event Filter",
    defaultDescription: "Filter events based on their properties.",
    dataSchema: nodeOptionEventFilterSchema,
    dataFields: ["event_filter_target", "event_filter_expression"],
  },
  suspend_response_modal: {
    color: suspendColor,
    icon: "picture-in-picture-2",
    defaultTitle: "Show Modal",
    defaultDescription:
      "Show a modal to the user and suspend the flow until the user submits the modal.",
    dataSchema: nodeSuspendResponseModalDataSchema,
    dataFields: ["modal_data", "custom_label"],
  },
};

const unknownNodeType: NodeValues = {
  color: "#ff0000",
  icon: "circle-help",
  defaultTitle: "Unknown",
  defaultDescription: "Unknown node type.",
  dataFields: [],
};

export function getNodeValues(nodeType: string): NodeValues {
  const values = nodeTypes[nodeType];
  if (!values) {
    return unknownNodeType;
  }
  return values;
}

export function useNodeValues(nodeType: string): NodeValues {
  return useMemo(() => getNodeValues(nodeType), [nodeType]);
}

const conditionChildType: Record<string, string> = {
  control_condition_compare: "control_condition_item_compare",
  control_condition_user: "control_condition_item_user",
  control_condition_channel: "control_condition_item_channel",
  control_condition_role: "control_condition_item_role",
};

export function createNode(
  type: string,
  position: XYPosition,
  props?: Partial<Node<NodeData>>
): [Node<NodeData>[], Edge[]] {
  const id = getNodeId();

  const nodes: Node<NodeData>[] = [
    {
      id,
      type,
      position,
      data: {},
      ...props,
    },
  ];
  const edges: Edge[] = [];

  // TODO?: connect option types to entry automatically?

  if (conditionChildType.hasOwnProperty(type)) {
    const [elseNodes, elseEdges] = createNode("control_condition_item_else", {
      x: position.x + 200,
      y: position.y + 200,
    });

    nodes.push(...elseNodes);
    edges.push({
      id: getEdgeId(),
      source: id,
      target: elseNodes[0].id,
      type: "fixed",
    });
    edges.push(...elseEdges);

    const [compareNodes, compareEdges] = createNode(conditionChildType[type], {
      x: position.x - 150,
      y: position.y + 200,
    });

    nodes.push(...compareNodes);
    edges.push({
      id: getEdgeId(),
      source: id,
      target: compareNodes[0].id,
      type: "fixed",
    });
    edges.push(...compareEdges);
  } else if (type === "control_loop") {
    const [endNodes, endEdges] = createNode("control_loop_end", {
      x: position.x + 200,
      y: position.y + 200,
    });

    nodes.push(...endNodes);
    edges.push({
      id: getEdgeId(),
      source: id,
      target: endNodes[0].id,
      type: "fixed",
    });
    edges.push(...endEdges);

    const [eachNodes, eachEdges] = createNode("control_loop_each", {
      x: position.x - 150,
      y: position.y + 200,
    });

    nodes.push(...eachNodes);
    edges.push({
      id: getEdgeId(),
      source: id,
      target: eachNodes[0].id,
      type: "fixed",
    });
    edges.push(...eachEdges);
  }

  return [nodes, edges];
}

export function getNodeId(): string {
  // This gives us a pool size of 75000
  // There is a small chance of collision, but reactflow handles it gracefully
  return humanId({
    separator: "",
    capitalize: false,
    addAdverb: false,
    adjectiveCount: 0,
  });
}

export function getEdgeId(): string {
  return getUniqueId().toString();
}
