import { FlowContextStoreProvider, FlowContextType } from "@/lib/flow/context";
import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import {
  Node,
  OnSelectionChangeParams,
  useStoreApi,
} from "@xyflow/react";
import { useCallback, useEffect, useRef, useState } from "react";
import FlowEditor, { FlowEditorApi, FlowSelectionSummary } from "./FlowEditor";
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
import FlowHelpPopup from "./FlowHelpPopup";
import { useIsMobile } from "@/lib/hooks/use-mobile";
import { track } from "@/lib/analytics";
import { useFlowParametersStore } from "@/lib/flow/parametersStore";
import { toast } from "sonner";

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
  saveCount,
}: Props) {
  const isMobile = useIsMobile();
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [mobileEditorOpen, setMobileEditorOpen] = useState(false);
  const [addBlockOpen, setAddBlockOpen] = useState(false);
  const [logsOpen, setLogsOpen] = useState(false);
  const [canUndo, setCanUndo] = useState(false);
  const [canRedo, setCanRedo] = useState(false);
  const [hasClipboard, setHasClipboard] = useState(false);
  const [mobileSelectMode, setMobileSelectMode] = useState(false);
  const [mobileMoveMode, setMobileMoveMode] = useState(false);

  const [selectionSummary, setSelectionSummary] =
    useState<FlowSelectionSummary>({
      topLevelCount: 0,
      actionableCount: 0,
      canDisconnect: false,
      singleSelectedNodeId: null,
    });

  const containerRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<FlowEditorApi>(null);
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

  const handleSelectionSummaryChange = useCallback(
    (summary: FlowSelectionSummary) => {
      setSelectionSummary(summary);
      if (summary.singleSelectedNodeId) {
        setSelectedNodeId(summary.singleSelectedNodeId);
      } else if (summary.topLevelCount === 0) {
        setSelectedNodeId(null);
        setMobileSelectMode(false);
        setMobileMoveMode(false);
      }
    },
    []
  );

  const onSelectionChange = useCallback(
    ({ nodes }: OnSelectionChangeParams) => {
      if (nodes.length === 1) {
        setSelectedNodeId(nodes[0].id);
      } else if (nodes.length === 0) {
        setSelectedNodeId(null);
        setMobileEditorOpen(false);
      }
    },
    []
  );

  const handleNodeTap = useCallback(
    (node: Node<NodeData>) => {
      setSelectedNodeId(node.id);
      if (isMobile && !mobileSelectMode && !mobileMoveMode) {
        setMobileEditorOpen(true);
      }
    },
    [isMobile, mobileSelectMode, mobileMoveMode]
  );

  const handleEnterMobileSelectMode = useCallback((node: Node<NodeData>) => {
    setMobileSelectMode(true);
    setMobileMoveMode(false);
    setSelectedNodeId(node.id);
  }, []);

  const handleExitMobileSelectMode = useCallback(() => {
    setMobileSelectMode(false);
    setMobileMoveMode(false);
    setSelectedNodeId(null);
  }, []);

  const handleHistoryChange = useCallback((u: boolean, r: boolean) => {
    setCanUndo(u);
    setCanRedo(r);
  }, []);

  const deselectNode = useCallback(() => {
    editorRef.current?.clearSelection();
    store.getState().addSelectedNodes([]);
    setSelectedNodeId(null);
    setMobileSelectMode(false);
    setMobileMoveMode(false);
    setMobileEditorOpen(false);
  }, [store]);

  const handleCenterFlow = useCallback(() => {
    editorRef.current?.centerFlow();
  }, []);

  const handleUndo = useCallback(() => {
    editorRef.current?.undo();
  }, []);

  const handleRedo = useCallback(() => {
    editorRef.current?.redo();
  }, []);

  const handleFormat = useCallback(() => {
    editorRef.current?.format();
  }, []);

  const handleCopy = useCallback(() => {
    const ok = editorRef.current?.copy();
    if (ok) {
      const count = selectionSummary.actionableCount;
      toast.success(
        count > 1 ? `Copied ${count} blocks` : "Copied block to clipboard"
      );
    }
  }, [selectionSummary.actionableCount]);

  const handleCut = useCallback(() => {
    const count = selectionSummary.actionableCount;
    const ok = editorRef.current?.cut();
    if (ok) {
      toast.success(count > 1 ? `Cut ${count} blocks` : "Cut block");
      deselectNode();
    }
  }, [selectionSummary.actionableCount, deselectNode]);

  const handleDuplicate = useCallback(() => {
    editorRef.current?.duplicate();
  }, []);

  const handlePaste = useCallback(() => {
    editorRef.current?.paste(null);
  }, []);

  const handleDisconnect = useCallback(() => {
    const count = selectionSummary.topLevelCount;
    const ok = editorRef.current?.disconnect();
    if (ok) {
      toast.success(
        count > 1
          ? `Disconnected ${count} blocks`
          : "Disconnected block connections"
      );
    }
  }, [selectionSummary.topLevelCount]);

  const handleDelete = useCallback(() => {
    const count = selectionSummary.actionableCount;
    const ok = editorRef.current?.deleteSelected();
    if (ok) {
      toast.success(count > 1 ? `Deleted ${count} blocks` : "Deleted block");
      deselectNode();
    }
  }, [selectionSummary.actionableCount, deselectNode]);

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
        onContextMenu={(e) => {
          const target = e.target as HTMLElement | null;
          if (target?.closest("input, textarea, [contenteditable]")) return;
          if (!target?.closest(".react-flow")) {
            e.preventDefault();
          }
        }}
      >
        <FlowMenu selectedNodeId={selectedNodeId} logs={logs} />

        <div className="flex-auto relative pb-14 md:pb-0">
          <FlowEditor
            initialData={flowData}
            onChange={onChange}
            onSelectionChange={onSelectionChange}
            onSelectionSummaryChange={handleSelectionSummaryChange}
            onClipboardChange={setHasClipboard}
            onNodeTap={handleNodeTap}
            onHistoryChange={handleHistoryChange}
            isMobile={isMobile}
            mobileSelectMode={mobileSelectMode}
            mobileMoveMode={mobileMoveMode}
            onEnterMobileSelectMode={handleEnterMobileSelectMode}
            onExitMobileSelectMode={handleExitMobileSelectMode}
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
          <FlowHelpPopup />
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
          topLevelCount={selectionSummary.topLevelCount}
          actionableCount={selectionSummary.actionableCount}
          canDisconnect={selectionSummary.canDisconnect}
          hasClipboard={hasClipboard}
          isSelectMode={mobileSelectMode}
          isMoveMode={mobileMoveMode}
          hidden={mobileEditorOpen || addBlockOpen || logsOpen}
          onAddBlock={() => setAddBlockOpen(true)}
          onEditNode={() => setMobileEditorOpen(true)}
          onDeselectNode={deselectNode}
          onStartMove={() => setMobileMoveMode(true)}
          onStopMove={() => setMobileMoveMode(false)}
          onCopy={handleCopy}
          onCut={handleCut}
          onDuplicate={handleDuplicate}
          onDisconnect={handleDisconnect}
          onDelete={handleDelete}
          onPaste={handlePaste}
          onOpenLogs={() => setLogsOpen(true)}
          onCenterFlow={handleCenterFlow}
          onUndo={handleUndo}
          onRedo={handleRedo}
          onFormat={handleFormat}
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
          onOpenChange={(open) => {
            setMobileEditorOpen(open);
            if (!open && !mobileSelectMode) {
              deselectNode();
            }
          }}
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
