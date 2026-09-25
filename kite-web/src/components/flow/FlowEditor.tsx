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
import { DragEvent, RefObject, useCallback, useEffect, useRef } from "react";

import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import { getFlowChangeKind, getFlowMergeKey } from "@/lib/flow/history";
import { getLayoutedElements } from "@/lib/flow/layout";
import {
  canConnect,
  createNode,
  getNodeValues,
  getDeletedNodeIds,
} from "@/lib/flow/nodes";
import { useFlowClipboard } from "@/lib/hooks/flowClipboard";
import { useFlowHistory } from "@/lib/hooks/flowHistory";
import { useHookedTheme } from "@/lib/hooks/theme";
import "@xyflow/react/dist/base.css";
import { ListTreeIcon, Redo2Icon, Undo2Icon } from "lucide-react";

interface Props {
  initialData?: FlowData;
  onChange: () => void;
  onSelectionChange?: OnSelectionChangeFunc;
  containerRef: RefObject<HTMLElement>;
}

export default function FlowEditor({
  initialData,
  onChange,
  onSelectionChange,
  containerRef,
}: Props) {
  const { theme } = useHookedTheme();

  // TODO: refactor?
  const [nodes, setNodes, onNodesChange] = useNodesState(
    initialData?.nodes || []
  );
  const [edges, setEdges, onEdgesChange] = useEdgesState(
    initialData?.edges || []
  );
  // Edits go through react-flow rather than the state setters above, so they
  // reach the change handlers below, which record them for undo.
  const {
    getEdge,
    getEdges,
    getNode,
    getNodes,
    screenToFlowPosition,
    fitView,
    setNodes: editNodes,
    setEdges: editEdges,
    addNodes,
    addEdges,
  } = useReactFlow<Node<NodeData>>();

  // onChange is called once the edit has been applied, so the nodes and edges
  // it reads through react-flow are up to date.
  const changed = useRef(false);
  const markChanged = useCallback(() => {
    changed.current = true;
  }, []);
  useEffect(() => {
    if (!changed.current) return;
    changed.current = false;
    onChange();
  }, [nodes, edges, onChange]);

  const { record, commit, undo, redo, canUndo, canRedo } = useFlowHistory({
    nodes,
    edges,
    setNodes,
    setEdges,
    onChange: markChanged,
    containerRef,
  });

  const onConnect = useCallback(
    (con: Connection) => editEdges((eds) => addEdge(con, eds)),
    [editEdges]
  );

  const isDragging = useRef(false);

  const wrappedOnNodesChange = useCallback(
    (changes: NodeChange<Node<NodeData>>[]) => {
      const filteredChanges = changes.filter((change) => {
        if (change.type === "remove") {
          const node = getNode(change.id);
          const values = getNodeValues(node!.type!);
          return !values.fixed;
        }

        return true;
      });

      if (filteredChanges.length === 0) return;

      if (filteredChanges.some((c) => getFlowChangeKind(c) === "drag")) {
        // The state from before the drag is recorded once, when it starts.
        if (!isDragging.current) record();
        isDragging.current = true;
      } else if (filteredChanges.some((c) => getFlowChangeKind(c) === "edit")) {
        if (isDragging.current) {
          // The drag has ended, which makes it an edit.
          isDragging.current = false;
          markChanged();
        } else {
          commit(getFlowMergeKey(filteredChanges));
        }
      }

      onNodesChange(filteredChanges);
    },
    [onNodesChange, getNode, record, commit, markChanged]
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

      if (filteredChanges.length === 0) return;

      if (filteredChanges.some((c) => getFlowChangeKind(c) === "edit")) {
        commit();
      }
      onEdgesChange(filteredChanges);
    },
    [getEdge, onEdgesChange, commit]
  );

  const onNodesDelete = useCallback(
    (deletedNodes: Node<NodeData>[]) => {
      // The change handlers keep fixed blocks, so the blocks owned by the
      // deleted ones, e.g. the else branch of a condition, are removed here.
      const nodes = [...deletedNodes, ...getNodes()];
      const types = new Map(nodes.map((n) => [n.id, n.type!]));
      const deletedIds = new Set(deletedNodes.map((n) => n.id));
      const removed = getDeletedNodeIds([...deletedIds], nodes, getEdges());

      // Nothing to do if the change handlers already removed everything.
      const handled = [...removed].every(
        (id) => deletedIds.has(id) && !getNodeValues(types.get(id)!).fixed
      );
      if (handled) return;

      commit();
      setEdges((edges) =>
        edges.filter((e) => !removed.has(e.source) && !removed.has(e.target))
      );
      setNodes((nodes) => nodes.filter((n) => !removed.has(n.id)));
    },
    [getNodes, getEdges, commit, setEdges, setNodes]
  );

  const format = useCallback(() => {
    const formattedNodes = getLayoutedElements(nodes, edges, {
      direction: "TB",
    });

    editNodes(formattedNodes.nodes);
    setTimeout(() => {
      fitView();
    }, 50);
  }, [nodes, edges, editNodes, fitView]);

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

      addNodes(newNodes);
      addEdges(newEdges);
    },
    [screenToFlowPosition, addNodes, addEdges]
  );

  const onMouseMove = useFlowClipboard();

  const isValidConnection = useCallback(
    (con: Connection | Edge) => {
      if (!con.source || !con.target) return false;

      const source = getNode(con.source)!;
      const target = getNode(con.target)!;
      if (!canConnect(source.type!, target.type!)) return false;

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
        <ControlButton onClick={undo} disabled={!canUndo} title="Undo">
          <Undo2Icon className="size-5 !fill-none" />
        </ControlButton>
        <ControlButton onClick={redo} disabled={!canRedo} title="Redo">
          <Redo2Icon className="size-5 !fill-none" />
        </ControlButton>
        <ControlButton onClick={format}>
          <ListTreeIcon className="size-5 !fill-none" />
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
