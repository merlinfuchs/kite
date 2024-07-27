import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeMarkers from "./FlowNodeMarkers";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeEntryEvent(props: NodeProps) {
  return (
    <FlowNodeBase {...props} highlight={true} showConnectedMarker={false}>
      <FlowNodeHandle type="source" position={Position.Bottom} />
      <FlowNodeMarkers id={props.id} type={props.type} data={props.data} />
    </FlowNodeBase>
  );
}
