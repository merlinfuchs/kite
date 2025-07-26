import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { controlColor } from "@/lib/flow/nodes";

export default function FlowNodeControlLoop(props: NodeProps) {
  return (
    <FlowNodeBase {...props}>
      <FlowNodeHandle type="target" position={Position.Top} />
      <FlowNodeHandle
        type="source"
        color={controlColor}
        position={Position.Bottom}
        size="small"
        isConnectable={false}
      />
    </FlowNodeBase>
  );
}
