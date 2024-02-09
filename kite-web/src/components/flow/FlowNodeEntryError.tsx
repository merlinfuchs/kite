import { NodeProps, Position } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { optionColor } from "@/lib/flow/nodes";

export default function FlowNodeEntryError(props: NodeProps<NodeData>) {
  return (
    <FlowNodeBase {...props} highlight={true}>
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
