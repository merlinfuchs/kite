import { cn } from "@/lib/utils";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { ReactFlow } from "@xyflow/react";
import { useHookedTheme } from "@/lib/hooks/theme";
import "@xyflow/react/dist/base.css";
import { forwardRef } from "react";

const initialNodes = [
  {
    id: "1",
    position: { x: 0, y: 0 },
    data: {},
    type: "entry_component_button",
  },
];

const FlowPreview = forwardRef<
  HTMLDivElement,
  {
    className?: string;
    onClick: () => void;
  }
>(({ className, onClick }, ref) => {
  const { theme } = useHookedTheme();

  return (
    <div
      className={cn("bg-muted/50 rounded-sm nowheel cursor-pointer", className)}
      onClick={onClick}
      role="button"
      ref={ref}
    >
      <ReactFlow
        nodes={initialNodes}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        elementsSelectable={false}
        nodesConnectable={false}
        nodesDraggable={false}
        connectOnClick={false}
        draggable={false}
        panOnDrag={false}
        zoomOnScroll={false}
        zoomOnPinch={false}
        colorMode={theme === "dark" ? "dark" : "light"}
        className="!bg-transparent hover:animate-shake"
        proOptions={{
          hideAttribution: true,
        }}
        fitView
      />
    </div>
  );
});

FlowPreview.displayName = "FlowPreview";
export default FlowPreview;
