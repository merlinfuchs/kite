import { Edge, Node } from "@xyflow/react";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { getNodeTitle } from "./nodes";
import { modalInputIsMultiValue, normalizeModalComponents } from "./modal";
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
export function interactionPlaceholders(
  contextType?: FlowContextType,
  prefix = "",
  labelPrefix = ""
): PlaceholderGroup[] {
  if (contextType === "event_schedule") {
    return [
      {
        label: `${labelPrefix}Schedule`,
        placeholders: [
          { label: "Scheduled Time (UTC)", value: `${prefix}schedule.time` },
          { label: "Scheduled Time (Unix)", value: `${prefix}schedule.unix` },
        ],
      },
    ];
  }

  if (contextType === "event_webhook") {
    return [
      {
        label: `${labelPrefix}Webhook`,
        placeholders: [
          { label: "Request Body", value: `${prefix}webhook.body` },
          {
            label: "Request Body Field (JSON)",
            value: `${prefix}webhook.data.name`,
          },
          {
            label: "Request Header",
            value: `${prefix}webhook.headers['content-type']`,
          },
          { label: "Query Parameter", value: `${prefix}webhook.query.name` },
        ],
      },
    ];
  }

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
        { label: "User Is Bot", value: `${prefix}user.is_bot` },
        { label: "User Created At", value: `${prefix}user.created_at` },
        { label: "User Joined At", value: `${prefix}user.joined_at` },
        { label: "User Top Role", value: `${prefix}user.top_role` },
        { label: "User Role Mentions", value: `${prefix}user.role_mentions` },
        { label: "User Role Names", value: `${prefix}user.role_names` },
        { label: "User Role Count", value: `${prefix}user.role_count` },
        { label: "User Color", value: `${prefix}user.color` },
        { label: "User Is Booster", value: `${prefix}user.is_booster` },
        { label: "User Boosting Since", value: `${prefix}user.boosting_since` },
        { label: "User Is Timed Out", value: `${prefix}user.is_timed_out` },
        { label: "User Timeout Until", value: `${prefix}user.timeout_until` },
        { label: "User Is Owner", value: `${prefix}user.is_owner` },
        { label: "User Is Admin", value: `${prefix}user.is_admin` },
        { label: "User Permissions", value: `${prefix}user.permissions` },
      ],
    },
    {
      label: `${labelPrefix}Server`,
      placeholders: [
        { label: "Server ID", value: `${prefix}guild.id` },
        { label: "Server Name", value: `${prefix}guild.name` },
        { label: "Server Icon URL", value: `${prefix}guild.icon_url` },
        { label: "Server Member Count", value: `${prefix}guild.member_count` },
        { label: "Server Boost Count", value: `${prefix}guild.boost_count` },
        { label: "Server Owner ID", value: `${prefix}guild.owner_id` },
        { label: "Server Boost Level", value: `${prefix}guild.boost_level` },
        { label: "Server Created At", value: `${prefix}guild.created_at` },
        { label: "Server Banner URL", value: `${prefix}guild.banner_url` },
        { label: "Server Description", value: `${prefix}guild.description` },
        { label: "Server Vanity URL", value: `${prefix}guild.vanity_url` },
        { label: "Server Role Count", value: `${prefix}guild.role_count` },
        {
          label: "Server Channel Count",
          value: `${prefix}guild.channel_count`,
        },
        { label: "Server Emoji Count", value: `${prefix}guild.emoji_count` },
        {
          label: "Server Rules Channel",
          value: `${prefix}guild.rules_channel`,
        },
        {
          label: "Server System Channel",
          value: `${prefix}guild.system_channel`,
        },
      ],
    },
    {
      label: `${labelPrefix}Channel`,
      placeholders: [
        { label: "Channel ID", value: `${prefix}channel.id` },
        { label: "Channel Name", value: `${prefix}channel.name` },
        { label: "Channel Mention", value: `${prefix}channel.mention` },
        { label: "Channel Type", value: `${prefix}channel.type` },
        { label: "Channel Category ID", value: `${prefix}channel.category_id` },
        {
          label: "Channel Category Name",
          value: `${prefix}channel.category_name`,
        },
      ],
    },
  ];

  if (contextType === "component_select_menu") {
    res.push({
      label: `${labelPrefix}Select Menu`,
      placeholders: [
        { label: "Selected Value or ID", value: `${prefix}interaction.value` },
        {
          label: "Selected Values or IDs",
          value: `${prefix}interaction.values`,
        },
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
    // Only set for reaction events.
    res.push({
      label: `${labelPrefix}Emoji`,
      placeholders: [
        { label: "Emoji", value: `${prefix}emoji` },
        { label: "Emoji ID", value: `${prefix}emoji.id` },
        { label: "Emoji Name", value: `${prefix}emoji.name` },
        { label: "Emoji Mention", value: `${prefix}emoji.mention` },
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
  const groups: PlaceholderGroup[] = [
    { label: "Modal Inputs", placeholders: [] },
    { label: "Temporary Variables", placeholders: [] },
    { label: "Node Results", placeholders: [] },
  ];
  const seen = new Set<string>();

  for (const parent of getUpstreamNodes(nodeId, nodes, edges)) {
    for (const { group, label, value } of getProvidedPlaceholders(parent)) {
      // The nearest block wins if several store the same variable.
      if (seen.has(value)) continue;
      seen.add(value);
      groups
        .find((g) => g.label === group)!
        .placeholders.push({
          label,
          value,
        });
    }
  }

  return groups.filter((g) => g.placeholders.length > 0);
}

// Returns the placeholders a block makes available to the blocks after it:
// its result, its temporary variable and the inputs of its modal.
export function getProvidedPlaceholders(node: Node<NodeData>) {
  const res: { group: string; label: string; value: string }[] = [];

  if (
    node.type?.startsWith("action_") ||
    node.type === "control_error_handler"
  ) {
    res.push({
      group: "Node Results",
      label: getNodeTitle(node),
      value: `result('${node.id}')`,
    });
  }

  if (node.data.temporary_name) {
    res.push({
      group: "Temporary Variables",
      label: `Temporary Variable '${node.data.temporary_name}'`,
      value: `var('${node.data.temporary_name}')`,
    });
  }

  if (node.type === "suspend_response_modal") {
    for (const row of normalizeModalComponents(
      node.data.modal_data?.components
    )) {
      for (const component of row.components ?? []) {
        res.push({
          group: "Modal Inputs",
          label: row.label ?? "Unknown Input",
          value: `input('${component.custom_id}')`,
        });
        if (modalInputIsMultiValue(component)) {
          res.push({
            group: "Modal Inputs",
            label: `${row.label ?? "Unknown Input"} (Selected Values)`,
            value: `inputs('${component.custom_id}')`,
          });
        }
      }
    }
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
