import { ExclamationTriangleIcon, LinkIcon } from "@heroicons/react/16/solid";
import { useMemo } from "react";
import { useEdges } from "reactflow";

interface Props {
  nodeId: string;
  showIsConnected?: boolean;
}

export default function FlowNodeMarkers({ nodeId, showIsConnected }: Props) {
  const edges = useEdges();

  const isConnected = useMemo(
    () => edges.some((edge) => edge.target === nodeId),
    [edges]
  );

  const hasError = true;

  return (
    <div className="absolute -top-2 -right-2 flex space-x-1">
      {!isConnected && showIsConnected && (
        <div className="h-4 w-4 bg-red-500 rounded-full flex items-center justify-center">
          <LinkIcon className="h-3 w-3 text-white" />
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
