import { Position } from "@xyflow/react";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { NodeProps } from "@/lib/flow/dataSchema";

export default function FlowNodeActionBase(props: NodeProps) {
  return (
    <FlowNodeBase {...props} showId>
      <FlowNodeHandle type="target" position={Position.Top} />
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
