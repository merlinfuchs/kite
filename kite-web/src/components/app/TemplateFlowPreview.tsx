import {
  Controls,
  Edge,
  Node,
  ReactFlow,
  useNodesInitialized,
  useReactFlow,
  useStore,
} from "@xyflow/react";
import { useEffect } from "react";
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

// FitView fits the flow in once its blocks are measured and the preview has
// a size, which it may not have yet when the item it's in opens.
function FitView() {
  const { fitView } = useReactFlow();
  const initialized = useNodesInitialized();
  const width = useStore((s) => s.width);
  const height = useStore((s) => s.height);

  useEffect(() => {
    if (initialized && width && height) fitView({ maxZoom: 1 });
  }, [initialized, width, height, fitView]);

  return null;
}
