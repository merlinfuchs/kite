import {
  addEdge,
  Background,
  BackgroundVariant,
  Connection,
  ControlButton,
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
import { DragEvent, MouseEvent, useCallback, useEffect, useRef } from "react";

import {
  copyFlowNodes,
  parseFlowClipboard,
  pasteFlowNodes,
} from "@/lib/flow/clipboard";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { useFlowContext } from "@/lib/flow/context";
import { FlowData } from "@/lib/flow/dataSchema";
import { getLayoutedElements } from "@/lib/flow/layout";
import { createNode, getNodeValues } from "@/lib/flow/nodes";
import { useHookedTheme } from "@/lib/hooks/theme";
import "@xyflow/react/dist/base.css";
import { ListTreeIcon } from "lucide-react";
import { toast } from "sonner";
import { isNodeTypeAvailable } from "./FlowNodeExplorer";

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
  const { getEdge, getNode, screenToFlowPosition, fitView } = useReactFlow();

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
    [onNodesChange, onChange, getNode]
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

  const onNodesDelete = useCallback(
    (deletedNodes: Node[]) => {
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
    },
    [edges, setEdges, setNodes]
  );

  const format = useCallback(() => {
    const formattedNodes = getLayoutedElements(nodes, edges, {
      direction: "TB",
    });

    setNodes(formattedNodes.nodes);
    setTimeout(() => {
      fitView();
    }, 50);
  }, [nodes, edges, setNodes, fitView]);

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

  const contextType = useFlowContext((c) => c.type);
  const mousePosition = useRef<{ x: number; y: number } | null>(null);

  const onMouseMove = useCallback((e: MouseEvent) => {
    mousePosition.current = { x: e.clientX, y: e.clientY };
  }, []);

  // Uses the native copy and paste events so the blocks go through the system
  // clipboard and can be pasted into other commands and event listeners.
  useEffect(() => {
    const shouldIgnore = (e: ClipboardEvent) => {
      const target = e.target as HTMLElement | null;
      return (
        !!target?.closest(
          "input, textarea, [contenteditable], [role=dialog]"
        ) || !!window.getSelection()?.toString()
      );
    };

    const onCopy = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      const clipboard = copyFlowNodes(nodes, edges);
      if (!clipboard) return;

      e.preventDefault();
      e.clipboardData?.setData("text/plain", JSON.stringify(clipboard));
    };

    const onPaste = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      const clipboard = parseFlowClipboard(
        e.clipboardData?.getData("text/plain") || ""
      );
      if (!clipboard) return;

      e.preventDefault();

      const position = mousePosition.current
        ? screenToFlowPosition(mousePosition.current)
        : {
            x: Math.min(...clipboard.nodes.map((n) => n.position.x)) + 50,
            y: Math.min(...clipboard.nodes.map((n) => n.position.y)) + 50,
          };

      const [newNodes, newEdges] = pasteFlowNodes(clipboard, position, (type) =>
        isNodeTypeAvailable(type, contextType)
      );
      if (newNodes.length < clipboard.nodes.length) {
        toast.warning("Some blocks aren't available here and weren't pasted.");
      }
      if (newNodes.length === 0) return;

      setNodes((nds) => [
        ...nds.map((n) => ({ ...n, selected: false })),
        ...newNodes,
      ]);
      setEdges((eds) => [
        ...eds.map((e) => ({ ...e, selected: false })),
        ...newEdges,
      ]);
      onChange();
    };

    document.addEventListener("copy", onCopy);
    document.addEventListener("paste", onPaste);
    return () => {
      document.removeEventListener("copy", onCopy);
      document.removeEventListener("paste", onPaste);
    };
  }, [
    nodes,
    edges,
    contextType,
    screenToFlowPosition,
    setNodes,
    setEdges,
    onChange,
  ]);

  const isValidConnection = useCallback(
    (con: Connection | Edge) => {
      if (!con.source || !con.target) return false;

      const source = getNode(con.source)!;
      const target = getNode(con.target)!;

      // This is a bit of a mess, but it works for now
      if (
        (target.type === "entry_command" || target.type === "entry_event") &&
        !source.type?.startsWith("option")
      )
        return false;
      if (
        source.type?.startsWith("option") &&
        target.type !== "entry_command" &&
        target.type !== "entry_event"
      )
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
    [getNode]
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
      onMouseMove={onMouseMove}
      onConnect={onConnect}
      isValidConnection={isValidConnection}
      onSelectionChange={onSelectionChange}
      colorMode={theme === "dark" ? "dark" : "light"}
      defaultEdgeOptions={{ type: "delete_button" }}
      multiSelectionKeyCode={null}
      deleteKeyCode={["Backspace", "Delete"]}
      proOptions={{
        hideAttribution: true,
      }}
      className="!bg-background flex-auto"
      fitView
    >
      <Controls
        showInteractive={false}
        position="bottom-right"
        className="scale-110"
      >
        <ControlButton onClick={format}>
          <ListTreeIcon className="size-5" />
        </ControlButton>
      </Controls>
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
