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
  Viewport,
} from "@xyflow/react";
import {
  DragEvent,
  MouseEvent as ReactMouseEvent,
  PointerEvent as ReactPointerEvent,
  RefObject,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from "react";

import { edgeTypes, nodeTypes } from "@/lib/flow/components";
import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import { getFlowChangeKind, getFlowMergeKey } from "@/lib/flow/history";
import { getLayoutedElements } from "@/lib/flow/layout";
import {
  canConnect,
  canSpliceEdge,
  createEdge,
  createNode,
  getActionableSelectedNodes,
  getDeletedNodeIds,
  getDisconnectedEdgeIds,
  getEdgeId,
  getNodeId,
  getNodeValues,
  getOwnerNodeId,
  getTopLevelSelectedNodes,
  isOwnedEdge,
  withOwnedNodes,
} from "@/lib/flow/nodes";
import { track } from "@/lib/analytics";
import { useFlowClipboard } from "@/lib/hooks/flowClipboard";
import { useFlowHistory } from "@/lib/hooks/flowHistory";
import { useHookedTheme } from "@/lib/hooks/theme";
import { useFlowViewportsStore } from "@/lib/state/flowViewports";
import { cn } from "@/lib/utils";
import "@xyflow/react/dist/base.css";
import { FocusIcon, ListTreeIcon, Redo2Icon, Undo2Icon } from "lucide-react";
import FlowContextMenu, { FlowContextMenuState } from "./FlowContextMenu";
import FlowInsertBlockDialog from "./FlowInsertBlockDialog";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

export interface FlowSelectionSummary {
  topLevelCount: number;
  actionableCount: number;
  canDisconnect: boolean;
  singleSelectedNodeId: string | null;
}

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
  centerFlow: () => void;
  copy: () => boolean;
  cut: () => boolean;
  duplicate: () => boolean;
  paste: (screenPos?: { x: number; y: number } | null) => Promise<boolean>;
  disconnect: () => boolean;
  deleteSelected: () => boolean;
  clearSelection: () => void;
}

interface Props {
  initialData?: FlowData;
  onChange: () => void;
  onSelectionChange?: OnSelectionChangeFunc;
  onSelectionSummaryChange?: (summary: FlowSelectionSummary) => void;
  onClipboardChange?: (hasClipboard: boolean) => void;
  onNodeTap?: (node: Node<NodeData>) => void;
  onHistoryChange?: (canUndo: boolean, canRedo: boolean) => void;
  isMobile?: boolean;
  mobileSelectMode?: boolean;
  mobileMoveMode?: boolean;
  onEnterMobileSelectMode?: (node: Node<NodeData>) => void;
  onExitMobileSelectMode?: () => void;
  containerRef: RefObject<HTMLElement>;
  apiRef?: RefObject<FlowEditorApi>;
  // Remembers the viewport under this key and restores it when opened again.
  viewportKey?: string;
}

const longPressDurationMs = 400;
const longPressMoveThresholdPx = 22;

export default function FlowEditor({
  initialData,
  onChange,
  onSelectionChange,
  onSelectionSummaryChange,
  onClipboardChange,
  onNodeTap,
  onHistoryChange,
  isMobile = false,
  mobileSelectMode = false,
  mobileMoveMode = false,
  onEnterMobileSelectMode,
  onExitMobileSelectMode,
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
    setCenter,
    getZoom,
    zoomTo,
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

  useEffect(() => {
    onHistoryChange?.(canUndo, canRedo);
  }, [canUndo, canRedo, onHistoryChange]);

  const { onMouseMove, copy, cut, duplicate, paste, hasClipboard } =
    useFlowClipboard({
      commit,
      setNodesState: setNodes,
      setEdgesState: setEdges,
    });

  useEffect(() => {
    onClipboardChange?.(hasClipboard);
  }, [hasClipboard, onClipboardChange]);

  const selectionSummary = useMemo<FlowSelectionSummary>(() => {
    const topLevel = getTopLevelSelectedNodes(nodes);
    const actionable = getActionableSelectedNodes(nodes);
    const topLevelIds = topLevel.map((n) => n.id);
    const disconnectedEdges =
      topLevelIds.length > 0
        ? getDisconnectedEdgeIds(topLevelIds, nodes, edges)
        : new Set<string>();

    return {
      topLevelCount: topLevel.length,
      actionableCount: actionable.length,
      canDisconnect: disconnectedEdges.size > 0,
      singleSelectedNodeId: topLevel.length === 1 ? topLevel[0].id : null,
    };
  }, [nodes, edges]);

  useEffect(() => {
    onSelectionSummaryChange?.(selectionSummary);
  }, [selectionSummary, onSelectionSummaryChange]);

  const format = useCallback(() => {
    const formattedNodes = getLayoutedElements(nodes, edges);

    editNodes(formattedNodes.nodes);
    setTimeout(() => {
      fitView();
    }, 50);
  }, [nodes, edges, editNodes, fitView]);

  const handleCenterFlow = useCallback(() => {
    const currentNodes = getNodes();
    if (currentNodes.length === 0) {
      fitView({ padding: 0.2, duration: 250 });
      return;
    }

    let minX = Infinity;
    let maxX = -Infinity;
    let minY = Infinity;
    let maxY = -Infinity;

    for (const n of currentNodes) {
      const x = n.position.x;
      const y = n.position.y;
      const w = n.measured?.width ?? 220;
      const h = n.measured?.height ?? 100;
      if (x < minX) minX = x;
      if (x + w > maxX) maxX = x + w;
      if (y < minY) minY = y;
      if (y + h > maxY) maxY = y + h;
    }

    const centerX = (minX + maxX) / 2;
    const centerY = (minY + maxY) / 2;
    setCenter(centerX, centerY, { zoom: getZoom(), duration: 250 });
  }, [getNodes, fitView, setCenter, getZoom]);

  const disconnect = useCallback(() => {
    const currentNodes = getNodes();
    const currentEdges = getEdges();
    const topLevel = getTopLevelSelectedNodes(currentNodes);
    if (topLevel.length === 0) return false;

    const edgeIds = getDisconnectedEdgeIds(
      topLevel.map((n) => n.id),
      currentNodes,
      currentEdges
    );
    if (edgeIds.size === 0) return false;

    commit();
    setEdges((eds) => eds.filter((e) => !edgeIds.has(e.id)));
    return true;
  }, [getNodes, getEdges, commit, setEdges]);

  const deleteSelected = useCallback(() => {
    const currentNodes = getNodes();
    const currentEdges = getEdges();
    const selectedIds = currentNodes.filter((n) => n.selected).map((n) => n.id);
    const removed = getDeletedNodeIds(selectedIds, currentNodes, currentEdges);
    if (removed.size === 0) return false;

    commit();
    setEdges((eds) =>
      eds.filter((e) => !removed.has(e.source) && !removed.has(e.target))
    );
    setNodes((nds) => nds.filter((n) => !removed.has(n.id)));
    return true;
  }, [getNodes, getEdges, commit, setEdges, setNodes]);

  const clearSelection = useCallback(() => {
    setNodes((nds) =>
      nds.some((n) => n.selected)
        ? nds.map((n) => (n.selected ? { ...n, selected: false } : n))
        : nds
    );
    setEdges((eds) =>
      eds.some((e) => e.selected)
        ? eds.map((e) => (e.selected ? { ...e, selected: false } : e))
        : eds
    );
    if (isMobile) {
      onExitMobileSelectMode?.();
    }
  }, [setNodes, setEdges, isMobile, onExitMobileSelectMode]);

  const resetZoom = useCallback(() => {
    zoomTo(1, { duration: 250 });
  }, [zoomTo]);

  const selectAllInView = useCallback(() => {
    if (!containerRef.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    const p1 = screenToFlowPosition({ x: rect.left, y: rect.top });
    const p2 = screenToFlowPosition({ x: rect.right, y: rect.bottom });
    const viewBounds = {
      minX: Math.min(p1.x, p2.x),
      maxX: Math.max(p1.x, p2.x),
      minY: Math.min(p1.y, p2.y),
      maxY: Math.max(p1.y, p2.y),
    };

    const currentNodes = getNodes();
    const currentEdges = getEdges();
    const inViewTopLevelIds = new Set<string>();

    for (const n of currentNodes) {
      const w = n.measured?.width ?? n.width ?? 200;
      const h = n.measured?.height ?? n.height ?? 80;
      const right = n.position.x + w;
      const bottom = n.position.y + h;
      if (
        right >= viewBounds.minX &&
        n.position.x <= viewBounds.maxX &&
        bottom >= viewBounds.minY &&
        n.position.y <= viewBounds.maxY
      ) {
        const topId = getOwnerNodeId(n.id, currentNodes, currentEdges) ?? n.id;
        inViewTopLevelIds.add(topId);
      }
    }

    const groupIds = withOwnedNodes(
      [...inViewTopLevelIds],
      currentNodes,
      currentEdges
    );
    setNodes((nds) => nds.map((n) => ({ ...n, selected: groupIds.has(n.id) })));
    setEdges((eds) =>
      eds.map((e) => (e.selected ? { ...e, selected: false } : e))
    );
    if (isMobile) {
      const firstTopLevel = currentNodes.find((n) =>
        inViewTopLevelIds.has(n.id)
      );
      if (firstTopLevel) {
        onEnterMobileSelectMode?.({ ...firstTopLevel, selected: true });
      }
    }
  }, [
    containerRef,
    screenToFlowPosition,
    getNodes,
    getEdges,
    setNodes,
    setEdges,
    isMobile,
    onEnterMobileSelectMode,
  ]);

  const selectAllInFlow = useCallback(() => {
    const currentNodes = getNodes();
    const currentEdges = getEdges();
    const allTopLevelIds = currentNodes
      .filter((n) => !getOwnerNodeId(n.id, currentNodes, currentEdges))
      .map((n) => n.id);
    const groupIds = withOwnedNodes(allTopLevelIds, currentNodes, currentEdges);
    setNodes((nds) => nds.map((n) => ({ ...n, selected: groupIds.has(n.id) })));
    setEdges((eds) =>
      eds.map((e) => (e.selected ? { ...e, selected: false } : e))
    );
    if (isMobile) {
      const firstTopLevel = currentNodes.find((n) =>
        allTopLevelIds.includes(n.id)
      );
      if (firstTopLevel) {
        onEnterMobileSelectMode?.({ ...firstTopLevel, selected: true });
      }
    }
  }, [
    getNodes,
    getEdges,
    setNodes,
    setEdges,
    isMobile,
    onEnterMobileSelectMode,
  ]);

  const duplicateWithConnections = useCallback(() => {
    const currentNodes = getNodes();
    const currentEdges = getEdges();
    const selectedIds = currentNodes
      .filter((n) => n.selected && !getNodeValues(n.type!).fixed)
      .map((n) => n.id);
    const copiedIds = withOwnedNodes(selectedIds, currentNodes, currentEdges);
    if (copiedIds.size === 0) return false;

    const sourceNodes = currentNodes.filter((n) => copiedIds.has(n.id));
    if (sourceNodes.length === 0) return false;

    const offset = 60;
    const newIds = new Map(sourceNodes.map((n) => [n.id, getNodeId()]));

    const newNodes: Node<NodeData>[] = sourceNodes.map((n) => ({
      id: newIds.get(n.id)!,
      type: n.type,
      position: {
        x: n.position.x + offset,
        y: n.position.y + offset,
      },
      data: structuredClone(n.data),
      selected: true,
    }));

    const internalEdges: Edge[] = currentEdges
      .filter((e) => newIds.has(e.source) && newIds.has(e.target))
      .map((e) => ({
        id: getEdgeId(),
        type: e.type,
        source: newIds.get(e.source)!,
        target: newIds.get(e.target)!,
        sourceHandle: e.sourceHandle,
        targetHandle: e.targetHandle,
      }));

    const incomingEdges: Edge[] = currentEdges
      .filter(
        (e) =>
          !copiedIds.has(e.source) &&
          newIds.has(e.target) &&
          !isOwnedEdge(e, currentNodes)
      )
      .map((e) => ({
        id: getEdgeId(),
        type: e.type,
        source: e.source,
        target: newIds.get(e.target)!,
        sourceHandle: e.sourceHandle,
        targetHandle: e.targetHandle,
      }));

    const outgoingEdges: Edge[] = currentEdges
      .filter(
        (e) =>
          newIds.has(e.source) &&
          !copiedIds.has(e.target) &&
          !isOwnedEdge(e, currentNodes)
      )
      .map((e) => ({
        id: getEdgeId(),
        type: e.type,
        source: newIds.get(e.source)!,
        target: e.target,
        sourceHandle: e.sourceHandle,
        targetHandle: e.targetHandle,
      }));

    commit();
    setNodes((nds) => [
      ...nds.map((n) => (n.selected ? { ...n, selected: false } : n)),
      ...newNodes,
    ]);
    setEdges((eds) => [
      ...eds.map((e) => (e.selected ? { ...e, selected: false } : e)),
      ...internalEdges,
      ...incomingEdges,
      ...outgoingEdges,
    ]);
    return true;
  }, [getNodes, getEdges, commit, setNodes, setEdges]);

  const deleteEdge = useCallback(
    (edgeId: string) => {
      commit();
      setEdges((eds) => eds.filter((e) => e.id !== edgeId));
    },
    [commit, setEdges]
  );

  const insertBlockIntoEdge = useCallback(
    (edgeId: string, blockType: string) => {
      const currentEdges = getEdges();
      const currentNodes = getNodes();
      const targetEdge = currentEdges.find((e) => e.id === edgeId);
      if (!targetEdge || !canSpliceEdge(targetEdge, currentNodes)) return;

      const sourceNode = currentNodes.find((n) => n.id === targetEdge.source);
      const targetNode = currentNodes.find((n) => n.id === targetEdge.target);
      if (!sourceNode || !targetNode) return;

      const midX = Math.round(
        (sourceNode.position.x + targetNode.position.x) / 2
      );
      const midY = Math.round(
        (sourceNode.position.y + targetNode.position.y) / 2
      );

      const [createdNodes, createdEdges] = createNode(blockType, {
        x: midX,
        y: midY,
      });
      if (createdNodes.length === 0) return;

      const primaryNode = createdNodes[0];
      const edgeIn = createEdge(
        sourceNode,
        primaryNode,
        targetEdge.sourceHandle
      );

      let outgoingSourceNode: Node<NodeData> | null = primaryNode;
      const primaryOutputs = getNodeValues(primaryNode.type!).outputs;
      if (primaryOutputs && primaryOutputs.length === 0) {
        outgoingSourceNode =
          createdNodes.find(
            (n) =>
              n.id !== primaryNode.id &&
              (getNodeValues(n.type!).outputs?.length ?? 1) > 0
          ) ?? null;
      }

      const newEdges = [...createdEdges, edgeIn];
      if (outgoingSourceNode) {
        newEdges.push(createEdge(outgoingSourceNode, targetNode));
      }

      commit();
      setEdges((eds) => [...eds.filter((e) => e.id !== edgeId), ...newEdges]);
      setNodes((nds) => [
        ...nds.map((n) => (n.selected ? { ...n, selected: false } : n)),
        ...createdNodes.map((n) =>
          n.id === primaryNode.id ? { ...n, selected: true } : n
        ),
      ]);
    },
    [getEdges, getNodes, commit, setEdges, setNodes]
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
      centerFlow: handleCenterFlow,
      copy,
      cut,
      duplicate,
      paste,
      disconnect,
      deleteSelected,
      clearSelection,
    }),
    [
      commit,
      setNodes,
      setEdges,
      undo,
      redo,
      format,
      handleCenterFlow,
      copy,
      cut,
      duplicate,
      paste,
      disconnect,
      deleteSelected,
      clearSelection,
    ]
  );

  const onConnect = useCallback(
    (con: Connection) => editEdges((eds) => addEdge(con, eds)),
    [editEdges]
  );

  const isDragging = useRef(false);
  const rightClickSelectedIdsRef = useRef<Set<string>>(new Set());
  const isRightClickingRef = useRef(false);
  const isShiftKeyDownRef = useRef(false);
  const companionDragIdsRef = useRef<Set<string>>(new Set());

  const wrappedOnNodesChange = useCallback(
    (changes: NodeChange<Node<NodeData>>[]) => {
      let filteredChanges = changes.filter((change) => {
        if (change.type === "remove") {
          const node = getNode(change.id);
          const values = getNodeValues(node!.type!);
          return !values.fixed;
        }
        if (
          change.type === "select" &&
          (isShiftKeyDownRef.current ||
            (isRightClickingRef.current &&
              rightClickSelectedIdsRef.current.size > 0))
        ) {
          return false;
        }

        return true;
      });

      if (isMobile) {
        if (!mobileMoveMode) {
          filteredChanges = filteredChanges.filter(
            (c) => c.type !== "position"
          );
        } else {
          const currentNodes = getNodes();
          const currentEdges = getEdges();
          const selectedIds = currentNodes
            .filter((n) => n.selected)
            .map((n) => n.id);
          const movableIds = withOwnedNodes(
            selectedIds,
            currentNodes,
            currentEdges
          );
          filteredChanges = filteredChanges.filter(
            (c) => c.type !== "position" || movableIds.has(c.id)
          );
        }
      }

      if (filteredChanges.length === 0) return;

      // Dragging a selected block or an owner block moves the rest of the
      // selection and any owned branch blocks along with it.
      const posChanges = filteredChanges.filter(
        (c): c is Extract<NodeChange<Node<NodeData>>, { type: "position" }> =>
          c.type === "position"
      );

      if (posChanges.length === 1) {
        const primary = posChanges[0];
        const primaryNode = getNode(primary.id);
        if (primaryNode && primary.position && primary.dragging) {
          const dx = primary.position.x - primaryNode.position.x;
          const dy = primary.position.y - primaryNode.position.y;
          const currentNodes = getNodes();
          const currentEdges = getEdges();
          const seedIds = primaryNode.selected
            ? currentNodes.filter((n) => n.selected).map((n) => n.id)
            : [primary.id];
          const groupIds = withOwnedNodes(seedIds, currentNodes, currentEdges);
          groupIds.delete(primary.id);

          if (groupIds.size > 0) {
            companionDragIdsRef.current = groupIds;
            const extraChanges: NodeChange<Node<NodeData>>[] = [];
            for (const id of groupIds) {
              const comp = getNode(id);
              if (comp) {
                extraChanges.push({
                  id,
                  type: "position",
                  dragging: true,
                  position: {
                    x: comp.position.x + dx,
                    y: comp.position.y + dy,
                  },
                });
              }
            }
            filteredChanges = [...filteredChanges, ...extraChanges];
          }
        } else if (!primary.dragging && companionDragIdsRef.current.size > 0) {
          const extraEndChanges: NodeChange<Node<NodeData>>[] = [
            ...companionDragIdsRef.current,
          ].map((id) => ({
            id,
            type: "position" as const,
            dragging: false,
          }));
          companionDragIdsRef.current = new Set();
          filteredChanges = [...filteredChanges, ...extraEndChanges];
        }
      }

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
    [
      onNodesChange,
      getNode,
      getNodes,
      getEdges,
      record,
      commit,
      markChanged,
      isMobile,
      mobileMoveMode,
    ]
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

  const longPressTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pointerStartRef = useRef<{
    x: number;
    y: number;
    nodeId?: string;
    edgeId?: string;
  } | null>(null);
  const longPressTriggeredRef = useRef(false);

  const clearLongPressTimer = useCallback(() => {
    if (longPressTimerRef.current) {
      clearTimeout(longPressTimerRef.current);
      longPressTimerRef.current = null;
    }
    pointerStartRef.current = null;
  }, []);

  useEffect(() => {
    return () => clearLongPressTimer();
  }, [clearLongPressTimer]);

  const selectMobileTopLevelNode = useCallback(
    (rawNodeId: string, additive: boolean) => {
      const currentNodes = getNodes();
      const currentEdges = getEdges();
      const targetId =
        getOwnerNodeId(rawNodeId, currentNodes, currentEdges) ?? rawNodeId;
      const targetNode = currentNodes.find((n) => n.id === targetId);
      if (!targetNode) return null;

      const groupIds = withOwnedNodes([targetId], currentNodes, currentEdges);

      setNodes((nds) =>
        nds.map((n) => {
          if (groupIds.has(n.id)) {
            return n.selected ? n : { ...n, selected: true };
          }
          if (!additive && n.selected) {
            return { ...n, selected: false };
          }
          return n;
        })
      );
      return targetNode;
    },
    [getNodes, getEdges, setNodes]
  );

  const findEdgeIdAtOrNear = useCallback(
    (e: ReactPointerEvent<HTMLDivElement>): string | null => {
      const target = e.target as HTMLElement | SVGElement | null;
      const direct =
        target
          ?.closest(".react-flow__edge[data-id]")
          ?.getAttribute("data-id") ||
        target?.closest("[data-edge-id]")?.getAttribute("data-edge-id") ||
        target
          ?.closest(".react-flow__edge-delete-btn")
          ?.getAttribute("data-id");
      if (direct) return direct;

      if (typeof document !== "undefined" && document.elementsFromPoint) {
        const offsets = [
          [0, -10],
          [0, 10],
          [-10, 0],
          [10, 0],
          [-8, -8],
          [8, -8],
          [-8, 8],
          [8, 8],
          [0, -6],
          [0, 6],
          [-6, 0],
          [6, 0],
          [0, -12],
          [0, 12],
          [-12, 0],
          [12, 0],
        ];
        for (const [ox, oy] of offsets) {
          const els = document.elementsFromPoint(
            e.clientX + ox,
            e.clientY + oy
          );
          for (const el of els) {
            const edgeEl =
              el.closest(".react-flow__edge[data-id]") ||
              el.closest("[data-edge-id]") ||
              el.closest(".react-flow__edge-delete-btn");
            if (edgeEl) {
              const id =
                edgeEl.getAttribute("data-id") ||
                edgeEl.getAttribute("data-edge-id");
              if (id) return id;
            }
          }
        }
      }
      return null;
    },
    []
  );

  const [contextMenu, setContextMenu] = useState<FlowContextMenuState>({
    open: false,
    x: 0,
    y: 0,
    kind: "node",
  });

  const closeContextMenu = useCallback(() => {
    setContextMenu((prev) => (prev.open ? { ...prev, open: false } : prev));
  }, []);

  const openContextMenu = useCallback(
    (menuState: Omit<FlowContextMenuState, "open" | "id">) => {
      setContextMenu({
        ...menuState,
        open: true,
        id: Date.now(),
      });
    },
    []
  );

  const handlePointerDown = useCallback(
    (e: ReactPointerEvent<HTMLDivElement>) => {
      if (!e.isPrimary) {
        clearLongPressTimer();
        return;
      }
      if (e.button !== 0) return;
      if (!isMobile && e.shiftKey) {
        e.preventDefault();
        e.stopPropagation();
        return;
      }

      const target = e.target as HTMLElement | null;
      const nodeEl = target?.closest(".react-flow__node[data-id]");

      if (isMobile && !mobileMoveMode && nodeEl) {
        const nodeId = nodeEl.getAttribute("data-id");
        if (!nodeId) return;

        clearLongPressTimer();
        longPressTriggeredRef.current = false;
        pointerStartRef.current = {
          x: e.clientX,
          y: e.clientY,
          nodeId,
        };

        longPressTimerRef.current = setTimeout(() => {
          longPressTimerRef.current = null;
          pointerStartRef.current = null;
          longPressTriggeredRef.current = true;

          if (typeof navigator !== "undefined" && navigator.vibrate) {
            navigator.vibrate(15);
          }

          const selectedTarget = selectMobileTopLevelNode(
            nodeId,
            mobileSelectMode
          );
          if (selectedTarget) {
            onEnterMobileSelectMode?.(selectedTarget);
          }
        }, longPressDurationMs);
        return;
      }

      if (!mobileMoveMode) {
        const edgeId = findEdgeIdAtOrNear(e);
        if (edgeId) {
          const currentEdges = getEdges();
          const currentNodes = getNodes();
          const targetEdge = currentEdges.find((ed) => ed.id === edgeId);
          if (!targetEdge || !canSpliceEdge(targetEdge, currentNodes)) return;

          clearLongPressTimer();
          longPressTriggeredRef.current = false;
          pointerStartRef.current = {
            x: e.clientX,
            y: e.clientY,
            edgeId,
          };

          longPressTimerRef.current = setTimeout(() => {
            longPressTimerRef.current = null;
            pointerStartRef.current = null;
            longPressTriggeredRef.current = true;

            if (typeof navigator !== "undefined" && navigator.vibrate) {
              navigator.vibrate(20);
            }

            setInsertBlockEdgeId(edgeId);
          }, 1000);
          return;
        }
      }

      if (
        isMobile &&
        !mobileMoveMode &&
        !target?.closest(".react-flow__handle") &&
        target?.closest(".react-flow__pane")
      ) {
        const clientX = e.clientX;
        const clientY = e.clientY;
        clearLongPressTimer();
        longPressTriggeredRef.current = false;
        pointerStartRef.current = {
          x: clientX,
          y: clientY,
        };

        longPressTimerRef.current = setTimeout(() => {
          longPressTimerRef.current = null;
          pointerStartRef.current = null;
          longPressTriggeredRef.current = true;

          if (typeof navigator !== "undefined" && navigator.vibrate) {
            navigator.vibrate(15);
          }

          openContextMenu({
            x: clientX,
            y: clientY,
            kind: "pane",
          });
        }, longPressDurationMs);
      }
    },
    [
      isMobile,
      mobileMoveMode,
      mobileSelectMode,
      clearLongPressTimer,
      selectMobileTopLevelNode,
      onEnterMobileSelectMode,
      getEdges,
      getNodes,
      findEdgeIdAtOrNear,
      openContextMenu,
    ]
  );

  const handlePointerMove = useCallback(
    (e: ReactPointerEvent<HTMLDivElement>) => {
      if (!pointerStartRef.current) return;
      const dx = e.clientX - pointerStartRef.current.x;
      const dy = e.clientY - pointerStartRef.current.y;
      const threshold = pointerStartRef.current.edgeId
        ? 30
        : longPressMoveThresholdPx;
      if (Math.hypot(dx, dy) > threshold) {
        clearLongPressTimer();
      }
    },
    [clearLongPressTimer]
  );

  const handlePointerUpOrCancel = useCallback(() => {
    clearLongPressTimer();
    if (longPressTriggeredRef.current) {
      setTimeout(() => {
        longPressTriggeredRef.current = false;
      }, 100);
    }
  }, [clearLongPressTimer]);

  const [marqueeRect, setMarqueeRect] = useState<{
    startX: number;
    startY: number;
    currentX: number;
    currentY: number;
  } | null>(null);
  const rightClickDragRef = useRef<{
    startX: number;
    startY: number;
    moved: boolean;
    isShiftKey?: boolean;
  } | null>(null);
  const marqueeCleanupRef = useRef<(() => void) | null>(null);
  const suppressNextContextMenuRef = useRef(false);
  const ignoreContainerContextMenuRef = useRef(false);

  const [insertBlockEdgeId, setInsertBlockEdgeId] = useState<string | null>(
    null
  );
  useEffect(() => {
    return () => {
      marqueeCleanupRef.current?.();
    };
  }, []);

  const [renameDialog, setRenameDialog] = useState<{
    open: boolean;
    nodeId: string;
    name: string;
  } | null>(null);

  const handleRenameClick = useCallback(() => {
    const currentNodes = getNodes();
    const targetId =
      contextMenu.nodeId ?? selectionSummary.singleSelectedNodeId;
    const targetNode = currentNodes.find((n) => n.id === targetId);
    if (!targetNode) return;
    setRenameDialog({
      open: true,
      nodeId: targetNode.id,
      name: (targetNode.data?.custom_label as string) || "",
    });
  }, [contextMenu.nodeId, selectionSummary.singleSelectedNodeId, getNodes]);

  const handleRenameSubmit = useCallback(
    (nodeId: string, name: string) => {
      const trimmed = name.trim();
      commit();
      setNodes((nds) =>
        nds.map((n) =>
          n.id === nodeId
            ? {
                ...n,
                data: {
                  ...n.data,
                  custom_label: trimmed || undefined,
                },
              }
            : n
        )
      );
      setRenameDialog(null);
    },
    [commit, setNodes]
  );

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Shift") {
        isShiftKeyDownRef.current = true;
      }
      if (e.key === "F2") {
        const activeEl = document.activeElement;
        if (
          activeEl?.tagName === "INPUT" ||
          activeEl?.tagName === "TEXTAREA" ||
          activeEl?.getAttribute("contenteditable") === "true"
        ) {
          return;
        }
        const currentNodes = getNodes();
        const selected = getActionableSelectedNodes(currentNodes);
        if (selected.length === 1) {
          e.preventDefault();
          const node = selected[0];
          setRenameDialog({
            open: true,
            nodeId: node.id,
            name: (node.data?.custom_label as string) || "",
          });
        }
      }
    };
    const onKeyUp = (e: KeyboardEvent) => {
      if (e.key === "Shift") {
        isShiftKeyDownRef.current = false;
      }
    };
    const onBlur = () => {
      isShiftKeyDownRef.current = false;
    };
    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("keyup", onKeyUp);
    window.addEventListener("blur", onBlur);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("keyup", onKeyUp);
      window.removeEventListener("blur", onBlur);
    };
  }, [getNodes]);

  const handleMouseDownCapture = useCallback(
    (e: ReactMouseEvent<HTMLDivElement>) => {
      if (isMobile) return;
      if (e.shiftKey && e.button === 0) {
        e.preventDefault();
        e.stopPropagation();
        return;
      }

      if (e.button !== 2) return;

      const target = e.target as HTMLElement | null;
      if (target?.closest("input, textarea, [contenteditable]")) return;

      marqueeCleanupRef.current?.();

      isRightClickingRef.current = true;
      setTimeout(() => {
        isRightClickingRef.current = false;
      }, 0);

      rightClickSelectedIdsRef.current = new Set(
        getNodes()
          .filter((n) => n.selected)
          .map((n) => n.id)
      );

      rightClickDragRef.current = {
        startX: e.clientX,
        startY: e.clientY,
        moved: false,
        isShiftKey: e.shiftKey,
      };

      const applyMarqueeSelection = (
        startX: number,
        startY: number,
        endX: number,
        endY: number,
        additive: boolean
      ) => {
        const minX = Math.min(startX, endX);
        const maxX = Math.max(startX, endX);
        const minY = Math.min(startY, endY);
        const maxY = Math.max(startY, endY);

        const p1 = screenToFlowPosition({ x: minX, y: minY });
        const p2 = screenToFlowPosition({ x: maxX, y: maxY });
        const box = {
          minX: Math.min(p1.x, p2.x),
          maxX: Math.max(p1.x, p2.x),
          minY: Math.min(p1.y, p2.y),
          maxY: Math.max(p1.y, p2.y),
        };

        const currentNodes = getNodes();
        const currentEdges = getEdges();
        const matchedTopLevelIds = new Set<string>();

        for (const n of currentNodes) {
          const w = n.measured?.width ?? n.width ?? 200;
          const h = n.measured?.height ?? n.height ?? 80;
          const right = n.position.x + w;
          const bottom = n.position.y + h;

          const marginX = Math.min(32, w * 0.35);
          const marginY = Math.min(24, h * 0.35);
          const centerX = n.position.x + w / 2;
          const centerY = n.position.y + h / 2;

          const centerInside =
            centerX >= box.minX &&
            centerX <= box.maxX &&
            centerY >= box.minY &&
            centerY <= box.maxY;

          const substantialOverlap =
            right - marginX >= box.minX &&
            n.position.x + marginX <= box.maxX &&
            bottom - marginY >= box.minY &&
            n.position.y + marginY <= box.maxY;

          if (centerInside || substantialOverlap) {
            const topId =
              getOwnerNodeId(n.id, currentNodes, currentEdges) ?? n.id;
            matchedTopLevelIds.add(topId);
          }
        }

        const seedIds = additive
          ? [...rightClickSelectedIdsRef.current, ...matchedTopLevelIds]
          : [...matchedTopLevelIds];
        const groupIds = withOwnedNodes(seedIds, currentNodes, currentEdges);

        setNodes((nds) => {
          let changedSelection = false;
          const next = nds.map((n) => {
            const shouldSelect = groupIds.has(n.id);
            if (!!n.selected !== shouldSelect) {
              changedSelection = true;
              return { ...n, selected: shouldSelect };
            }
            return n;
          });
          return changedSelection ? next : nds;
        });

        if (!additive) {
          setEdges((eds) =>
            eds.some((edge) => edge.selected)
              ? eds.map((edge) =>
                  edge.selected ? { ...edge, selected: false } : edge
                )
              : eds
          );
        }
      };

      const onWindowMouseMove = (moveEvent: MouseEvent) => {
        const drag = rightClickDragRef.current;
        if (!drag) return;
        const { startX, startY, moved } = drag;
        const dx = moveEvent.clientX - startX;
        const dy = moveEvent.clientY - startY;
        if (moved || Math.hypot(dx, dy) >= 6) {
          drag.moved = true;
          setMarqueeRect({
            startX,
            startY,
            currentX: moveEvent.clientX,
            currentY: moveEvent.clientY,
          });
          applyMarqueeSelection(
            startX,
            startY,
            moveEvent.clientX,
            moveEvent.clientY,
            moveEvent.shiftKey || !!drag.isShiftKey
          );
        }
      };

      const cleanupListeners = () => {
        window.removeEventListener("mousemove", onWindowMouseMove);
        window.removeEventListener("mouseup", onWindowMouseUp);
        marqueeCleanupRef.current = null;
      };

      const onWindowMouseUp = (upEvent: MouseEvent) => {
        if (upEvent.button !== 2) return;
        const drag = rightClickDragRef.current;
        cleanupListeners();
        rightClickDragRef.current = null;

        if (drag?.moved) {
          suppressNextContextMenuRef.current = true;
          applyMarqueeSelection(
            drag.startX,
            drag.startY,
            upEvent.clientX,
            upEvent.clientY,
            upEvent.shiftKey || !!drag.isShiftKey
          );
          setMarqueeRect(null);
          setTimeout(() => {
            suppressNextContextMenuRef.current = false;
          }, 100);
        } else {
          setMarqueeRect(null);
        }
      };

      marqueeCleanupRef.current = cleanupListeners;
      window.addEventListener("mousemove", onWindowMouseMove);
      window.addEventListener("mouseup", onWindowMouseUp);
    },
    [isMobile, getNodes, getEdges, screenToFlowPosition, setNodes, setEdges]
  );

  const handleNodeContextMenu = useCallback(
    (e: ReactMouseEvent, node: Node<NodeData>) => {
      e.preventDefault();
      e.stopPropagation();

      if (isMobile) {
        if (!mobileMoveMode && !longPressTriggeredRef.current) {
          longPressTriggeredRef.current = true;
          const selectedTarget = selectMobileTopLevelNode(
            node.id,
            mobileSelectMode
          );
          if (selectedTarget) {
            onEnterMobileSelectMode?.(selectedTarget);
          }
        }
        return;
      }

      if (suppressNextContextMenuRef.current) return;

      const currentNodes = getNodes();
      const currentEdges = getEdges();
      const targetId =
        getOwnerNodeId(node.id, currentNodes, currentEdges) ?? node.id;

      if (e.shiftKey) {
        const targetNode = currentNodes.find((n) => n.id === targetId);
        if (!targetNode) return;

        const groupIds = withOwnedNodes([targetId], currentNodes, currentEdges);
        const nextSelected = !targetNode.selected;
        const updatedNodes = currentNodes.map((n) =>
          groupIds.has(n.id) ? { ...n, selected: nextSelected } : n
        );
        setNodes(updatedNodes);

        if (nextSelected) {
          onNodeTap?.(node);
        }
        return;
      }

      ignoreContainerContextMenuRef.current = true;
      setTimeout(() => {
        ignoreContainerContextMenuRef.current = false;
      }, 0);

      const prevSelected = rightClickSelectedIdsRef.current;
      const isAlreadySelected =
        node.selected ||
        prevSelected.has(node.id) ||
        prevSelected.has(targetId);

      if (isAlreadySelected && prevSelected.size > 0) {
        const keepIds = withOwnedNodes(
          [...prevSelected],
          currentNodes,
          currentEdges
        );
        setNodes((nds) =>
          nds.map((n) =>
            !!n.selected === keepIds.has(n.id)
              ? n
              : { ...n, selected: keepIds.has(n.id) }
          )
        );
      } else {
        const groupIds = withOwnedNodes([targetId], currentNodes, currentEdges);
        setNodes((nds) =>
          nds.map((n) =>
            !!n.selected === groupIds.has(n.id)
              ? n
              : { ...n, selected: groupIds.has(n.id) }
          )
        );
        setEdges((eds) =>
          eds.some((edge) => edge.selected)
            ? eds.map((edge) =>
                edge.selected ? { ...edge, selected: false } : edge
              )
            : eds
        );
      }

      openContextMenu({
        x: e.clientX,
        y: e.clientY,
        kind: "node",
        nodeId: targetId,
      });
    },
    [
      isMobile,
      mobileMoveMode,
      mobileSelectMode,
      selectMobileTopLevelNode,
      onEnterMobileSelectMode,
      openContextMenu,
      getNodes,
      getEdges,
      setNodes,
      setEdges,
      onNodeTap,
    ]
  );

  const handleSelectionContextMenu = useCallback(
    (e: ReactMouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (isMobile) return;
      if (suppressNextContextMenuRef.current) return;

      if (e.shiftKey) {
        const target = e.target as HTMLElement | null;
        const targetNodeEl = target?.closest(".react-flow__node[data-id]");
        const targetNodeId = targetNodeEl?.getAttribute("data-id");
        if (targetNodeId) {
          const currentNodes = getNodes();
          const currentEdges = getEdges();
          const targetId =
            getOwnerNodeId(targetNodeId, currentNodes, currentEdges) ??
            targetNodeId;
          const targetNode = currentNodes.find((n) => n.id === targetId);
          if (targetNode) {
            const groupIds = withOwnedNodes(
              [targetId],
              currentNodes,
              currentEdges
            );
            const nextSelected = !targetNode.selected;
            const updatedNodes = currentNodes.map((n) =>
              groupIds.has(n.id) ? { ...n, selected: nextSelected } : n
            );
            setNodes(updatedNodes);
            if (nextSelected) {
              onNodeTap?.(targetNode);
            }
          }
        }
        return;
      }

      ignoreContainerContextMenuRef.current = true;
      setTimeout(() => {
        ignoreContainerContextMenuRef.current = false;
      }, 0);

      openContextMenu({
        x: e.clientX,
        y: e.clientY,
        kind: "node",
      });
    },
    [isMobile, openContextMenu, getNodes, getEdges, setNodes, onNodeTap]
  );

  const handlePaneContextMenu = useCallback(
    (e: ReactMouseEvent | MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (isMobile) return;
      if (e.shiftKey) return;
      if (suppressNextContextMenuRef.current) return;
      ignoreContainerContextMenuRef.current = true;
      setTimeout(() => {
        ignoreContainerContextMenuRef.current = false;
      }, 0);

      openContextMenu({
        x: e.clientX,
        y: e.clientY,
        kind: "pane",
      });
    },
    [isMobile, openContextMenu]
  );

  const handleEdgeContextMenu = useCallback(
    (e: ReactMouseEvent, edge: Edge) => {
      e.preventDefault();
      e.stopPropagation();
      if (isMobile) return;
      if (suppressNextContextMenuRef.current) return;

      const currentNodes = getNodes();
      if (isOwnedEdge(edge, currentNodes)) return;

      ignoreContainerContextMenuRef.current = true;
      setTimeout(() => {
        ignoreContainerContextMenuRef.current = false;
      }, 0);

      openContextMenu({
        x: e.clientX,
        y: e.clientY,
        kind: "edge",
        edgeId: edge.id,
        canInsertIntoEdge: canSpliceEdge(edge, currentNodes),
      });
    },
    [isMobile, openContextMenu, getNodes]
  );

  const handleNodeClick = useCallback(
    (e: ReactMouseEvent, node: Node<NodeData>) => {
      if (e.shiftKey) {
        e.preventDefault();
        e.stopPropagation();
        return;
      }

      closeContextMenu();

      if (longPressTriggeredRef.current) {
        longPressTriggeredRef.current = false;
        return;
      }

      if (isMobile && mobileMoveMode) {
        return;
      }

      if (isMobile && mobileSelectMode) {
        const currentNodes = getNodes();
        const currentEdges = getEdges();
        const targetId =
          getOwnerNodeId(node.id, currentNodes, currentEdges) ?? node.id;
        const targetNode = currentNodes.find((n) => n.id === targetId);
        if (!targetNode) return;

        const groupIds = withOwnedNodes([targetId], currentNodes, currentEdges);
        const nextSelected = !targetNode.selected;
        const updatedNodes = currentNodes.map((n) =>
          groupIds.has(n.id) ? { ...n, selected: nextSelected } : n
        );
        setNodes(updatedNodes);

        if (isMobile && getTopLevelSelectedNodes(updatedNodes).length === 0) {
          onExitMobileSelectMode?.();
        }

        if (nextSelected) {
          onNodeTap?.(node);
        }
        return;
      }

      onNodeTap?.(node);
    },
    [
      closeContextMenu,
      isMobile,
      mobileMoveMode,
      mobileSelectMode,
      getNodes,
      getEdges,
      setNodes,
      onExitMobileSelectMode,
      onNodeTap,
    ]
  );

  const handlePaneClick = useCallback(
    (e?: ReactMouseEvent) => {
      if (longPressTriggeredRef.current) {
        longPressTriggeredRef.current = false;
        return;
      }
      if (e?.shiftKey) {
        return;
      }
      closeContextMenu();
    },
    [closeContextMenu]
  );

  const handleClickCapture = useCallback(
    (e: ReactMouseEvent<HTMLDivElement>) => {
      if (isMobile && longPressTriggeredRef.current) {
        const target = e.target as HTMLElement | null;
        if (
          !target?.closest(".react-flow__node[data-id]") &&
          target?.closest(".react-flow__pane")
        ) {
          longPressTriggeredRef.current = false;
          e.preventDefault();
          e.stopPropagation();
          return;
        }
      }
      if (!isMobile && e.shiftKey && e.button === 0) {
        e.preventDefault();
        e.stopPropagation();
      }
    },
    [isMobile]
  );

  const handleContainerContextMenu = useCallback(
    (e: ReactMouseEvent<HTMLDivElement>) => {
      const target = e.target as HTMLElement | null;
      if (target?.closest("input, textarea, [contenteditable]")) {
        return;
      }

      e.preventDefault();

      if (isMobile) return;
      if (e.shiftKey) return;
      if (suppressNextContextMenuRef.current) return;
      if (ignoreContainerContextMenuRef.current) return;

      openContextMenu({
        x: e.clientX,
        y: e.clientY,
        kind: "pane",
      });
    },
    [isMobile, openContextMenu]
  );

  const handleEdgeClick = useCallback((_e: ReactMouseEvent, _edge: Edge) => {},
  []);

  const flowNodes = useMemo(() => {
    if (!isMobile) return nodes;
    if (!mobileMoveMode) {
      return nodes.map((n) =>
        n.draggable === false ? n : { ...n, draggable: false }
      );
    }
    const selectedIds = nodes.filter((n) => n.selected).map((n) => n.id);
    const movableIds = withOwnedNodes(selectedIds, nodes, edges);
    return nodes.map((n) => {
      const isMovable = movableIds.has(n.id);
      return n.draggable === isMovable ? n : { ...n, draggable: isMovable };
    });
  }, [isMobile, mobileMoveMode, nodes, edges]);

  return (
    <div
      className={cn(
        "w-full h-full",
        isMobile && !mobileMoveMode && "flow-mobile-pan-priority"
      )}
      onPointerDownCapture={handlePointerDown}
      onPointerMoveCapture={handlePointerMove}
      onPointerUpCapture={handlePointerUpOrCancel}
      onPointerCancelCapture={handlePointerUpOrCancel}
      onMouseDownCapture={handleMouseDownCapture}
      onClickCapture={handleClickCapture}
      onContextMenu={handleContainerContextMenu}
    >
      <style jsx>{`
        .flow-mobile-pan-priority :global(.react-flow__nodesselection-rect) {
          pointer-events: none !important;
        }
        :global(.react-flow__edge) {
          touch-action: none !important;
          pointer-events: visibleStroke !important;
        }
        :global(.react-flow__edge.inactive) {
          pointer-events: visibleStroke !important;
        }
        :global(.react-flow__edge-interaction) {
          stroke: #000000 !important;
          stroke-opacity: 0.0001 !important;
          stroke-width: 24px !important;
          pointer-events: stroke !important;
          touch-action: none !important;
          cursor: pointer;
        }
      `}</style>
      <ReactFlow
        nodes={flowNodes}
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
        onNodeClick={handleNodeClick}
        onEdgeClick={handleEdgeClick}
        onNodeContextMenu={handleNodeContextMenu}
        onSelectionContextMenu={handleSelectionContextMenu}
        onPaneContextMenu={handlePaneContextMenu}
        onEdgeContextMenu={handleEdgeContextMenu}
        onPaneClick={handlePaneClick}
        onMoveStart={() => {
          if (!longPressTriggeredRef.current) {
            closeContextMenu();
          }
        }}
        onMoveEnd={onMoveEnd}
        panOnDrag={isMobile ? undefined : [0, 1]}
        nodesDraggable={!isMobile || mobileMoveMode}
        elementsSelectable={!isMobile}
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
          showFitView={false}
          position="bottom-right"
          className="scale-110 !hidden md:!flex"
        >
          <ControlButton onClick={handleCenterFlow} title="Center flow">
            <FocusIcon className="size-5 !fill-none" />
          </ControlButton>
          <ControlButton onClick={undo} disabled={!canUndo} title="Undo">
            <Undo2Icon className="size-5 !fill-none" />
          </ControlButton>
          <ControlButton onClick={redo} disabled={!canRedo} title="Redo">
            <Redo2Icon className="size-5 !fill-none" />
          </ControlButton>
          <ControlButton onClick={format} title="Auto arrange layout">
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

      {marqueeRect && (
        <div
          className="pointer-events-none fixed z-50 rounded border border-primary/70 bg-primary/15 transition-none shadow-sm"
          style={{
            left: Math.min(marqueeRect.startX, marqueeRect.currentX),
            top: Math.min(marqueeRect.startY, marqueeRect.currentY),
            width: Math.abs(marqueeRect.currentX - marqueeRect.startX),
            height: Math.abs(marqueeRect.currentY - marqueeRect.startY),
          }}
        />
      )}

      <Dialog
        open={!!renameDialog?.open}
        onOpenChange={(open) => {
          if (!open) setRenameDialog(null);
        }}
      >
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Rename Block</DialogTitle>
            <DialogDescription>
              Enter a custom label for this block to easily identify it in the
              flow.
            </DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (renameDialog) {
                handleRenameSubmit(renameDialog.nodeId, renameDialog.name);
              }
            }}
            className="space-y-4 pt-2"
          >
            <Input
              value={renameDialog?.name ?? ""}
              onChange={(e) =>
                setRenameDialog((prev) =>
                  prev ? { ...prev, name: e.target.value } : null
                )
              }
              placeholder="Enter custom label (leave empty for default)..."
              autoFocus
            />
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setRenameDialog(null)}
              >
                Cancel
              </Button>
              <Button type="submit">Save</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <FlowInsertBlockDialog
        isMobile={isMobile}
        open={!!insertBlockEdgeId}
        onOpenChange={(open) => {
          if (!open) setInsertBlockEdgeId(null);
        }}
        onSelectBlock={(blockType) => {
          if (insertBlockEdgeId) {
            insertBlockIntoEdge(insertBlockEdgeId, blockType);
            setInsertBlockEdgeId(null);
          }
        }}
      />

      {(!isMobile || contextMenu.kind === "pane") && (
        <FlowContextMenu
          state={contextMenu}
          onClose={closeContextMenu}
          topLevelCount={selectionSummary.topLevelCount}
          actionableCount={selectionSummary.actionableCount}
          canDisconnect={selectionSummary.canDisconnect}
          hasClipboard={hasClipboard}
          canUndo={canUndo}
          canRedo={canRedo}
          onCopy={copy}
          onCut={cut}
          onDuplicate={duplicate}
          onDuplicateWithConnections={duplicateWithConnections}
          onRename={handleRenameClick}
          onDisconnect={disconnect}
          onDelete={deleteSelected}
          onPaste={paste}
          onUndo={undo}
          onRedo={redo}
          onFormat={format}
          onCenterFlow={handleCenterFlow}
          onResetZoom={resetZoom}
          onSelectAllInView={selectAllInView}
          onSelectAllInFlow={selectAllInFlow}
          onClearSelection={clearSelection}
          onDeleteEdge={deleteEdge}
          onOpenInsertDialog={(edgeId) => setInsertBlockEdgeId(edgeId)}
        />
      )}
    </div>
  );
}
