import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

// Nothing runs after a Return block, so it has no output.
export default function FlowNodeControlFunctionReturn(props: NodeProps) {
  return (
    <FlowNodeBase
      {...props}
      description={
        props.data.function_return_stop
          ? "End the function and stop the flow that called it."
          : undefined
      }
    >
      <FlowNodeHandle type="target" position={Position.Top} />
    </FlowNodeBase>
  );
}
