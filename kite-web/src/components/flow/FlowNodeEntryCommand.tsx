import { Position } from "@xyflow/react";
import { NodeProps } from "@/lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { optionColor } from "@/lib/flow/nodes";

export default function FlowNodeEntryCommand(props: NodeProps) {
  const isChatInput =
    !props.data.command_type || props.data.command_type === "chat_input";
  return (
    <FlowNodeBase
      {...props}
      title={(isChatInput ? "/" : "") + (props.data.name || "")}
      highlight={true}
      showConnectedMarker={false}
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
