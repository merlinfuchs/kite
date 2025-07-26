import { Position } from "@xyflow/react";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { NodeProps } from "@/lib/flow/dataSchema";

export default function FlowNodeControlSleep(props: NodeProps) {
  return (
    <FlowNodeBase {...props}>
      <FlowNodeHandle type="target" position={Position.Top} />
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
