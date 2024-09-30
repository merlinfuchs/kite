import { VariableIcon } from "lucide-react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";
import { Edge, getIncomers, Node, useEdges, useNodes } from "@xyflow/react";
import { useMemo } from "react";
import { getNodeValues } from "@/lib/flow/nodes";

const staticPlaceholders = [
  {
    label: "User",
    placeholders: [
      {
        label: "User ID",
        value: "interaction.user.id",
      },
      {
        label: "User Mention",
        value: "interaction.user.mention",
      },
      {
        label: "User Username",
        value: "interaction.user.username",
      },
      {
        label: "User Discriminator",
        value: "interaction.user.discriminator",
      },
      {
        label: "User Display Name",
        value: "interaction.user.display_name",
      },
      {
        label: "User Avatar URL",
        value: "interaction.user.avatar_url",
      },
      {
        label: "User Banner URL",
        value: "interaction.user.banner_url",
      },
    ],
  },
  {
    label: "Server",
    placeholders: [
      {
        label: "Server ID",
        value: "interaction.guild.id",
      },
    ],
  },
  {
    label: "Channel",
    placeholders: [
      {
        label: "Channel ID",
        value: "interaction.channel.id",
      },
    ],
  },
];

export default function FlowPlaceholderExplorer({
  onSelect,
}: {
  onSelect: (value: string) => void;
}) {
  const nodePlaceholders = useNodePlaceholders();
  const commandPlaceholders = useCommandPlaceholders();

  const placeholders = useMemo(
    () => [...commandPlaceholders, ...staticPlaceholders, ...nodePlaceholders],
    [commandPlaceholders, nodePlaceholders]
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

function useCommandPlaceholders() {
  const nodes = useNodes();

  const argNodes = useMemo(
    () => nodes.filter((n) => n.type === "option_command_argument"),
    [nodes]
  );

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

        // TODO: take node result type into account
        return {
          label,
          value: `nodes.${n.id}.result`,
        };
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
