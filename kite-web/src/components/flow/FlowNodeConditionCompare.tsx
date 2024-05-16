import { NodeProps, Position } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { conditionColor } from "@/lib/flow/nodes";

export default function FlowNodeConditionCompare(props: NodeProps<NodeData>) {
  return (
    <FlowNodeBase {...props}>
      <FlowNodeHandle type="target" position={Position.Top} />
      <FlowNodeHandle
        type="source"
        color={conditionColor}
        position={Position.Bottom}
        isConnectable={false}
        size="small"
      />
    </FlowNodeBase>
  );
}
