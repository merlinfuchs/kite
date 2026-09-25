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
}

export default function Flow({ flowData, logs, context, onChange }: Props) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<FlowEditorApi>(null);
  const [chatOpen, setChatOpen] = useState(false);
  const [aiBusy, setAIBusy] = useState(false);

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
          {/* The AI's edits apply to the flow as it was when it was asked. */}
          {aiBusy && <div className="absolute inset-0 z-10 cursor-wait" />}
          {!chatOpen && (
            <Button
              variant="secondary"
              size="sm"
              className="absolute top-3 right-3 z-10 gap-2"
              onClick={() => setChatOpen(true)}
            >
              <SparklesIcon className="size-4" />
              Ask AI
            </Button>
          )}
        </div>

        {/* Hidden rather than unmounted, so the chat is kept when closed. */}
        <div className={cn("flex", !chatOpen && "hidden")}>
          <FlowAIChat
            context={context}
            editorRef={editorRef}
            onBusyChange={setAIBusy}
            onClose={() => setChatOpen(false)}
          />
        </div>
      </div>
    </FlowContextStoreProvider>
  );
}
