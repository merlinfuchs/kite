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
import { canInsertIntoEdge, insertIntoEdge } from "@/lib/flow/insert";
import { getFlowLayout } from "@/lib/flow/layout";
import { createNoteNode, defaultNoteSize } from "@/lib/flow/notes";
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
import {
  ListTreeIcon,
  Redo2Icon,
  StickyNoteIcon,
  Undo2Icon,
} from "lucide-react";
import FlowContextMenu, {
  FlowContextMenuPoint,
  FlowContextMenuTarget,
} from "./FlowContextMenu";
import FlowInsertBlockDialog, {
  InsertIntoEdgeContext,
} from "./FlowInsertBlockDialog";

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
  // Zooms to the blocks.
  showNodes: (ids: string[]) => void;
  // Adds a note in the middle of the visible part of the flow.
  addNote: () => void;
}

interface Props {
  initialData?: FlowData;
  onChange: () => void;
  onSelectionChange?: OnSelectionChangeFunc;
  onNodeTap?: (node: Node<NodeData>) => void;
  onHistoryChange?: (canUndo: boolean, canRedo: boolean) => void;
  containerRef: RefObject<HTMLElement | null>;
  apiRef?: RefObject<FlowEditorApi | null>;
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
    editNodes(getFlowLayout(nodes, edges));
    setTimeout(() => {
      fitView();
    }, 50);
  }, [nodes, edges, editNodes, fitView]);

  const fitViewAnimated = useCallback(
    () => fitView({ padding: 0.2, duration: 250 }),
    [fitView]
  );

  const store = useStoreApi<Node<NodeData>>();

  const addNoteAt = useCallback(
    (position: { x: number; y: number }) => {
      editNodes((nodes) => [
        ...nodes.map((n) => (n.selected ? { ...n, selected: false } : n)),
        createNoteNode(position),
      ]);
      track("note_added");
    },
    [editNodes]
  );

  // Adds the note in the middle of the visible part of the flow.
  const addNote = useCallback(() => {
    const {
      width,
      height,
      transform: [x, y, zoom],
    } = store.getState();
    addNoteAt({
      x: (width / 2 - x) / zoom - defaultNoteSize.width / 2,
      y: (height / 2 - y) / zoom - defaultNoteSize.height / 2,
    });
  }, [store, addNoteAt]);

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
      showNodes: (ids) =>
        fitView({
          nodes: ids.map((id) => ({ id })),
          duration: 300,
          maxZoom: 1,
        }),
      addNote,
    }),
    [
      commit,
      setNodes,
      setEdges,
      undo,
      redo,
      format,
      fitViewAnimated,
      fitView,
      addNote,
    ]
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

  const getSelectionTarget = useCallback(
    (ids: string[]): FlowContextMenuTarget => {
      const nodes = getNodes();
      const edges = getEdges();
      return {
        type: "selection",
        ids,
        canCopy: !!copyFlowNodes(nodes, edges, ids),
        canDelete: getDeletedNodeIds(ids, nodes, edges).size > 0,
      };
    },
    [getNodes, getEdges]
  );

  // The menu is for the block, connection or canvas it's opened on. A block
  // outside of the selection is selected, the menu of one in it is for the
  // whole selection.
  const openMenuAt = useCallback(
    (
      el: Element,
      point: FlowContextMenuPoint
    ): FlowContextMenuTarget | null => {
      // Text fields, like a note that's edited, have no menu, so a long press
      // keeps the system's. Notes stop right clicks from reaching the menu.
      if (el.closest("input, textarea, [contenteditable]")) return null;

      const selectedIds = () =>
        getNodes()
          .filter((n) => n.selected)
          .map((n) => n.id);

      if (el.closest(".react-flow__nodesselection")) {
        return getSelectionTarget(selectedIds());
      }

      const nodeId = el.closest(".react-flow__node")?.getAttribute("data-id");
      const node = nodeId ? getNode(nodeId) : undefined;
      if (node) {
        if (node.selected) return getSelectionTarget(selectedIds());
        store.getState().addSelectedNodes([node.id]);
        return getSelectionTarget([node.id]);
      }

      const edgeId = el.closest(".react-flow__edge")?.getAttribute("data-id");
      const edge = edgeId ? getEdge(edgeId) : undefined;
      if (edge) {
        return edge.type === "fixed"
          ? null
          : {
              type: "edge",
              edgeId: edge.id,
              canInsert: canInsertIntoEdge(edge, getNode(edge.source)?.type),
            };
      }

      // Only the empty canvas, which contains the blocks and connections, gets
      // the canvas menu. Anything else, like the delete button of a connection
      // or the zoom controls, has none.
      return el.classList.contains("react-flow__pane")
        ? { type: "pane", position: point }
        : null;
    },
    [getNodes, getNode, getEdge, getSelectionTarget, store]
  );

  const deleteEdge = useCallback(
    (id: string) => deleteElements({ edges: [{ id }] }),
    [deleteElements]
  );

  // The connection the insert block dialog is open for.
  const [insertEdgeId, setInsertEdgeId] = useState<string | null>(null);

  const insertBlock = useCallback(
    (type: string) => {
      if (!insertEdgeId) return;
      const res = insertIntoEdge(type, insertEdgeId, getNodes(), getEdges());
      if (!res) return;

      commit();
      // Only the inserted block stays selected, so its settings open.
      setNodes(
        res.nodes.map((n) =>
          n.selected && n.id !== res.id ? { ...n, selected: false } : n
        )
      );
      setEdges(res.edges);
      track("block_added", { type, source: "insert" });
    },
    [insertEdgeId, getNodes, getEdges, commit, setNodes, setEdges]
  );

  const handleNodeClick = useCallback(
    (_e: React.MouseEvent, node: Node<NodeData>) => {
      onNodeTap?.(node);
    },
    [onNodeTap]
  );

  return (
    <InsertIntoEdgeContext.Provider value={setInsertEdgeId}>
      <FlowContextMenu
        onOpen={openMenuAt}
        canUndo={canUndo}
        canRedo={canRedo}
        onCopy={copy}
        onCut={cut}
        onDuplicate={duplicate}
        onDelete={deleteNodes}
        onDeleteEdge={deleteEdge}
        onInsertIntoEdge={setInsertEdgeId}
        onPaste={paste}
        onUndo={undo}
        onRedo={redo}
        onFormat={format}
        onFitView={fitViewAnimated}
        onAddNote={(p) => addNoteAt(screenToFlowPosition(p))}
      >
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
            <ControlButton onClick={format} title="Auto arrange">
              <ListTreeIcon className="size-5 !fill-none" />
            </ControlButton>
            <ControlButton onClick={addNote} title="Add note">
              <StickyNoteIcon className="size-5 !fill-none" />
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
      </FlowContextMenu>
      <FlowInsertBlockDialog
        open={!!insertEdgeId}
        onOpenChange={(open) => !open && setInsertEdgeId(null)}
        onPick={insertBlock}
      />
    </InsertIntoEdgeContext.Provider>
  );
}
