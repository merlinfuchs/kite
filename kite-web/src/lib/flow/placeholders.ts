import { Edge, Node } from "@xyflow/react";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import {
  ExpressionTypeName,
  expressionFields,
  isExpressionTypeName,
  resultFields,
  thingTypeName,
} from "./expressionTypes";
import { getNodeTitle, getNodeValues } from "./nodes";
import { modalInputIsMultiValue, normalizeModalComponents } from "./modal";
import { isResumeEdge } from "./resume";

export interface Placeholder {
  label: string;
  value: string;
  description?: string;
  // The fields of a value with a type, like the name of a role. Their labels
  // leave out the label of the value.
  fields?: Placeholder[];
}

export interface PlaceholderGroup {
  label: string;
  placeholders: Placeholder[];
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

const interactionCache = new Map<string, PlaceholderGroup[]>();

// interactionPlaceholders lists the placeholders of the interaction or event
// the flow runs with. Resumed sub-flows reach earlier ones through a prefix.
// They are built once, as the flow editor asks for them on every change.
export function interactionPlaceholders(
  contextType?: FlowContextType,
  prefix = "",
  labelPrefix = ""
): PlaceholderGroup[] {
  const key = `${contextType}|${prefix}|${labelPrefix}`;
  let res = interactionCache.get(key);
  if (!res) {
    res = buildInteractionPlaceholders(contextType, prefix, labelPrefix);
    interactionCache.set(key, res);
  }
  return res;
}

function buildInteractionPlaceholders(
  contextType: FlowContextType | undefined,
  prefix: string,
  labelPrefix: string
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
      placeholders: withFields(
        typedPlaceholder("User", `${prefix}user`, "member")
      ),
    },
    {
      label: `${labelPrefix}Server`,
      placeholders: withFields(
        typedPlaceholder("Server", `${prefix}guild`, "guild")
      ),
    },
    {
      label: `${labelPrefix}Channel`,
      placeholders: withFields(
        typedPlaceholder("Channel", `${prefix}channel`, "channel")
      ),
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
      placeholders: withFields(
        typedPlaceholder("Message", `${prefix}message`, "message")
      ),
    });
    // Only set for reaction events.
    res.push({
      label: `${labelPrefix}Emoji`,
      placeholders: withFields(
        typedPlaceholder("Emoji", `${prefix}emoji`, "emoji")
      ),
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

  return [
    {
      label: "Command",
      placeholders: nodes
        .filter((n) => n.type === "option_command_argument" && argIds.has(n.id))
        .map((n) => {
          const label = `Command Arg '${n.data.name}'`;
          const value = `arg('${n.data.name}')`;
          const type = argTypes[n.data.command_argument_type ?? ""];
          return typedPlaceholder(label, value, type);
        }),
    },
  ];
}

// What arg() returns for each argument type that isn't a plain value. A
// mentionable can be a member or a role.
const argTypes: Partial<Record<string, ExpressionTypeName>> = {
  user: "member",
  role: "role",
  channel: "channel",
  attachment: "attachment",
};

// typedPlaceholder is a value with its fields, and the fields of those of its
// fields that have a type themselves, like the top role of a user. The text
// fields of a list are joined, e.g. the names of all picked roles.
function typedPlaceholder(
  label: string,
  value: string,
  type: ExpressionTypeName | undefined,
  { list = false, result = false } = {}
): Placeholder {
  if (!type) return { label, value };

  const fields: Placeholder[] = [];
  for (const field of result ? resultFields(type) : expressionFields(type)) {
    const fieldLabel = placeholderLabel(field.name);
    const fieldType = isExpressionTypeName(field.type) ? field.type : undefined;

    if (list) {
      // join only takes strings.
      if (field.type !== "string" || field.list || field.call) continue;
      fields.push({
        label: fieldLabel,
        value: `join(map(${value}, #.${field.name}), ', ')`,
        description: field.description,
      });
    } else if (fieldType && !field.list) {
      fields.push({
        ...typedPlaceholder(fieldLabel, `${value}.${field.name}`, fieldType, {
          result,
        }),
        description: field.description,
      });
    } else {
      fields.push({
        label: fieldLabel,
        value: `${value}.${field.name}${field.call ? "()" : ""}`,
        description: field.description,
      });
    }
  }
  return { label, value, fields };
}

// withFields lists a value and its fields side by side, for values like the
// user, whose fields are used more often than the value itself.
function withFields({ fields = [], ...placeholder }: Placeholder) {
  return [
    placeholder,
    ...fields.map((f) => ({ ...f, label: `${placeholder.label} ${f.label}` })),
  ];
}

// flattenPlaceholders lists the placeholders and all of their fields, with
// the full label of each field, like User Top Role Name. Values keep their
// fields, so a search result can be opened.
export function flattenPlaceholders(
  placeholders: Placeholder[],
  labelPrefix = ""
): Placeholder[] {
  return placeholders.flatMap((placeholder) => {
    const label = labelPrefix + placeholder.label;
    return [
      { ...placeholder, label },
      ...flattenPlaceholders(placeholder.fields ?? [], `${label} `),
    ];
  });
}

// placeholderLabel turns the name of a field into a label, like avatar_url
// into Avatar URL.
function placeholderLabel(name: string) {
  return name
    .split("_")
    .map(
      (word) => labelWords[word] ?? word.charAt(0).toUpperCase() + word.slice(1)
    )
    .join(" ");
}

const labelWords: Record<string, string> = {
  id: "ID",
  ids: "IDs",
  url: "URL",
};

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
  if (depth === 1) return res;
  // Whether previous was a button, select menu or modal isn't tracked here.
  return [
    ...res,
    ...interactionPlaceholders(undefined, "previous.", "Previous "),
  ];
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
    for (const { group, ...placeholder } of getProvidedPlaceholders(parent)) {
      // The nearest block wins if several store the same variable.
      if (seen.has(placeholder.value)) continue;
      seen.add(placeholder.value);
      groups.find((g) => g.label === group)!.placeholders.push(placeholder);
    }
  }

  return groups.filter((g) => g.placeholders.length > 0);
}

// Returns the placeholders a block makes available to the blocks after it:
// its result, its temporary variable and the inputs of its modal.
export function getProvidedPlaceholders(node: Node<NodeData>) {
  const res: (Placeholder & { group: string })[] = [];

  if (
    node.type?.startsWith("action_") ||
    node.type === "control_error_handler"
  ) {
    const label = getNodeTitle(node);
    const value = `result('${node.id}')`;
    const result = getNodeValues(node.type).result;
    const type = result?.thing && thingTypeName(result.thing);
    res.push({
      group: "Node Results",
      ...typedPlaceholder(label, value, type, {
        list: result?.list,
        result: true,
      }),
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

        const picked = pickedNouns[component.type ?? ""];
        if (picked) {
          const [singular, plural, type] = picked;
          const label = `${row.label ?? "Unknown Input"} (${singular})`;
          const value = `picked('${component.custom_id}')`;
          res.push({
            group: "Modal Inputs",
            ...typedPlaceholder(label, value, type),
          });
          if (modalInputIsMultiValue(component)) {
            const label = `${row.label ?? "Unknown Input"} (All ${plural})`;
            const value = `pickedAll('${component.custom_id}')`;
            res.push({
              group: "Modal Inputs",
              ...typedPlaceholder(label, value, type, { list: true }),
            });
          }
        }
      }
    }
  }

  return res;
}

// What picked() and pickedAll() return for each select and file upload, as
// singular and plural, and its type. A mentionable can be a member or a role.
const pickedNouns: Partial<
  Record<string, [string, string, ExpressionTypeName?]>
> = {
  user_select: ["Member", "Members", "member"],
  role_select: ["Role", "Roles", "role"],
  mentionable_select: ["Member or Role", "Members and Roles"],
  channel_select: ["Channel", "Channels", "channel"],
  file_upload: ["File", "Files", "attachment"],
};

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
