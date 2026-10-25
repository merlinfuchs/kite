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
  NodePositionChange,
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
  useMemo,
  useRef,
  useState,
} from "react";

import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { copyFlowNodes } from "@/lib/flow/clipboard";
import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import { getFlowChangeKind, getFlowMergeKey } from "@/lib/flow/history";
import { getFlowLayout } from "@/lib/flow/layout";
import {
  createNoteNode,
  defaultNoteSize,
  noteNodeType,
  withNoteSizes,
} from "@/lib/flow/notes";
import {
  applyGroupVisibility,
  createGroupNode,
  defaultGroupSize,
  getGroupMemberIds,
  groupSelectedNodes,
  isGroupNode,
  withGroupSizes,
} from "@/lib/flow/groups";
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
  GroupIcon,
  ListTreeIcon,
  Redo2Icon,
  StickyNoteIcon,
  Undo2Icon,
} from "lucide-react";
import FlowContextMenu, { FlowContextMenuTarget } from "./FlowContextMenu";
import FlowGroupEdges from "./FlowGroupEdges";

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
  // Adds a note in the middle of the visible part of the flow.
  addNote: () => void;
  // Groups the selected blocks, or adds an empty group in the middle of the
  // visible part of the flow.
  group: () => void;
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
    withGroupSizes(withNoteSizes(initialData?.nodes || []))
  );
  // The blocks in collapsed groups are hidden. This is derived from the groups,
  // so undo, redo and AI edits can't leave blocks hidden by mistake.
  const visibleNodes = useMemo(() => applyGroupVisibility(nodes), [nodes]);
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

  // The middle of the visible part of the flow.
  const getViewCenter = useCallback(() => {
    const {
      width,
      height,
      transform: [x, y, zoom],
    } = store.getState();
    return { x: (width / 2 - x) / zoom, y: (height / 2 - y) / zoom };
  }, [store]);

  const addNoteAt = useCallback(
    (position: { x: number; y: number }) => {
      const note = createNoteNode(
        position,
        getNodes().map((n) => n.id)
      );
      editNodes((nodes) => [
        ...nodes.map((n) => (n.selected ? { ...n, selected: false } : n)),
        note,
      ]);
      track("note_added");
    },
    [getNodes, editNodes]
  );

  const addNote = useCallback(() => {
    const center = getViewCenter();
    addNoteAt({
      x: center.x - defaultNoteSize.width / 2,
      y: center.y - defaultNoteSize.height / 2,
    });
  }, [getViewCenter, addNoteAt]);

  // Adds an empty group with its center at the given position.
  const addGroupAt = useCallback(
    (center: { x: number; y: number }) => {
      editNodes((nodes) => [
        {
          ...createGroupNode(
            {
              x: center.x - defaultGroupSize.width / 2,
              y: center.y - defaultGroupSize.height / 2,
              ...defaultGroupSize,
            },
            nodes.map((n) => n.id)
          ),
          selected: true,
        },
        ...nodes.map((n) => (n.selected ? { ...n, selected: false } : n)),
      ]);
      track("group_added", { empty: true });
    },
    [editNodes]
  );

  const group = useCallback(() => {
    // The box around the grouped selection would cover the new group.
    store.setState({ nodesSelectionActive: false });

    const grouped = groupSelectedNodes(getNodes());
    if (grouped) {
      editNodes(grouped.nodes);
      track("group_added", { empty: false });
      return;
    }

    addGroupAt(getViewCenter());
  }, [store, getNodes, editNodes, getViewCenter, addGroupAt]);

  // Ctrl+G groups the selection, like in design tools.
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (!(e.metaKey || e.ctrlKey) || e.altKey || e.shiftKey) return;
      if (e.key.toLowerCase() !== "g") return;

      const target = e.target as HTMLElement;
      if (target.closest("input, textarea, [contenteditable]")) return;
      const dialog = target.closest("[role=dialog]");
      if (dialog && !dialog.contains(containerRef.current)) return;
      // Only the editor the pointer or focus is in, if there are several.
      if (!containerRef.current?.contains(target) && target !== document.body) {
        return;
      }

      e.preventDefault();
      e.stopPropagation();
      group();
    }

    document.addEventListener("keydown", onKeyDown, true);
    return () => document.removeEventListener("keydown", onKeyDown, true);
  }, [containerRef, group]);

  useImperativeHandle(
    apiRef,
    () => ({
      replaceFlow: (nodes, edges, mergeKey) => {
        commit(mergeKey, Infinity);
        setNodes(withGroupSizes(withNoteSizes(nodes)));
        setEdges(edges);
      },
      undo,
      redo,
      format,
      fitView: fitViewAnimated,
      addNote,
      group,
    }),
    [
      commit,
      setNodes,
      setEdges,
      undo,
      redo,
      format,
      fitViewAnimated,
      addNote,
      group,
    ]
  );

  const onConnect = useCallback(
    (con: Connection) => editEdges((eds) => addEdge(con, eds)),
    [editEdges]
  );

  const isDragging = useRef(false);
  const isResizing = useRef(false);

  // The nodes in the dragged groups, with where they were when the drag
  // started, so they move along with the groups.
  const groupDrag = useRef<{
    groupId: string;
    groupStart: { x: number; y: number };
    members: Map<string, { x: number; y: number }>;
  } | null>(null);

  const onDragStart = useCallback(
    (dragged: Node<NodeData>[]) => {
      groupDrag.current = null;
      const groups = dragged.filter(isGroupNode);
      if (groups.length === 0) return;

      const all = getNodes();
      const draggedIds = new Set(dragged.map((n) => n.id));
      const memberIds = new Set(
        groups.flatMap((g) => getGroupMemberIds(g, all))
      );
      const members = new Map<string, { x: number; y: number }>();
      for (const node of all) {
        if (memberIds.has(node.id) && !draggedIds.has(node.id)) {
          members.set(node.id, { ...node.position });
        }
      }

      groupDrag.current = {
        groupId: groups[0].id,
        groupStart: { ...groups[0].position },
        members,
      };
    },
    [getNodes]
  );

  const onNodeDragStart = useCallback(
    (_e: unknown, _node: Node<NodeData>, dragged: Node<NodeData>[]) =>
      onDragStart(dragged),
    [onDragStart]
  );

  // Dragging a selection box that has groups in it.
  const onSelectionDragStart = useCallback(
    (_e: unknown, dragged: Node<NodeData>[]) => onDragStart(dragged),
    [onDragStart]
  );

  const onDragStop = useCallback(() => {
    groupDrag.current = null;
  }, []);

  // Moves the nodes of the dragged groups by as much as the groups moved.
  const withGroupMoves = useCallback(
    (changes: NodeChange<Node<NodeData>>[]) => {
      const drag = groupDrag.current;
      if (!drag || drag.members.size === 0) return changes;

      const change = changes.find(
        (c): c is NodePositionChange =>
          c.type === "position" && c.id === drag.groupId
      );
      if (!change) return changes;

      const moved = new Set(
        changes.filter((c) => c.type === "position").map((c) => c.id)
      );
      const memberChanges: NodePositionChange[] = [];
      for (const [id, start] of drag.members) {
        if (moved.has(id)) continue;
        memberChanges.push({
          id,
          type: "position",
          dragging: change.dragging,
          position: change.position && {
            x: start.x + change.position.x - drag.groupStart.x,
            y: start.y + change.position.y - drag.groupStart.y,
          },
        });
      }
      return [...changes, ...memberChanges];
    },
    []
  );

  const wrappedOnNodesChange = useCallback(
    (allChanges: NodeChange<Node<NodeData>>[]) => {
      const changes = withGroupMoves(allChanges);
      const filteredChanges = changes.filter((change) => {
        if (change.type === "remove") {
          const node = getNode(change.id);
          const values = getNodeValues(node!.type!);
          return !values.fixed;
        }

        return true;
      });

      if (filteredChanges.length === 0) return;

      // Resizing a note or group is recorded like a drag, once when it
      // starts.
      const resize = filteredChanges.find(
        (c) => c.type === "dimensions" && typeof c.resizing === "boolean"
      );
      if (resize?.type === "dimensions") {
        if (resize.resizing) {
          if (!isResizing.current) record();
          isResizing.current = true;
        } else if (isResizing.current) {
          isResizing.current = false;
          markChanged();
        }
      } else if (filteredChanges.some((c) => getFlowChangeKind(c) === "drag")) {
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
    [onNodesChange, getNode, record, commit, markChanged, withGroupMoves]
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
      if (type === noteNodeType) {
        addNoteAt(position);
        return;
      }
      const [newNodes, newEdges] = createNode(type, position);

      addNodes(newNodes);
      addEdges(newEdges);
      track("block_added", { type, source: "drag" });
    },
    [screenToFlowPosition, addNodes, addEdges, addNoteAt]
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

  // Edges of blocks hidden in collapsed groups have nothing to attach to, which
  // isn't an error. They are drawn to the groups by FlowGroupEdges instead.
  const onError = useCallback((code: string, message: string) => {
    if (code === "008") return;
    if (process.env.NODE_ENV === "development") console.warn(message);
  }, []);

  const [menuTarget, setMenuTarget] = useState<FlowContextMenuTarget | null>(
    null
  );
  // Whether a handler below set the target for the current right click.
  const hasMenuTarget = useRef(false);
  const openMenu = useCallback((target: FlowContextMenuTarget) => {
    hasMenuTarget.current = true;
    setMenuTarget(target);
  }, []);
  // Right clicks on anything else, like the zoom controls, open no menu.
  const onContextMenu = useCallback((e: ReactMouseEvent) => {
    if (!hasMenuTarget.current) e.preventDefault();
    hasMenuTarget.current = false;
  }, []);

  // The context menu needs a mouse or trackpad, touch screens use the
  // block settings and the bottom bar instead.
  const [hasMouse, setHasMouse] = useState(false);
  useEffect(
    () => setHasMouse(window.matchMedia("(any-pointer: fine)").matches),
    []
  );

  const openSelectionMenu = useCallback(
    (ids: string[]) => {
      const nodes = getNodes();
      const edges = getEdges();
      openMenu({
        type: "selection",
        ids,
        canCopy: !!copyFlowNodes(nodes, edges, ids),
        canDelete: getDeletedNodeIds(ids, nodes, edges).size > 0,
      });
    },
    [getNodes, getEdges, openMenu]
  );

  const onNodeContextMenu = useCallback(
    (_e: ReactMouseEvent, node: Node<NodeData>) => {
      // Right clicking a block outside of the selection selects only it.
      if (node.selected) {
        openSelectionMenu(
          getNodes()
            .filter((n) => n.selected)
            .map((n) => n.id)
        );
      } else {
        store.getState().addSelectedNodes([node.id]);
        openSelectionMenu([node.id]);
      }
    },
    [openSelectionMenu, getNodes, store]
  );

  const onSelectionContextMenu = useCallback(
    (_e: ReactMouseEvent, nodes: Node<NodeData>[]) =>
      openSelectionMenu(nodes.map((n) => n.id)),
    [openSelectionMenu]
  );

  const onEdgeContextMenu = useCallback(
    (_e: ReactMouseEvent, edge: Edge) => {
      if (edge.type !== "fixed") openMenu({ type: "edge", edgeId: edge.id });
    },
    [openMenu]
  );

  const onPaneContextMenu = useCallback(
    (e: ReactMouseEvent | MouseEvent) =>
      openMenu({ type: "pane", position: { x: e.clientX, y: e.clientY } }),
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
    <FlowContextMenu
      target={menuTarget}
      disabled={!hasMouse}
      onContextMenu={onContextMenu}
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
      onGroup={group}
      onAddNote={(p) => addNoteAt(screenToFlowPosition(p))}
      onAddGroup={(p) => addGroupAt(screenToFlowPosition(p))}
    >
      <ReactFlow
        nodes={visibleNodes}
        edges={edges}
        onNodesChange={wrappedOnNodesChange}
        onEdgesChange={wrappedOnEdgesChange}
        onNodesDelete={onNodesDelete}
        onNodeDragStart={onNodeDragStart}
        onNodeDragStop={onDragStop}
        onSelectionDragStart={onSelectionDragStart}
        onSelectionDragStop={onDragStop}
        onError={onError}
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
          <ControlButton onClick={group} title="Group selected blocks (Ctrl+G)">
            <GroupIcon className="size-5 !fill-none" />
          </ControlButton>
        </Controls>
        <FlowGroupEdges />
        <Background
          variant={BackgroundVariant.Dots}
          gap={18}
          size={1}
          className="!bg-muted/20"
          color="#615d84"
        />
      </ReactFlow>
    </FlowContextMenu>
  );
}
