import { NodeProps, Position } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { optionColor } from "@/lib/flow/nodes";

export default function FlowNodeEntryCommand(props: NodeProps<NodeData>) {
  return (
    <FlowNodeBase
      {...props}
      title={"/" + (props.data.name || "")}
      highlight={true}
    >
      <FlowNodeHandle
        type="target"
        position={Position.Top}
        color={optionColor}
      />
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
