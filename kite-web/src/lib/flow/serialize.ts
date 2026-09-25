import { Edge, Node } from "@xyflow/react";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { normalizeHandle } from "./nodes";
import { walkDownstream } from "./placeholders";

// Writes a flow as compact text for the LLM flow editor: each block with its
// ID, type and settings, and the connections between them, in the order the
// flow runs. Positions and other editor state are left out.
export function serializeFlow(
  nodes: Node<NodeData>[],
  edges: Edge[],
  context: FlowContextType,
  selectedIds: string[] = []
) {
  const entryIds = nodes
    .filter((n) => n.type?.startsWith("entry_"))
    .map((n) => n.id);
  const order = [
    ...entryIds,
    ...nodes.filter((n) => n.type?.startsWith("option_")).map((n) => n.id),
    ...walkDownstream(entryIds, edges),
    ...nodes.map((n) => n.id),
  ];
  const rank = new Map<string, number>();
  order.forEach((id) => {
    if (!rank.has(id)) rank.set(id, rank.size);
  });

  // Connections to blocks that don't exist sort last.
  const byRank = (a: string, b: string) =>
    (rank.get(a) ?? rank.size) - (rank.get(b) ?? rank.size);
  const sortedNodes = [...nodes].sort((a, b) => byRank(a.id, b.id));
  const sortedEdges = [...edges].sort(
    (a, b) => byRank(a.source, b.source) || byRank(a.target, b.target)
  );

  const selected = new Set(selectedIds);
  const lines = [`Flow type: ${context}`, "", "Blocks:"];
  for (const node of sortedNodes) {
    const mark = selected.has(node.id) ? " (selected)" : "";
    const data = compact(node.data);
    lines.push(
      `- ${node.id} ${node.type}${mark}${
        data === undefined ? "" : ` ${JSON.stringify(data)}`
      }`
    );
  }

  lines.push("", "Connections:");
  for (const edge of sortedEdges) {
    const handle = normalizeHandle(edge.sourceHandle);
    lines.push(
      `- ${edge.source}${handle ? `[${handle}]` : ""} -> ${edge.target}`
    );
  }

  return lines.join("\n");
}

// Drops empty settings, which the editor leaves behind when fields are
// cleared, so they don't cost tokens. Arrays are kept as they are, as edits
// replace them as a whole.
function compact(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.length > 0 ? value : undefined;
  }
  if (value && typeof value === "object") {
    const entries = Object.entries(value)
      .map(([k, v]) => [k, compact(v)] as const)
      .filter(([, v]) => v !== undefined);
    return entries.length > 0 ? Object.fromEntries(entries) : undefined;
  }
  return value === null || value === "" ? undefined : value;
}
