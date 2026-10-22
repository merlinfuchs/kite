import { Position } from "@xyflow/react";
import { getBlockDefinition } from "../../lib/blocks";
import { NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

// A block with a labeled handle for each of its outputs, like the error and
// default branches of the error handler.
export default function FlowNodeBranches(props: NodeProps) {
  const block = getBlockDefinition(props.type);

  return (
    <div className="relative">
      <FlowNodeBase {...props}>
        <FlowNodeHandle type="target" position={Position.Top} />
      </FlowNodeBase>

      <div className="flex items-center justify-around gap-3">
        {block?.outputs?.map((output) => (
          <div className="relative" key={output}>
            <div className="px-2 py-1 shadow-md rounded-b bg-muted relative text-center flex items-center justify-center text-foreground gap-2">
              <div className="text-xs truncate">
                {block.output_labels?.[output]?.label ?? output}
              </div>
            </div>

            <FlowNodeHandle
              type="source"
              position={Position.Bottom}
              id={output}
              color={block.output_labels?.[output]?.color}
            />
          </div>
        ))}
      </div>
    </div>
  );
}
