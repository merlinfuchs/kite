import { VariableIcon } from "lucide-react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";
import { Edge, getIncomers, Node, useEdges, useNodes } from "@xyflow/react";
import { useMemo } from "react";
import { getNodeValues } from "@/lib/flow/nodes";
import { useFlowContext } from "@/lib/flow/context";

export default function FlowPlaceholderExplorer({
  onSelect,
}: {
  onSelect: (value: string) => void;
}) {
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
          label: "User Discriminator",
          value: `${baseKey}.user.discriminator`,
        },
        {
          label: "User Display Name",
          value: `${baseKey}.user.display_name`,
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
  const items = useMemo(() => {
    let parents: Node[] = [];
    for (const node of nodes) {
      if (node.selected) {
        parents = getParentNodes(node, nodes, edges);
        break;
      }
    }

    return parents
      .filter((n) => n.type?.startsWith("action_"))
      .map((n) => {
        let label = n.data.custom_label as string;
        if (!label) {
          const data = getNodeValues(n.type!);
          label = data.defaultTitle;
        }

        let value: string;
        if (numericRegex.test(n.id)) {
          value = `nodes[${n.id}].result`;
        } else {
          value = `nodes.${n.id}.result`;
        }

        // TODO: take node result type into account
        return { label, value };
      });
  }, [nodes, edges]);

  return [
    {
      label: "Node Results",
      placeholders: items,
    },
  ];
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
