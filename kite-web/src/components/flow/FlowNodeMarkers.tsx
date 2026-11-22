import { NodeData } from "@/lib/flow/dataSchema";
import {
  formatMultiRunWarning,
  getCachedMultiRunCounts,
} from "@/lib/flow/multiRun";
import { useNodeValues } from "@/lib/flow/nodes";
import { Node, useEdges, useStore } from "@xyflow/react";
import { LinkIcon } from "lucide-react";
import { useMemo } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipPortal,
  TooltipTrigger,
} from "../ui/tooltip";

interface Props {
  id: string;
  type: string;
  data: NodeData;
  showConnectedMarker?: boolean;
}

export default function FlowNodeMarkers({
  id,
  type,
  data,
  showConnectedMarker = true,
}: Props) {
  const edges = useEdges();

  const isConnected = useMemo(() => {
    if (!showConnectedMarker) return true;
    return edges.some((edge) => edge.target === id);
  }, [edges, id, showConnectedMarker]);

  const { dataSchema } = useNodeValues(type);

  const hasError = useMemo(() => {
    if (!dataSchema) return false;

    try {
      dataSchema.parse(data);
      return false;
    } catch {
      return true;
    }
  }, [dataSchema, data]);

  const multiRunCount = useStore(
    (s) =>
      getCachedMultiRunCounts(s.nodes as Node<NodeData>[], s.edges).get(id) ?? 0
  );

  return (
    <div className="absolute -top-2 -right-2 flex space-x-1">
      {multiRunCount > 1 && (
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              className="nodrag h-4 min-w-4 px-1 bg-amber-500 rounded-full flex items-center justify-center text-amber-950 text-[10px] font-semibold leading-none tabular-nums cursor-help focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              aria-label={`Can run up to ${multiRunCount} times per run`}
            >
              {multiRunCount}×
            </button>
          </TooltipTrigger>
          <TooltipPortal>
            <TooltipContent className="max-w-xs text-xs">
              {formatMultiRunWarning(multiRunCount)}
            </TooltipContent>
          </TooltipPortal>
        </Tooltip>
      )}
      {!isConnected && (
        <div className="h-4 w-4 bg-red-500 rounded-full flex items-center justify-center">
          <LinkIcon className="h-2.5 w-2.5 text-white" />
        </div>
      )}
      {hasError && (
        <div className="h-4 w-4 bg-red-500 rounded-full flex items-center justify-center text-white text-xs font-medium leading-4">
          !
        </div>
      )}
    </div>
  );
}
