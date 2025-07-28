import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { getNodeValues } from "@/lib/flow/nodes";
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

  const placeholders = useMemo(
    () => [...commandPlaceholders, ...globalPlaceholders, ...nodePlaceholders],
    [commandPlaceholders, globalPlaceholders, nodePlaceholders]
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

  const res = [
    {
      label: "User",
      placeholders: [
        {
          label: "User",
          value: `user`,
        },
        {
          label: "User ID",
          value: `user.id`,
        },
        {
          label: "User Mention",
          value: `user.mention`,
        },
        {
          label: "User Username",
          value: `user.username`,
        },
        {
          label: "User Display Name",
          value: `user.display_name`,
        },
        {
          label: "User Nickname",
          value: `user.nick`,
        },
        {
          label: "User Avatar URL",
          value: `user.avatar_url`,
        },
        {
          label: "User Banner URL",
          value: `user.banner_url`,
        },
      ],
    },
    {
      label: "Server",
      placeholders: [
        {
          label: "Server ID",
          value: `guild.id`,
        },
      ],
    },
    {
      label: "Channel",
      placeholders: [
        {
          label: "Channel ID",
          value: `channel.id`,
        },
      ],
    },
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

  if (contextType === "event_discord") {
    res.push({
      label: "Message",
      placeholders: [
        { label: "Message ID", value: `message.id` },
        { label: "Message Content", value: `message.content` },
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
      if (parent.type?.startsWith("action_")) {
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

  const incomers = getIncomers(current, nodes, edges);
  for (const incomer of incomers) {
    res.push(incomer);
    res.push(...getParentNodes(incomer, nodes, edges));
  }

  return res;
}
