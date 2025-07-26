import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeEntryEvent(props: NodeProps) {
  const eventName = props.data.event_type?.split("_").join(" ") || "";

  return (
    <FlowNodeBase
      {...props}
      title={`Listen for ${eventName}`}
      description={`Listens for ${eventName} events to trigger the flow. Drop different actions here!`}
      highlight={true}
      showConnectedMarker={false}
    >
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
