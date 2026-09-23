import { NodeProps } from "@/lib/flow/dataSchema";
import { suspendColor } from "@/lib/flow/nodes";
import {
  ButtonStyleLink,
  ComponentData,
  ComponentTypeActionRow,
  ComponentTypeButton,
  ComponentTypeStringSelect,
} from "@/lib/types/message.gen";
import { Position, useUpdateNodeInternals } from "@xyflow/react";
import { ListIcon, MousePointerClickIcon } from "lucide-react";
import { buttonColors } from "../message/MessageComponentButton";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { useEffect, useMemo } from "react";

export default function FlowNodeActionMessage(props: NodeProps) {
  const componentGroups = useMemo(
    () => collectComponentGroups(props.data.message_data?.components || []),
    [props.data.message_data]
  );
  const hasComponents = componentGroups.length > 0;

  // React Flow only re-measures handles when the node resizes, so swapping a
  // component for another of the same size would leave the new handle unknown.
  const updateNodeInternals = useUpdateNodeInternals();
  const handleIds = componentGroups.flat().map(buttonHandleId).join(",");
  useEffect(() => {
    updateNodeInternals(props.id);
  }, [handleIds, props.id, updateNodeInternals]);

  return (
    <div className="relative">
      <FlowNodeBase
        {...props}
        highlight={hasComponents}
        color={hasComponents ? suspendColor : undefined}
        showId
      >
        <FlowNodeHandle type="target" position={Position.Top} />
        <FlowNodeHandle
          type="source"
          position={hasComponents ? Position.Right : Position.Bottom}
        />
      </FlowNodeBase>

      <div className="flex flex-col mt-2 gap-5">
        {componentGroups.map((group) => (
          <div
            key={group[0].id}
            className="flex items-center justify-left gap-2"
          >
            {group.map((comp) => (
              <ComponentHandle comp={comp} key={comp.id} />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

// Groups interactive components the way Discord lays them out: one group per action row and one per section accessory.
function collectComponentGroups(
  components: ComponentData[]
): ComponentData[][] {
  const groups: ComponentData[][] = [];

  const isInteractive = (c: ComponentData) =>
    c.type === ComponentTypeStringSelect ||
    (c.type === ComponentTypeButton && c.style !== ButtonStyleLink);

  const walk = (c: ComponentData) => {
    if (c.type === ComponentTypeActionRow) {
      const interactive = (c.components || []).filter(isInteractive);
      if (interactive.length > 0) groups.push(interactive);
      return;
    }

    c.components?.forEach(walk);
    if (c.accessory && isInteractive(c.accessory)) groups.push([c.accessory]);
  };

  components.forEach(walk);
  return groups;
}

function ComponentHandle({ comp }: { comp: ComponentData }) {
  const isSelect = comp.type === ComponentTypeStringSelect;
  const color = isSelect
    ? buttonColors[2]
    : buttonColors[(comp.style ?? 1) as keyof typeof buttonColors];
  const Icon = isSelect ? ListIcon : MousePointerClickIcon;

  return (
    <div className="relative">
      <div
        className="px-2 shadow-md rounded-md relative max-w-32 min-w-16 text-center h-8 flex items-center justify-center text-white gap-2"
        style={{
          backgroundColor: color,
        }}
        key={comp.id}
      >
        <Icon className="w-4 h-4" />
        <div className="text-sm truncate">
          {isSelect ? comp.placeholder || "Select Menu" : comp.label}
        </div>
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
