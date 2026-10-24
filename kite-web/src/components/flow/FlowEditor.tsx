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
  useStoreApi,
  Viewport,
} from "@xyflow/react";
import {
  DragEvent,
  MouseEvent as ReactMouseEvent,
  RefObject,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";

import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { copyFlowNodes } from "@/lib/flow/clipboard";
import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import { getFlowChangeKind, getFlowMergeKey } from "@/lib/flow/history";
import { getLayoutedElements } from "@/lib/flow/layout";
import {
  canConnect,
  createNode,
  getNodeValues,
  getDeletedNodeIds,
} from "@/lib/flow/nodes";
import { track } from "@/lib/analytics";
import { useFlowClipboard } from "@/lib/hooks/flowClipboard";
import { useFlowHistory } from "@/lib/hooks/flowHistory";
import { useHookedTheme } from "@/lib/hooks/theme";
import { useFlowViewportsStore } from "@/lib/state/flowViewports";
import "@xyflow/react/dist/base.css";
import { ListTreeIcon, Redo2Icon, Undo2Icon } from "lucide-react";
import FlowContextMenu, {
  FlowContextMenuState,
  FlowContextMenuTarget,
} from "./FlowContextMenu";

export interface FlowEditorApi {
  // Replaces the flow in one undo step, or in the undo step of the previous
  // replacement with the same merge key.
  replaceFlow: (
    nodes: Node<NodeData>[],
    edges: Edge[],
    mergeKey?: string
  ) => void;
  undo: () => void;
  redo: () => void;
  format: () => void;
  fitView: () => void;
}

interface Props {
  initialData?: FlowData;
  onChange: () => void;
  onSelectionChange?: OnSelectionChangeFunc;
  onNodeTap?: (node: Node<NodeData>) => void;
  onHistoryChange?: (canUndo: boolean, canRedo: boolean) => void;
  containerRef: RefObject<HTMLElement>;
  apiRef?: RefObject<FlowEditorApi>;
  // Remembers the viewport under this key and restores it when opened again.
  viewportKey?: string;
}

export default function FlowEditor({
  initialData,
  onChange,
  onSelectionChange,
  onNodeTap,
  onHistoryChange,
  containerRef,
  apiRef,
  viewportKey,
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
    deleteElements,
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

  useEffect(() => {
    onHistoryChange?.(canUndo, canRedo);
  }, [canUndo, canRedo, onHistoryChange]);

  const format = useCallback(() => {
    const formattedNodes = getLayoutedElements(nodes, edges);

    editNodes(formattedNodes.nodes);
    setTimeout(() => {
      fitView();
    }, 50);
  }, [nodes, edges, editNodes, fitView]);

  const fitViewAnimated = useCallback(
    () => fitView({ padding: 0.2, duration: 250 }),
    [fitView]
  );

  useImperativeHandle(
    apiRef,
    () => ({
      replaceFlow: (nodes, edges, mergeKey) => {
        commit(mergeKey, Infinity);
        setNodes(nodes);
        setEdges(edges);
      },
      undo,
      redo,
      format,
      fitView: fitViewAnimated,
    }),
    [commit, setNodes, setEdges, undo, redo, format, fitViewAnimated]
  );

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
      track("block_added", { type, source: "drag" });
    },
    [screenToFlowPosition, addNodes, addEdges]
  );

  const { onMouseMove, copy, cut, duplicate, deleteNodes, paste } =
    useFlowClipboard(containerRef);

  // Only read on mount, react-flow ignores later changes.
  const [savedViewport] = useState(() =>
    viewportKey
      ? useFlowViewportsStore.getState().viewports[viewportKey]
      : undefined
  );
  // Programmatic moves like fitView have no event and aren't saved, so flows
  // the user never moved keep fitting the view.
  const onMoveEnd = useCallback(
    (event: unknown, viewport: Viewport) => {
      if (viewportKey && event)
        useFlowViewportsStore.getState().save(viewportKey, viewport);
    },
    [viewportKey]
  );

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

  const store = useStoreApi();
  const [contextMenu, setContextMenu] = useState<FlowContextMenuState | null>(
    null
  );
  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  // The context menu needs a mouse or trackpad, touch screens use the
  // block settings and the bottom bar instead.
  const openMenu = useCallback(
    (e: ReactMouseEvent | MouseEvent, target: FlowContextMenuTarget) => {
      if (!window.matchMedia("(any-pointer: fine)").matches) return false;
      e.preventDefault();
      setContextMenu({ target, x: e.clientX, y: e.clientY });
      return true;
    },
    []
  );

  const openSelectionMenu = useCallback(
    (e: ReactMouseEvent, ids: string[]) => {
      const nodes = getNodes();
      const edges = getEdges();
      return openMenu(e, {
        type: "selection",
        ids,
        canCopy: !!copyFlowNodes(nodes, edges, ids),
        canDelete: getDeletedNodeIds(ids, nodes, edges).size > 0,
      });
    },
    [getNodes, getEdges, openMenu]
  );

  const onNodeContextMenu = useCallback(
    (e: ReactMouseEvent, node: Node<NodeData>) => {
      // Right clicking a block outside of the selection selects only it.
      const ids = node.selected
        ? getNodes()
            .filter((n) => n.selected)
            .map((n) => n.id)
        : [node.id];
      if (openSelectionMenu(e, ids) && !node.selected) {
        store.getState().addSelectedNodes(ids);
      }
    },
    [openSelectionMenu, getNodes, store]
  );

  const onSelectionContextMenu = useCallback(
    (e: ReactMouseEvent, nodes: Node<NodeData>[]) =>
      openSelectionMenu(
        e,
        nodes.map((n) => n.id)
      ),
    [openSelectionMenu]
  );

  const onEdgeContextMenu = useCallback(
    (e: ReactMouseEvent, edge: Edge) => {
      if (edge.type !== "fixed") openMenu(e, { type: "edge", edgeId: edge.id });
    },
    [openMenu]
  );

  const onPaneContextMenu = useCallback(
    (e: ReactMouseEvent | MouseEvent) => openMenu(e, { type: "pane" }),
    [openMenu]
  );

  const deleteEdge = useCallback(
    (id: string) => deleteElements({ edges: [{ id }] }),
    [deleteElements]
  );

  const handleNodeClick = useCallback(
    (_e: React.MouseEvent, node: Node<NodeData>) => {
      onNodeTap?.(node);
    },
    [onNodeTap]
  );

  return (
    <>
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
        onMoveEnd={onMoveEnd}
        onConnect={onConnect}
        isValidConnection={isValidConnection}
        onSelectionChange={onSelectionChange}
        onNodeClick={handleNodeClick}
        onNodeContextMenu={onNodeContextMenu}
        onSelectionContextMenu={onSelectionContextMenu}
        onEdgeContextMenu={onEdgeContextMenu}
        onPaneContextMenu={onPaneContextMenu}
        onMoveStart={closeContextMenu}
        colorMode={theme === "dark" ? "dark" : "light"}
        defaultEdgeOptions={{ type: "delete_button" }}
        multiSelectionKeyCode={null}
        deleteKeyCode={["Backspace", "Delete"]}
        proOptions={{
          hideAttribution: true,
        }}
        className="!bg-background flex-auto"
        defaultViewport={savedViewport}
        fitView={!savedViewport}
      >
        <Controls
          showInteractive={false}
          position="bottom-right"
          className="scale-110 !hidden md:!flex"
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
      <FlowContextMenu
        menu={contextMenu}
        onClose={closeContextMenu}
        canUndo={canUndo}
        canRedo={canRedo}
        onCopy={copy}
        onCut={cut}
        onDuplicate={duplicate}
        onDelete={deleteNodes}
        onDeleteEdge={deleteEdge}
        onPaste={paste}
        onUndo={undo}
        onRedo={redo}
        onFormat={format}
        onFitView={fitViewAnimated}
      />
    </>
  );
}
