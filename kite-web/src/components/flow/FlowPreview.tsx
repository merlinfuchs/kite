import { cn } from "@/lib/utils";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { ReactFlow } from "@xyflow/react";
import { useHookedTheme } from "@/lib/hooks/theme";
import "@xyflow/react/dist/base.css";

const initialNodes = [
  {
    id: "1",
    position: { x: 0, y: 0 },
    data: { name: "ban", description: "Ban a user" },
    type: "entry_command",
  },
];

export default function FlowPreview({
  className,
  onClick,
}: {
  className?: string;
  onClick: () => void;
}) {
  const { theme } = useHookedTheme();

  return (
    <div
      className={cn("bg-muted/50 rounded-sm nowheel cursor-pointer", className)}
      onClick={onClick}
      role="button"
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
}
