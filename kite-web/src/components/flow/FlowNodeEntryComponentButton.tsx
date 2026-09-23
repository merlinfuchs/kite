import { Position } from "@xyflow/react";
import { NodeProps } from "@/lib/flow/dataSchema";
import { useFlowContext } from "@/lib/flow/context";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";

// Buttons and select menus share this entry node, so it names whichever the flow belongs to.
export default function FlowNodeEntryComponentButton(props: NodeProps) {
  const selectMenu = useFlowContext((c) => c.type) === "component_select_menu";

  return (
    <FlowNodeBase
      {...props}
      highlight={true}
      showConnectedMarker={false}
      title={selectMenu ? "Select Menu" : undefined}
      description={
        selectMenu
          ? "This gets triggered when a user picks an option, available as {{interaction.value}}. Drop different actions here!"
          : undefined
      }
    >
      <FlowNodeHandle type="source" position={Position.Bottom} />
    </FlowNodeBase>
  );
}
