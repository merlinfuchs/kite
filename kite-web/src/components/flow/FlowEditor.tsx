import { useCallback, DragEvent } from "react";
import ReactFlow, {
  addEdge,
  Connection,
  Controls,
  Background,
  BackgroundVariant,
  Node,
  useReactFlow,
  getOutgoers,
  Edge,
  useNodesState,
  useEdgesState,
  NodeChange,
  EdgeChange,
} from "reactflow";
import FlowEdgeButton from "./FlowEdgeButton";
import FlowNodeActionBase from "./FlowNodeActionBase";
import FlowNodeConditionBase from "./FlowNodeConditionBase";
import FlowNodeOptionBase from "./FlowNodeOptionBase";
import FlowNodeEntryCommand from "./FlowNodeEntryCommand";
import FlowNodeEntryEvent from "./FlowNodeEntryEvent";
import FlowNodeEntryError from "./FlowNodeEntryError";

import "reactflow/dist/base.css";
import { FlowData, NodeData } from "../../lib/flow/data";
import { getId } from "@/lib/flow/util";

const nodeTypes = {
  action_response_text: FlowNodeActionBase,
  action_message_create: FlowNodeActionBase,
  action_log: FlowNodeActionBase,
  entry_command: FlowNodeEntryCommand,
  entry_event: FlowNodeEntryEvent,
  entry_error: FlowNodeEntryError,
  condition: FlowNodeConditionBase,
  option_text: FlowNodeOptionBase,
  option_number: FlowNodeOptionBase,
  option_user: FlowNodeOptionBase,
  option_channel: FlowNodeOptionBase,
  option_role: FlowNodeOptionBase,
  option_attachment: FlowNodeOptionBase,
};

const edgeTypes = {
  buttonedge: FlowEdgeButton,
};

interface Props {
  initialData?: FlowData;
  onChange: () => void;
}

export default function FlowEditor({ initialData, onChange }: Props) {
  const [nodes, setNodes, onNodesChange] = useNodesState(
    initialData?.nodes || []
  );
  const [edges, setEdges, onEdgesChange] = useEdgesState(
    initialData?.edges || []
  );
  const rfInstance = useReactFlow();

  const onConnect = useCallback(
    (con: Connection) => setEdges((eds) => addEdge(con, eds)),
    [setEdges]
  );

  const { getNodes, getEdges } = useReactFlow();

  const isValidConnection = useCallback(
    (con: Connection) => {
      const nodes = getNodes();
      const edges = getEdges();

      const source = nodes.find((node) => node.id === con.source)!;
      const target = nodes.find((node) => node.id === con.target)!;

      // TODO: This is a bit of a mess, but it works for now
      if (target.type === "entry_command" && !source.type?.startsWith("option"))
        return false;
      if (source.type?.startsWith("option") && target.type !== "entry_command")
        return false;

      // Prevent cycles
      /*const hasCycle = (node: Node, visited = new Set()) => {
        if (visited.has(node.id)) return false;

        visited.add(node.id);

        for (const outgoer of getOutgoers(node, nodes, edges)) {
          if (outgoer.id === con.source) return true;
          if (hasCycle(outgoer, visited)) return true;
        }
      };

      if (target.id === con.source) return false;
      return !hasCycle(target);*/
      return true;
    },
    [getNodes, getEdges]
  );

  const onDragOver = useCallback((e: DragEvent) => {
    e.preventDefault();
    e.dataTransfer!.dropEffect = "move";
  }, []);

  const onDrop = useCallback(
    (e: DragEvent) => {
      e.preventDefault();

      const type = e.dataTransfer?.getData("application/reactflow");
      if (!type) {
        return;
      }

      const position = rfInstance.screenToFlowPosition({
        x: e.clientX,
        y: e.clientY,
      });
      const newNode = {
        id: getId(),
        type,
        position,
        data: {},
      };

      setNodes((nds) => nds.concat(newNode));
    },
    [rfInstance]
  );

  function wrappedOnNodesChange(changes: NodeChange[]) {
    onNodesChange(changes);
    onChange();
  }

  function wrappedOnEdgesChange(changes: EdgeChange[]) {
    onEdgesChange(changes);
    onChange();
  }

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={wrappedOnNodesChange}
      onEdgesChange={wrappedOnEdgesChange}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      onConnect={onConnect}
      onDrop={onDrop}
      onDragOver={onDragOver}
      fitView
      className="bg-dark-4"
      defaultEdgeOptions={{ type: "buttonedge" }}
      isValidConnection={isValidConnection}
      multiSelectionKeyCode={null}
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
