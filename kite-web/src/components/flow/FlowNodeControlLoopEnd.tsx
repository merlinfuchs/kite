import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { controlColor } from "@/lib/flow/nodes";

export default function FlowNodeControlLoopEnd(props: NodeProps) {
  return (
    <FlowNodeBase {...props}>
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
