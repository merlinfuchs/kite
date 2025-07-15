import { NodeValues, createNode, nodeTypes } from "@/lib/flow/nodes";
import clsx from "clsx";
import { DragEvent, useMemo, useState } from "react";
import { useReactFlow } from "@xyflow/react";
import { getUniqueId } from "@/lib/utils";
import { useFlowContext } from "@/lib/flow/context";
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
        "action_response_defer",
        "action_response_edit",
        "action_response_delete",
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
      ],
      contextTypes: null,
    },
    {
      title: "Roles",
      nodeTypes: ["action_member_role_add", "action_member_role_remove"],
      contextTypes: null,
    },
    {
      title: "Variables",
      nodeTypes: [
        "action_variable_set",
        "action_variable_delete",
        "action_variable_get",
      ],
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
      title: "Others",
      nodeTypes: ["control_sleep"],
      contextTypes: null,
    },
  ],
};

type NodeCategory = keyof typeof nodeCategories;

export default function FlowNodeExplorer() {
  const [category, setCategory] = useState<NodeCategory>("action");

  const contextType = useFlowContext((c) => c.type);

  const sections = useMemo(() => {
    const sections = nodeCategories[category];
    if (!sections) return [];

    return sections.filter(
      (s) => !s.contextTypes || s.contextTypes.includes(contextType)
    );
  }, [category, contextType]);

  return (
    <div className="w-96 h-full flex flex-col bg-muted/40">
      <div className="p-5 flex-none mb-2">
        <div className="text-xl font-bold text-foreground mb-2">
          Block Explorer
        </div>
        <div className="text-muted-foreground">
          With Blocks you define what your bot does and how it works.
        </div>
      </div>
      <NodeCategories category={category} setCategory={setCategory} />
      <ScrollArea className="flex-auto mr-1">
        <div className="space-y-3 pl-2 pr-1 pb-5">
          {sections.map((section, i) => (
            <div key={i}>
              <div className="text-foreground font-medium mb-2 px-1">
                {section.title}
              </div>
              <div className="space-y-2">
                {section.nodeTypes.map((type) => (
                  <AvailableNode
                    key={type}
                    type={type}
                    values={nodeTypes[type]}
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

function NodeCategories({
  category,
  setCategory,
}: {
  category: NodeCategory;
  setCategory: (tab: NodeCategory) => void;
}) {
  return (
    <div className="flex space-x-3 text-lg mb-3 px-5 justify-between text-muted-foreground border-b-2 border-dark-5">
      <div onClick={() => setCategory("action")} className="cursor-pointer">
        <div
          className={clsx(
            "pb-2 px-3 hover:text-foreground",
            category === "action" && "text-foreground"
          )}
        >
          Actions
        </div>
        <div
          className={clsx("h-1 rounde", category === "action" && "bg-primary")}
        ></div>
      </div>
      <div
        onClick={() => setCategory("control_flow")}
        className="cursor-pointer"
      >
        <div
          className={clsx(
            "pb-2 px-3 hover:text-foreground",
            category === "control_flow" && "text-foreground"
          )}
        >
          Control Flow
        </div>
        <div
          className={clsx(
            "h-1 rounde",
            category === "control_flow" && "bg-primary"
          )}
        ></div>
      </div>
      <div onClick={() => setCategory("option")} className="cursor-pointer">
        <div
          className={clsx(
            "pb-2 px-3 hover:text-foreground",
            category === "option" && "text-foreground"
          )}
        >
          Options
        </div>
        <div
          className={clsx("h-1 rounde", category === "option" && "bg-primary")}
        ></div>
      </div>
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
          <values.icon className="h-5 w-5 text-white" />
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
