import { Position, useReactFlow } from "@xyflow/react";
import { NodeProps, NodeType } from "../../lib/flow/data";
import FlowNodeBase from "./FlowNodeBase";
import FlowNodeHandle from "./FlowNodeHandle";
import { controlColor, createNode } from "@/lib/flow/nodes";
import { Button } from "../ui/button";
import { PlusIcon } from "lucide-react";
import { getUniqueId } from "@/lib/utils";

export default function FlowNodeConditionUser(props: NodeProps) {
  const { setNodes, setEdges, getNode } = useReactFlow<NodeType>();

  function addItem() {
    const node = getNode(props.id);
    if (!node) return;

    const [newNodes, newEdges] = createNode("control_condition_item_user", {
      x: node.position.x + 50,
      y: node.position.y + 300,
    });

    setNodes((nodes) => [...nodes, ...newNodes]);
    setEdges((edges) => [
      ...edges,
      {
        id: getUniqueId().toString(),
        source: props.id,
        target: newNodes[0].id,
        type: "fixed",
      },
    ]);
  }

  return (
    <FlowNodeBase {...props}>
      <Button
        size="icon"
        variant="outline"
        className="text-xs leading-none h-7 w-7 absolute -right-2 -bottom-2 hover:bg-background hover:border-primary"
        onClick={addItem}
      >
        <PlusIcon className="h-4 w-4" />
      </Button>
      <FlowNodeHandle type="target" position={Position.Top} />
      <FlowNodeHandle
        type="source"
        color={controlColor}
        position={Position.Bottom}
        isConnectable={false}
        size="small"
      />
    </FlowNodeBase>
  );
}
