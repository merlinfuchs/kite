import { FlowContextType, useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { getNodeValues } from "@/lib/flow/nodes";
import { isResumeEdge } from "@/lib/flow/resume";
import { Edge, getIncomers, Node, useEdges, useNodes } from "@xyflow/react";
import { VariableIcon } from "lucide-react";
import { useMemo } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";

export default function FlowPlaceholderExplorer({
  onSelect,
  hideBrackets,
}: {
  onSelect: (value: string) => void;
  hideBrackets?: boolean;
}) {
  // TODO: only compute when explorer is open
  const nodePlaceholders = useNodePlaceholders();
  const commandPlaceholders = useCommandPlaceholders();
  const globalPlaceholders = useGlobalPlaceholders();
  const resumePlaceholders = useResumePlaceholders();

  const placeholders = useMemo(
    () => [
      ...commandPlaceholders,
      ...globalPlaceholders,
      ...resumePlaceholders,
      ...nodePlaceholders,
    ],
    [
      commandPlaceholders,
      globalPlaceholders,
      resumePlaceholders,
      nodePlaceholders,
    ]
  );

  return (
    <div className="absolute top-1.5 right-1.5 z-20">
      <PlaceholderExplorer
        onSelect={onSelect}
        placeholders={placeholders}
        hideBrackets={hideBrackets}
      >
        <VariableIcon
          className="h-5.5 w-5.5 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
        />
      </PlaceholderExplorer>
    </div>
  );
}

function useGlobalPlaceholders() {
  const contextType = useFlowContext((c) => c.type);

  return [
    ...interactionPlaceholders(contextType),
    {
      label: "App",
      placeholders: [
        {
          label: "App User ID",
          value: `app.user.id`,
        },
        {
          label: "App User Mention",
          value: `app.user.mention`,
        },
      ],
    },
  ];
}

// interactionPlaceholders lists the placeholders of the interaction or event
// the flow runs with. Resumed sub-flows reach earlier ones through a prefix.
function interactionPlaceholders(
  contextType?: FlowContextType,
  prefix = "",
  labelPrefix = ""
) {
  const res = [
    {
      label: `${labelPrefix}User`,
      placeholders: [
        {
          label: "User",
          value: `${prefix}user`,
        },
        {
          label: "User ID",
          value: `${prefix}user.id`,
        },
        {
          label: "User Mention",
          value: `${prefix}user.mention`,
        },
        {
          label: "User Username",
          value: `${prefix}user.username`,
        },
        {
          label: "User Display Name",
          value: `${prefix}user.display_name`,
        },
        {
          label: "User Nickname",
          value: `${prefix}user.nick`,
        },
        {
          label: "User Avatar URL",
          value: `${prefix}user.avatar_url`,
        },
        {
          label: "User Banner URL",
          value: `${prefix}user.banner_url`,
        },
      ],
    },
    {
      label: `${labelPrefix}Server`,
      placeholders: [
        {
          label: "Server ID",
          value: `${prefix}guild.id`,
        },
      ],
    },
    {
      label: `${labelPrefix}Channel`,
      placeholders: [
        {
          label: "Channel ID",
          value: `${prefix}channel.id`,
        },
      ],
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

function useCommandPlaceholders() {
  const nodes = useNodes();

  const argNodes = useMemo(
    () => nodes.filter((n) => n.type === "option_command_argument"),
    [nodes]
  );

  const contextType = useFlowContext((c) => c.type);
  if (contextType !== "command") {
    return [];
  }

  // TODO: take arg type into account
  return [
    {
      label: "Command",
      placeholders: argNodes.map((n) => ({
        label: `Command Arg '${n.data.name}'`,
        value: `arg('${n.data.name}')`,
      })),
    },
  ];
}

// Sub-flows run with the interaction that resumed them, so placeholders of the
// interaction or event before a resume point are only reachable via origin and
// previous.
function useResumePlaceholders() {
  const nodes = useNodes();
  const edges = useEdges();
  const contextType = useFlowContext((c) => c.type);

  return useMemo(() => {
    const selected = nodes.find((n) => n.selected);
    if (!selected) {
      return [];
    }

    const depth = getResumeDepth(selected.id, nodes, edges);
    if (depth === 0) {
      return [];
    }

    const res = interactionPlaceholders(contextType, "origin.", "Original ");
    if (depth > 1) {
      // Whether previous was a button, select menu or modal isn't tracked here.
      res.push(...interactionPlaceholders(undefined, "previous.", "Previous "));
    }

    return res;
  }, [nodes, edges, contextType]);
}

// getResumeDepth counts the resume points between the root and a node.
function getResumeDepth(nodeId: string, nodes: Node[], edges: Edge[]) {
  const nodeTypes = new Map(nodes.map((n) => [n.id, n.type]));
  const incoming = new Map<string, Edge[]>();
  for (const edge of edges) {
    const targetEdges = incoming.get(edge.target) ?? [];
    targetEdges.push(edge);
    incoming.set(edge.target, targetEdges);
  }

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

function useNodePlaceholders() {
  const nodes = useNodes();
  const edges = useEdges();

  // Optimize or debounce this?
  return useMemo(() => {
    let parents: Node<NodeData>[] = [];
    for (const node of nodes) {
      if (node.selected) {
        parents = getParentNodes(node, nodes, edges);
        break;
      }
    }

    const nodeItems: { label: string; value: string }[] = [];
    const resultKeyItems: { label: string; value: string }[] = [];
    const componentItems: { label: string; value: string }[] = [];

    const seenResultKeys = new Set<string>();

    for (const parent of parents) {
      if (
        parent.type?.startsWith("action_") ||
        parent.type === "control_error_handler"
      ) {
        let label = parent.data.custom_label;
        if (!label) {
          const data = getNodeValues(parent.type!);
          label = data.defaultTitle;
        }

        nodeItems.push({ label, value: `result('${parent.id}')` });
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

      if (parent?.type === "suspend_response_modal") {
        if (!parent.data.modal_data?.components) {
          continue;
        }

        for (const row of parent.data.modal_data.components) {
          if (!row?.components) {
            continue;
          }

          for (const component of row.components) {
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
      res.push({
        label: "Modal Inputs",
        placeholders: componentItems,
      });
    }

    if (resultKeyItems.length > 0) {
      res.push({
        label: "Temporary Variables",
        placeholders: resultKeyItems,
      });
    }

    if (nodeItems.length > 0) {
      res.push({
        label: "Node Results",
        placeholders: nodeItems,
      });
    }

    return res;
  }, [nodes, edges]);
}

function getParentNodes(current: Node, nodes: Node[], edges: Edge[]) {
  const res: Node[] = [];
  const visited = new Set<string>();

  function traverse(node: Node) {
    if (visited.has(node.id)) {
      return;
    }
    visited.add(node.id);

    const incomers = getIncomers(node, nodes, edges);
    for (const incomer of incomers) {
      res.push(incomer);
      traverse(incomer);
    }
  }

  traverse(current);
  return res;
}
