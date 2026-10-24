import { Controls, Edge, Node, ReactFlow } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import {
  FlowContextStoreProvider,
  FlowContextType,
  FlowReadOnlyContext,
} from "@/lib/flow/context";
import { useHookedTheme } from "@/lib/hooks/theme";

// Read-only view of a template's flow. It can be panned and zoomed with the
// controls, the scroll wheel keeps scrolling the page.
export default function TemplateFlowPreview({
  nodes,
  edges,
  context,
}: {
  nodes: Node[];
  edges: Edge[];
  context: FlowContextType;
}) {
  const { theme } = useHookedTheme();

  return (
    <div className="h-80 rounded-md border bg-muted/30">
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
