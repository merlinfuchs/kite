import { useCallback, DragEvent } from "react";
import ReactFlow, {
  addEdge,
  Connection,
  Controls,
  Background,
  BackgroundVariant,
  useReactFlow,
  useNodesState,
  useEdgesState,
  NodeChange,
  EdgeChange,
  Edge,
  Node,
} from "reactflow";
import FlowEdgeDeleteButton from "./FlowEdgeDeleteButton";
import FlowEdgeFixed from "./FlowEdgeFixed";
import FlowNodeActionBase from "./FlowNodeActionBase";
import FlowNodeConditionChildCompare from "./FlowNodeConditionItemCompare";
import FlowNodeConditionCompare from "./FlowNodeConditionCompare";
import FlowNodeOptionBase from "./FlowNodeOptionBase";
import FlowNodeEntryCommand from "./FlowNodeEntryCommand";
import FlowNodeEntryEvent from "./FlowNodeEntryEvent";
import FlowNodeEntryError from "./FlowNodeEntryError";

import "reactflow/dist/base.css";
import { FlowData } from "../../lib/flow/data";
import { createNode, getNodeValues } from "@/lib/flow/nodes";
import FlowNodeConditionItemElse from "./FlowNodeConditionItemElse";

const nodeTypes = {
  action_response_text: FlowNodeActionBase,
  action_message_create: FlowNodeActionBase,
  action_log: FlowNodeActionBase,
  entry_command: FlowNodeEntryCommand,
  entry_event: FlowNodeEntryEvent,
  entry_error: FlowNodeEntryError,
  condition_compare: FlowNodeConditionCompare,
  condition_item_compare: FlowNodeConditionChildCompare,
  condition_item_else: FlowNodeConditionItemElse,
  option_text: FlowNodeOptionBase,
  option_number: FlowNodeOptionBase,
  option_user: FlowNodeOptionBase,
  option_channel: FlowNodeOptionBase,
  option_role: FlowNodeOptionBase,
  option_attachment: FlowNodeOptionBase,
};

const edgeTypes = {
  delete_button: FlowEdgeDeleteButton,
  fixed: FlowEdgeFixed,
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

  const { getNodes, getEdges, getEdge, getNode } = useReactFlow();

  const isValidConnection = useCallback(
    (con: Connection) => {
      if (!con.source || !con.target) return false;

      const source = getNode(con.source)!;
      const target = getNode(con.target)!;

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
      const [newNodes, newEdges] = createNode(type, position);

      setNodes((nds) => nds.concat(newNodes));
      setEdges((eds) => eds.concat(newEdges));
    },
    [rfInstance]
  );

  function wrappedOnNodesChange(changes: NodeChange[]) {
    onNodesChange(changes);
    onChange();
  }

  function wrappedOnEdgesChange(changes: EdgeChange[]) {
    const filteredChanges = changes.filter((change) => {
      if (change.type === "remove") {
        const edge = getEdge(change.id);
        return edge?.type !== "fixed";
      }

      return true;
    });

    if (filteredChanges.length > 0) {
      onEdgesChange(filteredChanges);
      onChange();
    }
  }

  function onNodesDelete(deletedNodes: Node[]) {
    for (const node of deletedNodes) {
      const nodeValues = getNodeValues(node.type!);

      // delete children if this node owns them
      if (nodeValues.ownsChildren) {
        const childIds = edges
          .filter((edge) => edge.source === node.id)
          .map((edge) => edge.target);

        setEdges((edges) => edges.filter((edge) => edge.source !== node.id));
        setNodes((nodes) =>
          nodes.filter((n) => n.id !== node.id && !childIds.includes(n.id))
        );
      }
    }
  }

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={wrappedOnNodesChange}
      onEdgesChange={wrappedOnEdgesChange}
      onNodesDelete={onNodesDelete}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      onConnect={onConnect}
      onDrop={onDrop}
      onDragOver={onDragOver}
      fitView
      className="bg-dark-4"
      defaultEdgeOptions={{ type: "delete_button" }}
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
