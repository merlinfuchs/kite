import { NodeProps, Position } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { conditionColor } from "@/lib/flow/nodes";

export default function FlowNodeConditionItemElse(props: NodeProps<NodeData>) {
  return (
    <FlowNodeBase {...props}>
      <FlowNodeHandle
        type="target"
        color={conditionColor}
        position={Position.Top}
        size="small"
        isConnectable={false}
      />
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
