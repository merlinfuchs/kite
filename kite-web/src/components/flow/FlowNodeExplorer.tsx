import { NodeValues, nodeTypes } from "@/lib/flow/nodes";
import { getId } from "@/lib/flow/util";
import clsx from "clsx";
import { DragEvent, useState } from "react";
import { useReactFlow } from "reactflow";

const nodeCategories = {
  entry: [
    {
      title: "Commands",
      nodeTypes: [
        "entry_command",
        "option_text",
        "option_number",
        "option_user",
        "option_channel",
        "option_role",
        "option_attachment",
      ],
    },
    {
      title: "Events",
      nodeTypes: ["entry_event"],
    },
    {
      title: "Errors",
      nodeTypes: ["entry_error"],
    },
  ],
  action: [
    {
      title: "Messages",
      nodeTypes: ["action_response_text", "action_message_create"],
    },
    {
      title: "Debugging",
      nodeTypes: ["action_log"],
    },
  ],
  condition: [
    {
      title: "Conditions",
      nodeTypes: ["condition"],
    },
  ],
};

type NodeCategory = keyof typeof nodeCategories;

export default function FlowNodeExplorer() {
  const [category, setCategory] = useState<NodeCategory>("entry");

  const sections = nodeCategories[category];

  return (
    <div className="w-96 h-full flex flex-col">
      <div className="p-5 flex-none mb-2">
        <div className="text-xl font-bold text-gray-100 mb-2">
          Block Explorer
        </div>
        <div className="text-gray-300">
          With Blocks you define what your bot does and how it works.
        </div>
      </div>
      <NodeCategories category={category} setCategory={setCategory} />
      <div className="overflow-y-auto flex-auto space-y-3 px-2 pb-5">
        {sections.map((section, i) => (
          <div key={i}>
            <div className="text-gray-300 font-medium mb-2 px-1">
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
    <div className="flex space-x-3 text-lg mb-3 px-5 justify-between text-gray-300 border-b-2 border-dark-5">
      <div onClick={() => setCategory("entry")} className="cursor-pointer">
        <div
          className={clsx(
            "pb-2 px-3 hover:text-white",
            category === "entry" && "text-white"
          )}
        >
          Entries
        </div>
        <div
          className={clsx("h-1 rounde", category === "entry" && "bg-primary")}
        ></div>
      </div>
      <div onClick={() => setCategory("action")} className="cursor-pointer">
        <div
          className={clsx(
            "pb-2 px-3 hover:text-white",
            category === "action" && "text-white"
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
            "pb-2 px-3 hover:text-white",
            category === "condition" && "text-white"
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
        id: getId(),
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
          <div className="font-medium text-gray-100 leading-5 mb-1 truncate">
            {values.defaultTitle}
          </div>
          <div className="text-sm text-gray-300">
            {values.defaultDescription}
          </div>
        </div>
      </div>
    </div>
  );
}
