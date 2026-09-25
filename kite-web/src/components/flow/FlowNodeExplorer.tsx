import {
  isNodeTypeAvailable,
  NodeCategory,
  nodeCategories,
} from "@/lib/flow/categories";
import { useFlowContext } from "@/lib/flow/context";
import { NodeValues, createNode, getNodeValues } from "@/lib/flow/nodes";
import { useReactFlow, useStore } from "@xyflow/react";
import { SearchIcon } from "lucide-react";
import { DragEvent, useMemo, useState } from "react";
import DynamicIcon from "../icons/DynamicIcon";
import { Input } from "../ui/input";
import { ScrollArea } from "../ui/scroll-area";

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
            isNodeTypeAvailable(n.type, contextType) &&
            (n.values.defaultTitle.toLowerCase().includes(normalizedSearch) ||
              n.values.defaultDescription
                .toLowerCase()
                .includes(normalizedSearch))
        ),
      }))
      .filter((s) => s.nodes.length > 0);
  }, [sections, contextType, search]);

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
  const { addNodes, addEdges, getViewport } = useReactFlow();
  // The canvas is not the window: it sits right of the w-96 block explorer,
  // inside a dialog. Ask react-flow for its own pane size rather than reading
  // window dimensions, so blocks don't spawn off-centre or under the menu.
  const paneWidth = useStore((s) => s.width);
  const paneHeight = useStore((s) => s.height);

  function onStartDrag(e: DragEvent) {
    e.dataTransfer.setData("application/reactflow", type);
    e.dataTransfer.effectAllowed = "move";
  }

  function onClick() {
    // Spawn in the middle of what the user is currently looking at, with a
    // small jitter so repeated clicks don't stack blocks exactly on top of
    // each other. Using flow coordinates directly would put the block next to
    // the entry node, off-screen for anything but a freshly opened flow.
    const { x, y, zoom } = getViewport();
    const center = {
      x: (-x + paneWidth / 2) / zoom,
      y: (-y + paneHeight / 2) / zoom,
    };

    const [nodes, edges] = createNode(type, {
      x: center.x + 200 * Math.random() - 100,
      y: center.y + 100 * Math.random() - 50,
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
