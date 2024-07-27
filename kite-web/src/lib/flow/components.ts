import FlowNodeEntryCommand from "@/components/flow/FlowNodeEntryCommand";
import FlowEdgeDeleteButton from "@/components/flow/FlowEdgeDeleteButton";
import FlowEdgeFixed from "@/components/flow/FlowEdgeFixed";
import FlowNodeActionBase from "@/components/flow/FlowNodeActionBase";
import FlowNodeEntryEvent from "@/components/flow/FlowNodeEntryEvent";
import FlowNodeConditionCompare from "@/components/flow/FlowNodeConditionCompare";
import FlowNodeConditionItemCompare from "@/components/flow/FlowNodeConditionItemCompare";
import FlowNodeConditionItemElse from "@/components/flow/FlowNodeConditionItemElse";
import FlowNodeOptionBase from "@/components/flow/FlowNodeOptionBase";
import FlowNodeConditionPermissions from "@/components/flow/FlowNodeConditionPermissions";
import FlowNodeConditionItemPermissions from "@/components/flow/FlowNodeConditionItemPermissions";

export const nodeTypes = {
  entry_command: FlowNodeEntryCommand,
  entry_event: FlowNodeEntryEvent,
  action_response_create: FlowNodeActionBase,
  action_message_create: FlowNodeActionBase,
  action_log: FlowNodeActionBase,
  condition_compare: FlowNodeConditionCompare,
  condition_item_compare: FlowNodeConditionItemCompare,
  condition_permissions: FlowNodeConditionPermissions,
  condition_item_permissions: FlowNodeConditionItemPermissions,
  condition_item_else: FlowNodeConditionItemElse,
  option_command_text: FlowNodeOptionBase,
  option_command_number: FlowNodeOptionBase,
  option_command_user: FlowNodeOptionBase,
  option_command_channel: FlowNodeOptionBase,
  option_command_role: FlowNodeOptionBase,
  option_command_attachment: FlowNodeOptionBase,
};

export const edgeTypes = {
  delete_button: FlowEdgeDeleteButton,
  fixed: FlowEdgeFixed,
};
