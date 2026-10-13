import FlowNodeEditor from "@/components/flow/FlowNodeEditor";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Background,
  BackgroundVariant,
  ReactFlow,
  ReactFlowProvider,
} from "@xyflow/react";
import { BracesIcon, Maximize2Icon } from "lucide-react";
import { useState } from "react";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import {
  FlowContextStoreProvider,
  type FlowContextType,
} from "@/lib/flow/context";
import type { FlowData } from "@/lib/flow/dataSchema";
import { useHookedTheme } from "@/lib/hooks/theme";
import "@xyflow/react/dist/base.css";

function ReadOnlyFlowCanvas({
  flowData,
  onSelectedNodeChange,
}: {
  flowData: FlowData;
  onSelectedNodeChange: (nodeId: string | null) => void;
}) {
  const { theme } = useHookedTheme();

  return (
    <ReactFlow
      defaultNodes={flowData.nodes}
      defaultEdges={flowData.edges.map((edge) => ({ ...edge, type: "fixed" }))}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      nodesDraggable={false}
      nodesConnectable={false}
      connectOnClick={false}
      deleteKeyCode={null}
      elementsSelectable
      edgesFocusable={false}
      colorMode={theme === "dark" ? "dark" : "light"}
      onNodeClick={(_, node) => onSelectedNodeChange(node.id)}
      onPaneClick={() => onSelectedNodeChange(null)}
      proOptions={{ hideAttribution: true }}
      className="!bg-background"
      fitView
    >
      <Background variant={BackgroundVariant.Dots} gap={18} size={1} />
    </ReactFlow>
  );
}

function FlowWorkspace({
  flowData,
  context,
  selectedNodeId,
  onSelectedNodeChange,
  onOpenJson,
  expanded = false,
}: {
  flowData: FlowData;
  context: FlowContextType;
  selectedNodeId: string | null;
  onSelectedNodeChange: (nodeId: string | null) => void;
  onOpenJson: () => void;
  expanded?: boolean;
}) {
  const selectedNode = flowData.nodes.find(
    (node) => node.id === selectedNodeId
  );

  return (
    <ReactFlowProvider>
      <FlowContextStoreProvider type={context} readOnly>
        <div
          className={`grid min-h-0 overflow-hidden bg-background lg:grid-cols-[minmax(0,1fr)_22rem] ${
            expanded ? "h-full" : "mt-4 min-h-[560px] rounded-xl border"
          }`}
        >
          <div
            className={`relative min-w-0 bg-muted/20 ${
              expanded ? "h-full" : "h-[560px]"
            }`}
          >
            <ReadOnlyFlowCanvas
              flowData={flowData}
              onSelectedNodeChange={onSelectedNodeChange}
            />
            <Button
              type="button"
              variant="secondary"
              size="sm"
              className="absolute bottom-4 left-4 z-10 shadow-md"
              onClick={onOpenJson}
            >
              <BracesIcon className="mr-2 size-4" />
              Flow JSON
            </Button>
          </div>
          <aside className="flex min-h-0 min-w-0 flex-col border-t bg-background lg:border-l lg:border-t-0">
            <div className="min-h-0 flex-1 overflow-hidden">
              {selectedNode ? (
                <FlowNodeEditor
                  key={selectedNode.id}
                  nodeId={selectedNode.id}
                  className="h-full"
                  hideTitle
                  readOnly
                />
              ) : (
                <div className="flex h-full min-h-[280px] items-center justify-center p-6 text-center text-sm text-muted-foreground">
                  Select a block on the canvas to view its settings.
                </div>
              )}
            </div>
          </aside>
        </div>
      </FlowContextStoreProvider>
    </ReactFlowProvider>
  );
}

export default function SupportFlowViewer({
  flowData,
  context,
}: {
  flowData: FlowData;
  context: FlowContextType;
}) {
  const [open, setOpen] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [jsonOpen, setJsonOpen] = useState(false);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  return (
    <div className="mt-5">
      <Button
        type="button"
        variant="outline"
        onClick={() => setOpen((value) => !value)}
        aria-expanded={open}
      >
        {open ? "Hide editor" : "View editor"}
      </Button>

      {open && (
        <>
          <div className="mt-4 flex items-center justify-between gap-3">
            <p className="text-sm text-muted-foreground">
              Read-only flow inspector
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setExpanded(true)}
            >
              <Maximize2Icon className="mr-2 size-4" />
              Expand editor
            </Button>
          </div>
          <FlowWorkspace
            flowData={flowData}
            context={context}
            selectedNodeId={selectedNodeId}
            onSelectedNodeChange={setSelectedNodeId}
            onOpenJson={() => setJsonOpen(true)}
          />
        </>
      )}

      <Dialog open={expanded} onOpenChange={setExpanded}>
        <DialogContent className="flex h-[92dvh] w-[94vw] max-h-[92dvh] max-w-none flex-col gap-0 overflow-hidden rounded-xl p-0">
          <div className="min-h-0 flex-1">
            <FlowWorkspace
              flowData={flowData}
              context={context}
              selectedNodeId={selectedNodeId}
              onSelectedNodeChange={setSelectedNodeId}
              onOpenJson={() => setJsonOpen(true)}
              expanded
            />
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={jsonOpen} onOpenChange={setJsonOpen}>
        <DialogContent className="flex max-h-[85dvh] max-w-4xl flex-col overflow-hidden">
          <DialogHeader>
            <DialogTitle>Flow JSON</DialogTitle>
            <DialogDescription>
              Read-only configuration for this flow.
            </DialogDescription>
          </DialogHeader>
          <pre className="min-h-0 overflow-auto rounded-md border bg-muted/30 p-4 text-xs leading-relaxed">
            {JSON.stringify(flowData, null, 2)}
          </pre>
        </DialogContent>
      </Dialog>
    </div>
  );
}
