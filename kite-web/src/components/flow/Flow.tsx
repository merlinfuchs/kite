import { FlowData, NodeType } from "@/lib/flow/data";
import FlowEditor from "./FlowEditor";
import FlowNav from "./FlowNav";
import FlowNodeEditor from "./FlowNodeEditor";
import FlowNodeExplorer from "./FlowNodeExplorer";
import {
  OnSelectionChangeParams,
  ReactFlowProvider,
  useReactFlow,
} from "@xyflow/react";
import { useCallback, useState } from "react";

interface Props {
  flowData: FlowData;
  hasUnsavedChanges: boolean;
  onChange: () => void;
  isSaving: boolean;
  onSave: (data: FlowData) => void;
  onExit: () => void;
}

function InnerFlow({
  flowData,
  hasUnsavedChanges,
  onChange,
  isSaving,
  onSave,
  onExit,
}: Props) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  const { getNodes, getEdges } = useReactFlow<NodeType>();

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

  const save = useCallback(() => {
    onSave({
      nodes: getNodes(),
      edges: getEdges(),
    });
  }, [getNodes, getEdges, onSave]);

  return (
    <div className="h-[100dvh] w-[100dvw] flex flex-col">
      <div className="flex-none">
        <FlowNav
          hasUnsavedChanges={hasUnsavedChanges}
          isSaving={isSaving}
          onSave={save}
          onExit={onExit}
        />
      </div>
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
    </div>
  );
}

export default function Flow(props: Props) {
  return (
    <ReactFlowProvider>
      <InnerFlow {...props} />
    </ReactFlowProvider>
  );
}
