import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeControlErrorHandler(props: NodeProps) {
  return (
    <div className="relative">
      <FlowNodeBase {...props}>
        <FlowNodeHandle type="target" position={Position.Top} />
      </FlowNodeBase>

      <div className="flex items-center justify-around gap-3">
        <div className="relative">
          <div className="px-2 py-1 shadow-md rounded-b bg-muted relative text-center flex items-center justify-center text-white gap-2">
            <div className="text-xs truncate">Handle Error</div>
          </div>

          <FlowNodeHandle
            type="source"
            position={Position.Bottom}
            id="error"
            color="#ef4444"
          />
        </div>

        <div className="relative">
          <div className="px-2 py-1 shadow-md rounded-b bg-muted relative text-center flex items-center justify-center text-white gap-2">
            <div className="text-xs truncate">Try Blocks</div>
          </div>

          <FlowNodeHandle
            type="source"
            position={Position.Bottom}
            id="default"
          />
        </div>
      </div>
    </div>
  );
}
