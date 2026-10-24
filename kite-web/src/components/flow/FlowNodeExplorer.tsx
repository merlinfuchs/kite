import {
  isNodeTypeAvailable,
  NodeCategory,
  nodeCategories,
} from "@/lib/flow/categories";
import { useFlowContext } from "@/lib/flow/context";
import {
  canConnect,
  createEdge,
  createNode,
  getNodeValues,
  getTopLevelSelectedNodes,
  NodeValues,
} from "@/lib/flow/nodes";
import { NodeData } from "@/lib/flow/dataSchema";
import { FlowClipboard } from "@/lib/flow/clipboard";
import { setSharedClipboard } from "@/lib/hooks/flowClipboard";
import { Edge, Node, useReactFlow, useStore } from "@xyflow/react";
import {
  ChevronDownIcon,
  ChevronRightIcon,
  CopyIcon,
  CopyPlusIcon,
  LinkIcon,
  MinusIcon,
  PlusIcon,
  SearchIcon,
} from "lucide-react";
import {
  DragEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import DynamicIcon from "../icons/DynamicIcon";
import Link from "next/link";
import { useAppId } from "@/lib/hooks/params";
import { track } from "@/lib/analytics";
import { useMissingIntegrations } from "@/lib/integrations/hooks";
import { Input } from "../ui/input";
import { ScrollArea } from "../ui/scroll-area";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

export default function FlowNodeExplorer({
  category,
  onNodeSelect,
  hideHeader,
}: {
  category: NodeCategory;
  onNodeSelect?: () => void;
  hideHeader?: boolean;
}) {
  const contextType = useFlowContext((c) => c.type);

  const [search, setSearch] = useState("");
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const toggleCollapsed = (title: string) =>
    setCollapsed((c) => ({ ...c, [title]: !c[title] }));

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
    <div className="w-full h-full flex flex-col min-h-0 overflow-hidden">
      <div className={cn("flex-none", hideHeader ? "p-3 pb-2" : "p-5")}>
        {!hideHeader && (
          <>
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
          {filteredSections.map((section, i) => {
            const isCollapsed = !search.trim() && collapsed[section.title];
            return (
              <div key={i}>
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
}: {
  type: string;
  values: NodeValues;
  onSelect?: () => void;
}) {
  const appId = useAppId();
  const missingIntegrations = useMissingIntegrations(type);
  const {
    addNodes,
    addEdges,
    getViewport,
    getNodes,
    setNodes: editNodes,
  } = useReactFlow();
  // The canvas is not the window: it sits right of the w-96 block explorer,
  // inside a dialog. Ask react-flow for its own pane size rather than reading
  // window dimensions, so blocks don't spawn off-centre or under the menu.
  const paneWidth = useStore((s) => s.width);
  const paneHeight = useStore((s) => s.height);

  const [menuState, setMenuState] = useState<{
    open: boolean;
    x: number;
    y: number;
  }>({ open: false, x: 0, y: 0 });
  const [customCount, setCustomCount] = useState<number>(5);
  const [showCustom, setShowCustom] = useState<boolean>(false);

  const pointerStartRef = useRef<{ x: number; y: number } | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const longPressTriggeredRef = useRef(false);

  const pointerThresholdPx = 22;
  const pointerDurationMs = 400;

  const clearTimer = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    pointerStartRef.current = null;
  }, []);

  useEffect(() => {
    return () => clearTimer();
  }, [clearTimer]);

  const spawnNodes = useCallback(
    (count: number, connectToSelected = false) => {
      const clampedCount = Math.max(1, Math.min(50, count));
      const { x, y, zoom } = getViewport();
      const currentNodes = getNodes() as Node<NodeData>[];
      const topLevelSelected = getTopLevelSelectedNodes(currentNodes);
      const singleSelected =
        topLevelSelected.length === 1 ? topLevelSelected[0] : null;

      const allNewNodes: Node<NodeData>[] = [];
      const allNewEdges: Edge[] = [];

      if (connectToSelected && singleSelected) {
        if (type.startsWith("option_")) {
          const [newNodes, newEdges] = createNode(
            type,
            {
              x: singleSelected.position.x - 240,
              y: singleSelected.position.y,
            },
            { selected: true }
          );
          allNewNodes.push(...newNodes);
          allNewEdges.push(...newEdges);
          allNewEdges.push(createEdge(newNodes[0], singleSelected));
        } else {
          let previousNode: Node<NodeData> | null = singleSelected;
          for (let i = 0; i < clampedCount; i++) {
            const [newNodes, newEdges] = createNode(
              type,
              {
                x:
                  singleSelected.position.x +
                  (i > 0 ? (i % 2 === 0 ? 0 : 30) : 0),
                y: singleSelected.position.y + (i + 1) * 160,
              },
              { selected: true }
            );
            allNewNodes.push(...newNodes);
            allNewEdges.push(...newEdges);
            if (previousNode) {
              allNewEdges.push(createEdge(previousNode, newNodes[0]));
            }
            previousNode =
              (getNodeValues(type).outputs?.length ?? 1) > 0
                ? newNodes[0]
                : null;
          }
        }
      } else {
        // Spawn in the middle of what the user is currently looking at, with a
        // small jitter so repeated clicks don't stack blocks exactly on top of
        // each other. Using flow coordinates directly would put the block next to
        // the entry node, off-screen for anything but a freshly opened flow.
        const center = {
          x: (-x + paneWidth / 2) / zoom,
          y: (-y + paneHeight / 2) / zoom,
        };

        if (clampedCount === 1) {
          const [newNodes, newEdges] = createNode(
            type,
            {
              x: center.x + 200 * Math.random() - 100,
              y: center.y + 100 * Math.random() - 50,
            },
            { selected: true }
          );
          allNewNodes.push(...newNodes);
          allNewEdges.push(...newEdges);
        } else {
          const cols = clampedCount > 6 ? 2 : 1;
          const colWidth = 260;
          const rowHeight = 120;
          const rows = Math.ceil(clampedCount / cols);
          const startX = center.x - ((cols - 1) * colWidth) / 2;
          const startY = center.y - ((rows - 1) * rowHeight) / 2;

          for (let i = 0; i < clampedCount; i++) {
            const c = i % cols;
            const r = Math.floor(i / cols);
            const stagger = cols === 1 ? i * 20 : 0;
            const [newNodes, newEdges] = createNode(
              type,
              {
                x: startX + c * colWidth + stagger,
                y: startY + r * rowHeight,
              },
              { selected: true }
            );
            allNewNodes.push(...newNodes);
            allNewEdges.push(...newEdges);
          }
        }
      }

      editNodes((nds) =>
        nds.map((n) => (n.selected ? { ...n, selected: false } : n))
      );
      addNodes(allNewNodes);
      addEdges(allNewEdges);
      track("block_added", {
        type,
        count: clampedCount,
        source: onSelect ? "drawer" : "click",
      });

      if (clampedCount > 1) {
        toast.success(`Added ${clampedCount} ${values.defaultTitle} blocks`);
      }

      setMenuState((prev) => ({ ...prev, open: false }));
      onSelect?.();
    },
    [
      getViewport,
      getNodes,
      paneWidth,
      paneHeight,
      type,
      values.defaultTitle,
      editNodes,
      addNodes,
      addEdges,
      onSelect,
    ]
  );

  const handlePointerDown = (e: React.PointerEvent) => {
    if (e.pointerType === "mouse" && e.button !== 0) return;
    longPressTriggeredRef.current = false;
    pointerStartRef.current = { x: e.clientX, y: e.clientY };
    clearTimer();
    timerRef.current = setTimeout(() => {
      longPressTriggeredRef.current = true;
      if (typeof navigator !== "undefined" && navigator.vibrate) {
        navigator.vibrate(15);
      }
      setShowCustom(false);
      setMenuState({
        open: true,
        x: pointerStartRef.current?.x ?? e.clientX,
        y: pointerStartRef.current?.y ?? e.clientY,
      });
      clearTimer();
    }, pointerDurationMs);
  };

  const handlePointerMove = (e: React.PointerEvent) => {
    if (!pointerStartRef.current) return;
    const dx = e.clientX - pointerStartRef.current.x;
    const dy = e.clientY - pointerStartRef.current.y;
    if (Math.hypot(dx, dy) > pointerThresholdPx) {
      clearTimer();
    }
  };

  const handlePointerUpOrCancel = () => {
    clearTimer();
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    clearTimer();
    setShowCustom(false);
    setMenuState({
      open: true,
      x: e.clientX,
      y: e.clientY,
    });
  };

  const handleClick = () => {
    if (longPressTriggeredRef.current) {
      longPressTriggeredRef.current = false;
      return;
    }
    spawnNodes(1);
  };

  const copyBlock = useCallback(() => {
    const [newNodes, newEdges] = createNode(type, { x: 0, y: 0 });
    const clipboard: FlowClipboard = {
      type: "kite-flow-clipboard",
      nodes: newNodes.map(({ id, type, position, data }) => ({
        id,
        type,
        position,
        data,
      })),
      edges: newEdges.map(
        ({ id, type, source, target, sourceHandle, targetHandle }) => ({
          id,
          type,
          source,
          target,
          sourceHandle,
          targetHandle,
        })
      ),
    };

    setSharedClipboard(clipboard);
    toast.success(`Copied "${values.defaultTitle}" block`);
    setMenuState((prev) => ({ ...prev, open: false }));
  }, [type, values.defaultTitle]);

  const canConnectToSelected = useMemo(() => {
    if (!menuState.open) return false;
    const currentNodes = getNodes() as Node<NodeData>[];
    const topLevelSelected = getTopLevelSelectedNodes(currentNodes);
    if (topLevelSelected.length !== 1) return false;
    const selected = topLevelSelected[0];
    if (!selected.type) return false;
    if (type.startsWith("option_")) {
      return canConnect(type, selected.type);
    }
    return (
      (getNodeValues(selected.type).outputs?.length ?? 1) > 0 &&
      canConnect(selected.type, type)
    );
  }, [menuState.open, getNodes, type]);

  function onStartDrag(e: DragEvent) {
    e.dataTransfer.setData("application/reactflow", type);
    e.dataTransfer.effectAllowed = "move";
  }
  return (
    <>
      <div
        className="p-2 hover:bg-muted rounded-md relative select-none cursor-grab"
        onDragStart={onStartDrag}
        onClick={handleClick}
        onContextMenu={handleContextMenu}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUpOrCancel}
        onPointerCancel={handlePointerUpOrCancel}
        draggable={!onSelect}
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
                <div className="bg-primary/15 text-primary rounded px-1.5 text-xs font-medium flex-none">
                  Premium
                </div>
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

      <DropdownMenu
        open={menuState.open}
        onOpenChange={(open) => {
          if (!open) setMenuState((prev) => ({ ...prev, open: false }));
        }}
        modal={false}
      >
        <DropdownMenuTrigger asChild>
          <div
            className="fixed w-0 h-0 pointer-events-none"
            style={{ left: menuState.x, top: menuState.y }}
            aria-hidden
          />
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align="start"
          side="bottom"
          sideOffset={4}
          className="w-56 select-none z-[80]"
          onCloseAutoFocus={(e) => e.preventDefault()}
        >
          <div className="flex items-center gap-2 px-2 py-1.5 border-b border-border/40 mb-1">
            <div
              className="size-5 rounded flex items-center justify-center flex-none"
              style={{ backgroundColor: values.color }}
            >
              <DynamicIcon
                name={values.icon as any}
                className="size-3 text-white"
              />
            </div>
            <div className="text-xs font-semibold text-foreground truncate">
              {values.defaultTitle}
            </div>
          </div>

          <DropdownMenuItem onClick={() => spawnNodes(1)}>
            <PlusIcon className="size-4 mr-2" />
            <span>Add 1 block</span>
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => spawnNodes(2)}>
            <CopyPlusIcon className="size-4 mr-2" />
            <span>Add 2 blocks</span>
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => spawnNodes(3)}>
            <CopyPlusIcon className="size-4 mr-2" />
            <span>Add 3 blocks</span>
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          <button
            type="button"
            className="w-full flex items-center justify-between px-2 py-1.5 text-xs font-medium rounded hover:bg-accent hover:text-accent-foreground text-muted-foreground transition-colors"
            onClick={(e) => {
              e.stopPropagation();
              setShowCustom((prev) => !prev);
            }}
          >
            <span>Custom quantity...</span>
            {showCustom ? (
              <ChevronDownIcon className="size-3.5" />
            ) : (
              <ChevronRightIcon className="size-3.5" />
            )}
          </button>

          {showCustom && (
            <div
              className="px-2 py-1.5 space-y-1.5 bg-muted/40 rounded-md my-1"
              onClick={(e) => e.stopPropagation()}
              onKeyDown={(e) => e.stopPropagation()}
            >
              <div className="flex items-center gap-1.5">
                <Button
                  variant="outline"
                  size="icon"
                  className="size-7 flex-none"
                  onClick={() => setCustomCount((c) => Math.max(1, c - 1))}
                >
                  <MinusIcon className="size-3" />
                </Button>
                <Input
                  type="number"
                  min={1}
                  max={50}
                  value={customCount}
                  onChange={(e) =>
                    setCustomCount(
                      Math.min(50, Math.max(1, parseInt(e.target.value) || 1))
                    )
                  }
                  className="h-7 w-12 text-center px-1 text-xs"
                />
                <Button
                  variant="outline"
                  size="icon"
                  className="size-7 flex-none"
                  onClick={() => setCustomCount((c) => Math.min(50, c + 1))}
                >
                  <PlusIcon className="size-3" />
                </Button>
                <Button
                  size="sm"
                  className="h-7 text-xs px-2 ml-auto"
                  onClick={() => spawnNodes(customCount)}
                >
                  Add {customCount}
                </Button>
              </div>
            </div>
          )}

          {canConnectToSelected && (
            <>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => spawnNodes(1, true)}>
                <LinkIcon className="size-4 mr-2" />
                <span>Add & Connect to selected</span>
              </DropdownMenuItem>
            </>
          )}

          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={copyBlock}>
            <CopyIcon className="size-4 mr-2" />
            <span>Copy block</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
