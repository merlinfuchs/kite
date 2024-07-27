import { NodeValues, nodeTypes } from "@/lib/flow/nodes";
import clsx from "clsx";
import { DragEvent, useState } from "react";
import { useReactFlow } from "@xyflow/react";
import { getUniqueId } from "@/lib/utils";

const nodeCategories = {
  option: [
    {
      title: "Commands",
      nodeTypes: [
        "option_command_text",
        "option_command_number",
        "option_command_user",
        "option_command_channel",
        "option_command_role",
        "option_command_attachment",
      ],
    },
  ],
  action: [
    {
      title: "Messages",
      nodeTypes: ["action_response_create", "action_message_create"],
    },
    {
      title: "Debugging",
      nodeTypes: ["action_log"],
    },
  ],
  condition: [
    {
      title: "Conditions",
      nodeTypes: ["condition_compare", "condition_permissions"],
    },
  ],
};

type NodeCategory = keyof typeof nodeCategories;

export default function FlowNodeExplorer() {
  const [category, setCategory] = useState<NodeCategory>("action");

  const sections = nodeCategories[category];

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
      <div className="overflow-y-auto flex-auto space-y-3 px-2 pb-5">
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
      <div onClick={() => setCategory("condition")} className="cursor-pointer">
        <div
          className={clsx(
            "pb-2 px-3 hover:text-foreground",
            category === "condition" && "text-foreground"
          )}
        >
          Conditions
        </div>
        <div
          className={clsx(
            "h-1 rounde",
            category === "condition" && "bg-primary"
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
  const { addNodes } = useReactFlow();

  function onStartDrag(e: DragEvent) {
    e.dataTransfer.setData("application/reactflow", type);
    e.dataTransfer.effectAllowed = "move";
  }

  function onClick() {
    addNodes([
      {
        id: getUniqueId().toString(),
        type,
        position: { x: 0, y: 0 },
        data: {},
      },
    ]);
  }

  return (
    <div
      className="p-2 hover:bg-dark-4 rounded relative select-none cursor-grab"
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
