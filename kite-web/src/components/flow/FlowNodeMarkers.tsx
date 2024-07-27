import { NodeData } from "@/lib/flow/data";
import { LinkIcon } from "lucide-react";
import { useMemo } from "react";
import { useEdges } from "@xyflow/react";
import { useNodeValues } from "@/lib/flow/nodes";

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

  return (
    <div className="absolute -top-2 -right-2 flex space-x-1">
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
