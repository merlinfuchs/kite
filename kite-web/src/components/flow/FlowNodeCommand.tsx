import {
  Connection,
  Handle,
  NodeProps,
  Position,
  useReactFlow,
} from "reactflow";
import { NodeData } from "./types";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeMarkers from "./FlowNodeMarkers";
import { CommandLineIcon } from "@heroicons/react/24/solid";

export default function FlowNodeCommand(props: NodeProps<NodeData>) {
  const { id } = props;

  return (
    <FlowNodeBase
      {...props}
      title="/"
      description="Command entry. Drop different actions and options here!"
      icon={CommandLineIcon}
      highlight={true}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="w-2 h-2 rounded-full !bg-purple-500"
      />
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-2 h-2 rounded-full !bg-primary"
      />

      <FlowNodeMarkers nodeId={id} />
    </FlowNodeBase>
  );
}
