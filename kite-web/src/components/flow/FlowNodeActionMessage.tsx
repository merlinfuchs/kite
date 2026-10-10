import { NodeProps } from "@/lib/flow/dataSchema";
import { suspendColor } from "@/lib/flow/nodes";
import {
  collectComponentGroups,
  componentHandleId,
  getComponentHandleIds,
  isSelectMenu,
} from "@/lib/flow/resume";
import { selectMenuType } from "@/lib/message/selectMenu";
import {
  ComponentData,
  ComponentTypeChannelSelect,
  ComponentTypeMentionableSelect,
  ComponentTypeRoleSelect,
  ComponentTypeUserSelect,
} from "@/lib/types/message.gen";
import { Position, useUpdateNodeInternals } from "@xyflow/react";
import {
  AtSignIcon,
  HashIcon,
  ListIcon,
  LucideIcon,
  MousePointerClickIcon,
  ShieldIcon,
  UserIcon,
} from "lucide-react";
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
  const handleIds = useMemo(
    () =>
      getComponentHandleIds(props.data.message_data?.components || []).join(
        ","
      ),
    [props.data.message_data]
  );
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

const selectMenuIcons: Record<number, LucideIcon> = {
  [ComponentTypeUserSelect]: UserIcon,
  [ComponentTypeRoleSelect]: ShieldIcon,
  [ComponentTypeMentionableSelect]: AtSignIcon,
  [ComponentTypeChannelSelect]: HashIcon,
};

function ComponentHandle({ comp }: { comp: ComponentData }) {
  const isSelect = isSelectMenu(comp);
  const color = isSelect
    ? buttonColors[2]
    : buttonColors[(comp.style ?? 1) as keyof typeof buttonColors];
  const Icon = isSelect
    ? selectMenuIcons[comp.type ?? 0] ?? ListIcon
    : MousePointerClickIcon;

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
          {isSelect
            ? comp.placeholder || selectMenuType(comp.type).label
            : comp.label}
        </div>
      </div>

      <FlowNodeHandle
        type="source"
        position={Position.Bottom}
        id={componentHandleId(comp.id)}
        size="small"
      />
    </div>
  );
}
