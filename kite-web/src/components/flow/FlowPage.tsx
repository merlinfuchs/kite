import { FlowData, NodeType } from "@/lib/flow/dataSchema";
import FlowNav from "./FlowNav";
import { ReactFlowProvider, useReactFlow } from "@xyflow/react";
import { useCallback, useState } from "react";
import Flow from "./Flow";
import { FlowContextType } from "@/lib/flow/context";
import { LogEntry } from "@/lib/types/wire.gen";
import UnsavedChangesDialog from "@/components/common/UnsavedChangesDialog";
import { bypassUnsavedChangesWarning } from "@/lib/hooks/exit";

interface Props {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  hasUnsavedChanges: boolean;
  hasUndeployedChanges?: boolean;
  isDeploying?: boolean;
  onDeploy?: () => void;
  onChange: () => void;
  isSaving: boolean;
  onSave: (data: FlowData, options?: { onSuccess?: () => void }) => void;
  onExit: () => void;
}

function InnerFlowPage({
  flowData,
  logs,
  context,
  hasUnsavedChanges,
  onChange,
  isSaving,
  hasUndeployedChanges,
  onDeploy,
  onSave,
  onExit,
}: Props) {
  const { getNodes, getEdges } = useReactFlow<NodeType>();
  const [chatOpen, setChatOpen] = useState(false);
  const [unsavedDialogOpen, setUnsavedDialogOpen] = useState(false);

  const save = useCallback(
    (options?: { onSuccess?: () => void }) => {
      onSave(
        {
          nodes: getNodes(),
          edges: getEdges(),
        },
        options
      );
    },
    [getNodes, getEdges, onSave]
  );

  const handleExitRequest = useCallback(() => {
    if (hasUnsavedChanges) {
      setUnsavedDialogOpen(true);
    } else {
      onExit();
    }
  }, [hasUnsavedChanges, onExit]);

  const handleDiscardAndExit = useCallback(() => {
    setUnsavedDialogOpen(false);
    bypassUnsavedChangesWarning(() => {
      onExit();
    });
  }, [onExit]);

  const handleSaveAndExit = useCallback(() => {
    save({
      onSuccess: () => {
        setUnsavedDialogOpen(false);
        bypassUnsavedChangesWarning(() => {
          onExit();
        });
      },
    });
  }, [save, onExit]);

  return (
    <div className="h-[100dvh] w-[100dvw] flex flex-col">
      <div className="flex-none">
        <FlowNav
          hasUnsavedChanges={hasUnsavedChanges}
          isSaving={isSaving}
          hasUndeployedChanges={hasUndeployedChanges}
          onDeploy={onDeploy}
          onSave={onSave}
          onExit={handleExitRequest}
          chatOpen={chatOpen}
          onChatOpenChange={setChatOpen}
        />
      </div>
      <Flow
        flowData={flowData}
        logs={logs}
        context={context}
        onChange={onChange}
        chatOpen={chatOpen}
        onChatOpenChange={setChatOpen}
      />
      <UnsavedChangesDialog
        open={unsavedDialogOpen}
        onOpenChange={setUnsavedDialogOpen}
        onSaveAndExit={handleSaveAndExit}
        onDiscardAndExit={handleDiscardAndExit}
        isSaving={isSaving}
      />
    </div>
  );
}

export default function FlowPage(props: Props) {
  return (
    <ReactFlowProvider>
      <style jsx global>{`
        @media (max-width: 767px) {
          [data-sonner-toaster][data-y-position="top"] {
            top: 56px !important;
            --offset: 56px !important;
          }
        }
      `}</style>
      <InnerFlowPage {...props} />
    </ReactFlowProvider>
  );
}
