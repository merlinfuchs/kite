import {
  ReactFlowProvider,
  useOnSelectionChange,
  useReactFlow,
} from "reactflow";
import FlowEditor from "./FlowEditor";
import FlowNodeExplorer from "./FlowNodeExplorer";
import FlowNodeEditor from "./FlowNodeEditor";
import FlowNav from "./FlowNav";
import { FlowData, NodeData } from "../../lib/flow/data";
import { useEffect, useMemo, useState } from "react";
import { FlatFile } from "@/lib/code/filetree";

interface Props {
  files: FlatFile[];
  openFilePath: string | null;
  setOpenFilePath: (path: string | null) => void;
  hasUnsavedChanges: boolean;
  onChange: () => void;
  isSaving: boolean;
  onSave: () => void;
  isDeploying: boolean;
  onDeploy: () => void;
  onExit: () => void;
}

function Flow({
  files,
  openFilePath,
  setOpenFilePath,
  hasUnsavedChanges,
  onChange,
  isSaving,
  onSave,
  isDeploying,
  onDeploy,
  onExit,
}: Props) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const { setNodes, setEdges } = useReactFlow<NodeData>();

  useEffect(() => {
    const file = files.find((f) => f.path === openFilePath);
    if (!file) {
      console.warn(`File ${openFilePath} not found.`);
      return;
    }

    const data: FlowData = JSON.parse(file.content);
    setNodes(data.nodes);
    setEdges(data.edges);
  }, [files]);

  useOnSelectionChange({
    onChange: ({ nodes }) => {
      if (nodes.length === 1) setSelectedNodeId(nodes[0].id);
      else setSelectedNodeId(null);
    },
  });

  function save(d: FlowData) {
    const file = files.find((f) => f.path === openFilePath);
    if (file) {
      // We intentionally don't trigger a state update here, the state is primarily managed by reactflow
      file.content = JSON.stringify(d);
    }
    onSave();
  }

  return (
    <div className="h-[100dvh] w-[100dvw] flex flex-col">
      <div className="flex-none">
        <FlowNav
          hasUnsavedChanges={hasUnsavedChanges}
          isSaving={isSaving}
          onSave={save}
          isDeploying={isDeploying}
          onDeploy={onDeploy}
          onExit={onExit}
        />
      </div>
      <div className="flex flex-auto overflow-y-hidden relative">
        <div className="flex-none">
          <FlowNodeExplorer />
          {selectedNodeId && <FlowNodeEditor nodeId={selectedNodeId} />}
        </div>
        <div className="flex-auto">
          <FlowEditor onChange={onChange} />
        </div>
      </div>
    </div>
  );
}

export default (props: Props) => (
  <ReactFlowProvider>
    <Flow {...props} />
  </ReactFlowProvider>
);
