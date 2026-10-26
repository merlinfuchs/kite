import { Controls, Edge, Node, ReactFlow } from "@xyflow/react";
import { useState } from "react";
import { Maximize2Icon } from "lucide-react";
import { isNoteNode } from "@/lib/flow/notes";
import { Button } from "../ui/button";
import dynamic from "next/dynamic";
import FitView from "./TemplateFlowFitView";

// The full view has the settings of blocks, which load with it.
const TemplateFlowDialog = dynamic(() => import("./TemplateFlowDialog"), {
  ssr: false,
});
import "@xyflow/react/dist/base.css";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import {
  FlowContextStoreProvider,
  FlowContextType,
  FlowReadOnlyContext,
} from "@/lib/flow/context";
import { useHookedTheme } from "@/lib/hooks/theme";

// Read-only view of a template's flow. It can be panned and zoomed with the
// controls, the scroll wheel keeps scrolling the page. Clicking a block, or
// the button in the corner, opens the flow on the whole screen with the
// settings of its blocks.
export default function TemplateFlowPreview({
  name,
  nodes,
  edges,
  context,
}: {
  name: string;
  nodes: Node[];
  edges: Edge[];
  context: FlowContextType;
}) {
  const { theme } = useHookedTheme();
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<string>();
  // The dialog only loads once it's opened.
  const [opened, setOpened] = useState(false);

  function openFull(nodeId?: string) {
    setSelected(nodeId);
    setOpen(true);
    setOpened(true);
  }

  return (
    <div className="h-80 rounded-md border bg-muted/30 relative">
      <Button
        variant="outline"
        size="sm"
        className="absolute top-2 right-2 z-10"
        onClick={() => openFull()}
      >
        <Maximize2Icon className="mr-2 h-4 w-4" />
        Full screen
      </Button>
      {opened && (
        <TemplateFlowDialog
          name={name}
          nodes={nodes}
          edges={edges}
          context={context}
          selectedNodeId={selected}
          open={open}
          onOpenChange={setOpen}
        />
      )}
      <FlowContextStoreProvider type={context}>
        <FlowReadOnlyContext.Provider value={true}>
          <ReactFlow
            defaultNodes={nodes}
            defaultEdges={edges}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            elementsSelectable={false}
            nodesConnectable={false}
            nodesDraggable={false}
            connectOnClick={false}
            deleteKeyCode={null}
            onNodeClick={(_, node) => !isNoteNode(node) && openFull(node.id)}
            zoomOnScroll={false}
            preventScrolling={false}
            colorMode={theme === "dark" ? "dark" : "light"}
            className="!bg-transparent"
            proOptions={{
              hideAttribution: true,
            }}
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
        </FlowReadOnlyContext.Provider>
      </FlowContextStoreProvider>
    </div>
  );
}
