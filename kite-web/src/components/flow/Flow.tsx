import { FlowContextStoreProvider, FlowContextType } from "@/lib/flow/context";
import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import { Node, OnSelectionChangeParams, useStoreApi } from "@xyflow/react";
import { RefObject, useCallback, useEffect, useRef, useState } from "react";
import FlowEditor, { FlowEditorApi } from "./FlowEditor";
import FlowMenu from "./FlowMenu";
import { LogEntry } from "@/lib/types/wire.gen";
import FlowAIChat from "./FlowAIChat";
import { Button } from "../ui/button";
import { SparklesIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import FlowMobileBottomBar from "./FlowMobileBottomBar";
import FlowAddBlockDrawer from "./FlowAddBlockDrawer";
import FlowNodeEditorDrawer from "./FlowNodeEditorDrawer";
import FlowLogsDrawer from "./FlowLogsDrawer";
import { useIsMobile } from "@/lib/hooks/use-mobile";
import { track } from "@/lib/analytics";
import { useFlowParametersStore } from "@/lib/flow/parametersStore";
import { isNoteNode } from "@/lib/flow/notes";

interface Props {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  viewportKey?: string;
  onChange: () => void;
  // Whether the AI chat is open, if it's opened from outside, like the page's
  // header. Otherwise the editor shows its own button for it.
  chatOpen?: boolean;
  onChatOpenChange?: (open: boolean) => void;
  // Lets the page edit the flow, e.g. to restore a save.
  editorRef?: RefObject<FlowEditorApi | null>;
  // Goes up when the flow is saved, which keeps the AI's changes.
  saveCount?: number;
}

export default function Flow({
  flowData,
  logs,
  context,
  viewportKey,
  onChange,
  chatOpen: controlledChatOpen,
  onChatOpenChange,
  editorRef: outerEditorRef,
  saveCount,
}: Props) {
  const isMobile = useIsMobile();
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [mobileEditorOpen, setMobileEditorOpen] = useState(false);
  const [addBlockOpen, setAddBlockOpen] = useState(false);
  const [logsOpen, setLogsOpen] = useState(false);
  const [canUndo, setCanUndo] = useState(false);
  const [canRedo, setCanRedo] = useState(false);

  const containerRef = useRef<HTMLDivElement>(null);
  const ownEditorRef = useRef<FlowEditorApi>(null);
  const editorRef = outerEditorRef ?? ownEditorRef;
  const [ownChatOpen, setOwnChatOpen] = useState(false);
  const chatOpen = controlledChatOpen ?? ownChatOpen;
  const setChatOpen = onChatOpenChange ?? setOwnChatOpen;
  // The chat is mounted when first opened and then kept when closed.
  const [chatMounted, setChatMounted] = useState(false);
  if (chatOpen && !chatMounted) setChatMounted(true);
  const closeChat = useCallback(() => setChatOpen(false), [setChatOpen]);

  useEffect(() => {
    if (chatOpen) track("flow_ai_opened", { context });
  }, [chatOpen, context]);

  const store = useStoreApi();

  // Changing a parameter is a change of the flow like changing a block.
  const parametersStore = useFlowParametersStore();
  useEffect(
    () => parametersStore.subscribe(() => onChange()),
    [parametersStore, onChange]
  );

  const onSelectionChange = useCallback(
    ({ nodes }: OnSelectionChangeParams) => {
      if (nodes.length === 1 && !isNoteNode(nodes[0])) {
        setSelectedNodeId(nodes[0].id);
      } else {
        setSelectedNodeId(null);
        setMobileEditorOpen(false);
      }
    },
    []
  );

  const handleNodeTap = useCallback(
    (node: Node<NodeData>) => {
      // Notes are edited on the canvas, they have no settings.
      if (isNoteNode(node)) return;
      setSelectedNodeId(node.id);
      if (isMobile) {
        setMobileEditorOpen(true);
      }
    },
    [isMobile]
  );

  const handleHistoryChange = useCallback((u: boolean, r: boolean) => {
    setCanUndo(u);
    setCanRedo(r);
  }, []);

  const deselectNode = useCallback(() => {
    store.getState().addSelectedNodes([]);
    setSelectedNodeId(null);
    setMobileEditorOpen(false);
  }, [store]);

  const handleFitView = useCallback(() => {
    editorRef.current?.fitView();
  }, [editorRef]);

  const handleUndo = useCallback(() => {
    editorRef.current?.undo();
  }, [editorRef]);

  const handleRedo = useCallback(() => {
    editorRef.current?.redo();
  }, [editorRef]);

  const handleFormat = useCallback(() => {
    editorRef.current?.format();
  }, [editorRef]);

  const handleAddNote = useCallback(() => {
    editorRef.current?.addNote();
  }, [editorRef]);

  return (
    <FlowContextStoreProvider type={context}>
      {/* Keep toasts below the nav bar on mobile */}
      <style jsx global>{`
        @media (max-width: 767px) {
          [data-sonner-toaster][data-y-position="top"] {
            top: 56px !important;
            --offset: 56px !important;
          }
        }
      `}</style>
      <div
        ref={containerRef}
        className="flex flex-auto overflow-y-hidden relative"
      >
        <FlowMenu selectedNodeId={selectedNodeId} logs={logs} />

        <div className="flex-auto relative pb-14 md:pb-0">
          <FlowEditor
            initialData={flowData}
            onChange={onChange}
            onSelectionChange={onSelectionChange}
            onNodeTap={handleNodeTap}
            onHistoryChange={handleHistoryChange}
            containerRef={containerRef}
            apiRef={editorRef}
            viewportKey={viewportKey}
          />
          {controlledChatOpen === undefined && !chatOpen && (
            <Button
              variant="secondary"
              size="sm"
              className="absolute top-3 left-3 z-10 gap-2"
              onClick={() => setChatOpen(true)}
            >
              <SparklesIcon className="size-4" />
              Ask AI
            </Button>
          )}
        </div>

        {chatMounted && (
          <div
            className={cn(
              "flex z-40 fixed inset-0 md:relative md:inset-auto",
              !chatOpen && "hidden"
            )}
          >
            <FlowAIChat
              editorRef={editorRef}
              saveCount={saveCount}
              onClose={closeChat}
            />
          </div>
        )}

        <FlowMobileBottomBar
          selectedNodeId={selectedNodeId}
          hidden={mobileEditorOpen || addBlockOpen || logsOpen}
          onAddBlock={() => setAddBlockOpen(true)}
          onEditNode={() => setMobileEditorOpen(true)}
          onDeselectNode={deselectNode}
          onOpenLogs={() => setLogsOpen(true)}
          onFitView={handleFitView}
          onUndo={handleUndo}
          onRedo={handleRedo}
          onFormat={handleFormat}
          onAddNote={handleAddNote}
          canUndo={canUndo}
          canRedo={canRedo}
          logCount={logs?.length}
        />

        <FlowAddBlockDrawer
          open={addBlockOpen}
          onOpenChange={setAddBlockOpen}
        />

        <FlowNodeEditorDrawer
          nodeId={selectedNodeId}
          open={mobileEditorOpen && !!selectedNodeId}
          onOpenChange={setMobileEditorOpen}
        />

        <FlowLogsDrawer
          logs={logs}
          open={logsOpen}
          onOpenChange={setLogsOpen}
        />
      </div>
    </FlowContextStoreProvider>
  );
}
