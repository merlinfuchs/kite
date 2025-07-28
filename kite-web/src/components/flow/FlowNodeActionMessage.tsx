import { NodeProps } from "@/lib/flow/dataSchema";
import { suspendColor } from "@/lib/flow/nodes";
import { ComponentData } from "@/lib/types/message.gen";
import { Position } from "@xyflow/react";
import { MousePointerClickIcon } from "lucide-react";
import { buttonColors } from "../message/MessageComponentButton";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { useMemo } from "react";

export default function FlowNodeActionMessage(props: NodeProps) {
  const components = useMemo(() => {
    const messageData = props.data.message_data;
    return messageData?.components || [];
  }, [props.data.message_data]);

  return (
    <div className="relative">
      <FlowNodeBase
        {...props}
        highlight={components?.length > 0}
        color={components?.length > 0 ? suspendColor : undefined}
        showId
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

  const color = buttonColors[(comp.style ?? 1) as keyof typeof buttonColors];

  return (
    <div className="relative">
      <div
        className="px-2 shadow-md rounded-md relative max-w-32 min-w-16 text-center h-8 flex items-center justify-center text-white gap-2"
        style={{
          backgroundColor: color,
        }}
        key={comp.id}
      >
        <MousePointerClickIcon className="w-4 h-4" />
        <div className="text-sm truncate">{comp.label}</div>
      </div>

      <FlowNodeHandle
        type="source"
        position={Position.Bottom}
        id={buttonHandleId(comp)}
        size="small"
      />
    </div>
  );
}

function buttonHandleId(comp: ComponentData) {
  // NOTE: The format has to match with the backend for the resume point to work
  return `component_${comp.id}`;
}
