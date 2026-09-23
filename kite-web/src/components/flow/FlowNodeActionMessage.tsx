import { NodeProps } from "@/lib/flow/dataSchema";
import { suspendColor } from "@/lib/flow/nodes";
import {
  ButtonStyleLink,
  ComponentData,
  ComponentTypeActionRow,
  ComponentTypeButton,
} from "@/lib/types/message.gen";
import { Position } from "@xyflow/react";
import { MousePointerClickIcon } from "lucide-react";
import { buttonColors } from "../message/MessageComponentButton";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { useMemo } from "react";

export default function FlowNodeActionMessage(props: NodeProps) {
  const buttonGroups = useMemo(
    () => collectButtonGroups(props.data.message_data?.components || []),
    [props.data.message_data]
  );
  const hasButtons = buttonGroups.length > 0;

  return (
    <div className="relative">
      <FlowNodeBase
        {...props}
        highlight={hasButtons}
        color={hasButtons ? suspendColor : undefined}
        showId
      >
        <FlowNodeHandle type="target" position={Position.Top} />
        <FlowNodeHandle
          type="source"
          position={hasButtons ? Position.Right : Position.Bottom}
        />
      </FlowNodeBase>

      <div className="flex flex-col mt-2 gap-5">
        {buttonGroups.map((group) => (
          <div
            key={group[0].id}
            className="flex items-center justify-left gap-2"
          >
            {group.map((comp) => (
              <ButtonHandle comp={comp} key={comp.id} />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

// Groups clickable buttons the way Discord lays them out: one group per action row and one per section accessory.
function collectButtonGroups(components: ComponentData[]): ComponentData[][] {
  const groups: ComponentData[][] = [];

  const isClickable = (c: ComponentData) =>
    c.type === ComponentTypeButton && c.style !== ButtonStyleLink;

  const walk = (c: ComponentData) => {
    if (c.type === ComponentTypeActionRow) {
      const buttons = (c.components || []).filter(isClickable);
      if (buttons.length > 0) groups.push(buttons);
      return;
    }

    c.components?.forEach(walk);
    if (c.accessory && isClickable(c.accessory)) groups.push([c.accessory]);
  };

  components.forEach(walk);
  return groups;
}

function ButtonHandle({ comp }: { comp: ComponentData }) {
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
