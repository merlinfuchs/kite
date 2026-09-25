import { Edge, EdgeChange, Node, NodeChange } from "@xyflow/react";
import { NodeData } from "./dataSchema";

export interface FlowSnapshot {
  nodes: Node<NodeData>[];
  edges: Edge[];
}

export interface FlowHistory {
  past: FlowSnapshot[];
  future: FlowSnapshot[];
}

export const emptyFlowHistory: FlowHistory = { past: [], future: [] };

const historyLimit = 100;

// Records the state from before an edit. A new edit makes the undone states
// unreachable, so they are dropped.
export function pushFlowHistory(
  history: FlowHistory,
  snapshot: FlowSnapshot
): FlowHistory {
  return {
    past: [...history.past, snapshot].slice(-historyLimit),
    future: [],
  };
}

// Returns the new history and the state to restore, or null if there is
// nothing to undo.
export function undoFlowHistory(
  history: FlowHistory,
  current: FlowSnapshot
): [FlowHistory, FlowSnapshot] | null {
  const previous = history.past.at(-1);
  if (!previous) return null;

  return [
    { past: history.past.slice(0, -1), future: [...history.future, current] },
    previous,
  ];
}

export function redoFlowHistory(
  history: FlowHistory,
  current: FlowSnapshot
): [FlowHistory, FlowSnapshot] | null {
  const next = history.future.at(-1);
  if (!next) return null;

  return [
    { past: [...history.past, current], future: history.future.slice(0, -1) },
    next,
  ];
}

// Consecutive edits with the same key are merged into one undo step, so typing
// in the block settings or nudging blocks with the arrow keys isn't undone one
// keystroke at a time.
export function getFlowMergeKey(
  changes: NodeChange<Node<NodeData>>[]
): string | undefined {
  if (changes.every((c) => c.type === "position")) return "move";
  if (changes.length === 1 && changes[0].type === "replace") {
    return `data:${changes[0].id}`;
  }
}

export type FlowChangeKind = "edit" | "drag" | "ignore";

// Selecting and measuring nodes doesn't change the flow. Positions only count
// as an edit once a drag has ended, or when nodes are moved with the keyboard.
export function getFlowChangeKind(
  change: NodeChange<Node<NodeData>> | EdgeChange
): FlowChangeKind {
  switch (change.type) {
    case "select":
    case "dimensions":
      return "ignore";
    case "position":
      return change.dragging ? "drag" : "edit";
    default:
      return "edit";
  }
}
