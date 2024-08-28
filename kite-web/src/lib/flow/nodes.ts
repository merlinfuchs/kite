import { ExoticComponent, useMemo } from "react";
import {
  nodeActionChannelCreateDataSchema,
  nodeActionChannelDeleteDataSchema,
  nodeActionChannelEditDataSchema,
  nodeActionHttpRequestDataSchema,
  nodeActionLogDataSchema,
  nodeActionMemberBanDataSchema,
  nodeActionMemberKickDataSchema,
  nodeActionMemberTimeoutDataSchema,
  nodeActionMessageCreateDataSchema,
  nodeActionMessageDeleteDataSchema,
  nodeActionMessageEditDataSchema,
  nodeActionResponseCreateDataSchema,
  nodeActionResponseDeleteDataSchema,
  nodeActionResponseEditDataSchema,
  nodeActionRoleCreateDataSchema,
  nodeActionRoleDeleteDataSchema,
  nodeActionRoleEditDataSchema,
  nodeActionThreadCreateDataSchema,
  nodeActionVariableDeleteDataSchema,
  nodeActionVariableSetDataSchema,
  nodeConditionCompareDataSchema,
  nodeConditionItemCompareDataSchema,
  nodeControlLoopDataSchema,
  NodeData,
  nodeEntryCommandDataSchema,
  nodeEntryEventDataSchema,
  nodeOptionCommandArgumentDataSchema,
  nodeOptionCommandContextsSchema,
  nodeOptionCommandPermissionsSchema,
  nodeOptionEventFilterSchema,
} from "./data";
import { ZodSchema } from "zod";
import { Edge, Node, XYPosition } from "@xyflow/react";
import { getUniqueId } from "../utils";
import {
  ArrowLeftRightIcon,
  BookmarkIcon,
  BookmarkMinusIcon,
  BookmarkPlusIcon,
  CircleHelpIcon,
  CornerDownRightIcon,
  FilterIcon,
  FolderMinusIcon,
  FolderPenIcon,
  FolderPlusIcon,
  FolderSearchIcon,
  LogOutIcon,
  MapPinIcon,
  MessageCircleOffIcon,
  MessageCirclePlusIcon,
  MessageCircleReply,
  MessageCircleXIcon,
  PenIcon,
  Repeat2Icon,
  SatelliteDishIcon,
  ScrollTextIcon,
  ShieldCheckIcon,
  SlashSquareIcon,
  TextCursorInputIcon,
  UserRoundMinusIcon,
  UserRoundXIcon,
  UserSearchIcon,
  VariableIcon,
  WebhookIcon,
  XCircleIcon,
} from "lucide-react";

export const primaryColor = "#3B82F6";

export const actionColor = "#3b82f6";
export const entryColor = "#eab308";
export const errorColor = "#ef4444";
export const controlColor = "#22c55e";
export const optionColor = "#a855f7";

export interface NodeValues {
  color: string;
  icon: ExoticComponent<{ className: string }>;
  defaultTitle: string;
  defaultDescription: string;
  dataSchema?: ZodSchema;
  dataFields: string[];
  ownsChildren?: boolean;
  fixed?: boolean;
}

export const nodeTypes: Record<string, NodeValues> = {
  entry_command: {
    color: entryColor,
    icon: SlashSquareIcon,
    defaultTitle: "Command",
    defaultDescription:
      "Command entry. Drop different actions and options here!",
    dataSchema: nodeEntryCommandDataSchema,
    dataFields: ["name", "description"],
    fixed: true,
  },
  entry_event: {
    color: entryColor,
    icon: SatelliteDishIcon,
    defaultTitle: "Listen for Event",
    defaultDescription:
      "Listens for an event to trigger the flow. Drop different actions and options here!",
    dataSchema: nodeEntryEventDataSchema,
    dataFields: ["event_type"],
    fixed: true,
  },
  action_response_create: {
    color: actionColor,
    icon: MessageCircleReply,
    defaultTitle: "Create response message",
    defaultDescription: "Bot replies to the interaction with a message",
    dataSchema: nodeActionResponseCreateDataSchema,
    dataFields: [
      "message_template_id",
      "message_data",
      "message_ephemeral",
      "custom_label",
    ],
  },
  action_response_edit: {
    color: actionColor,
    icon: PenIcon,
    defaultTitle: "Edit response message",
    defaultDescription: "Bot edits an existing interaction response message",
    dataSchema: nodeActionResponseEditDataSchema,
    dataFields: [
      "message_target",
      "message_template_id",
      "message_data",
      "message_ephemeral",
      "custom_label",
    ],
  },
  action_response_delete: {
    color: actionColor,
    icon: MessageCircleXIcon,
    defaultTitle: "Delete response message",
    defaultDescription: "Bot deletes an existing interaction response message",
    dataSchema: nodeActionResponseDeleteDataSchema,
    dataFields: ["message_target", "custom_label"],
  },
  action_message_create: {
    color: actionColor,
    icon: MessageCirclePlusIcon,
    defaultTitle: "Create channel message",
    defaultDescription: "Bot sends a message to a channel",
    dataSchema: nodeActionMessageCreateDataSchema,
    dataFields: ["message_template_id", "message_data", "custom_label"],
  },
  action_message_edit: {
    color: actionColor,
    icon: PenIcon,
    defaultTitle: "Edit channel message",
    defaultDescription: "Bot edits an existing message in a channel",
    dataSchema: nodeActionMessageEditDataSchema,
    dataFields: [
      "message_target",
      "message_template_id",
      "message_data",
      "custom_label",
    ],
  },
  action_message_delete: {
    color: actionColor,
    icon: MessageCircleXIcon,
    defaultTitle: "Delete channel message",
    defaultDescription: "Bot deletes an existing message in a channel",
    dataSchema: nodeActionMessageDeleteDataSchema,
    dataFields: ["message_target", "audit_log_reason", "custom_label"],
  },
  action_member_ban: {
    color: actionColor,
    icon: UserRoundXIcon,
    defaultTitle: "Ban member",
    defaultDescription: "Ban a member from the server",
    dataSchema: nodeActionMemberBanDataSchema,
    dataFields: [
      "member_target",
      "member_ban_delete_message_duration",
      "audit_log_reason",
      "custom_label",
    ],
  },
  action_member_kick: {
    color: actionColor,
    icon: UserRoundMinusIcon,
    defaultTitle: "Kick member",
    defaultDescription: "Kick a member from the server",
    dataSchema: nodeActionMemberKickDataSchema,
    dataFields: ["member_target", "audit_log_reason", "custom_label"],
  },
  action_member_timeout: {
    color: actionColor,
    icon: MessageCircleOffIcon,
    defaultTitle: "Timeout member",
    defaultDescription: "Timeout a member in the server",
    dataSchema: nodeActionMemberTimeoutDataSchema,
    dataFields: [
      "member_target",
      "member_timeout_duration",
      "audit_log_reason",
      "custom_label",
    ],
  },
  action_channel_create: {
    color: actionColor,
    icon: FolderPlusIcon,
    defaultTitle: "Create channel",
    defaultDescription: "Create a new channel in the server",
    dataSchema: nodeActionChannelCreateDataSchema,
    dataFields: ["channel_data", "audit_log_reason", "custom_label"],
  },
  action_channel_edit: {
    color: actionColor,
    icon: FolderPenIcon,
    defaultTitle: "Edit channel",
    defaultDescription: "Edit an existing channel in the server",
    dataSchema: nodeActionChannelEditDataSchema,
    dataFields: [
      "channel_target",
      "channel_data",
      "audit_log_reason",
      "custom_label",
    ],
  },
  action_channel_delete: {
    color: actionColor,
    icon: FolderMinusIcon,
    defaultTitle: "Delete channel",
    defaultDescription: "Delete an existing channel in the server",
    dataSchema: nodeActionChannelDeleteDataSchema,
    dataFields: ["channel_target", "audit_log_reason", "custom_label"],
  },
  action_thread_create: {
    color: actionColor,
    icon: FolderPlusIcon,
    defaultTitle: "Start thread",
    defaultDescription: "Start a new thread under a message or in a channel",
    dataSchema: nodeActionThreadCreateDataSchema,
    dataFields: ["message_target", "audit_log_reason", "custom_label"],
  },
  action_role_create: {
    color: actionColor,
    icon: BookmarkPlusIcon,
    defaultTitle: "Create role",
    defaultDescription: "Create a new role in the server",
    dataSchema: nodeActionRoleCreateDataSchema,
    dataFields: ["role_data", "audit_log_reason", "custom_label"],
  },
  action_role_edit: {
    color: actionColor,
    icon: BookmarkIcon,
    defaultTitle: "Edit role",
    defaultDescription: "Edit an existing role in the server",
    dataSchema: nodeActionRoleEditDataSchema,
    dataFields: [
      "role_target",
      "role_data",
      "audit_log_reason",
      "custom_label",
    ],
  },
  action_role_delete: {
    color: actionColor,
    icon: BookmarkMinusIcon,
    defaultTitle: "Delete role",
    defaultDescription: "Delete an existing role in the server",
    dataSchema: nodeActionRoleDeleteDataSchema,
    dataFields: ["role_target", "audit_log_reason", "custom_label"],
  },
  action_variable_set: {
    color: actionColor,
    icon: VariableIcon,
    defaultTitle: "Set variable",
    defaultDescription: "Set a variable to a value",
    dataSchema: nodeActionVariableSetDataSchema,
    dataFields: ["variable_name", "variable_value", "custom_label"],
  },
  action_variable_delete: {
    color: actionColor,
    icon: VariableIcon,
    defaultTitle: "Delete variable",
    defaultDescription: "Delete a variable",
    dataSchema: nodeActionVariableDeleteDataSchema,
    dataFields: ["variable_name", "custom_label"],
  },
  action_http_request: {
    color: actionColor,
    icon: WebhookIcon,
    defaultTitle: "Send API Request",
    defaultDescription: "Send an API request to an external server",
    dataSchema: nodeActionHttpRequestDataSchema,
    dataFields: ["http_request_data", "custom_label"],
  },
  action_log: {
    color: actionColor,
    icon: ScrollTextIcon,
    defaultTitle: "Log Message",
    defaultDescription:
      "Log some text which is only visible in the deployment logs",
    dataSchema: nodeActionLogDataSchema,
    dataFields: ["log_level", "log_message", "custom_label"],
  },
  control_condition_compare: {
    color: controlColor,
    icon: ArrowLeftRightIcon,
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
    icon: CircleHelpIcon,
    defaultTitle: "Match Condition",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the two values are equal.",
    dataFields: ["condition_item_compare_mode", "condition_item_compare_value"],
  },
  control_condition_user: {
    color: controlColor,
    icon: UserSearchIcon,
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
    icon: CircleHelpIcon,
    defaultTitle: "Match User",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the user meets the criteria.",
    dataFields: ["condition_item_user_mode", "condition_item_user_value"],
  },
  control_condition_channel: {
    color: controlColor,
    icon: FolderSearchIcon,
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
    icon: CircleHelpIcon,
    defaultTitle: "Match Channel",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the channel meets the criteria.",
    dataFields: ["condition_item_channel_mode", "condition_item_channel_value"],
  },
  control_condition_role: {
    color: controlColor,
    icon: BookmarkIcon,
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
    icon: CircleHelpIcon,
    defaultTitle: "Match Role",
    dataSchema: nodeConditionItemCompareDataSchema,
    defaultDescription: "Run actions if the role meets the criteria.",
    dataFields: ["condition_item_role_mode", "condition_item_role_value"],
  },
  control_condition_item_else: {
    color: errorColor,
    icon: XCircleIcon,
    defaultTitle: "Else",
    defaultDescription: "Run actions if no other conditions are met.",
    dataFields: [],
    fixed: true,
  },
  control_loop: {
    color: controlColor,
    icon: Repeat2Icon,
    defaultTitle: "Run a loop",
    dataSchema: nodeControlLoopDataSchema,
    defaultDescription: "Run a set of actions multiple times.",
    dataFields: ["loop_count", "custom_label"],
    ownsChildren: true,
  },
  control_loop_each: {
    color: controlColor,
    icon: Repeat2Icon,
    defaultTitle: "Each loop iteration",
    defaultDescription: "Run actions for each iteration of the loop.",
    dataFields: [],
    fixed: true,
  },
  control_loop_end: {
    color: controlColor,
    icon: CornerDownRightIcon,
    defaultTitle: "After loop",
    defaultDescription: "Run actions after the loop has finished.",
    dataFields: [],
    fixed: true,
  },
  control_loop_exit: {
    color: controlColor,
    icon: LogOutIcon,
    defaultTitle: "Exit loop",
    defaultDescription: "Exit out of the loop.",
    dataFields: [],
  },
  option_command_argument: {
    color: optionColor,
    icon: TextCursorInputIcon,
    defaultTitle: "Command Argument",
    defaultDescription: "Argument for a command.",
    dataSchema: nodeOptionCommandArgumentDataSchema,
    dataFields: [
      "name",
      "description",
      "command_argument_type",
      "command_argument_required",
    ],
  },
  option_command_permissions: {
    color: optionColor,
    icon: ShieldCheckIcon,
    defaultTitle: "Command Permissions",
    defaultDescription:
      "Make the command only available to users with the specified permissions.",
    dataSchema: nodeOptionCommandPermissionsSchema,
    dataFields: ["command_permissions"],
  },
  option_command_contexts: {
    color: optionColor,
    icon: MapPinIcon,
    defaultTitle: "Command Contexts",
    defaultDescription:
      "Define if the command should be available in direct messages or just in servers.",
    dataSchema: nodeOptionCommandContextsSchema,
    dataFields: ["command_contexts"],
  },
  option_event_filter: {
    color: optionColor,
    icon: FilterIcon,
    defaultTitle: "Event Filter",
    defaultDescription: "Filter events based on their properties.",
    dataSchema: nodeOptionEventFilterSchema,
    dataFields: ["event_filter_target", "event_filter_expression"],
  },
};

const unknownNodeType: NodeValues = {
  color: "#ff0000",
  icon: CircleHelpIcon,
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
  const id = getUniqueId().toString();

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

  // TODO: connect option types to entry automatically?

  if (conditionChildType.hasOwnProperty(type)) {
    const [elseNodes, elseEdges] = createNode("control_condition_item_else", {
      x: position.x + 200,
      y: position.y + 200,
    });

    nodes.push(...elseNodes);
    edges.push({
      id: getUniqueId().toString(),
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
      id: getUniqueId().toString(),
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
      id: getUniqueId().toString(),
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
      id: getUniqueId().toString(),
      source: id,
      target: eachNodes[0].id,
      type: "fixed",
    });
    edges.push(...eachEdges);
  }

  return [nodes, edges];
}
