import { Handle, NodeProps, Position } from "reactflow";
import { NodeData } from "./types";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeMarkers from "./FlowNodeMarkers";
import { LanguageIcon } from "@heroicons/react/24/solid";

export default function FlowNodeConditionBase(props: NodeProps<NodeData>) {
  const { id } = props;

  return (
    <FlowNodeBase
      {...props}
      title="Text"
      description="A plain text option"
      color="#a855f7"
      icon={LanguageIcon}
      highlight={false}
    >
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-2 h-2 rounded-full !bg-purple-500"
      />

      <FlowNodeMarkers nodeId={id} />
    </FlowNodeBase>
  );
}
