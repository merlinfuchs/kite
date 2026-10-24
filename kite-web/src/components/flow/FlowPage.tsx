import { FlowData, NodeType } from "@/lib/flow/dataSchema";
import FlowNav from "./FlowNav";
import { ReactFlowProvider, useReactFlow } from "@xyflow/react";
import { useCallback, useState } from "react";
import Flow from "./Flow";
import { FlowContextType } from "@/lib/flow/context";
import { Command, LogEntry } from "@/lib/types/wire.gen";
import UnsavedChangesDialog from "@/components/common/UnsavedChangesDialog";

interface Props {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  viewportKey?: string;
  hasUnsavedChanges: boolean;
  command?: Command;
  onChange: () => void;
  isSaving: boolean;
  onSave: (data: FlowData, options?: { onSuccess?: () => void }) => void;
  onExit: () => void;
}

function InnerFlowPage({
  flowData,
  logs,
  context,
  viewportKey,
  hasUnsavedChanges,
  onChange,
  isSaving,
  command,
  onSave,
  onExit,
}: Props) {
  const { getNodes, getEdges } = useReactFlow<NodeType>();
  const [chatOpen, setChatOpen] = useState(false);
  const [unsavedDialogOpen, setUnsavedDialogOpen] = useState(false);
  const [saveCount, setSaveCount] = useState(0);

  const handleSave = useCallback<Props["onSave"]>(
    (data, options) => {
      setSaveCount((c) => c + 1);
      onSave(data, options);
    },
    [onSave]
  );

  const save = useCallback(
    (options?: { onSuccess?: () => void }) => {
      handleSave(
        {
          nodes: getNodes(),
          edges: getEdges(),
        },
        options
      );
    },
    [getNodes, getEdges, handleSave]
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
    onExit();
  }, [onExit]);

  const handleSaveAndExit = useCallback(() => {
    save({
      onSuccess: () => {
        setUnsavedDialogOpen(false);
        onExit();
      },
    });
  }, [save, onExit]);

  return (
    <div className="h-[100dvh] w-[100dvw] flex flex-col">
      <div className="flex-none">
        <FlowNav
          hasUnsavedChanges={hasUnsavedChanges}
          isSaving={isSaving}
          command={command}
          onSave={handleSave}
          onExit={handleExitRequest}
          chatOpen={chatOpen}
          onChatOpenChange={setChatOpen}
        />
      </div>
      <Flow
        flowData={flowData}
        logs={logs}
        context={context}
        viewportKey={viewportKey}
        onChange={onChange}
        chatOpen={chatOpen}
        onChatOpenChange={setChatOpen}
        saveCount={saveCount}
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
      <InnerFlowPage {...props} />
    </ReactFlowProvider>
  );
}
