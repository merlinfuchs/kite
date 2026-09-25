import { NodeData } from "@/lib/flow/dataSchema";
import {
  emptyFlowHistory,
  FlowHistory,
  FlowSnapshot,
  pushFlowHistory,
  redoFlowHistory,
  undoFlowHistory,
} from "@/lib/flow/history";
import { Edge, Node } from "@xyflow/react";
import {
  Dispatch,
  RefObject,
  SetStateAction,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";

const mergeWindowMs = 1000;

// Undo and redo for the flow editor. record() and commit() must be called
// before an edit is applied, so the current nodes and edges are the state to
// go back to. The history only lives in memory and is kept across saves.
export function useFlowHistory({
  nodes,
  edges,
  setNodes,
  setEdges,
  onChange,
  containerRef,
}: {
  nodes: Node<NodeData>[];
  edges: Edge[];
  setNodes: Dispatch<SetStateAction<Node<NodeData>[]>>;
  setEdges: Dispatch<SetStateAction<Edge[]>>;
  onChange: () => void;
  containerRef: RefObject<HTMLElement>;
}) {
  const current = useRef<FlowSnapshot>({ nodes, edges });
  current.current = { nodes, edges };

  // The ref is read by the callbacks below, the state re-renders the buttons.
  const history = useRef<FlowHistory>(emptyFlowHistory);
  const [{ past, future }, setHistoryState] = useState(emptyFlowHistory);

  const setHistory = useCallback((next: FlowHistory) => {
    history.current = next;
    setHistoryState(next);
  }, []);

  const lastRecord = useRef<{ key?: string; time: number } | null>(null);
  const recordedThisTick = useRef(false);

  const record = useCallback(
    (mergeKey?: string) => {
      // A single action can be reported as several changes, e.g. a deleted
      // condition removes its items and edges too. They all make one step.
      if (recordedThisTick.current) return;

      const now = Date.now();
      const last = lastRecord.current;
      lastRecord.current = { key: mergeKey, time: now };
      if (
        mergeKey &&
        last?.key === mergeKey &&
        now - last.time < mergeWindowMs
      ) {
        return;
      }

      recordedThisTick.current = true;
      queueMicrotask(() => {
        recordedThisTick.current = false;
      });

      setHistory(pushFlowHistory(history.current, current.current));
    },
    [setHistory]
  );

  const commit = useCallback(
    (mergeKey?: string) => {
      record(mergeKey);
      onChange();
    },
    [record, onChange]
  );

  const restore = useCallback(
    (result: [FlowHistory, FlowSnapshot] | null) => {
      if (!result) return;

      const [next, snapshot] = result;
      setHistory(next);
      lastRecord.current = null;

      // Keep the current selection so the open block settings don't close.
      const selected = new Set(
        current.current.nodes.filter((n) => n.selected).map((n) => n.id)
      );
      setNodes(
        snapshot.nodes.map((n) =>
          !!n.selected === selected.has(n.id)
            ? n
            : { ...n, selected: selected.has(n.id) }
        )
      );
      setEdges(
        snapshot.edges.map((e) => (e.selected ? { ...e, selected: false } : e))
      );
      onChange();
    },
    [setHistory, setNodes, setEdges, onChange]
  );

  const undo = useCallback(
    () => restore(undoFlowHistory(history.current, current.current)),
    [restore]
  );
  const redo = useCallback(
    () => restore(redoFlowHistory(history.current, current.current)),
    [restore]
  );

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (!(e.metaKey || e.ctrlKey) || e.altKey) return;

      const key = e.key.toLowerCase();
      const isUndo = key === "z" && !e.shiftKey;
      const isRedo =
        (key === "z" && e.shiftKey) ||
        (key === "y" && e.ctrlKey && !e.shiftKey);
      if (!isUndo && !isRedo) return;

      // Text fields keep the browser's own undo, and dialogs opened on top of
      // the flow (e.g. the message editor) handle the shortcut themselves.
      const target = e.target as HTMLElement;
      if (target.closest("input, textarea, [contenteditable]")) return;
      const dialog = target.closest("[role=dialog]");
      if (dialog && !dialog.contains(containerRef.current)) return;

      e.preventDefault();
      // Keeps the message editor from also undoing when the flow of a button
      // is edited inside it.
      e.stopPropagation();
      isUndo ? undo() : redo();
    }

    document.addEventListener("keydown", onKeyDown, true);
    return () => document.removeEventListener("keydown", onKeyDown, true);
  }, [containerRef, undo, redo]);

  return {
    record,
    commit,
    undo,
    redo,
    canUndo: past.length > 0,
    canRedo: future.length > 0,
  };
}
