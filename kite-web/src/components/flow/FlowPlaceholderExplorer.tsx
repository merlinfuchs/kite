import { VariableIcon } from "lucide-react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";
import { Edge, getIncomers, Node, useEdges, useNodes } from "@xyflow/react";
import { useMemo } from "react";
import { getNodeValues } from "@/lib/flow/nodes";
import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/data";

export default function FlowPlaceholderExplorer({
  onSelect,
}: {
  onSelect: (value: string) => void;
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
      <PlaceholderExplorer onSelect={onSelect} placeholders={placeholders}>
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

  const baseKey = useMemo(
    () => (contextType === "event_discord" ? "event" : "interaction"),
    [contextType]
  );

  const res = [
    {
      label: "User",
      placeholders: [
        {
          label: "User ID",
          value: `${baseKey}.user.id`,
        },
        {
          label: "User Mention",
          value: `${baseKey}.user.mention`,
        },
        {
          label: "User Username",
          value: `${baseKey}.user.username`,
        },
        {
          label: "User Display Name",
          value: `${baseKey}.user.display_name`,
        },
        {
          label: "User Nickname",
          value: `${baseKey}.user.nick`,
        },
        {
          label: "User Avatar URL",
          value: `${baseKey}.user.avatar_url`,
        },
        {
          label: "User Banner URL",
          value: `${baseKey}.user.banner_url`,
        },
      ],
    },
    {
      label: "Server",
      placeholders: [
        {
          label: "Server ID",
          value: `${baseKey}.guild.id`,
        },
      ],
    },
    {
      label: "Channel",
      placeholders: [
        {
          label: "Channel ID",
          value: `${baseKey}.channel.id`,
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
        { label: "Message ID", value: `${baseKey}.message.id` },
        { label: "Message Content", value: `${baseKey}.message.content` },
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
        value: `interaction.command.args.${n.data.name}`,
      })),
    },
  ];
}

const numericRegex = /^[0-9]+$/;

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
    const componentItems: { label: string; value: string }[] = [];

    for (const parent of parents) {
      if (parent.type?.startsWith("action_")) {
        let label = parent.data.custom_label;
        if (!label) {
          const data = getNodeValues(parent.type!);
          label = data.defaultTitle;
        }

        let value: string;
        if (numericRegex.test(parent.id)) {
          value = `nodes[${parent.id}].result`;
        } else {
          value = `nodes.${parent.id}.result`;
        }

        nodeItems.push({ label, value });
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
              value: `interaction.components.${component.custom_id}`,
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
