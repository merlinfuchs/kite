import { FlowData } from "@/lib/flow/data";
import FlowEditor from "./FlowEditor";
import FlowNodeEditor from "./FlowNodeEditor";
import FlowNodeExplorer from "./FlowNodeExplorer";
import { OnSelectionChangeParams } from "@xyflow/react";
import { useCallback, useState } from "react";

interface Props {
  flowData: FlowData;
  onChange: () => void;
}

export default function Flow({ flowData, onChange }: Props) {
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
  );
}
