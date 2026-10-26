import FlowNodeEntryCommand from "@/components/flow/FlowNodeEntryCommand";
import FlowEdgeDeleteButton from "@/components/flow/FlowEdgeDeleteButton";
import FlowEdgeFixed from "@/components/flow/FlowEdgeFixed";
import FlowNodeActionBase from "@/components/flow/FlowNodeActionBase";
import { blockDefinitions } from "../blocks";
import { BlockComponent } from "../blocks/types";
import { ComponentType } from "react";
import FlowNodeEntryEvent from "@/components/flow/FlowNodeEntryEvent";
import FlowNodeConditionItem from "@/components/flow/FlowNodeConditionItem";
import FlowNodeOptionBase from "@/components/flow/FlowNodeOptionBase";
import FlowNodeOptionCommandArgument from "@/components/flow/FlowNodeOptionCommandArgument";
import FlowNodeControlLoop from "@/components/flow/FlowNodeControlLoop";
import FlowNodeControlLoopEach from "@/components/flow/FlowNodeControlLoopEach";
import FlowNodeControlLoopEnd from "@/components/flow/FlowNodeControlLoopEnd";
import FlowNodeTerminal from "@/components/flow/FlowNodeTerminal";
import FlowNodeCondition from "@/components/flow/FlowNodeCondition";
import FlowNodeControlSleep from "@/components/flow/FlowNodeControlSleep";
import FlowNodeEntryComponentButton from "@/components/flow/FlowNodeEntryComponentButton";
import FlowNodeSuspendBase from "@/components/flow/FlowNodeSuspendBase";
import FlowNodeActionMessage from "@/components/flow/FlowNodeActionMessage";
import FlowNodeBase from "@/components/flow/FlowNodeBase";
import FlowNodeBranches from "@/components/flow/FlowNodeBranches";
import FlowNodeNote from "@/components/flow/FlowNodeNote";
import { noteNodeType } from "./notes";

const blockComponents: Record<BlockComponent, ComponentType<any>> = {
  action: FlowNodeActionBase,
  action_message: FlowNodeActionMessage,
  option: FlowNodeOptionBase,
  option_command_argument: FlowNodeOptionCommandArgument,
  entry_command: FlowNodeEntryCommand,
  entry_event: FlowNodeEntryEvent,
  entry_component_button: FlowNodeEntryComponentButton,
  condition: FlowNodeCondition,
  condition_item: FlowNodeConditionItem,
  control_loop: FlowNodeControlLoop,
  control_loop_each: FlowNodeControlLoopEach,
  control_loop_end: FlowNodeControlLoopEnd,
  control_sleep: FlowNodeControlSleep,
  terminal: FlowNodeTerminal,
  branches: FlowNodeBranches,
  suspend: FlowNodeSuspendBase,
};

export const nodeTypes = {
  ...Object.fromEntries(
    blockDefinitions.map((block) => [
      block.type,
      blockComponents[block.component ?? "action"],
    ])
  ),
  // Notes aren't blocks, see notes.ts.
  [noteNodeType]: FlowNodeNote,
};

export const edgeTypes = {
  delete_button: FlowEdgeDeleteButton,
  fixed: FlowEdgeFixed,
};
