import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { controlColor } from "@/lib/flow/nodes";

export default function FlowNodeControlLoopExit(props: NodeProps) {
  return (
    <FlowNodeBase {...props}>
      <FlowNodeHandle type="target" position={Position.Top} />
    </FlowNodeBase>
  );
}
