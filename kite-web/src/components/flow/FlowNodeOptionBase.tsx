import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import { optionColor } from "@/lib/flow/nodes";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeOptionBase(props: NodeProps) {
  return (
    <FlowNodeBase {...props} showConnectedMarker={false}>
      <FlowNodeHandle
        type="source"
        position={Position.Bottom}
        color={optionColor}
      />
    </FlowNodeBase>
  );
}
