import { useState } from "react";
import { ReactFlow } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { FlowContextStoreProvider, FlowContextType } from "@/lib/flow/context";
import { useHookedTheme } from "@/lib/hooks/theme";
import { FlowData, FlowNode } from "@/lib/types/flow.gen";
import { getBlockTitle } from "@/lib/marketplace";

// Read-only view of a listing's flow. Clicking a block shows its raw data, so
// things like request URLs can be checked before importing.
export default function MarketplaceFlowViewer({
  flow,
  context,
}: {
  flow: FlowData;
  context: FlowContextType;
}) {
  const { theme } = useHookedTheme();
  const [selected, setSelected] = useState<FlowNode | null>(null);

  return (
    <div className="space-y-3">
      <div className="h-80 rounded-md border bg-muted/30 nowheel">
        <FlowContextStoreProvider type={context}>
          <ReactFlow
            nodes={flow.nodes ?? []}
            edges={flow.edges ?? []}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            nodesConnectable={false}
            nodesDraggable={false}
            connectOnClick={false}
            deleteKeyCode={null}
            onNodeClick={(_, node) => setSelected(node as FlowNode)}
            onPaneClick={() => setSelected(null)}
            colorMode={theme === "dark" ? "dark" : "light"}
            className="!bg-transparent"
            proOptions={{
              hideAttribution: true,
            }}
            minZoom={0.1}
            fitView
          />
        </FlowContextStoreProvider>
      </div>
      {selected ? (
        <div className="rounded-md border p-3">
          <div className="text-sm font-medium mb-2">
            {getBlockTitle(selected.type ?? "")}
            <span className="text-muted-foreground font-normal ml-2">
              {selected.type}
            </span>
          </div>
          <pre className="text-xs font-mono whitespace-pre-wrap break-all max-h-60 overflow-y-auto">
            {JSON.stringify(selected.data, null, 2)}
          </pre>
        </div>
      ) : (
        <div className="text-xs text-muted-foreground">
          Click a block to see its settings.
        </div>
      )}
    </div>
  );
}
