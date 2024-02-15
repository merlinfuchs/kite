import { NodeProps, Position } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import { optionColor } from "@/lib/flow/nodes";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeOptionBase(props: NodeProps<NodeData>) {
  return (
    <FlowNodeBase
      {...props}
      title={props.data.name}
      description={props.data.description}
    >
      <FlowNodeHandle
        type="source"
        position={Position.Bottom}
        color={optionColor}
      />
    </FlowNodeBase>
  );
}
