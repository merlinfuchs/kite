import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import { optionColor } from "@/lib/flow/nodes";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeOptionCommandArgument(props: NodeProps) {
  return (
    <FlowNodeBase
      {...props}
      title={props.data.name}
      description={props.data.description}
      showConnectedMarker={false}
    >
      <FlowNodeHandle
        type="source"
        position={Position.Bottom}
        color={optionColor}
      />
    </FlowNodeBase>
  );
}
