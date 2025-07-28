import { FlowData } from "@/lib/flow/dataSchema";
import FlowEditor from "./FlowEditor";
import FlowNodeEditor from "./FlowNodeEditor";
import FlowNodeExplorer from "./FlowNodeExplorer";
import { OnSelectionChangeParams } from "@xyflow/react";
import { useCallback, useState } from "react";
import { FlowContextStoreProvider, FlowContextType } from "@/lib/flow/context";

interface Props {
  flowData: FlowData;
  context: FlowContextType;
  onChange: () => void;
}

export default function Flow({ flowData, context, onChange }: Props) {
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
        <div className="flex-none">
          <FlowNodeExplorer />
          {selectedNodeId && <FlowNodeEditor nodeId={selectedNodeId} />}
        </div>
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
