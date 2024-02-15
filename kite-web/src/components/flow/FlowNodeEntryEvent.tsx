import { NodeProps, Position } from "reactflow";
import { NodeData } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeMarkers from "./FlowNodeMarkers";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeEntryEvent(props: NodeProps<NodeData>) {
  return (
    <FlowNodeBase {...props} highlight={true}>
      <FlowNodeHandle type="source" position={Position.Bottom} />
      <FlowNodeMarkers id={props.id} type={props.type} data={props.data} />
    </FlowNodeBase>
  );
}
