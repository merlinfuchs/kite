import { useCallback, useEffect, useMemo, useState } from "react";
import ReactFlow, {
  useEdgesState,
  useNodesState,
  addEdge,
  Connection,
  Controls,
  Background,
  BackgroundVariant,
  Node,
  useReactFlow,
  getOutgoers,
  ReactFlowProvider,
  ReactFlowInstance,
} from "reactflow";
import FlowEdgeButton from "./FlowEdgeButton";
import FlowNodeActionBase from "./FlowNodeActionBase";
import FlowNodeConditionBase from "./FlowNodeConditionBase";
import FlowNodeOptionBase from "./FlowNodeOptionBase";
import FlowNodeCommand from "./FlowNodeCommand";

import "reactflow/dist/base.css";
import { NodeData } from "./types";

const initialNodes: Node<NodeData>[] = [
  {
    id: "1",
    type: "command",
    position: { x: 0, y: 200 },
    data: {},
  },
  {
    id: "2",
    type: "action",
    position: { x: 100, y: 600 },
    data: {},
  },
  {
    id: "3",
    type: "option",
    position: { x: 300, y: 0 },
    data: {},
  },
  {
    id: "4",
    type: "condition",
    position: { x: -100, y: 400 },
    data: {},
  },
];
const initialEdges = [
  { id: "e1-4", source: "1", target: "4" },
  { id: "e4-2", source: "4", target: "2" },
  { id: "e3-1", source: "3", target: "1" },
];

const nodeTypes = {
  action: FlowNodeActionBase,
  command: FlowNodeCommand,
  condition: FlowNodeConditionBase,
  option: FlowNodeOptionBase,
};

const edgeTypes = {
  buttonedge: FlowEdgeButton,
};

function FlowEditor() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [rfInstance, setRfInstance] = useState<ReactFlowInstance | null>(null);
  const { setViewport } = useReactFlow();

  const onConnect = useCallback(
    (con: Connection) => setEdges((eds) => addEdge(con, eds)),
    [setEdges]
  );

  /* const selectedNodeId = useMemo(
    () => nodes.find((n) => n.selected)?.id,
    [nodes]
  );

  useEffect(() => {
    console.log("selected node", selectedNodeId);
  }, [selectedNodeId]); */

  const { getNodes, getEdges } = useReactFlow();

  const isValidConnection = useCallback(
    (con: Connection) => {
      const nodes = getNodes();
      const edges = getEdges();

      const source = nodes.find((node) => node.id === con.source)!;
      const target = nodes.find((node) => node.id === con.target)!;

      // TODO: This is a bit of a mess, but it works for now
      if (target.type === "command" && source.type !== "option") return false;
      if (source.type === "option" && target.type !== "command") return false;

      // Prevent cycles
      const hasCycle = (node: Node, visited = new Set()) => {
        if (visited.has(node.id)) return false;

        visited.add(node.id);

        for (const outgoer of getOutgoers(node, nodes, edges)) {
          if (outgoer.id === con.source) return true;
          if (hasCycle(outgoer, visited)) return true;
        }
      };

      if (target.id === con.source) return false;
      return !hasCycle(target);
    },
    [getNodes, getEdges]
  );

  function save() {
    if (rfInstance) {
      const flow = rfInstance.toObject();
      localStorage.setItem("flow", JSON.stringify(flow));
    }
  }

  function restore() {
    const flow = localStorage.getItem("flow");
    if (flow) {
      const flowObj = JSON.parse(flow);
      const { x = 0, y = 0, zoom = 1 } = flowObj.viewport;
      setNodes(flowObj.nodes);
      setEdges(flowObj.edges);
      setViewport({ x, y, zoom });
    }
  }

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      onConnect={onConnect}
      onInit={setRfInstance}
      fitView
      className="bg-dark-4"
      defaultEdgeOptions={{ type: "buttonedge" }}
      isValidConnection={isValidConnection}
    >
      <Controls showInteractive={false} />
      <Background
        variant={BackgroundVariant.Dots}
        gap={12}
        size={1}
        color="#615d84"
      />
    </ReactFlow>
  );
}

export default () => (
  <ReactFlowProvider>
    <FlowEditor />
  </ReactFlowProvider>
);
