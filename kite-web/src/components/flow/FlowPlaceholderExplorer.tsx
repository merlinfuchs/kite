import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { getAvailablePlaceholders } from "@/lib/flow/placeholders";
import { Node, useEdges, useNodes } from "@xyflow/react";
import { VariableIcon } from "lucide-react";
import { useMemo } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";

export default function FlowPlaceholderExplorer({
  onSelect,
  hideBrackets,
}: {
  onSelect: (value: string) => void;
  hideBrackets?: boolean;
}) {
  const nodes = useNodes<Node<NodeData>>();
  const edges = useEdges();
  const contextType = useFlowContext((c) => c.type);

  // TODO: only compute when explorer is open
  const placeholders = useMemo(() => {
    const selected = nodes.find((n) => n.selected);
    return getAvailablePlaceholders(selected?.id, nodes, edges, contextType);
  }, [nodes, edges, contextType]);

  return (
    <div className="absolute top-1.5 right-1.5 z-20">
      <PlaceholderExplorer
        onSelect={onSelect}
        placeholders={placeholders}
        hideBrackets={hideBrackets}
      >
        <VariableIcon
          className="h-5.5 w-5.5 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
        />
      </PlaceholderExplorer>
    </div>
  );
}
