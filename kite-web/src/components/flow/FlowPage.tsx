import { FlowData, NodeType } from "@/lib/flow/dataSchema";
import FlowNav from "./FlowNav";
import { ReactFlowProvider, useReactFlow } from "@xyflow/react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Flow from "./Flow";
import { FlowContextType } from "@/lib/flow/context";
import { FlowVersion, LogEntry } from "@/lib/types/wire.gen";
import { FlowEditorApi } from "./FlowEditor";
import {
  FlowVersionTarget,
  getFlowVersion,
  useFlowVersionsQuery,
} from "@/lib/api/queries";
import { useFlowAutoSave } from "@/lib/hooks/flowAutoSave";
import FlowHistory from "./FlowHistory";
import { toast } from "sonner";

export interface FlowSaveOptions {
  // The save was made by auto-save rather than by the user.
  auto?: boolean;
}

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
  // Resolves to whether the save succeeded.
  onSave: (data: FlowData, options?: FlowSaveOptions) => Promise<boolean>;
  onExit: () => void;
  // Enables the save history, which needs the flow to be saved already.
  versionTarget?: FlowVersionTarget;
}

// How long auto-save waits after the last edit, so a burst of edits is saved
// once.
const autoSaveDelayMs = 2000;

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
  versionTarget,
}: Props) {
  const { getNodes, getEdges } = useReactFlow<NodeType>();
  const [chatOpen, setChatOpen] = useState(false);
  const editorRef = useRef<FlowEditorApi>(null);

  const save = useCallback(
    (options?: FlowSaveOptions) =>
      onSave(
        {
          nodes: getNodes(),
          edges: getEdges(),
        },
        options
      ),
    [getNodes, getEdges, onSave]
  );

  const saveNow = useCallback(() => save(), [save]);

  // Counts edits, so auto-save restarts its delay on each one and doesn't
  // retry a failed save until the flow is edited again.
  const [changeCount, setChangeCount] = useState(0);
  const failedAutoSave = useRef<number | null>(null);
  // Set by a restore, so the restored flow is saved once it's in the editor.
  const saveOnNextChange = useRef(false);
  // The version the last "Undo save" went back to, so pressing it again goes
  // further back instead of undoing the undo.
  const [undoChainId, setUndoChainId] = useState<string | null>(null);

  const handleChange = useCallback(() => {
    onChange();
    setChangeCount((c) => c + 1);

    if (saveOnNextChange.current) {
      saveOnNextChange.current = false;
      save();
    } else {
      setUndoChainId(null);
    }
  }, [onChange, save]);

  // Auto-save is turned on for each command or event listener on its own.
  const autoSaveKey = versionTarget
    ? "commandId" in versionTarget
      ? `command:${versionTarget.commandId}`
      : `event_listener:${versionTarget.eventListenerId}`
    : undefined;
  const [autoSave, setAutoSave] = useFlowAutoSave(autoSaveKey);

  useEffect(() => {
    if (!autoSave || !hasUnsavedChanges || isSaving) return;
    if (failedAutoSave.current === changeCount) return;

    const timeout = setTimeout(async () => {
      const ok = await save({ auto: true });
      if (!ok) failedAutoSave.current = changeCount;
    }, autoSaveDelayMs);
    return () => clearTimeout(timeout);
  }, [autoSave, hasUnsavedChanges, isSaving, changeCount, save]);

  const versionsQuery = useFlowVersionsQuery(
    versionTarget ?? { appId: "", commandId: "" },
    !!versionTarget
  );
  const versions = useMemo(
    () =>
      versionsQuery.data?.success
        ? versionsQuery.data.data.filter((v): v is FlowVersion => !!v)
        : undefined,
    [versionsQuery.data]
  );

  const [restoringId, setRestoringId] = useState<string | null>(null);

  const restore = useCallback(
    async (versionId: string, isUndo: boolean) => {
      if (!versionTarget) return;
      if (
        hasUnsavedChanges &&
        !confirm(
          "Restoring replaces your unsaved changes. You can still get them back with Undo (Ctrl+Z). Continue?"
        )
      ) {
        return;
      }

      setRestoringId(versionId);
      try {
        const res = await getFlowVersion(versionTarget, versionId);
        if (!res.success) {
          toast.error(
            `Failed to load save: ${res.error.message} (${res.error.code})`
          );
          return;
        }

        const data = res.data.flow_source as FlowData;
        // Goes through the editor so it's one step of its undo history, and
        // is saved once applied.
        saveOnNextChange.current = true;
        editorRef.current?.replaceFlow(data.nodes, data.edges);
        setUndoChainId(isUndo ? versionId : null);
      } catch (err) {
        toast.error(`Failed to load save: ${err}`);
      } finally {
        setRestoringId(null);
      }
    },
    [versionTarget, hasUnsavedChanges]
  );

  // The newest version is the current save, so undoing goes to the one after
  // it, or after the one the previous undo went to.
  const undoTarget = useMemo(() => {
    if (!versions) return undefined;

    const chainIndex = undoChainId
      ? versions.findIndex((v) => v.id === undoChainId)
      : -1;
    return versions[chainIndex === -1 ? 1 : chainIndex + 1];
  }, [versions, undoChainId]);

  const undoSave = useCallback(() => {
    if (undoTarget) restore(undoTarget.id, true);
  }, [undoTarget, restore]);

  return (
    <div className="h-[100dvh] w-[100dvw] flex flex-col">
      <div className="flex-none">
        <FlowNav
          hasUnsavedChanges={hasUnsavedChanges}
          isSaving={isSaving}
          hasUndeployedChanges={hasUndeployedChanges}
          onDeploy={onDeploy}
          onSave={saveNow}
          onExit={onExit}
          chatOpen={chatOpen}
          onChatOpenChange={setChatOpen}
          autoSave={autoSave}
          onAutoSaveChange={setAutoSave}
          history={
            versionTarget && (
              <FlowHistory
                versions={versions}
                hasUnsavedChanges={hasUnsavedChanges}
                isLoading={versionsQuery.isLoading}
                restoringId={restoringId}
                canUndoSave={!!undoTarget && !isSaving}
                onUndoSave={undoSave}
                onRestore={(v) => restore(v.id, false)}
              />
            )
          }
        />
      </div>
      <Flow
        flowData={flowData}
        logs={logs}
        context={context}
        onChange={handleChange}
        chatOpen={chatOpen}
        onChatOpenChange={setChatOpen}
        editorRef={editorRef}
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
