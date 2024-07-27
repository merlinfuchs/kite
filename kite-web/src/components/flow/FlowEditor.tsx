import React, { DragEvent, useCallback } from "react";
import {
  addEdge,
  Background,
  BackgroundVariant,
  Connection,
  Controls,
  Edge,
  EdgeChange,
  Node,
  NodeChange,
  OnSelectionChangeFunc,
  ReactFlow,
  useEdgesState,
  useNodesState,
  useReactFlow,
} from "@xyflow/react";

import "@xyflow/react/dist/base.css";
import { useHookedTheme } from "@/lib/hooks/theme";
import { FlowData } from "@/lib/flow/data";
import { createNode, getNodeValues } from "@/lib/flow/nodes";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";

interface Props {
  initialData?: FlowData;
  onChange: () => void;
  onSelectionChange?: OnSelectionChangeFunc;
}

export default function FlowEditor({
  initialData,
  onChange,
  onSelectionChange,
}: Props) {
  const { theme } = useHookedTheme();

  // TODO: refactor?
  const [nodes, setNodes, onNodesChange] = useNodesState(
    initialData?.nodes || []
  );
  const [edges, setEdges, onEdgesChange] = useEdgesState(
    initialData?.edges || []
  );
  const { getNodes, getEdges, getEdge, getNode, screenToFlowPosition } =
    useReactFlow();

  const onConnect = useCallback(
    (con: Connection) => setEdges((eds) => addEdge(con, eds)),
    [setEdges]
  );

  const wrappedOnNodesChange = useCallback(
    (changes: NodeChange[]) => {
      const filteredChanges = changes.filter((change) => {
        if (change.type === "remove") {
          const node = getNode(change.id);
          const values = getNodeValues(node!.type!);
          return !values.fixed;
        }

        return true;
      });

      if (filteredChanges.length > 0) {
        onNodesChange(filteredChanges);
        onChange();
      }
    },
    [onNodesChange, onChange]
  );

  const wrappedOnEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
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
    },
    [getEdge, onEdgesChange, onChange]
  );

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

      const position = screenToFlowPosition({
        x: e.clientX,
        y: e.clientY,
      });
      const [newNodes, newEdges] = createNode(type, position);

      setNodes((nds) => nds.concat(newNodes));
      setEdges((eds) => eds.concat(newEdges));
    },
    [screenToFlowPosition, setNodes, setEdges]
  );

  const isValidConnection = useCallback(
    (con: Connection | Edge) => {
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

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={wrappedOnNodesChange}
      onEdgesChange={wrappedOnEdgesChange}
      onNodesDelete={onNodesDelete}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      onDrop={onDrop}
      onDragOver={onDragOver}
      onConnect={onConnect}
      isValidConnection={isValidConnection}
      onSelectionChange={onSelectionChange}
      colorMode={theme === "dark" ? "dark" : "light"}
      defaultEdgeOptions={{ type: "delete_button" }}
      className="!bg-background flex-auto"
      fitView
    >
      <Controls showInteractive={false} />
      <Background
        variant={BackgroundVariant.Dots}
        gap={18}
        size={1}
        className="!bg-muted/20"
        color="#615d84"
      />
    </ReactFlow>
  );
}
