import { FlowData } from "@/lib/flow/dataSchema";
import FlowNav from "./FlowNav";
import { ReactFlowProvider } from "@xyflow/react";
import { useCallback, useRef, useState } from "react";
import Flow from "./Flow";
import { FlowContextType } from "@/lib/flow/context";
import { Command, LogEntry } from "@/lib/types/wire.gen";
import { FlowEditorApi } from "./FlowEditor";
import { ItemVersionTarget } from "@/lib/api/queries";
import { useVersionHistory } from "@/lib/hooks/versionHistory";
import VersionHistory from "@/components/common/VersionHistory";
import UnsavedChangesDialog from "@/components/common/UnsavedChangesDialog";
import {
  FlowParametersProvider,
  useFlowParametersStore,
  useGetFlowData,
} from "@/lib/flow/parametersStore";

interface Props {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  viewportKey?: string;
  hasUnsavedChanges: boolean;
  command?: Command;
  // The install of the template or bundle the flow was added with.
  bundleInstallId?: string | null;
  onChange: () => void;
  isSaving: boolean;
  // Resolves to whether the save succeeded.
  onSave: (data: FlowData) => Promise<boolean>;
  onExit: () => void;
  versionTarget: ItemVersionTarget;
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
  bundleInstallId,
  onSave,
  onExit,
  versionTarget,
}: Props) {
  const getFlowData = useGetFlowData();
  const [chatOpen, setChatOpen] = useState(false);
  const editorRef = useRef<FlowEditorApi>(null);
  const [unsavedDialogOpen, setUnsavedDialogOpen] = useState(false);
  const [saveCount, setSaveCount] = useState(0);

  const handleSave = useCallback<Props["onSave"]>(
    async (data) => {
      const ok = await onSave(data);
      if (ok) setSaveCount((c) => c + 1);
      return ok;
    },
    [onSave]
  );

  const save = useCallback(
    () => handleSave(getFlowData()),
    [getFlowData, handleSave]
  );

  // Set by a restore, so the restored flow is saved once it's in the editor.
  const saveOnNextChange = useRef(false);
  const parametersStore = useFlowParametersStore();

  const history = useVersionHistory({
    target: versionTarget,
    hasUnsavedChanges,
    apply: useCallback(
      (content: Record<string, any>) => {
        const data = content as FlowData;
        parametersStore.getState().setParameters(data.parameters ?? []);
        // Goes through the editor so it's one step of its undo history, and
        // is saved once applied.
        saveOnNextChange.current = true;
        editorRef.current?.replaceFlow(data.nodes, data.edges);
      },
      [parametersStore]
    ),
  });
  const { resetUndoChain } = history;

  const handleChange = useCallback(() => {
    onChange();

    if (saveOnNextChange.current) {
      saveOnNextChange.current = false;
      save();
    } else {
      resetUndoChain();
    }
  }, [onChange, save, resetUndoChain]);

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

  const handleSaveAndExit = useCallback(async () => {
    if (await save()) {
      setUnsavedDialogOpen(false);
      onExit();
    }
  }, [save, onExit]);

  return (
    <div className="h-[100dvh] w-[100dvw] flex flex-col">
      <div className="flex-none">
        <FlowNav
          hasUnsavedChanges={hasUnsavedChanges}
          isSaving={isSaving}
          command={command}
          bundleInstallId={bundleInstallId}
          onSave={handleSave}
          onExit={handleExitRequest}
          chatOpen={chatOpen}
          onChatOpenChange={setChatOpen}
          history={<VersionHistory history={history} isSaving={isSaving} />}
        />
      </div>
      <Flow
        flowData={flowData}
        logs={logs}
        context={context}
        viewportKey={viewportKey}
        onChange={handleChange}
        chatOpen={chatOpen}
        onChatOpenChange={setChatOpen}
        editorRef={editorRef}
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
      <FlowParametersProvider initial={props.flowData.parameters}>
        <InnerFlowPage {...props} />
      </FlowParametersProvider>
    </ReactFlowProvider>
  );
}
