import {
  Controls,
  Edge,
  Node,
  ReactFlow,
  ReactFlowProvider,
} from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import { MousePointerClickIcon } from "lucide-react";
import { useState } from "react";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import {
  FlowContextStoreProvider,
  FlowContextType,
  FlowReadOnlyContext,
} from "@/lib/flow/context";
import { isNoteNode } from "@/lib/flow/notes";
import { useHookedTheme } from "@/lib/hooks/theme";
import FlowNodeEditor from "../flow/FlowNodeEditor";
import { Dialog, DialogContent, DialogTitle } from "../ui/dialog";
import FitView from "./TemplateFlowFitView";

// TemplateFlowDialog shows a flow of a template on the whole screen. Clicking
// a block shows its settings next to it, like in the editor, but they can't
// be changed.
export default function TemplateFlowDialog({
  name,
  nodes,
  edges,
  context,
  selectedNodeId,
  open,
  onOpenChange,
}: {
  name: string;
  nodes: Node[];
  edges: Edge[];
  context: FlowContextType;
  // The block to show first, like the one clicked in the small preview.
  selectedNodeId?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[95vw] w-[95vw] h-[90dvh] p-0 gap-0 flex flex-col overflow-hidden">
        <div className="px-5 py-4 border-b flex-none">
          <DialogTitle>{name}</DialogTitle>
        </div>
        {open && (
          <ReactFlowProvider>
            <FlowContextStoreProvider type={context}>
              <FlowReadOnlyContext.Provider value={true}>
                <FlowView
                  nodes={nodes}
                  edges={edges}
                  selectedNodeId={selectedNodeId}
                />
              </FlowReadOnlyContext.Provider>
            </FlowContextStoreProvider>
          </ReactFlowProvider>
        )}
      </DialogContent>
    </Dialog>
  );
}

function FlowView({
  nodes,
  edges,
  selectedNodeId,
}: {
  nodes: Node[];
  edges: Edge[];
  selectedNodeId?: string;
}) {
  const { theme } = useHookedTheme();
  const [selected, setSelected] = useState(selectedNodeId);
  const [initialNodes] = useState(() =>
    nodes.map((n) => ({ ...n, selected: n.id === selectedNodeId }))
  );

  return (
    <div className="flex-1 min-h-0 flex">
      <div className="flex-1 min-w-0 bg-muted/30">
        <ReactFlow
          defaultNodes={initialNodes}
          defaultEdges={edges}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          nodesConnectable={false}
          nodesDraggable={false}
          connectOnClick={false}
          deleteKeyCode={null}
          multiSelectionKeyCode={null}
          selectionKeyCode={null}
          onSelectionChange={({ nodes }) => {
            const node = nodes.find((n) => !isNoteNode(n));
            setSelected(node?.id);
          }}
          colorMode={theme === "dark" ? "dark" : "light"}
          className="!bg-transparent"
          proOptions={{ hideAttribution: true }}
          minZoom={0.1}
          fitView
          fitViewOptions={{ maxZoom: 1 }}
        >
          <FitView />
          <Controls
            showInteractive={false}
            position="bottom-right"
            className="rounded-md border overflow-hidden [&_button]:bg-background [&_button]:text-foreground [&_button:hover]:bg-muted [&_button+button]:border-t"
          />
        </ReactFlow>
      </div>
      {/* The panel is always there, so the flow doesn't move when a block is
      selected. */}
      {selected ? (
        <FlowNodeEditor
          nodeId={selected}
          className="relative w-80 md:w-96 flex-none border-l h-full"
        />
      ) : (
        <div className="w-80 md:w-96 flex-none border-l h-full flex flex-col items-center justify-center gap-3 p-8 text-center">
          <MousePointerClickIcon className="h-8 w-8 text-muted-foreground" />
          <div className="text-sm text-muted-foreground">
            Click on a block to see how it&apos;s set up.
          </div>
        </div>
      )}
    </div>
  );
}
