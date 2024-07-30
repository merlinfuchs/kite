import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { controlColor } from "@/lib/flow/nodes";

export default function FlowNodeControlLoopEach(props: NodeProps) {
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
