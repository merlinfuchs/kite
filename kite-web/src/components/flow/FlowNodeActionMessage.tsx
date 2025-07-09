import { NodeProps } from "@/lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { Position } from "@xyflow/react";
import { suspendColor } from "@/lib/flow/nodes";
import { ComponentData } from "@/lib/types/message.gen";
import { cn } from "@/lib/utils";
import { buttonColors } from "../message/MessageComponentButton";

export default function FlowNodeActionMessage(props: NodeProps) {
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
        <FlowNodeHandle
          type="source"
          position={components?.length > 0 ? Position.Right : Position.Bottom}
        />
      </FlowNodeBase>

      <div className="flex flex-col mt-2 gap-5">
        {components?.map((row) => (
          <div key={row.id} className="flex items-center justify-left gap-2">
            {row.components?.map((comp) => (
              <ButtonHandle comp={comp} key={comp.id} />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

function ButtonHandle({ comp }: { comp: ComponentData }) {
  if (comp.style === 5) {
    return null;
  }

  return (
    <div className="relative">
      <div
        className="px-2 shadow-md rounded-md relative max-w-32 min-w-16 text-center h-8 flex items-center justify-center text-white"
        style={{ backgroundColor: buttonColors[comp.style] }}
        key={comp.id}
      >
        <div className="text-sm truncate">{comp.label}</div>
      </div>

      <FlowNodeHandle
        type="source"
        position={Position.Bottom}
        id={`component_${comp.id}`}
        size="small"
      />
    </div>
  );
}
