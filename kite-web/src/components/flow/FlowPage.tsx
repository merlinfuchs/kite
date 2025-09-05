import { FlowData, NodeType } from "@/lib/flow/dataSchema";
import FlowNav from "./FlowNav";
import { ReactFlowProvider, useReactFlow } from "@xyflow/react";
import { useCallback } from "react";
import Flow from "./Flow";
import { FlowContextType } from "@/lib/flow/context";
import { LogEntry } from "@/lib/types/wire.gen";

interface Props {
  flowData: FlowData;
  logs?: LogEntry[];
  context: FlowContextType;
  hasUnsavedChanges: boolean;
  onChange: () => void;
  isSaving: boolean;
  onSave: (data: FlowData) => void;
  onExit: () => void;
}

function InnerFlowPage({
  flowData,
  logs,
  context,
  hasUnsavedChanges,
  onChange,
  isSaving,
  onSave,
  onExit,
}: Props) {
  const { getNodes, getEdges } = useReactFlow<NodeType>();

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
      <Flow
        flowData={flowData}
        logs={logs}
        context={context}
        onChange={onChange}
      />
    </div>
  );
}

export default function FlowPage(props: Props) {
  return (
    <ReactFlowProvider>
      <InnerFlowPage {...props} />
    </ReactFlowProvider>
  );
}
