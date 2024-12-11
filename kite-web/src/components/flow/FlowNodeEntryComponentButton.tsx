import { Position } from "@xyflow/react";
import { NodeProps } from "@/lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { optionColor } from "@/lib/flow/nodes";

export default function FlowNodeEntryComponentButton(props: NodeProps) {
  return (
    <FlowNodeBase {...props} highlight={true} showConnectedMarker={false}>
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
