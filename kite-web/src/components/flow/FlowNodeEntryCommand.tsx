import { Position } from "@xyflow/react";
import { NodeProps } from "@/lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { optionColor } from "@/lib/flow/nodes";
import { isSlashCommand } from "@/lib/flow/commandType";

export default function FlowNodeEntryCommand(props: NodeProps) {
  return (
    <FlowNodeBase
      {...props}
      title={
        (isSlashCommand(props.data.command_type) ? "/" : "") +
        (props.data.name || "")
      }
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
