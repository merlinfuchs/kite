import {
  isNodeTypeAvailable,
  NodeCategory,
  nodeCategories,
} from "@/lib/flow/categories";
import { useFlowContext } from "@/lib/flow/context";
import { NodeValues, createNode, getNodeValues } from "@/lib/flow/nodes";
import { useReactFlow, useStore } from "@xyflow/react";
import { ChevronDownIcon, ChevronRightIcon, SearchIcon } from "lucide-react";
import { DragEvent, useMemo, useState } from "react";
import DynamicIcon from "../icons/DynamicIcon";
import Link from "next/link";
import { useAppId } from "@/lib/hooks/params";
import { track } from "@/lib/analytics";
import { useMissingIntegrations } from "@/lib/integrations/hooks";
import { Input } from "../ui/input";
import { ScrollArea } from "../ui/scroll-area";
import { cn } from "@/lib/utils";
import { useFeaturePlan } from "@/lib/hooks/api";
import { Features } from "@/lib/types/wire.gen";

const allCategories = Object.keys(nodeCategories) as NodeCategory[];

const categoryInfo: Record<
  NodeCategory,
  { title: string; tab: string; description: string }
> = {
  action: {
    title: "Action Blocks",
    tab: "Action",
    description: "With Action Blocks you can perform actions with your app.",
  },
  control_flow: {
    title: "Control Flow Blocks",
    tab: "Control",
    description: "With Control Flow Blocks you define how your app behaves.",
  },
  option: {
    title: "Option Blocks",
    tab: "Option",
    description: "With Option Blocks you add option to other blocks.",
  },
  integration: {
    title: "Integration Blocks",
    tab: "Integrations",
    description:
      "With Integration Blocks you can use other services in your app.",
  },
};

export default function FlowNodeExplorer({
  category,
  searchCategories = allCategories,
  onNodeSelect,
  onPick,
  hideHeader,
}: {
  category: NodeCategory;
  // Searching finds blocks of these categories too, so a block is found
  // without knowing its tab.
  searchCategories?: NodeCategory[];
  onNodeSelect?: () => void;
  // Called with the type of the clicked block instead of adding it.
  onPick?: (type: string) => void;
  hideHeader?: boolean;
}) {
  const appId = useAppId();
  const contextType = useFlowContext((c) => c.type);

  const [search, setSearch] = useState("");
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const toggleCollapsed = (title: string) =>
    setCollapsed((c) => ({ ...c, [title]: !c[title] }));

  const isSearching = search.trim() !== "";

  const sections = useMemo(() => {
    const categories = isSearching
      ? [category, ...searchCategories.filter((c) => c !== category)]
      : [category];
    return categories
      .flatMap((c) => nodeCategories[c].map((s) => ({ ...s, category: c })))
      .map((s) => ({
        ...s,
        nodes: s.nodeTypes.map((t) => {
          const nodeValues = getNodeValues(t);
          return {
            values: nodeValues,
            type: t,
          };
        }),
      }));
  }, [category, searchCategories, isSearching]);

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
    <div className="w-full h-full flex flex-col min-h-0 overflow-hidden">
      <div className={cn("flex-none", hideHeader ? "p-3 pb-2" : "p-5")}>
        {!hideHeader && (
          <>
            <div className="text-xl font-bold text-foreground mb-2">
              {categoryInfo[category].title}
            </div>
            <div className="text-muted-foreground mb-5">
              {categoryInfo[category].description}
              {category === "integration" && (
                <>
                  {" "}
                  <Link
                    href={{
                      pathname: "/apps/[appId]/integrations",
                      query: { appId },
                    }}
                    target="_blank"
                    className="text-primary hover:underline"
                  >
                    Manage integrations
                  </Link>
                </>
              )}
            </div>
          </>
        )}
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
      <ScrollArea className="flex-1 min-h-0 mr-1">
        <div className="space-y-3 pl-3 pr-1 pb-5">
          {filteredSections.map((section) => {
            const isCollapsed = !isSearching && collapsed[section.title];
            return (
              <div key={section.title}>
                <button
                  type="button"
                  onClick={() => toggleCollapsed(section.title)}
                  className="w-full flex items-center gap-1 text-foreground font-medium mb-2 px-2 hover:text-foreground/80"
                >
                  {isCollapsed ? (
                    <ChevronRightIcon className="h-4 w-4 flex-none" />
                  ) : (
                    <ChevronDownIcon className="h-4 w-4 flex-none" />
                  )}
                  <span>{section.title}</span>
                  {section.category !== category && (
                    <span className="ml-1 text-muted-foreground text-sm font-normal">
                      {categoryInfo[section.category].tab}
                    </span>
                  )}
                  <span className="ml-auto text-muted-foreground text-sm font-normal">
                    {section.nodes.length}
                  </span>
                </button>
                {!isCollapsed && (
                  <div className="space-y-2">
                    {section.nodes.map((node) => (
                      <AvailableNode
                        key={node.type}
                        type={node.type}
                        values={node.values}
                        onSelect={onNodeSelect}
                        onPick={onPick}
                      />
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}

function AvailableNode({
  type,
  values,
  onSelect,
  onPick,
}: {
  type: string;
  values: NodeValues;
  onSelect?: () => void;
  onPick?: (type: string) => void;
}) {
  const appId = useAppId();
  const missingIntegrations = useMissingIntegrations(type);
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
    if (onPick) {
      onPick(type);
      return;
    }

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
    track("block_added", { type, source: onSelect ? "drawer" : "click" });
    onSelect?.();
  }

  return (
    <div
      className="p-2 hover:bg-muted rounded-md relative select-none cursor-grab"
      onDragStart={onStartDrag}
      onClick={onClick}
      draggable={!onSelect && !onPick}
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
          <div className="flex items-center gap-2 mb-1">
            <div className="font-medium text-foreground leading-5 truncate">
              {values.defaultTitle}
            </div>
            {values.premiumFeature && (
              <PlanBadge feature={values.premiumFeature} />
            )}
            {missingIntegrations.map((integration) => (
              <Link
                key={integration.id}
                href={{
                  pathname: "/apps/[appId]/integrations",
                  query: { appId },
                }}
                target="_blank"
                onClick={(e) => e.stopPropagation()}
                className="bg-amber-500/15 text-amber-600 dark:text-amber-400 rounded px-1.5 text-xs font-medium flex-none hover:underline"
              >
                Enable {integration.name}
              </Link>
            ))}
          </div>
          <div className="text-sm text-muted-foreground">
            {values.defaultDescription}
          </div>
        </div>
      </div>
    </div>
  );
}

// PlanBadge names the cheapest plan that includes a block's feature.
function PlanBadge({ feature }: { feature: keyof Features }) {
  const plan = useFeaturePlan(feature);

  return (
    <div className="bg-primary/15 text-primary rounded px-1.5 text-xs font-medium flex-none">
      {plan?.title ?? "Premium"}
    </div>
  );
}
