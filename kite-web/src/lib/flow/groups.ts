import { Edge, Node, XYPosition } from "@xyflow/react";
import { NodeData } from "./dataSchema";

// Groups are labeled frames around blocks, so big flows can be organized,
// moved in one piece and collapsed to a single bar. Like notes they're kept
// with the blocks, so they're saved, exported and shared with the flow, but
// the service never reaches them when it runs it. A group's blocks are the
// ones inside its frame, so blocks join and leave it by being moved.
export const groupNodeType = "flow_group";

export const groupHeaderHeight = 44;
export const groupPadding = 40;
export const groupMinWidth = 240;
export const groupMinHeight = 140;
export const groupMaxSize = 20000;
export const groupLabelMaxLength = 100;
export const defaultGroupSize = { width: 400, height: 260 };

const estimatedBlockSize = { width: 240, height: 64 };

// Selected nodes are raised by 1000, so groups stay below blocks and edges
// even when selected.
const groupZIndex = -2000;

export const groupColors = {
  blurple: "#5865f2",
  red: "#ef4444",
  orange: "#f59e0b",
  green: "#22c55e",
  cyan: "#06b6d4",
  purple: "#a855f7",
  pink: "#ec4899",
  gray: "#71717a",
} as const;

export type GroupColor = keyof typeof groupColors;

export const defaultGroupColor: GroupColor = "blurple";

export interface GroupData {
  label: string;
  color: string;
  collapsed: boolean;
  // The nodes hidden by the collapsed group. They are taken when it's
  // collapsed, so blocks moved over a collapsed group don't disappear.
  hiddenIds: string[];
}

export interface Rect {
  x: number;
  y: number;
  width: number;
  height: number;
}

export function isGroupNode(node: { type?: string }) {
  return node.type === groupNodeType;
}

export function getGroupData(node: { data?: NodeData }): GroupData {
  const data = (node.data ?? {}) as Record<string, unknown>;
  const color = data.group_color;
  return {
    label: typeof data.group_label === "string" ? data.group_label : "Group",
    color:
      typeof color === "string" && Object.hasOwn(groupColors, color)
        ? groupColors[color as GroupColor]
        : groupColors[defaultGroupColor],
    collapsed: data.group_collapsed === true,
    hiddenIds: Array.isArray(data.group_hidden_ids)
      ? data.group_hidden_ids.filter(
          (id): id is string => typeof id === "string"
        )
      : [],
  };
}

function clamp(value: unknown, min: number, fallback: number) {
  const n = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(n) || n <= 0) return fallback;
  return Math.round(Math.min(groupMaxSize, Math.max(min, n)));
}

// The size of the expanded group as it's saved. In the editor react-flow
// keeps it as the width and height of the node, which resizing changes and
// undo restores. A collapsed group is only as high as its header, and keeps
// its expanded height in its data.
function getStoredGroupSize(data: NodeData = {}) {
  return {
    width: clamp(data.group_width, groupMinWidth, defaultGroupSize.width),
    height: clamp(data.group_height, groupMinHeight, defaultGroupSize.height),
  };
}

function getNodeSize(node: Node) {
  return {
    width: node.width ?? node.measured?.width ?? 0,
    height: node.height ?? node.measured?.height ?? 0,
  };
}

// The bounds of a group as if it was expanded.
export function getGroupBounds(group: Node<NodeData>): Rect {
  const stored = getStoredGroupSize(group.data);
  const size = getNodeSize(group);
  return {
    x: group.position.x,
    y: group.position.y,
    width: size.width || stored.width,
    height: getGroupData(group).collapsed
      ? stored.height
      : size.height || stored.height,
  };
}

function getNodeRect(node: Node<NodeData>): Rect {
  if (isGroupNode(node)) return getGroupBounds(node);
  return { ...node.position, ...getNodeSize(node) };
}

function containsPoint(rect: Rect, point: XYPosition) {
  return (
    point.x >= rect.x &&
    point.x <= rect.x + rect.width &&
    point.y >= rect.y &&
    point.y <= rect.y + rect.height
  );
}

function containsRect(outer: Rect, inner: Rect) {
  return (
    inner.x >= outer.x &&
    inner.y >= outer.y &&
    inner.x + inner.width <= outer.x + outer.width &&
    inner.y + inner.height <= outer.y + outer.height
  );
}

// Blocks and notes are in a group when their center is, so they can stick out
// a bit. Groups are in a group when they are completely inside it, so two
// groups can never contain each other.
function isInGroupBounds(rect: Rect, node: Node<NodeData>) {
  if (isGroupNode(node)) return containsRect(rect, getGroupBounds(node));

  // Blocks hidden since the flow was loaded were never measured.
  const size = getNodeSize(node);
  const width = size.width || estimatedBlockSize.width;
  const height = size.height || estimatedBlockSize.height;
  return containsPoint(rect, {
    x: node.position.x + width / 2,
    y: node.position.y + height / 2,
  });
}

// The IDs of the nodes in a group, including the groups inside it and their
// nodes.
export function getGroupMemberIds(
  group: Node<NodeData>,
  nodes: Node<NodeData>[]
): string[] {
  const data = getGroupData(group);
  if (data.collapsed) {
    const ids = new Set(nodes.map((n) => n.id));
    return data.hiddenIds.filter((id) => id !== group.id && ids.has(id));
  }

  const rect = getGroupBounds(group);
  return nodes
    .filter((n) => n.id !== group.id && isInGroupBounds(rect, n))
    .map((n) => n.id);
}

// The given nodes plus the nodes in the groups among them, so a group is
// copied, cut and deleted together with its blocks.
export function withGroupMembers(
  ids: Iterable<string>,
  nodes: Node<NodeData>[]
): string[] {
  const res = new Set(ids);
  for (const node of nodes) {
    if (res.has(node.id) && isGroupNode(node)) {
      getGroupMemberIds(node, nodes).forEach((id) => res.add(id));
    }
  }
  return [...res];
}

// The nodes hidden by collapsed groups. A collapsed group inside a collapsed
// group is hidden too, and keeps its own nodes hidden when the outer one is
// expanded.
export function getHiddenNodeIds(nodes: Node<NodeData>[]): Set<string> {
  const hidden = new Set<string>();
  for (const node of nodes) {
    if (!isGroupNode(node)) continue;
    const data = getGroupData(node);
    if (!data.collapsed) continue;
    for (const id of data.hiddenIds) {
      if (id !== node.id) hidden.add(id);
    }
  }
  return hidden;
}

// The visible collapsed group each hidden node is drawn as.
export function getHiddenNodeOwners(
  nodes: Node<NodeData>[],
  hidden: Set<string> = getHiddenNodeIds(nodes)
): Map<string, string> {
  const owners = new Map<string, string>();
  for (const node of nodes) {
    if (!isGroupNode(node) || hidden.has(node.id)) continue;
    const data = getGroupData(node);
    if (!data.collapsed) continue;
    for (const id of data.hiddenIds) {
      if (!owners.has(id)) owners.set(id, node.id);
    }
  }
  return owners;
}

const hiddenClones = new WeakMap<object, Node>();
const shownClones = new WeakMap<object, Node>();

// Sets hidden on the nodes of collapsed groups and clears it on the rest. This
// is derived from the groups, so undo, redo and AI edits can't leave blocks
// hidden by mistake. The copies are kept, so nodes that stay hidden don't
// change on every edit.
export function applyGroupVisibility<T extends Node<NodeData>>(
  nodes: T[]
): T[] {
  const hidden = getHiddenNodeIds(nodes);
  if (hidden.size === 0 && !nodes.some((n) => n.hidden)) return nodes;

  return nodes.map((node) => {
    const shouldHide = hidden.has(node.id);
    if (!!node.hidden === shouldHide) return node;

    const clones = shouldHide ? hiddenClones : shownClones;
    let clone = clones.get(node);
    if (!clone) {
      clone = { ...node, hidden: shouldHide };
      clones.set(node, clone);
    }
    return clone as T;
  });
}

// Collapsing hides the nodes in the group and shrinks it to its header.
// Hidden nodes are deselected, so deleting the selection can't remove them
// unseen.
export function setGroupCollapsed<T extends Node<NodeData>>(
  nodes: T[],
  groupId: string,
  collapsed: boolean
): T[] {
  const group = nodes.find((n) => n.id === groupId);
  if (!group || !isGroupNode(group)) return nodes;

  const data = getGroupData(group);
  if (data.collapsed === collapsed) return nodes;

  const bounds = getGroupBounds(group);
  if (collapsed) {
    const memberIds = getGroupMemberIds(group, nodes);
    const members = new Set(memberIds);

    return nodes.map((n) => {
      if (n.id === groupId) {
        return {
          ...n,
          width: bounds.width,
          height: groupHeaderHeight,
          data: {
            ...n.data,
            group_collapsed: true,
            group_width: bounds.width,
            group_height: bounds.height,
            group_hidden_ids: memberIds,
          },
        };
      }
      return members.has(n.id) && n.selected ? { ...n, selected: false } : n;
    });
  }

  return nodes.map((n) => {
    if (n.id !== groupId) return n;
    const {
      group_collapsed: _collapsed,
      group_hidden_ids: _hiddenIds,
      ...rest
    } = n.data as Record<string, unknown>;
    return {
      ...n,
      width: bounds.width,
      height: bounds.height,
      data: rest as NodeData,
    };
  });
}

let groupIdCounter = 0;

function getGroupId(taken: Set<string>) {
  const chars = "abcdefghijklmnopqrstuvwxyz0123456789";
  for (;;) {
    let suffix = "";
    for (let i = 0; i < 3; i++) {
      suffix += chars[Math.floor(Math.random() * chars.length)];
    }
    // Falls back to a counter in the unlikely case all short IDs are taken.
    const id =
      taken.size > 30000 ? `group_${++groupIdCounter}` : `group_${suffix}`;
    if (!taken.has(id)) return id;
  }
}

export function createGroupNode(
  rect: Rect,
  existingIds: Iterable<string> = [],
  label = "Group"
): Node<NodeData> {
  const width = Math.round(Math.max(rect.width, groupMinWidth));
  const height = Math.round(Math.max(rect.height, groupMinHeight));
  return withGroupStyle({
    id: getGroupId(new Set(existingIds)),
    type: groupNodeType,
    position: { x: Math.round(rect.x), y: Math.round(rect.y) },
    width,
    height,
    data: {
      group_label: label,
      group_color: defaultGroupColor,
      group_width: width,
      group_height: height,
    },
  });
}

function withGroupStyle<T extends Node<NodeData>>(node: T): T {
  return { ...node, zIndex: groupZIndex, className: "flow-group" };
}

// The bounds of a group around the given nodes, with room for its header.
export function getGroupRectForNodes(nodes: Node<NodeData>[]): Rect | null {
  if (nodes.length === 0) return null;

  const rects = nodes.map(getNodeRect);
  const minX = Math.min(...rects.map((r) => r.x));
  const minY = Math.min(...rects.map((r) => r.y));
  const maxX = Math.max(...rects.map((r) => r.x + r.width));
  const maxY = Math.max(...rects.map((r) => r.y + r.height));

  return {
    x: minX - groupPadding,
    y: minY - groupPadding - groupHeaderHeight,
    width: maxX - minX + groupPadding * 2,
    height: maxY - minY + groupPadding * 2 + groupHeaderHeight,
  };
}

// Groups the selected nodes, or returns null if there is nothing to group.
// The new group goes first, so it's drawn below the groups inside it.
export function groupSelectedNodes<T extends Node<NodeData>>(
  nodes: T[]
): { nodes: T[]; group: Node<NodeData> } | null {
  const selected = nodes.filter((n) => n.selected && !n.hidden);
  const rect = getGroupRectForNodes(selected);
  if (!rect) return null;

  const group = {
    ...createGroupNode(
      rect,
      nodes.map((n) => n.id)
    ),
    selected: true,
  };
  return {
    group,
    nodes: [
      group as T,
      ...nodes.map((n) => (n.selected ? { ...n, selected: false } : n)),
    ],
  };
}

// The nodes the editor starts with. Groups get their saved size and are drawn
// below the blocks, and collapsed groups take the nodes inside them. Groups
// that already have a size in the editor keep it, e.g. when the flow AI's
// edits are applied.
export function withGroupSizes<N extends Node<NodeData>>(nodes: N[]): N[] {
  if (!nodes.some(isGroupNode)) return nodes;

  const sized = nodes.map((node) => {
    if (!isGroupNode(node)) return node;
    if (node.width && node.height) return withGroupStyle(node);
    const stored = getStoredGroupSize(node.data);
    return withGroupStyle({
      ...node,
      width: stored.width,
      height: stored.height,
    });
  });

  // Groups go first, so the outer ones are drawn below the inner ones.
  const ordered = [
    ...sized
      .filter(isGroupNode)
      .sort(
        (a, b) =>
          b.width! * b.height! - a.width! * a.height! ||
          sized.indexOf(a) - sized.indexOf(b)
      ),
    ...sized.filter((n) => !isGroupNode(n)),
  ];

  return ordered.map((node) => {
    if (!isGroupNode(node)) return node;
    const data = getGroupData(node);
    if (!data.collapsed || Array.isArray(node.data.group_hidden_ids)) {
      return node;
    }
    // Hidden IDs aren't saved, so a collapsed group that's loaded takes
    // what's inside its frame.
    const expanded = {
      ...node,
      data: { ...node.data, group_collapsed: false },
    };
    return {
      ...node,
      height: groupHeaderHeight,
      data: {
        ...node.data,
        group_hidden_ids: getGroupMemberIds(expanded, ordered),
      },
    };
  });
}

// storeGroupSizes writes the size groups have in the editor into their data,
// so it's saved and copied with them, and drops the editor-only state.
export function storeGroupSizes<N extends Node<NodeData>>(nodes: N[]): N[] {
  return nodes.map((node) => {
    if (!isGroupNode(node)) {
      if (node.hidden === undefined) return node;
      const { hidden: _hidden, ...rest } = node;
      return rest as N;
    }

    const bounds = getGroupBounds(node);
    const { group_hidden_ids: _hiddenIds, ...data } = node.data as Record<
      string,
      unknown
    >;
    const {
      zIndex: _zIndex,
      className: _className,
      hidden: _hidden,
      ...rest
    } = node;
    return {
      ...rest,
      data: {
        ...data,
        group_width: Math.round(bounds.width),
        group_height: Math.round(bounds.height),
      } as NodeData,
    } as N;
  });
}

// Copies of groups are pasted expanded, with the blocks that were in them.
export function expandGroupsForCopy<N extends Node<NodeData>>(nodes: N[]): N[] {
  return nodes.map((node) => {
    if (!isGroupNode(node) || !getGroupData(node).collapsed) return node;
    const { group_collapsed: _collapsed, ...data } = node.data as Record<
      string,
      unknown
    >;
    return { ...node, data: data as NodeData };
  });
}

export interface GroupProxyEdge {
  id: string;
  source: string;
  sourceHandle: string | null;
  target: string;
  targetHandle: string | null;
}

// The connections of the hidden nodes, drawn to the collapsed groups they are
// in instead. Connections inside a collapsed group aren't drawn.
export function getGroupProxyEdges(
  nodes: Node<NodeData>[],
  edges: Edge[]
): GroupProxyEdge[] {
  const hidden = getHiddenNodeIds(nodes);
  if (hidden.size === 0) return [];

  const owners = getHiddenNodeOwners(nodes, hidden);
  const res = new Map<string, GroupProxyEdge>();

  for (const edge of edges) {
    const sourceHidden = hidden.has(edge.source);
    const targetHidden = hidden.has(edge.target);
    if (!sourceHidden && !targetHidden) continue;

    const source = sourceHidden ? owners.get(edge.source) : edge.source;
    const target = targetHidden ? owners.get(edge.target) : edge.target;
    if (!source || !target || source === target) continue;

    const sourceHandle = sourceHidden ? null : edge.sourceHandle ?? null;
    const targetHandle = targetHidden ? null : edge.targetHandle ?? null;
    const id = `${source}:${sourceHandle ?? ""}->${target}:${
      targetHandle ?? ""
    }`;
    if (!res.has(id)) {
      res.set(id, { id, source, sourceHandle, target, targetHandle });
    }
  }

  return [...res.values()];
}

// Fits an expanded group around the nodes in it, or returns the nodes as they
// are if it's empty.
export function fitGroupToMembers<T extends Node<NodeData>>(
  nodes: T[],
  groupId: string
): T[] {
  const group = nodes.find((n) => n.id === groupId);
  if (!group || !isGroupNode(group) || getGroupData(group).collapsed) {
    return nodes;
  }

  const members = new Set(getGroupMemberIds(group, nodes));
  const rect = getGroupRectForNodes(nodes.filter((n) => members.has(n.id)));
  if (!rect) return nodes;

  const width = Math.round(Math.max(rect.width, groupMinWidth));
  const height = Math.round(Math.max(rect.height, groupMinHeight));
  return nodes.map((n) =>
    n.id === groupId
      ? {
          ...n,
          position: { x: Math.round(rect.x), y: Math.round(rect.y) },
          width,
          height,
        }
      : n
  );
}
