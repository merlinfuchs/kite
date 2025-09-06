import { useFlowContext } from "@/lib/flow/context";
import { NodeValues, createNode, getNodeValues } from "@/lib/flow/nodes";
import { useReactFlow } from "@xyflow/react";
import { SearchIcon } from "lucide-react";
import { DragEvent, useMemo, useState } from "react";
import DynamicIcon from "../icons/DynamicIcon";
import { Input } from "../ui/input";
import { ScrollArea } from "../ui/scroll-area";

const nodeCategories = {
  option: [
    {
      title: "Commands",
      nodeTypes: [
        "option_command_argument",
        "option_command_permissions",
        "option_command_contexts",
      ],
      contextTypes: ["command"],
    },
    {
      title: "Events",
      nodeTypes: ["option_command_contexts"],
      contextTypes: ["event_discord"],
    },
    /* {
      title: "Events",
      nodeTypes: ["option_event_filter"],
    }, */
  ],
  action: [
    {
      title: "Responses",
      nodeTypes: [
        "action_response_create",
        "action_response_edit",
        "action_response_delete",
        "action_response_defer",
        "suspend_response_modal",
      ],
      contextTypes: ["command", "component_button"],
    },
    {
      title: "Messages",
      nodeTypes: [
        "action_message_create",
        "action_message_edit",
        "action_message_delete",
        "action_message_get",
        "action_private_message_create",
        "action_message_reaction_create",
        "action_message_reaction_delete",
      ],
      contextTypes: null,
    },
    {
      title: "Members",
      nodeTypes: [
        "action_member_ban",
        "action_member_unban",
        "action_member_kick",
        "action_member_timeout",
        "action_member_edit",
        "action_member_get",
      ],
      contextTypes: null,
    },
    {
      title: "Users",
      nodeTypes: ["action_user_get"],
      contextTypes: null,
    },
    {
      title: "Roles",
      nodeTypes: [
        "action_member_role_add",
        "action_member_role_remove",
        "action_role_get",
      ],
      contextTypes: null,
    },

    {
      title: "Servers",
      nodeTypes: ["action_guild_get"],
      contextTypes: null,
    },
    {
      title: "Channels",
      nodeTypes: [
        "action_channel_create",
        "action_channel_edit",
        "action_channel_delete",
        "action_channel_get",
        "action_thread_create",
        "action_thread_member_add",
        "action_thread_member_remove",
      ],
      contextTypes: null,
    },
    {
      title: "Stored Variables",
      nodeTypes: [
        "action_variable_set",
        "action_variable_delete",
        "action_variable_get",
      ],
      contextTypes: null,
    },
    {
      title: "Roblox",
      nodeTypes: ["action_roblox_user_get"],
      contextTypes: null,
    },
    {
      title: "Other Actions",
      nodeTypes: [
        "action_expression_evaluate",
        "action_ai_chat_completion",
        "action_ai_web_search",
        "action_http_request",
        "action_random_generate",
        "action_log",
      ],
      contextTypes: null,
    },
  ],
  control_flow: [
    {
      title: "Conditions",
      nodeTypes: [
        "control_condition_compare",
        "control_condition_user",
        "control_condition_channel",
        "control_condition_role",
      ],
      contextTypes: null,
    },
    {
      title: "Loops",
      nodeTypes: ["control_loop", "control_loop_exit"],
      contextTypes: null,
    },
    {
      title: "Errors",
      nodeTypes: ["control_error_handler"],
      contextTypes: null,
    },
    {
      title: "Others",
      nodeTypes: ["control_sleep"],
      contextTypes: null,
    },
  ],
};

type NodeCategory = keyof typeof nodeCategories;

export default function FlowNodeExplorer({
  category,
}: {
  category: NodeCategory;
}) {
  const contextType = useFlowContext((c) => c.type);

  const [search, setSearch] = useState("");

  const sections = useMemo(() => {
    return nodeCategories[category].map((s) => ({
      ...s,
      nodes: s.nodeTypes.map((t) => {
        const nodeValues = getNodeValues(t);
        return {
          values: nodeValues,
          type: t,
        };
      }),
    }));
  }, [category]);

  const filteredSections = useMemo(() => {
    const normalizedSearch = search.toLowerCase().trim();
    return sections
      .map((s) => ({
        ...s,
        nodes: s.nodes.filter(
          (n) =>
            n.values.defaultTitle.toLowerCase().includes(normalizedSearch) ||
            n.values.defaultDescription.toLowerCase().includes(normalizedSearch)
        ),
      }))
      .filter(
        (s) =>
          s.nodes.length > 0 &&
          (!s.contextTypes || s.contextTypes.includes(contextType))
      );
  }, [category, contextType, search]);

  return (
    <div className="w-full h-full flex flex-col">
      <div className="p-5 flex-none">
        <div className="text-xl font-bold text-foreground mb-2">
          {category === "action"
            ? "Action"
            : category === "control_flow"
            ? "Control Flow"
            : "Option"}{" "}
          Blocks
        </div>
        <div className="text-muted-foreground mb-5">
          {category === "action"
            ? "With Action Blocks you can perform actions with your app."
            : category === "control_flow"
            ? "With Control Flow Blocks you define how your app behaves."
            : "With Option Blocks you add option to other blocks."}
        </div>
        <div className="relative">
          <Input
            placeholder="Search ..."
            className="pl-10"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <SearchIcon className="absolute size-5 left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        </div>
      </div>
      <ScrollArea className="flex-auto mr-1">
        <div className="space-y-3 pl-3 pr-1 pb-5">
          {filteredSections.map((section, i) => (
            <div key={i}>
              <div className="text-foreground font-medium mb-2 px-2">
                {section.title}
              </div>
              <div className="space-y-2">
                {section.nodes.map((node) => (
                  <AvailableNode
                    key={node.type}
                    type={node.type}
                    values={node.values}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}

function AvailableNode({ type, values }: { type: string; values: NodeValues }) {
  const { addNodes, addEdges } = useReactFlow();

  function onStartDrag(e: DragEvent) {
    e.dataTransfer.setData("application/reactflow", type);
    e.dataTransfer.effectAllowed = "move";
  }

  function onClick() {
    const [nodes, edges] = createNode(type, {
      x: 0 + 200 * Math.random() - 100,
      y: 0 + 100 * Math.random() + 200,
    });
    addNodes(nodes);
    addEdges(edges);
  }

  return (
    <div
      className="p-2 hover:bg-muted rounded-md relative select-none cursor-grab"
      onDragStart={onStartDrag}
      onClick={onClick}
      draggable
    >
      <div className="flex items-start space-x-3">
        <div
          className="rounded-md w-8 h-8 flex justify-center items-center flex-none"
          style={{ backgroundColor: values.color }}
        >
          <DynamicIcon
            name={values.icon as any}
            className="h-5 w-5 text-white"
          />
        </div>
        <div className="overflow-hidden">
          <div className="font-medium text-foreground leading-5 mb-1 truncate">
            {values.defaultTitle}
          </div>
          <div className="text-sm text-muted-foreground">
            {values.defaultDescription}
          </div>
        </div>
      </div>
    </div>
  );
}
