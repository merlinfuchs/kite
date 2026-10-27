import { FlowData } from "@/lib/flow/dataSchema";
import FlowNav from "./FlowNav";
import { ReactFlowProvider } from "@xyflow/react";
import { useCallback, useRef, useState } from "react";
import Flow from "./Flow";
import { FlowContextType } from "@/lib/flow/context";
import { Command, LogEntry } from "@/lib/types/wire.gen";
import { FlowEditorApi } from "./FlowEditor";
import { ItemVersionTarget } from "@/lib/api/queries";
import {
  useEditorSession,
  useVersionHistory,
} from "@/lib/hooks/versionHistory";
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
  // Resolves to whether the save succeeded. See useEditorSession.
  onSave: (data: FlowData, editorSession?: string) => Promise<boolean>;
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

  const editorSession = useEditorSession();
  // Set by a restore until it's saved, so it's saved without the session even
  // if the first try fails.
  const restorePending = useRef(false);

  // Saves run one after the other, so each sends the version the one before
  // saved instead of failing as if the item changed elsewhere.
  const lastSave = useRef<Promise<unknown>>(Promise.resolve());

  const handleSave = useCallback(
    (data: FlowData) => {
      const save = lastSave.current.then(async () => {
        const ok = await onSave(
          data,
          restorePending.current ? undefined : editorSession
        );
        if (ok) {
          restorePending.current = false;
          setSaveCount((c) => c + 1);
        }
        return ok;
      });
      lastSave.current = save;
      return save;
    },
    [onSave, editorSession]
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
        restorePending.current = true;
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
      handleSave(getFlowData());
    } else {
      resetUndoChain();
    }
  }, [onChange, handleSave, getFlowData, resetUndoChain]);

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
    if (await handleSave(getFlowData())) {
      setUnsavedDialogOpen(false);
      onExit();
    }
  }, [handleSave, getFlowData, onExit]);

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
