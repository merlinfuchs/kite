import { Position } from "@xyflow/react";
import { NodeProps } from "@/lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeEntryFunction(props: NodeProps) {
  const parameters = (props.data.function_parameters ?? [])
    .map((p) => p.name)
    .filter(Boolean);

  return (
    <FlowNodeBase
      {...props}
      title={props.data.name || "Function"}
      description={
        parameters.length > 0
          ? `Receives ${parameters.join(", ")}. Drop different actions here!`
          : "Runs when a Call function block calls it. Drop different actions here!"
      }
      highlight={true}
      showConnectedMarker={false}
    >
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
