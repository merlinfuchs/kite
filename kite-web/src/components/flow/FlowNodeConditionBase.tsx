import { Handle, NodeProps, Position } from "reactflow";
import { NodeData } from "./types";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeMarkers from "./FlowNodeMarkers";
import { ArrowsRightLeftIcon } from "@heroicons/react/24/solid";

export default function FlowNodeConditionBase(props: NodeProps<NodeData>) {
  const { id } = props;

  return (
    <FlowNodeBase
      {...props}
      title="Comparison Condition"
      description="Run actions based on the difference between two values."
      color="#22c55e"
      icon={ArrowsRightLeftIcon}
      highlight={false}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="w-2 h-2 rounded-full !bg-primary"
      />
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-2 h-2 rounded-full !bg-primary"
      />

      <FlowNodeMarkers nodeId={id} showIsConnected={true} />
    </FlowNodeBase>
  );
}
