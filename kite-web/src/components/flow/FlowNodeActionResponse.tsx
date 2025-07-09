import { NodeProps } from "@/lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { Position } from "@xyflow/react";
import { suspendColor } from "@/lib/flow/nodes";

export default function FlowNodeActionResponse(props: NodeProps) {
  const messageData = props.data.message_data;

  const components = messageData?.components || [];

  console.log(components);

  return (
    <div className="relative">
      <FlowNodeBase
        {...props}
        highlight={components?.length > 0}
        color={components?.length > 0 ? suspendColor : undefined}
      >
        <FlowNodeHandle type="target" position={Position.Top} />
        <FlowNodeHandle type="source" position={Position.Bottom} />
      </FlowNodeBase>

      <div className="flex flex-col mt-4 gap-5">
        {components?.map((row) => (
          <div key={row.id} className="flex items-center justify-left gap-2">
            {row.components?.map((comp) => (
              <div
                className="px-2 py-1.5 shadow-md rounded-sm bg-slate-300 relative max-w-32 min-w-16 text-center"
                key={comp.id}
              >
                <div className="text-sm">{comp.label}</div>

                <FlowNodeHandle
                  type="source"
                  position={Position.Bottom}
                  id={`component_${row.id}_${comp.id}`}
                  size="small"
                />
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
