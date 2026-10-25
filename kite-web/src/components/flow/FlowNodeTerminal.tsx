import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

// A block without outputs, like Exit loop or End flow.
export default function FlowNodeTerminal(props: NodeProps) {
  return (
    <FlowNodeBase {...props}>
      <FlowNodeHandle type="target" position={Position.Top} />
    </FlowNodeBase>
  );
}
