import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { controlColor } from "@/lib/flow/nodes";

export default function FlowNodeConditionItem(props: NodeProps) {
  return (
    // Options of an AI Choice show their name, as they all have the same
    // title.
    <FlowNodeBase {...props} title={props.data.ai_choice_name || undefined}>
      <FlowNodeHandle
        type="target"
        color={controlColor}
        position={Position.Top}
        size="small"
        isConnectable={false}
      />
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
