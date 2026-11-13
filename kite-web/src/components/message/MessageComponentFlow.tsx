import { memo, useCallback, useMemo } from "react";
import { useShallow } from "zustand/react/shallow";
import { FlowData } from "@/lib/flow/dataSchema";
import { FlowContextType } from "@/lib/flow/context";
import {
  getMessage,
  useCurrentFlow,
  useCurrentFlowStore,
  useDocumentStoreApi,
} from "@/lib/message/state";
import { FlowTestTarget } from "@/lib/flow/flowTest";
import { newComponentFlow } from "@/lib/flow/newFlows";
import FlowDialog from "../flow/FlowDialog";
import FlowPreview from "../flow/FlowPreview";
import { useMessageId } from "@/lib/hooks/params";

const initialFlow = newComponentFlow();

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

  const messageId = useMessageId();
  const documentStore = useDocumentStoreApi();
  const flowStore = useCurrentFlowStore();
  // Only message templates can be tested, not messages of the message
  // creator.
  const testTarget = useMemo<FlowTestTarget | undefined>(
    () =>
      messageId
        ? {
            kind: "message",
            id: messageId,
            flowSourceId,
            getMessage: () => ({
              data: getMessage(documentStore),
              flow_sources: flowStore.getState().flowSources,
            }),
          }
        : undefined,
    [messageId, flowSourceId, documentStore, flowStore]
  );

  return (
    <FlowDialog
      flowData={flowData || initialFlow}
      logFilter={{ messageId }}
      context={context}
      viewportKey={`component:${messageId}:${flowSourceId}`}
      testTarget={testTarget}
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
