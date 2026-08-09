import { NodeProps } from "@/lib/flow/dataSchema";
import { suspendColor } from "@/lib/flow/nodes";
import { ComponentData, ComponentSelectOptionData } from "@/lib/types/message.gen";
import { Position } from "@xyflow/react";
import { ChevronDownIcon, MousePointerClickIcon } from "lucide-react";
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
            {row.components?.map((comp) =>
              comp.type === 3 ? (
                <SelectMenuOptions comp={comp} key={comp.id} />
              ) : (
                <ButtonHandle comp={comp} key={comp.id} />
              )
            )}
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
        id={`component_${comp.id}`}
        size="small"
      />
    </div>
  );
}

function SelectMenuOptions({ comp }: { comp: ComponentData }) {
  if (!comp.options?.length) return null;

  return (
    <div className="flex flex-col gap-2">
      {comp.options.map((opt) => (
        <SelectOptionHandle key={opt.id} opt={opt} />
      ))}
    </div>
  );
}

function SelectOptionHandle({ opt }: { opt: ComponentSelectOptionData }) {
  return (
    <div className="relative">
      <div className="px-2 shadow-md rounded-md relative max-w-40 min-w-16 text-center h-8 flex items-center justify-center text-white gap-2 bg-[#4E5058]">
        <ChevronDownIcon className="w-4 h-4 flex-none" />
        <div className="text-sm truncate">{opt.label || "Option"}</div>
      </div>

      <FlowNodeHandle
        type="source"
        position={Position.Bottom}
        id={`component_${opt.flow_source_id || opt.id}`}
        size="small"
      />
    </div>
  );
}
