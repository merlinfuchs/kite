import { Handle, NodeProps, Position } from "reactflow";
import { NodeData } from "./types";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeMarkers from "./FlowNodeMarkers";
import { ChatBubbleBottomCenterIcon } from "@heroicons/react/24/solid";

export default function FlowNodeActionBase(props: NodeProps<NodeData>) {
  const { id } = props;

  return (
    <FlowNodeBase
      {...props}
      title="Plain text response"
      description="Bot replies with a plain text response"
      color="#3b82f6"
      icon={ChatBubbleBottomCenterIcon}
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
