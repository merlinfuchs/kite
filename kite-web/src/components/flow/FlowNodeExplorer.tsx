import { NodeValues, nodeTypes } from "@/lib/flow/nodes";
import { getId } from "@/lib/flow/util";
import { DragEvent } from "react";
import { useReactFlow } from "reactflow";

export default function FlowNodeExplorer() {
  return (
    <div className="w-96 h-full flex flex-col">
      <div className="p-5 flex-none">
        <div className="text-xl font-bold text-gray-100 mb-2">
          Block Explorer
        </div>
        <div className="text-gray-300">
          With Blocks you define what your bot does and how it works.
        </div>
      </div>
      <div className="overflow-y-auto flex-auto space-y-2 px-2">
        {Object.entries(nodeTypes).map(([type, values]) => (
          <AvailableNode key={type} type={type} values={values} />
        ))}
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
