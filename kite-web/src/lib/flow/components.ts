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
import FlowNodeControlSleep from "@/components/flow/FlowNodeControlSleep";
import FlowNodeEntryComponentButton from "@/components/flow/FlowNodeEntryComponentButton";
import FlowNodeSuspendBase from "@/components/flow/FlowNodeSuspendBase";
import FlowNodeActionMessage from "@/components/flow/FlowNodeActionMessage";

export const nodeTypes = {
  entry_command: FlowNodeEntryCommand,
  entry_event: FlowNodeEntryEvent,
  entry_component_button: FlowNodeEntryComponentButton,

  option_command_argument: FlowNodeOptionCommandArgument,
  option_command_permissions: FlowNodeOptionBase,
  option_command_contexts: FlowNodeOptionBase,
  option_event_filter: FlowNodeOptionBase,

  action_response_create: FlowNodeActionMessage,
  action_response_edit: FlowNodeActionMessage,
  action_response_delete: FlowNodeActionBase,
  action_response_defer: FlowNodeActionBase,
  action_message_create: FlowNodeActionMessage,
  action_message_edit: FlowNodeActionMessage,
  action_message_delete: FlowNodeActionBase,
  action_private_message_create: FlowNodeActionMessage,
  action_message_reaction_create: FlowNodeActionBase,
  action_message_reaction_delete: FlowNodeActionBase,
  action_member_ban: FlowNodeActionBase,
  action_member_unban: FlowNodeActionBase,
  action_member_kick: FlowNodeActionBase,
  action_member_timeout: FlowNodeActionBase,
  action_member_edit: FlowNodeActionBase,
  action_member_role_add: FlowNodeActionBase,
  action_member_role_remove: FlowNodeActionBase,
  action_member_get: FlowNodeActionBase,
  action_user_get: FlowNodeActionBase,
  action_channel_get: FlowNodeActionBase,
  action_role_get: FlowNodeActionBase,
  action_guild_get: FlowNodeActionBase,
  action_message_get: FlowNodeActionBase,
  action_roblox_user_get: FlowNodeActionBase,
  action_variable_set: FlowNodeActionBase,
  action_variable_delete: FlowNodeActionBase,
  action_variable_get: FlowNodeActionBase,
  action_http_request: FlowNodeActionBase,
  action_ai_chat_completion: FlowNodeActionBase,
  action_ai_web_search: FlowNodeActionBase,
  action_expression_evaluate: FlowNodeActionBase,
  action_random_generate: FlowNodeActionBase,
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
  control_sleep: FlowNodeControlSleep,

  suspend_response_modal: FlowNodeSuspendBase,
};

export const edgeTypes = {
  delete_button: FlowEdgeDeleteButton,
  fixed: FlowEdgeFixed,
};
