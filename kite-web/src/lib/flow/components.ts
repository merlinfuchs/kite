import FlowNodeEntryCommand from "@/components/flow/FlowNodeEntryCommand";
import FlowEdgeDeleteButton from "@/components/flow/FlowEdgeDeleteButton";
import FlowEdgeFixed from "@/components/flow/FlowEdgeFixed";
import FlowNodeActionBase from "@/components/flow/FlowNodeActionBase";
import FlowNodeEntryEvent from "@/components/flow/FlowNodeEntryEvent";
import FlowNodeConditionCompare from "@/components/flow/FlowNodeConditionCompare";
import FlowNodeConditionItem from "@/components/flow/FlowNodeConditionItem";
import FlowNodeOptionBase from "@/components/flow/FlowNodeOptionBase";
import FlowNodeConditionUser from "@/components/flow/FlowNodeConditionUser";
import FlowNodeOptionCommandArgument from "@/components/flow/FlowNodeOptionCommandArgument";
import FlowNodeControlLoop from "@/components/flow/FlowNodeControlLoop";
import FlowNodeControlLoopEach from "@/components/flow/FlowNodeControlLoopEach";
import FlowNodeControlLoopEnd from "@/components/flow/FlowNodeControlLoopEnd";
import FlowNodeControlLoopExit from "@/components/flow/FlowNodeControlLoopExit";
import FlowNodeConditionChannel from "@/components/flow/FlowNodeConditionChannel";
import FlowNodeConditionRole from "@/components/flow/FlowNodeConditionRole";

export const nodeTypes = {
  entry_command: FlowNodeEntryCommand,
  entry_event: FlowNodeEntryEvent,

  option_command_argument: FlowNodeOptionCommandArgument,
  option_command_permissions: FlowNodeOptionBase,
  option_command_contexts: FlowNodeOptionBase,
  option_event_filter: FlowNodeOptionBase,

  action_response_create: FlowNodeActionBase,
  action_response_edit: FlowNodeActionBase,
  action_response_delete: FlowNodeActionBase,
  action_message_create: FlowNodeActionBase,
  action_message_edit: FlowNodeActionBase,
  action_message_delete: FlowNodeActionBase,
  action_member_ban: FlowNodeActionBase,
  action_member_kick: FlowNodeActionBase,
  action_member_timeout: FlowNodeActionBase,
  action_channel_create: FlowNodeActionBase,
  action_channel_edit: FlowNodeActionBase,
  action_channel_delete: FlowNodeActionBase,
  action_thread_create: FlowNodeActionBase,
  action_role_create: FlowNodeActionBase,
  action_role_edit: FlowNodeActionBase,
  action_role_delete: FlowNodeActionBase,
  action_variable_set: FlowNodeActionBase,
  action_variable_delete: FlowNodeActionBase,
  action_http_request: FlowNodeActionBase,
  action_log: FlowNodeActionBase,

  control_condition_compare: FlowNodeConditionCompare,
  control_condition_item_compare: FlowNodeConditionItem,
  control_condition_user: FlowNodeConditionUser,
  control_condition_item_user: FlowNodeConditionItem,
  control_condition_channel: FlowNodeConditionChannel,
  control_condition_item_channel: FlowNodeConditionItem,
  control_condition_role: FlowNodeConditionRole,
  control_condition_item_role: FlowNodeConditionItem,
  control_condition_item_else: FlowNodeConditionItem,
  control_loop: FlowNodeControlLoop,
  control_loop_each: FlowNodeControlLoopEach,
  control_loop_end: FlowNodeControlLoopEnd,
  control_loop_exit: FlowNodeControlLoopExit,
};

export const edgeTypes = {
  delete_button: FlowEdgeDeleteButton,
  fixed: FlowEdgeFixed,
};
