import { FlowContextStoreProvider, FlowContextType } from "@/lib/flow/context";
import { FlowData } from "@/lib/flow/dataSchema";
import { OnSelectionChangeParams } from "@xyflow/react";
import { useCallback, useRef, useState } from "react";
import FlowEditor, { FlowEditorApi } from "./FlowEditor";
import FlowMenu from "./FlowMenu";
import { LogEntry } from "@/lib/types/wire.gen";
import FlowAIChat from "./FlowAIChat";
import { Button } from "../ui/button";
import { SparklesIcon } from "lucide-react";
import { cn } from "@/lib/utils";

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
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<FlowEditorApi>(null);
  const [ownChatOpen, setOwnChatOpen] = useState(false);
  const chatOpen = controlledChatOpen ?? ownChatOpen;
  const setChatOpen = onChatOpenChange ?? setOwnChatOpen;
  // The chat is mounted when first opened and then kept when closed.
  const [chatMounted, setChatMounted] = useState(false);
  if (chatOpen && !chatMounted) setChatMounted(true);
  const closeChat = useCallback(() => setChatOpen(false), [setChatOpen]);

  const onSelectionChange = useCallback(
    ({ nodes }: OnSelectionChangeParams) => {
      if (nodes.length === 1) {
        setSelectedNodeId(nodes[0].id);
      } else {
        setSelectedNodeId(null);
      }
    },
    []
  );

  return (
    <FlowContextStoreProvider type={context}>
      <div
        ref={containerRef}
        className="flex flex-auto overflow-y-hidden relative"
      >
        <FlowMenu selectedNodeId={selectedNodeId} logs={logs} />

        <div className="flex-auto relative">
          <FlowEditor
            initialData={flowData}
            onChange={onChange}
            onSelectionChange={onSelectionChange}
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
          <div className={cn("flex", !chatOpen && "hidden")}>
            <FlowAIChat editorRef={editorRef} onClose={closeChat} />
          </div>
        )}
      </div>
    </FlowContextStoreProvider>
  );
}
