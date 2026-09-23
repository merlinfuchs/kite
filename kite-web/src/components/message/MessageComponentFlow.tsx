import { memo, useCallback } from "react";
import { useShallow } from "zustand/react/shallow";
import { FlowData } from "@/lib/flow/dataSchema";
import { FlowContextType } from "@/lib/flow/context";
import { useCurrentFlow } from "@/lib/message/state";
import { getUniqueId } from "@/lib/utils";
import FlowDialog from "../flow/FlowDialog";
import FlowPreview from "../flow/FlowPreview";

const initialFlow = {
  nodes: [
    {
      id: getUniqueId().toString(),
      position: { x: 0, y: 0 },
      data: {},
      type: "entry_component_button",
    },
  ],
  edges: [],
};

/**
 * The flow a button or select menu triggers, edited in a dialog. Memoized
 * because the preview is a full ReactFlow instance that editing the component
 * would otherwise re-render on every keystroke.
 */
export default memo(function MessageComponentFlow({
  flowSourceId,
  context,
}: {
  flowSourceId: string;
  context: FlowContextType;
}) {
  const [flowData, replaceFlow] = useCurrentFlow(
    useShallow((s) => [s.getFlow(flowSourceId), s.replaceFlow])
  );

  const onClose = useCallback(
    (d: FlowData) => replaceFlow(flowSourceId, d),
    [replaceFlow, flowSourceId]
  );

  return (
    <FlowDialog
      flowData={flowData || initialFlow}
      context={context}
      onClose={onClose}
    >
      <FlowPreview className="h-64 p-16 w-full" onClick={() => {}} />
    </FlowDialog>
  );
});
