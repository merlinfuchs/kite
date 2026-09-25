import { Edge, Node } from "@xyflow/react";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { getNodeTitle } from "./nodes";
import { isResumeEdge } from "./resume";

export interface PlaceholderGroup {
  label: string;
  placeholders: { label: string; value: string }[];
}

// Returns the placeholders a block can use: those of the interaction or event
// the flow runs with, plus the results, temporary variables and modal inputs
// of the blocks before it. Without a block, only the flow wide ones.
export function getAvailablePlaceholders(
  nodeId: string | undefined,
  nodes: Node<NodeData>[],
  edges: Edge[],
  contextType: FlowContextType
): PlaceholderGroup[] {
  const res = [
    ...commandPlaceholders(nodes, edges, contextType),
    ...interactionPlaceholders(contextType),
    {
      label: "App",
      placeholders: [
        { label: "App User ID", value: "app.user.id" },
        { label: "App User Mention", value: "app.user.mention" },
      ],
    },
  ];
  if (!nodeId) return res;

  return [
    ...res,
    ...resumePlaceholders(nodeId, nodes, edges, contextType),
    ...upstreamPlaceholders(nodeId, nodes, edges),
  ];
}

// interactionPlaceholders lists the placeholders of the interaction or event
// the flow runs with. Resumed sub-flows reach earlier ones through a prefix.
function interactionPlaceholders(
  contextType?: FlowContextType,
  prefix = "",
  labelPrefix = ""
): PlaceholderGroup[] {
  const res = [
    {
      label: `${labelPrefix}User`,
      placeholders: [
        { label: "User", value: `${prefix}user` },
        { label: "User ID", value: `${prefix}user.id` },
        { label: "User Mention", value: `${prefix}user.mention` },
        { label: "User Username", value: `${prefix}user.username` },
        { label: "User Display Name", value: `${prefix}user.display_name` },
        { label: "User Nickname", value: `${prefix}user.nick` },
        { label: "User Avatar URL", value: `${prefix}user.avatar_url` },
        { label: "User Banner URL", value: `${prefix}user.banner_url` },
      ],
    },
    {
      label: `${labelPrefix}Server`,
      placeholders: [{ label: "Server ID", value: `${prefix}guild.id` }],
    },
    {
      label: `${labelPrefix}Channel`,
      placeholders: [{ label: "Channel ID", value: `${prefix}channel.id` }],
    },
  ];

  if (contextType === "component_select_menu") {
    res.push({
      label: `${labelPrefix}Select Menu`,
      placeholders: [
        { label: "Selected Value", value: `${prefix}interaction.value` },
        { label: "Selected Values", value: `${prefix}interaction.values` },
      ],
    });
  }

  if (contextType === "event_discord") {
    res.push({
      label: `${labelPrefix}Message`,
      placeholders: [
        { label: "Message ID", value: `${prefix}message.id` },
        { label: "Message Content", value: `${prefix}message.content` },
      ],
    });
  }

  return res;
}

function commandPlaceholders(
  nodes: Node<NodeData>[],
  edges: Edge[],
  contextType: FlowContextType
): PlaceholderGroup[] {
  if (contextType !== "command") return [];

  // Only arguments connected to the entry are registered with Discord.
  const entryIds = new Set(
    nodes.filter((n) => n.type === "entry_command").map((n) => n.id)
  );
  const argIds = new Set(
    edges.filter((e) => entryIds.has(e.target)).map((e) => e.source)
  );

  // TODO: take arg type into account
  return [
    {
      label: "Command",
      placeholders: nodes
        .filter((n) => n.type === "option_command_argument" && argIds.has(n.id))
        .map((n) => ({
          label: `Command Arg '${n.data.name}'`,
          value: `arg('${n.data.name}')`,
        })),
    },
  ];
}

// Sub-flows run with the interaction that resumed them, so placeholders of the
// interaction or event before a resume point are only reachable via origin and
// previous.
function resumePlaceholders(
  nodeId: string,
  nodes: Node<NodeData>[],
  edges: Edge[],
  contextType: FlowContextType
): PlaceholderGroup[] {
  const depth = getResumeDepth(nodeId, nodes, edges);
  if (depth === 0) return [];

  const res = interactionPlaceholders(contextType, "origin.", "Original ");
  if (depth > 1) {
    // Whether previous was a button, select menu or modal isn't tracked here.
    res.push(...interactionPlaceholders(undefined, "previous.", "Previous "));
  }
  return res;
}

// getResumeDepth counts the resume points between the root and a node.
function getResumeDepth(nodeId: string, nodes: Node[], edges: Edge[]) {
  const nodeTypes = new Map(nodes.map((n) => [n.id, n.type]));
  const { incoming } = indexEdges(edges);

  const visited = new Set<string>();

  function traverse(id: string): number {
    if (visited.has(id)) {
      return 0;
    }
    visited.add(id);

    let depth = 0;
    for (const edge of incoming.get(id) ?? []) {
      const isResume = isResumeEdge(edge, nodeTypes.get(edge.source));
      depth = Math.max(depth, traverse(edge.source) + (isResume ? 1 : 0));
    }
    return depth;
  }

  return traverse(nodeId);
}

function upstreamPlaceholders(
  nodeId: string,
  nodes: Node<NodeData>[],
  edges: Edge[]
): PlaceholderGroup[] {
  const nodeItems: { label: string; value: string }[] = [];
  const resultKeyItems: { label: string; value: string }[] = [];
  const componentItems: { label: string; value: string }[] = [];

  const seenResultKeys = new Set<string>();

  for (const parent of getUpstreamNodes(nodeId, nodes, edges)) {
    if (
      parent.type?.startsWith("action_") ||
      parent.type === "control_error_handler"
    ) {
      nodeItems.push({
        label: getNodeTitle(parent),
        value: `result('${parent.id}')`,
      });
    }

    if (
      parent.data.temporary_name &&
      !seenResultKeys.has(parent.data.temporary_name)
    ) {
      seenResultKeys.add(parent.data.temporary_name);

      resultKeyItems.push({
        label: `Temporary Variable '${parent.data.temporary_name}'`,
        value: `var('${parent.data.temporary_name}')`,
      });
    }

    if (parent.type === "suspend_response_modal") {
      for (const row of parent.data.modal_data?.components ?? []) {
        for (const component of row?.components ?? []) {
          componentItems.push({
            label: component.label ?? "Unknown Input",
            value: `input('${component.custom_id}')`,
          });
        }
      }
    }
  }

  const res = [];

  if (componentItems.length > 0) {
    res.push({ label: "Modal Inputs", placeholders: componentItems });
  }

  if (resultKeyItems.length > 0) {
    res.push({ label: "Temporary Variables", placeholders: resultKeyItems });
  }

  if (nodeItems.length > 0) {
    res.push({ label: "Node Results", placeholders: nodeItems });
  }

  return res;
}

// Returns all blocks that run before the given one, nearest first. Besides
// the blocks leading up to it, a loop's iterations run before what comes after
// the loop, and an error handler's blocks run before its error branch.
function getUpstreamNodes(
  nodeId: string,
  nodes: Node<NodeData>[],
  edges: Edge[]
) {
  const nodesById = new Map(nodes.map((n) => [n.id, n]));
  const { incoming, outgoing } = indexEdges(edges);

  const res: Node<NodeData>[] = [];
  const visited = new Set([nodeId]);
  const queue = [nodeId];
  const add = (id: string) => {
    const node = nodesById.get(id);
    if (!node || visited.has(id)) return;
    visited.add(id);
    res.push(node);
    queue.push(id);
  };

  for (let i = 0; i < queue.length; i++) {
    const current = nodesById.get(queue[i]);
    for (const edge of incoming.get(queue[i]) ?? []) {
      add(edge.source);

      let earlierBranch: Edge[] = [];
      if (
        nodesById.get(edge.source)?.type === "control_error_handler" &&
        edge.sourceHandle === "error"
      ) {
        earlierBranch = (outgoing.get(edge.source) ?? []).filter(
          (e) => (e.sourceHandle || "default") === "default"
        );
      } else if (current?.type === "control_loop_end") {
        earlierBranch = (outgoing.get(edge.source) ?? []).filter(
          (e) => nodesById.get(e.target)?.type === "control_loop_each"
        );
      }

      const starts = earlierBranch.map((e) => e.target);
      [...starts, ...walk(starts, outgoing, "target")].forEach(add);
    }
  }

  return res;
}

// Returns the IDs of the blocks reached from the given ones by following
// edges forward, nearest first.
export function walkDownstream(startIds: string[], edges: Edge[]) {
  return walk(startIds, indexEdges(edges).outgoing, "target");
}

// Returns the IDs of the blocks leading up to the given ones, nearest first.
export function walkUpstream(startIds: string[], edges: Edge[]) {
  return walk(startIds, indexEdges(edges).incoming, "source");
}

function indexEdges(edges: Edge[]) {
  const incoming = new Map<string, Edge[]>();
  const outgoing = new Map<string, Edge[]>();
  for (const edge of edges) {
    if (!incoming.has(edge.target)) incoming.set(edge.target, []);
    if (!outgoing.has(edge.source)) outgoing.set(edge.source, []);
    incoming.get(edge.target)!.push(edge);
    outgoing.get(edge.source)!.push(edge);
  }
  return { incoming, outgoing };
}

function walk(
  startIds: string[],
  next: Map<string, Edge[]>,
  follow: "source" | "target"
) {
  const visited = new Set(startIds);
  const res: string[] = [];
  const queue = [...startIds];
  for (let i = 0; i < queue.length; i++) {
    for (const edge of next.get(queue[i]) ?? []) {
      const id = edge[follow];
      if (visited.has(id)) continue;
      visited.add(id);
      res.push(id);
      queue.push(id);
    }
  }
  return res;
}
