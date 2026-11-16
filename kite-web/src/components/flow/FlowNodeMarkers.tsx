import { NodeData } from "@/lib/flow/dataSchema";
import {
  AlertTriangleIcon,
  CheckIcon,
  LinkIcon,
  LucideIcon,
  MinusIcon,
  XIcon,
} from "lucide-react";
import { TestNodeStatus, useTestNodeStatus } from "@/lib/flow/context";
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

  const testStatus = useTestNodeStatus(id);
  const test = testStatus && testStatusMarkers[testStatus];

  return (
    <>
      {test && (
        <div
          className={`absolute -top-2 -left-2 h-5 w-5 rounded-full flex items-center justify-center text-white ${test.className}`}
          title={test.title}
        >
          <test.Icon className="h-3 w-3" />
        </div>
      )}
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
    </>
  );
}

const testStatusMarkers: Record<
  TestNodeStatus,
  { className: string; Icon: LucideIcon; title: string }
> = {
  ok: { className: "bg-green-600", Icon: CheckIcon, title: "Ran in the test" },
  skipped: {
    className: "bg-zinc-500",
    Icon: MinusIcon,
    title: "Didn't run in the test, it calls a service other than Discord",
  },
  warning: {
    className: "bg-amber-500",
    Icon: AlertTriangleIcon,
    title: "Ran in the test, Discord would likely reject what it does",
  },
  error: { className: "bg-red-500", Icon: XIcon, title: "Failed in the test" },
};
