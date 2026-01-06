import { Position } from "@xyflow/react";
import { NodeProps } from "@/lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeEntryComponentSelect(props: NodeProps) {
  return (
    <FlowNodeBase {...props} highlight={true} showConnectedMarker={false}>
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
