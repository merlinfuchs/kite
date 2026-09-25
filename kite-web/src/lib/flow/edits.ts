import { Edge, Node, XYPosition } from "@xyflow/react";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import {
  canConnect,
  createEdge,
  createNode,
  getNodeId,
  getNodeValues,
  getOwnedChildTypes,
  getOwnerTypes,
  isKnownNodeType,
  normalizeHandle,
  withOwnedNodes,
} from "./nodes";
import { walkDownstream } from "./placeholders";
import { FlowIssue, validateFlow } from "./validate";

// A change to a flow, as produced by the LLM flow editor. Blocks are referred
// to by their ID, or by the ref of a block added earlier in the same batch,
// e.g. "$check". Conditions and loops also create the blocks they own, which
// are referred to as "$check.item0", "$check.else", "$loop.each" and
// "$loop.end".
export type FlowEdit =
  | {
      op: "add_node";
      ref: string;
      type: string;
      data?: NodeData;
      // The block and output the new block is connected after. With before,
      // the new block is put into that existing connection.
      after?: string;
      handle?: string;
      before?: string;
      // The settings of each branch of a new condition.
      items?: NodeData[];
    }
  // Objects in data are merged into the current settings, null removes one.
  | { op: "update_node"; id: string; data: Record<string, unknown> }
  | { op: "remove_node"; id: string; reconnect?: boolean }
  | { op: "connect"; source: string; target: string; handle?: string }
  | { op: "disconnect"; source: string; target: string; handle?: string };

export interface FlowEditResult {
  nodes: Node<NodeData>[];
  edges: Edge[];
  // The IDs of the blocks added by the batch, by their ref.
  refs: Record<string, string>;
  issues: FlowIssue[];
}

const verticalSpacing = 250;
const horizontalSpacing = 450;

// Applies a batch of edits and validates the result. Edits that can't be
// applied are skipped and reported, so all problems can be fixed at once.
export function applyFlowEdits(
  flow: { nodes: Node<NodeData>[]; edges: Edge[] },
  edits: FlowEdit[],
  context: FlowContextType
): FlowEditResult {
  let nodes = [...flow.nodes];
  let edges = [...flow.edges];
  const refs: Record<string, string> = {};
  const issues: FlowIssue[] = [];
  const added = new Set<string>();

  const getNode = (id: string) => {
    const resolved = id.startsWith("$") ? refs[id] : id;
    const node = nodes.find((n) => n.id === resolved);
    if (!node) throw new Error(`There is no block '${id}'.`);
    return node;
  };
  const findEdges = (source: string, target: string, handle?: string | null) =>
    edges.filter(
      (e) =>
        e.source === source &&
        e.target === target &&
        normalizeHandle(e.sourceHandle) === normalizeHandle(handle)
    );
  const connect = (
    source: Node<NodeData>,
    target: Node<NodeData>,
    handle?: string | null
  ) => {
    if (findEdges(source.id, target.id, handle).length === 0) {
      edges.push(createEdge(source, target, handle));
    }
  };

  // Adds a block with the blocks it owns. Generated IDs can collide with
  // existing ones, so those are generated again.
  const addNodes = (type: string, data: NodeData) => {
    const [newNodes, newEdges] = createNode(type, { x: 0, y: 0 }, { data });
    const taken = new Set(nodes.map((n) => n.id));
    const ids = new Map<string, string>();
    for (const node of newNodes) {
      let id = node.id;
      while (taken.has(id)) id = getNodeId();
      taken.add(id);
      ids.set(node.id, id);
      added.add(id);
    }
    const res = newNodes.map((n) => ({ ...n, id: ids.get(n.id)! }));
    nodes.push(...res);
    edges.push(
      ...newEdges.map((e) => ({
        ...e,
        source: ids.get(e.source)!,
        target: ids.get(e.target)!,
      }))
    );
    return res;
  };

  edits.forEach((edit, i) => {
    try {
      switch (edit.op) {
        case "add_node": {
          if (!isKnownNodeType(edit.type)) {
            throw new Error(`Unknown block type '${edit.type}'.`);
          }
          if (!/^\$[A-Za-z0-9_]+$/.test(edit.ref) || refs[edit.ref]) {
            throw new Error(
              `The ref '${edit.ref}' must be unique and look like '$name', with only letters, numbers and underscores.`
            );
          }
          if (getNodeValues(edit.type).fixed) {
            throw new Error(
              `'${edit.type}' is created together with the block it belongs to.`
            );
          }

          // Checked before anything is added, so a failed edit changes
          // nothing.
          const itemType = getOwnedChildTypes(edit.type).find(
            (t) => !getNodeValues(t).fixed
          );
          if (
            edit.items !== undefined &&
            (!itemType ||
              !Array.isArray(edit.items) ||
              !edit.items.every(isPlainObject))
          ) {
            throw new Error(
              "items must be a list of branch settings, and only conditions have branches."
            );
          }

          const isOption = edit.type.startsWith("option_");
          const entry = nodes.find((n) => canConnect(edit.type, n.type!));
          if (isOption && !entry) {
            throw new Error(
              "The flow has no entry block options can connect to."
            );
          }
          if (isOption && (edit.after || edit.before || edit.handle)) {
            throw new Error(
              "Options are always connected to the entry block, so leave out after, before and handle."
            );
          }

          const after = edit.after ? getNode(edit.after) : undefined;
          const before = edit.before ? getNode(edit.before) : undefined;
          if (before && getOwnerTypes(before.type!).length > 0) {
            throw new Error(
              `'${edit.before}' belongs to another block, so nothing can be put in front of it. Add the block after it instead.`
            );
          }
          if (before && getNodeValues(edit.type).outputs?.length === 0) {
            throw new Error(
              `'${edit.type}' has no outputs, so nothing can come after it. Connect '${edit.before}' to one of its branches instead.`
            );
          }
          const split =
            after && before ? findEdges(after.id, before.id, edit.handle) : [];
          if (after && before && split.length === 0) {
            throw new Error(
              `'${edit.after}' isn't connected to '${edit.before}'.`
            );
          }

          const [owner, ...owned] = addNodes(edit.type, { ...edit.data });
          refs[edit.ref] = owner.id;

          // A new condition comes with one empty branch, which is replaced
          // by the given items.
          let items = owned.filter((n) => n.type === itemType);
          if (itemType && edit.items) {
            nodes = nodes.filter((n) => !items.includes(n));
            edges = edges.filter((e) => !items.some((n) => n.id === e.target));
            items = edit.items.map((data) => {
              const [item] = addNodes(itemType, { ...data });
              connect(owner, item);
              return item;
            });
          }
          items.forEach((item, j) => (refs[`${edit.ref}.item${j}`] = item.id));

          // The else branch and the each and end blocks of loops are named
          // after the last part of their type.
          for (const child of owned) {
            if (getNodeValues(child.type!).fixed) {
              refs[`${edit.ref}.${child.type!.split("_").pop()}`] = child.id;
            }
          }

          if (isOption) {
            connect(owner, entry!);
            break;
          }

          edges = edges.filter((e) => !split.includes(e));
          if (after) connect(after, owner, edit.handle);
          if (before) connect(owner, before);
          break;
        }
        case "update_node": {
          const id = getNode(edit.id).id;
          nodes = nodes.map((n) =>
            n.id === id ? { ...n, data: mergeData(n.data, edit.data) } : n
          );
          break;
        }
        case "remove_node": {
          const node = getNode(edit.id);
          if (getNodeValues(node.type!).fixed) {
            throw new Error(
              `'${edit.id}' can't be removed on its own. Remove the block it belongs to instead.`
            );
          }

          // The blocks a condition or loop owns go with it.
          const removed = withOwnedNodes([node.id], nodes, edges);
          const ownedTypes = getOwnedChildTypes(node.type!);

          // Blocks after the removed one's default output are connected to
          // the block before it, unless it was part of a condition or loop.
          // Blocks after its other outputs, e.g. an error branch or a button,
          // only ran in a different situation, so they are left unconnected.
          const reconnect =
            (edit.reconnect ?? true) &&
            ownedTypes.length === 0 &&
            getOwnerTypes(node.type!).length === 0;
          const parentEdges = edges.filter(
            (e) => e.target === node.id && e.source !== node.id
          );
          const childIds = edges
            .filter(
              (e) =>
                e.source === node.id &&
                e.target !== node.id &&
                !normalizeHandle(e.sourceHandle)
            )
            .map((e) => e.target);

          nodes = nodes.filter((n) => !removed.has(n.id));
          edges = edges.filter(
            (e) => !removed.has(e.source) && !removed.has(e.target)
          );
          if (reconnect) {
            for (const parentEdge of parentEdges) {
              const parent = nodes.find((n) => n.id === parentEdge.source);
              for (const childId of childIds) {
                const child = nodes.find((n) => n.id === childId);
                if (parent && child) {
                  connect(parent, child, parentEdge.sourceHandle);
                }
              }
            }
          }
          break;
        }
        case "connect": {
          connect(getNode(edit.source), getNode(edit.target), edit.handle);
          break;
        }
        case "disconnect": {
          const source = getNode(edit.source);
          const target = getNode(edit.target);
          if (getOwnedChildTypes(source.type!).includes(target.type!)) {
            throw new Error(
              `'${edit.target}' belongs to '${edit.source}' and can't be disconnected. Remove it instead.`
            );
          }
          const matching = findEdges(source.id, target.id, edit.handle);
          if (matching.length === 0) {
            throw new Error(
              `'${edit.source}' isn't connected to '${edit.target}'.`
            );
          }
          edges = edges.filter((e) => !matching.includes(e));
          break;
        }
      }
    } catch (err) {
      issues.push({
        severity: "error",
        message: `Edit ${i + 1} (${edit.op}): ${(err as Error).message}`,
      });
    }
  });

  nodes = layoutAddedNodes(nodes, edges, added);

  return {
    nodes,
    edges,
    refs,
    issues: [...issues, ...validateFlow(nodes, edges, context)],
  };
}

// Merges partial settings into a block's settings. Objects are merged, other
// values replaced, and null removes a setting.
function mergeData(
  data: Record<string, unknown>,
  patch: Record<string, unknown>
): NodeData {
  const res: Record<string, unknown> = { ...data };
  for (const [key, value] of Object.entries(patch)) {
    const current = res[key];
    if (value === null) {
      delete res[key];
    } else if (isPlainObject(value) && isPlainObject(current)) {
      res[key] = mergeData(
        current as Record<string, unknown>,
        value as Record<string, unknown>
      );
    } else {
      res[key] = value;
    }
  }
  return res as NodeData;
}

function isPlainObject(value: unknown) {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// Places added blocks below the block they are connected after, next to
// their new siblings, and moves the existing blocks after an inserted block
// further down to make room. Unlike the editor's format button, which lays out
// the whole flow with dagre, this keeps every existing block where the user
// put it.
function layoutAddedNodes(
  nodes: Node<NodeData>[],
  edges: Edge[],
  added: Set<string>
) {
  const existing = new Map<string, XYPosition>(
    nodes.filter((n) => !added.has(n.id)).map((n) => [n.id, n.position])
  );

  // Existing blocks that come right after an added one are pushed below it,
  // together with everything after them, and the added blocks are placed
  // again against the moved ones.
  let positions = placeAddedNodes(nodes, edges, added, existing);
  let shift = 0;
  const pushed = new Set<string>();
  for (const edge of edges) {
    if (added.has(edge.source) && existing.has(edge.target)) {
      const source = nodes.find((n) => n.id === edge.source)!;
      if (source.type!.startsWith("option_")) continue;

      pushed.add(edge.target);
      shift = Math.max(
        shift,
        positions.get(edge.source)!.y +
          verticalSpacing -
          existing.get(edge.target)!.y
      );
    }
  }
  if (shift > 0) {
    const moved = new Set([...pushed, ...walkDownstream([...pushed], edges)]);
    for (const id of moved) {
      const position = existing.get(id);
      if (position) existing.set(id, { x: position.x, y: position.y + shift });
    }
    positions = placeAddedNodes(nodes, edges, added, existing);
  }

  return nodes.map((n) => {
    const position = positions.get(n.id)!;
    return position === n.position ? n : { ...n, position };
  });
}

// Places each added block below the block before it, or above the block
// after it if there is none, e.g. for options. Added blocks wait until the
// block they are placed against has a position.
function placeAddedNodes(
  nodes: Node<NodeData>[],
  edges: Edge[],
  added: Set<string>,
  existing: Map<string, XYPosition>
) {
  const positions = new Map(existing);
  const siblings = new Map<string, number>();

  // New blocks go right of the existing blocks around the same anchor. The
  // branch that runs last goes rightmost, like in the editor.
  const siblingStart = (anchor: string, below: boolean) =>
    edges.filter((e) =>
      below
        ? e.source === anchor && !added.has(e.target)
        : e.target === anchor && !added.has(e.source)
    ).length;
  const runsLast = (n: Node<NodeData>) =>
    n.type === "control_condition_item_else" || n.type === "control_loop_end"
      ? 1
      : 0;
  let pending = nodes
    .filter((n) => added.has(n.id))
    .sort((a, b) => runsLast(a) - runsLast(b));
  while (pending.length > 0) {
    const next: Node<NodeData>[] = [];
    for (const node of pending) {
      const parent = edges.find((e) => e.target === node.id)?.source;
      const anchor = parent ?? edges.find((e) => e.source === node.id)?.target;
      const anchorPosition = anchor ? positions.get(anchor) : undefined;
      if (anchor && !anchorPosition) {
        next.push(node);
        continue;
      }

      const below = !!parent || !anchor;
      const key = `${anchor}:${below}`;
      const index =
        siblings.get(key) ?? (anchor ? siblingStart(anchor, below) : 0);
      siblings.set(key, index + 1);

      const base = anchorPosition ?? { x: 0, y: 0 };
      positions.set(node.id, {
        x: base.x + index * horizontalSpacing,
        y: base.y + (below ? verticalSpacing : -verticalSpacing),
      });
    }

    // Blocks that wait on each other end up at the origin.
    if (next.length === pending.length) {
      next.forEach((n) => positions.set(n.id, { x: 0, y: 0 }));
      break;
    }
    pending = next;
  }

  return positions;
}
