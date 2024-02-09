import {
  Node,
  ReactFlowProvider,
  useOnSelectionChange,
  useReactFlow,
} from "reactflow";
import FlowEditor from "./FlowEditor";
import FlowNodeExplorer from "./FlowNodeExplorer";
import FlowNodeEditor from "./FlowNodeEditor";
import { NodeData } from "../../lib/flow/data";
import { useEffect, useState } from "react";

const initialNodes: Node<NodeData>[] = [
  {
    id: "1",
    type: "entry_command",
    position: { x: 0, y: 200 },
    data: {
      name: "ban",
      description: "Ban a user from the server",
    },
  },
  {
    id: "2",
    type: "action_response_text",
    position: { x: 100, y: 600 },
    data: {},
  },
  {
    id: "3",
    type: "option_user",
    position: { x: -50, y: 0 },
    data: {
      name: "user",
      description: "The user you want to ban",
    },
  },
  {
    id: "4",
    type: "condition",
    position: { x: -100, y: 400 },
    data: {},
  },
  {
    id: "5",
    type: "entry_error",
    position: { x: 300, y: 50 },
    data: {},
  },
  {
    id: "6",
    type: "action_log",
    position: { x: 350, y: 325 },
    data: {},
  },
];
const initialEdges = [
  { id: "e1-4", source: "1", target: "4" },
  { id: "e4-2", source: "4", target: "2" },
  { id: "e3-1", source: "3", target: "1" },
  { id: "e5-6", source: "5", target: "6" },
];

function Flow() {
  const { getEdges, getNodes } = useReactFlow();

  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  useOnSelectionChange({
    onChange: ({ nodes }) => {
      if (nodes.length === 1) setSelectedNodeId(nodes[0].id);
      else setSelectedNodeId(null);
    },
  });

  function onSave() {
    console.log(getEdges());
    console.log(getNodes());
  }

  useEffect(() => {
    async function onKeyDown(e: KeyboardEvent) {
      if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        onSave();
      }
    }

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onSave]);

  return (
    <div className="h-[100dvh] w-[100dvw] flex">
      <div className="flex-none">
        <FlowNodeExplorer />
        {selectedNodeId && <FlowNodeEditor nodeId={selectedNodeId} />}
      </div>
      <div className="flex-auto h-full">
        <FlowEditor initialNodes={initialNodes} initialEdges={initialEdges} />
      </div>
    </div>
  );
}

export default () => (
  <ReactFlowProvider>
    <Flow />
  </ReactFlowProvider>
);
