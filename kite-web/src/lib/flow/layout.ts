import { Edge, Node } from "@xyflow/react";
import { NodeData } from "./dataSchema";
import { getNodeValues, getOwnedChildTypes, normalizeHandle } from "./nodes";
import { collectComponentGroups, getComponentHandleIds } from "./resume";
import {
  getGroupBounds,
  getGroupData,
  getGroupMemberIds,
  getHiddenNodeIds,
  getHiddenNodeOwners,
  groupHeaderHeight,
  groupMinHeight,
  groupMinWidth,
  groupNodeType,
  groupPadding,
  isGroupNode,
} from "./groups";
import { isNoteNode } from "./notes";

// Space between blocks next to each other, between rows, and between separate
// flows (e.g. blocks that aren't connected to the entry).
const columnGap = 60;
const rowGap = 100;
const treeGap = 200;

// Blocks that haven't been rendered yet have no measured size.
const defaultWidth = 350;
const defaultHeight = 110;
const componentRowHeight = 52;
const branchLabelsHeight = 30;

type AnyNode = Node<any>;

// Lays a flow out as a tree from top to bottom. Unlike a generic graph layout
// it knows which output each connection leaves from, so the blocks below a
// block are put in the same left-to-right order as its outputs: buttons and
// select menus in the order of the message, the error handler's outputs, the
// branches of a condition with else last, and the loop body before what runs
// after the loop. Blocks connected from more than one block are placed under
// the first one.
export const getLayoutedElements = <N extends AnyNode>(
  nodes: N[],
  edges: Edge[]
): { nodes: N[]; edges: Edge[] } => {
  if (nodes.length === 0) return { nodes, edges };

  const byId = new Map(nodes.map((n) => [n.id, n]));
  const isOption = (n: AnyNode) => !!n.type?.startsWith("option_");
  const isEntry = (n: AnyNode) => !!n.type?.startsWith("entry_");

  // Options sit in a row above the entry they connect to.
  const optionsOf = new Map<string, string[]>();
  const placedOptions = new Set<string>();
  // Outgoing connections of each block, in the order of its outputs.
  const outgoing = new Map<string, string[]>();
  const hasParent = new Set<string>();

  const sortable: { edge: Edge; rank: number }[] = [];
  edges.forEach((edge) => {
    const source = byId.get(edge.source);
    const target = byId.get(edge.target);
    if (!source || !target || source.id === target.id) return;

    if (isOption(source)) {
      if (isEntry(target) && !placedOptions.has(source.id)) {
        placedOptions.add(source.id);
        append(optionsOf, target.id, source.id);
      }
      return;
    }
    sortable.push({ edge, rank: outputRank(source, edge, target) });
  });

  // Stable, so connections with the same rank keep their order.
  sortable
    .sort((a, b) => a.rank - b.rank)
    .forEach(({ edge }) => {
      append(outgoing, edge.source, edge.target);
      hasParent.add(edge.target);
    });

  // Roots: entries first, then blocks nothing connects into (loose blocks),
  // each in their current order on the canvas so the result is predictable.
  const byPosition = (a: AnyNode, b: AnyNode) =>
    (a.position?.y ?? 0) - (b.position?.y ?? 0) ||
    (a.position?.x ?? 0) - (b.position?.x ?? 0);
  const candidates = nodes
    .filter((n) => !placedOptions.has(n.id))
    .sort(byPosition);
  const roots = [
    ...candidates.filter(isEntry),
    ...candidates.filter((n) => !isEntry(n) && !hasParent.has(n.id)),
    // Whatever is left is only reachable through a cycle.
    ...candidates,
  ];

  // Builds the trees. A block reachable from several blocks goes under the
  // first one that reaches it.
  const visited = new Set<string>();
  const children = new Map<string, string[]>();
  const trees: string[] = [];
  for (const root of roots) {
    if (visited.has(root.id)) continue;
    trees.push(root.id);
    visited.add(root.id);
    const queue = [root.id];
    for (let i = 0; i < queue.length; i++) {
      const id = queue[i];
      const kids: string[] = [];
      for (const k of outgoing.get(id) ?? []) {
        if (visited.has(k)) continue;
        visited.add(k);
        kids.push(k);
      }
      children.set(id, kids);
      queue.push(...kids);
    }
  }

  const sizes = new Map(nodes.map((n) => [n.id, nodeSize(n)]));
  const size = (id: string) => sizes.get(id)!;
  const positions = new Map<string, { x: number; y: number }>();

  let cursorX = 0;
  for (const root of trees) {
    // Centers horizontally and tops vertically, relative to the root.
    const placed = layoutTree(root, children, size);

    // Options are centered in a row above the entry.
    const options = optionsOf.get(root) ?? [];
    if (options.length > 0) {
      const widths = options.map((id) => size(id).width);
      const rowWidth =
        widths.reduce((a, b) => a + b, 0) + columnGap * (options.length - 1);
      const rowHeight = Math.max(...options.map((id) => size(id).height));
      let x = -rowWidth / 2;
      options.forEach((id, i) => {
        placed.set(id, { x: x + widths[i] / 2, y: -rowGap - rowHeight });
        x += widths[i] + columnGap;
      });
    }

    let left = Infinity;
    let right = -Infinity;
    let top = Infinity;
    for (const [id, p] of placed) {
      left = Math.min(left, p.x - size(id).width / 2);
      right = Math.max(right, p.x + size(id).width / 2);
      top = Math.min(top, p.y);
    }

    for (const [id, p] of placed) {
      positions.set(id, {
        x: cursorX + (p.x - left) - size(id).width / 2,
        y: p.y - top,
      });
    }
    cursorX += right - left + treeGap;
  }

  return {
    nodes: nodes.map((node) => {
      const position = positions.get(node.id);
      return position ? { ...node, position } : node;
    }),
    edges,
  };
};

// Where a connection's output is on its block, from left to right.
function outputRank(
  source: AnyNode,
  edge: Pick<Edge, "sourceHandle">,
  target: AnyNode
): number {
  const owned = getOwnedChildTypes(source.type ?? "");
  if (target.type && owned.includes(target.type)) {
    // Blocks created with their owner: the first owned type goes to the
    // right (the else branch, what runs after a loop), the last to the left
    // (the condition items, the loop body).
    return owned.length - 1 - owned.indexOf(target.type);
  }

  const handle = normalizeHandle(edge.sourceHandle);

  // Buttons and select menus are below the message in its order, and the
  // default output is on the right of the block, so it goes last.
  const components = getComponentHandleIds(
    (source.data as NodeData | undefined)?.message_data?.components ?? []
  );
  if (components.length > 0) {
    const index = handle ? components.indexOf(handle) : -1;
    return index >= 0 ? index : components.length;
  }

  // Blocks with labeled outputs, like the error handler, show them in order.
  const outputs = getNodeValues(source.type ?? "").outputs;
  if (outputs && outputs.length > 0) {
    const index = outputs.indexOf(handle ?? "default");
    return index >= 0 ? index : outputs.length;
  }

  return 0;
}

function append<K, V>(map: Map<K, V[]>, key: K, value: V) {
  const list = map.get(key);
  if (list) list.push(value);
  else map.set(key, [value]);
}

function nodeSize(node: AnyNode) {
  const width = node.measured?.width ?? node.width ?? defaultWidth;
  const measuredHeight = node.measured?.height ?? node.height;
  if (measuredHeight) return { width, height: measuredHeight };

  let height = defaultHeight;
  height +=
    collectComponentGroups(
      (node.data as NodeData | undefined)?.message_data?.components ?? []
    ).length * componentRowHeight;
  if (getNodeValues(node.type ?? "").outputs?.length) {
    height += branchLabelsHeight;
  }
  return { width, height };
}

type Contour = { left: number[]; right: number[] };

// Places a tree: siblings are packed as tightly as their subtrees allow, level
// by level, and each block is centered above its first and last child. Each
// level is one row, as tall as its tallest block, so rows line up.
function layoutTree(
  root: string,
  children: Map<string, string[]>,
  size: (id: string) => { width: number; height: number }
) {
  // Horizontal offset of each block's center from its parent's center.
  const offsets = new Map<string, number>();
  const depths = new Map<string, number>();
  const rowHeights: number[] = [];

  // Iterative so very long flows can't overflow the call stack.
  const order: string[] = [];
  const stack: [string, number][] = [[root, 0]];
  while (stack.length > 0) {
    const [id, depth] = stack.pop()!;
    order.push(id);
    depths.set(id, depth);
    rowHeights[depth] = Math.max(rowHeights[depth] ?? 0, size(id).height);
    for (const kid of children.get(id) ?? []) stack.push([kid, depth + 1]);
  }

  // Children before parents.
  const contours = new Map<string, Contour>();
  for (const id of order.reverse()) {
    const { width } = size(id);
    const kids = children.get(id) ?? [];

    if (kids.length === 0) {
      contours.set(id, { left: [-width / 2], right: [width / 2] });
      continue;
    }

    // Packs the children from left to right, relative to the first one.
    const merged: Contour = { left: [], right: [] };
    const placed: number[] = [];
    for (const kid of kids) {
      const c = contours.get(kid)!;
      let offset = 0;
      if (placed.length > 0) {
        offset = -Infinity;
        const shared = Math.min(c.left.length, merged.right.length);
        for (let d = 0; d < shared; d++) {
          offset = Math.max(offset, merged.right[d] - c.left[d] + columnGap);
        }
      }
      placed.push(offset);
      for (let d = 0; d < c.left.length; d++) {
        if (d >= merged.left.length) merged.left[d] = c.left[d] + offset;
        merged.right[d] = c.right[d] + offset;
      }
      contours.delete(kid);
    }

    const middle = (placed[0] + placed[placed.length - 1]) / 2;
    kids.forEach((kid, i) => offsets.set(kid, placed[i] - middle));
    contours.set(id, {
      left: [-width / 2, ...merged.left.map((x) => x - middle)],
      right: [width / 2, ...merged.right.map((x) => x - middle)],
    });
  }

  const rowTops: number[] = [];
  rowHeights.forEach((_, d) => {
    rowTops[d] = d === 0 ? 0 : rowTops[d - 1] + rowHeights[d - 1] + rowGap;
  });

  const positions = new Map<string, { x: number; y: number }>();
  const walk: [string, number][] = [[root, 0]];
  while (walk.length > 0) {
    const [id, x] = walk.pop()!;
    positions.set(id, { x, y: rowTops[depths.get(id)!] });
    for (const kid of children.get(id) ?? []) {
      walk.push([kid, x + offsets.get(kid)!]);
    }
  }

  return positions;
}

type Point = { x: number; y: number };
type Size = { width: number; height: number };

// Lays out a flow with its notes and groups. Notes stay where they were put,
// or move along with the group they're in. The blocks of each expanded group
// are laid out on their own, inner groups first, and then the groups like
// single blocks around them, so the blocks of a group stay together and the
// groups are fitted around them. Collapsed groups are laid out like blocks and
// the nodes hidden in them move with them.
export function getFlowLayout<N extends Node<NodeData>>(
  nodes: N[],
  edges: Edge[]
): N[] {
  if (!nodes.some(isGroupNode)) {
    const laid = getLayoutedElements(
      nodes.filter((n) => !isNoteNode(n)),
      edges
    ).nodes;
    const byId = new Map(laid.map((n) => [n.id, n]));
    return moveCoveredNotes(nodes.map((n) => byId.get(n.id) ?? n));
  }

  const hidden = getHiddenNodeIds(nodes);
  const owners = getHiddenNodeOwners(nodes, hidden);
  const byId = new Map(nodes.map((n) => [n.id, n]));
  const visible = nodes.filter((n) => !hidden.has(n.id));
  const isExpandedGroup = (n: N) =>
    isGroupNode(n) && !getGroupData(n).collapsed;

  // The innermost expanded group of each visible node. Groups are claimed
  // from smallest to largest, so inner groups get their nodes first.
  const area = (n: N) => {
    const b = getGroupBounds(n);
    return b.width * b.height;
  };
  const parentOf = new Map<string, string>();
  for (const group of visible
    .filter(isExpandedGroup)
    .sort((a, b) => area(a) - area(b))) {
    for (const id of getGroupMemberIds(group, visible)) {
      if (!parentOf.has(id)) parentOf.set(id, group.id);
    }
  }

  // What each container (a group, or the flow for null) lays out: its blocks
  // and the groups directly in it. Notes aren't laid out.
  const children = new Map<string | null, string[]>();
  for (const node of visible) {
    if (isNoteNode(node)) continue;
    append(children, parentOf.get(node.id) ?? null, node.id);
  }

  // The direct child of the container that a node is drawn as, if any.
  const childIn = (id: string, container: string | null) => {
    let current = hidden.has(id) ? owners.get(id) : id;
    while (current !== undefined) {
      const parent = parentOf.get(current) ?? null;
      if (parent === container) return current;
      if (parent === null) return undefined;
      current = parent;
    }
    return undefined;
  };

  const groupSizes = new Map<string, Size>();
  const sizeOf = (id: string): Size => {
    const size = groupSizes.get(id);
    if (size) return size;
    const node = byId.get(id)!;
    if (isGroupNode(node)) {
      const bounds = getGroupBounds(node);
      return {
        width: bounds.width,
        height: getGroupData(node).collapsed
          ? groupHeaderHeight
          : bounds.height,
      };
    }
    return nodeSize(node);
  };

  // Positions in the container, from the top left of its content.
  const relative = new Map<string, Point>();

  const layoutContainer = (container: string | null): Size => {
    const ids = children.get(container) ?? [];
    for (const id of ids) {
      if (!isGroupNode(byId.get(id)!) || !isExpandedGroup(byId.get(id)!)) {
        continue;
      }
      const content = layoutContainer(id);
      groupSizes.set(id, {
        width: Math.max(content.width + groupPadding * 2, groupMinWidth),
        height: Math.max(
          content.height + groupPadding * 2 + groupHeaderHeight,
          groupMinHeight
        ),
      });
    }
    if (ids.length === 0) return { width: 0, height: 0 };

    // Groups are laid out as blocks of their size, connected where blocks in
    // them are.
    const proxies = ids.map((id) => {
      const node = byId.get(id)!;
      return isGroupNode(node)
        ? ({
            id,
            type: groupNodeType,
            position: node.position,
            data: {},
            measured: sizeOf(id),
          } as Node<NodeData>)
        : node;
    });
    const proxyEdges = new Map<string, Edge>();
    for (const edge of edges) {
      const source = childIn(edge.source, container);
      const target = childIn(edge.target, container);
      if (!source || !target || source === target) continue;
      const sourceHandle =
        source === edge.source ? edge.sourceHandle ?? null : null;
      const targetHandle =
        target === edge.target ? edge.targetHandle ?? null : null;
      const key = `${source}:${sourceHandle}->${target}:${targetHandle}`;
      if (!proxyEdges.has(key)) {
        proxyEdges.set(key, {
          id: key,
          source,
          target,
          sourceHandle,
          targetHandle,
        });
      }
    }

    const laid = getLayoutedElements(proxies, [...proxyEdges.values()]).nodes;
    const minX = Math.min(...laid.map((n) => n.position.x));
    const minY = Math.min(...laid.map((n) => n.position.y));
    let maxX = -Infinity;
    let maxY = -Infinity;
    for (const node of laid) {
      const { width, height } = sizeOf(node.id);
      relative.set(node.id, {
        x: node.position.x - minX,
        y: node.position.y - minY,
      });
      maxX = Math.max(maxX, node.position.x + width);
      maxY = Math.max(maxY, node.position.y + height);
    }
    return { width: maxX - minX, height: maxY - minY };
  };

  layoutContainer(null);

  // Turns the positions in groups into positions in the flow. The flow keeps
  // the top left corner it had.
  const positions = new Map<string, Point>();
  const place = (container: string | null, origin: Point) => {
    for (const id of children.get(container) ?? []) {
      const rel = relative.get(id)!;
      const pos = { x: origin.x + rel.x, y: origin.y + rel.y };
      positions.set(id, pos);
      if (children.has(id)) {
        place(id, {
          x: pos.x + groupPadding,
          y: pos.y + groupPadding + groupHeaderHeight,
        });
      }
    }
  };
  const laidOut = visible.filter((n) => !isNoteNode(n));
  place(null, {
    x: Math.min(...laidOut.map((n) => n.position.x)),
    y: Math.min(...laidOut.map((n) => n.position.y)),
  });

  // Nodes that weren't laid out move as much as the group they're in: notes
  // with their innermost expanded group, hidden nodes with their collapsed
  // group.
  const delta = (id: string): Point | null => {
    const node = byId.get(id)!;
    const pos = positions.get(id);
    return pos
      ? { x: pos.x - node.position.x, y: pos.y - node.position.y }
      : null;
  };
  for (const node of nodes) {
    if (positions.has(node.id)) continue;
    const owner = hidden.has(node.id)
      ? owners.get(node.id)
      : parentOf.get(node.id);
    const d = owner && delta(owner);
    if (d) {
      positions.set(node.id, {
        x: node.position.x + d.x,
        y: node.position.y + d.y,
      });
    }
  }

  // The notes in a group stay in it: they're kept inside its frame, which
  // grows to fit them.
  for (const node of visible) {
    if (!isNoteNode(node)) continue;
    const groupId = parentOf.get(node.id);
    const groupPos = groupId && positions.get(groupId);
    const groupSize = groupId && groupSizes.get(groupId);
    const pos = positions.get(node.id);
    if (!groupPos || !groupSize || !pos) continue;

    const { width, height } = nodeSize(node);
    pos.x = Math.max(pos.x, groupPos.x + groupPadding / 2);
    pos.y = Math.max(pos.y, groupPos.y + groupHeaderHeight + groupPadding / 2);
    groupSize.width = Math.max(
      groupSize.width,
      pos.x + width - groupPos.x + groupPadding / 2
    );
    groupSize.height = Math.max(
      groupSize.height,
      pos.y + height - groupPos.y + groupPadding / 2
    );
  }

  return moveCoveredNotes(
    nodes.map((node) => {
      const position = positions.get(node.id);
      if (!position) return node;
      const size = groupSizes.get(node.id);
      return size
        ? { ...node, position, width: size.width, height: size.height }
        : { ...node, position };
    })
  );
}

// Notes that aren't in a group stay where they were put, unless the arranged
// blocks now cover them. Those move to the right of the blocks together, so
// they stay where they were to each other.
function moveCoveredNotes<N extends Node<NodeData>>(nodes: N[]): N[] {
  const hidden = getHiddenNodeIds(nodes);
  const rects = (filter: (n: N) => boolean) =>
    nodes
      .filter((n) => filter(n) && !hidden.has(n.id))
      .map((n) => {
        const size = isGroupNode(n) ? getGroupBounds(n) : nodeSize(n);
        return {
          id: n.id,
          x: n.position.x,
          y: n.position.y,
          width: size.width,
          height:
            isGroupNode(n) && getGroupData(n).collapsed
              ? groupHeaderHeight
              : size.height,
        };
      });

  const groups = nodes.filter((n) => isGroupNode(n) && !hidden.has(n.id));
  const inGroup = new Set(groups.flatMap((g) => getGroupMemberIds(g, nodes)));
  const content = rects((n) => !isNoteNode(n));
  if (content.length === 0) return nodes;

  const overlaps = (a: Rect, b: Rect) =>
    a.x < b.x + b.width &&
    b.x < a.x + a.width &&
    a.y < b.y + b.height &&
    b.y < a.y + a.height;
  const covered = rects((n) => isNoteNode(n) && !inGroup.has(n.id)).filter(
    (note) => content.some((c) => overlaps(note, c))
  );
  if (covered.length === 0) return nodes;

  const right = Math.max(...content.map((c) => c.x + c.width));
  const dx = right + columnGap - Math.min(...covered.map((n) => n.x));
  const moved = new Set(covered.map((n) => n.id));
  return nodes.map((n) =>
    moved.has(n.id)
      ? { ...n, position: { x: n.position.x + dx, y: n.position.y } }
      : n
  );
}

type Rect = { x: number; y: number; width: number; height: number };
