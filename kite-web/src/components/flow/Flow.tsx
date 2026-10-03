import { FlowContextStoreProvider, FlowContextType } from "@/lib/flow/context";
import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import {
  Node,
  OnSelectionChangeParams,
  useReactFlow,
  useStoreApi,
} from "@xyflow/react";
import { useCallback, useRef, useState } from "react";
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

interface Props {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  onChange: () => void;
  // Whether the AI chat is open, if it's opened from outside, like the page's
  // header. Otherwise the editor shows its own button for it.
  chatOpen?: boolean;
  onChatOpenChange?: (open: boolean) => void;
}

export default function Flow({
  flowData,
  logs,
  context,
  onChange,
  chatOpen: controlledChatOpen,
  onChatOpenChange,
}: Props) {
  const isMobile = useIsMobile();
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [mobileEditorOpen, setMobileEditorOpen] = useState(false);
  const [addBlockOpen, setAddBlockOpen] = useState(false);
  const [logsOpen, setLogsOpen] = useState(false);
  const [canUndo, setCanUndo] = useState(false);
  const [canRedo, setCanRedo] = useState(false);

  const containerRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<FlowEditorApi>(null);
  const [ownChatOpen, setOwnChatOpen] = useState(false);
  const chatOpen = controlledChatOpen ?? ownChatOpen;
  const setChatOpen = onChatOpenChange ?? setOwnChatOpen;
  // The chat is mounted when first opened and then kept when closed.
  const [chatMounted, setChatMounted] = useState(false);
  if (chatOpen && !chatMounted) setChatMounted(true);
  const closeChat = useCallback(() => setChatOpen(false), [setChatOpen]);

  const { fitView } = useReactFlow();
  const store = useStoreApi();

  const onSelectionChange = useCallback(
    ({ nodes }: OnSelectionChangeParams) => {
      if (nodes.length === 1) {
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
    fitView({ padding: 0.2, duration: 250 });
  }, [fitView]);

  const handleUndo = useCallback(() => {
    editorRef.current?.undo();
  }, []);

  const handleRedo = useCallback(() => {
    editorRef.current?.redo();
  }, []);

  const handleFormat = useCallback(() => {
    editorRef.current?.format();
  }, []);

  return (
    <FlowContextStoreProvider type={context}>
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
            <FlowAIChat editorRef={editorRef} onClose={closeChat} />
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
