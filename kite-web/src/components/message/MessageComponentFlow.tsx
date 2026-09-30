import { memo, useCallback } from "react";
import { useShallow } from "zustand/react/shallow";
import { FlowData } from "@/lib/flow/dataSchema";
import { FlowContextType } from "@/lib/flow/context";
import { useCurrentFlow } from "@/lib/message/state";
import { getUniqueId } from "@/lib/utils";
import FlowDialog from "../flow/FlowDialog";
import FlowPreview from "../flow/FlowPreview";
import { useLogEntriesQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { useAppId, useMessageId } from "@/lib/hooks/params";

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

  const logsQuery = useLogEntriesQuery(useAppId(), {
    limit: 10,
    messageId: useMessageId(),
    refetchInterval: 10000,
  });
  const logs = useResponseData(logsQuery);

  return (
    <FlowDialog
      flowData={flowData || initialFlow}
      logs={logs}
      context={context}
      onClose={onClose}
    >
      <FlowPreview
        className="h-64 p-16 w-full"
        onClick={() => {}}
        context={context}
      />
    </FlowDialog>
  );
});
