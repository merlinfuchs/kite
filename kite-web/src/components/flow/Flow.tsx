import { FlowContextStoreProvider, FlowContextType } from "@/lib/flow/context";
import { FlowData } from "@/lib/flow/dataSchema";
import { OnSelectionChangeParams } from "@xyflow/react";
import { useCallback, useState } from "react";
import FlowEditor from "./FlowEditor";
import FlowMenu from "./FlowMenu";
import { LogEntry } from "@/lib/types/wire.gen";

interface Props {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  onChange: () => void;
}

export default function Flow({ flowData, logs, context, onChange }: Props) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  const onSelectionChange = useCallback(
    ({ nodes }: OnSelectionChangeParams) => {
      if (nodes.length === 1) {
        setSelectedNodeId(nodes[0].id);
      } else {
        setSelectedNodeId(null);
      }
    },
    []
  );

  return (
    <FlowContextStoreProvider type={context}>
      <div className="flex flex-auto overflow-y-hidden relative">
        <FlowMenu selectedNodeId={selectedNodeId} logs={logs} />

        <div className="flex-auto">
          <FlowEditor
            initialData={flowData}
            onChange={onChange}
            onSelectionChange={onSelectionChange}
          />
        </div>
      </div>
    </FlowContextStoreProvider>
  );
}
