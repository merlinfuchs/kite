import { Position } from "@xyflow/react";
import { NodeProps } from "../../lib/flow/dataSchema";
import { errorColor } from "../../lib/blocks/colors";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeControlWebhookVerify(props: NodeProps) {
  return (
    <div className="relative">
      <FlowNodeBase {...props}>
        <FlowNodeHandle type="target" position={Position.Top} />
      </FlowNodeBase>

      <div className="flex items-center justify-around gap-3">
        <div className="relative">
          <div className="px-2 py-1 shadow-md rounded-b bg-muted relative text-center flex items-center justify-center text-foreground gap-2">
            <div className="text-xs truncate">Valid</div>
          </div>

          <FlowNodeHandle type="source" position={Position.Bottom} id="valid" />
        </div>

        <div className="relative">
          <div className="px-2 py-1 shadow-md rounded-b bg-muted relative text-center flex items-center justify-center text-foreground gap-2">
            <div className="text-xs truncate">Invalid</div>
          </div>

          <FlowNodeHandle
            type="source"
            position={Position.Bottom}
            id="invalid"
            color={errorColor}
          />
        </div>
      </div>
    </div>
  );
}
