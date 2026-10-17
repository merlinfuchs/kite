import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  isNodeTypeAvailable,
  NodeCategory,
  nodeCategories,
} from "@/lib/flow/categories";
import { useFlowContext } from "@/lib/flow/context";
import { getNodeValues, NodeValues } from "@/lib/flow/nodes";
import {
  BoxIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  GitCompareIcon,
  LayersIcon,
  SearchIcon,
  XIcon,
} from "lucide-react";
import { useMemo, useState } from "react";
import DynamicIcon from "../icons/DynamicIcon";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelectBlock: (type: string) => void;
  isMobile?: boolean;
}

type FilterCategory = "all" | "action" | "control_flow";

export default function FlowInsertBlockDialog({
  open,
  onOpenChange,
  onSelectBlock,
  isMobile,
}: Props) {
  const contextType = useFlowContext((c) => c.type);
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<FilterCategory>("all");
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const toggleCollapsed = (title: string) =>
    setCollapsed((c) => ({ ...c, [title]: !c[title] }));

  const categoriesToInclude: NodeCategory[] = useMemo(() => {
    if (category === "action") return ["action"];
    if (category === "control_flow") return ["control_flow"];
    return ["action", "control_flow"];
  }, [category]);

  const allSections = useMemo(() => {
    return categoriesToInclude.flatMap((cat) =>
      nodeCategories[cat].map((sec) => ({
        ...sec,
        category: cat,
        nodes: sec.nodeTypes
          .map((type) => {
            const values = getNodeValues(type);
            return { type, values };
          })
          .filter(
            ({ type, values }) =>
              !values.fixed &&
              !type.startsWith("entry_") &&
              !type.startsWith("option_") &&
              isNodeTypeAvailable(type, contextType)
          ),
      }))
    );
  }, [categoriesToInclude, contextType]);

  const filteredSections = useMemo(() => {
    const q = search.trim().toLowerCase();
    return allSections
      .map((sec) => ({
        ...sec,
        nodes: sec.nodes.filter(
          (n) =>
            !q ||
            n.values.defaultTitle.toLowerCase().includes(q) ||
            n.values.defaultDescription.toLowerCase().includes(q)
        ),
      }))
      .filter((sec) => sec.nodes.length > 0);
  }, [allSections, search]);

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      setSearch("");
    }
    onOpenChange(nextOpen);
  };

  const handleSelect = (type: string) => {
    onSelectBlock(type);
    handleOpenChange(false);
  };

  const headerControls = (
    <div className="flex flex-col gap-2">
      <div className="relative">
        <Input
          placeholder="Search ..."
          className="pl-10"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          autoFocus={!isMobile}
        />
        <SearchIcon className="absolute size-5 left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
      </div>
      <Tabs
        value={category}
        onValueChange={(val) => setCategory(val as FilterCategory)}
      >
        <TabsList className="grid w-full grid-cols-3 bg-muted/60 p-1 rounded-lg">
          <TabsTrigger value="all" className="gap-1.5 text-xs">
            <LayersIcon className="size-3.5" />
            <span>All</span>
          </TabsTrigger>
          <TabsTrigger value="action" className="gap-1.5 text-xs">
            <BoxIcon className="size-3.5" />
            <span>Action</span>
          </TabsTrigger>
          <TabsTrigger value="control_flow" className="gap-1.5 text-xs">
            <GitCompareIcon className="size-3.5" />
            <span>Control</span>
          </TabsTrigger>
        </TabsList>
      </Tabs>
    </div>
  );

  const blockList = (
    <div className="p-3 space-y-3">
      {filteredSections.length === 0 ? (
        <div className="text-center py-12 text-sm text-muted-foreground">
          No blocks found matching &quot;{search}&quot;.
        </div>
      ) : (
        filteredSections.map((sec, secIdx) => {
          const isCollapsed = !search.trim() && collapsed[sec.title];
          return (
            <div key={secIdx}>
              <button
                type="button"
                onClick={() => toggleCollapsed(sec.title)}
                className="w-full flex items-center gap-1.5 text-foreground font-medium mb-1.5 px-2 py-1 hover:text-foreground/80 rounded select-none text-left transition-colors"
              >
                {isCollapsed ? (
                  <ChevronRightIcon className="h-4 w-4 flex-none text-muted-foreground" />
                ) : (
                  <ChevronDownIcon className="h-4 w-4 flex-none text-muted-foreground" />
                )}
                <span className="text-sm font-medium">{sec.title}</span>
                <span className="ml-auto text-muted-foreground text-xs font-normal">
                  {sec.nodes.length}
                </span>
              </button>
              {!isCollapsed && (
                <div className="space-y-1">
                  {sec.nodes.map((node) => (
                    <BlockItem
                      key={node.type}
                      values={node.values}
                      onClick={() => handleSelect(node.type)}
                    />
                  ))}
                </div>
              )}
            </div>
          );
        })
      )}
    </div>
  );

  if (isMobile) {
    return (
      <Drawer open={open} onOpenChange={handleOpenChange}>
        <DrawerContent className="max-h-[85dvh] h-[85dvh] flex flex-col after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-background">
          <div className="flex items-center justify-between px-4 py-3 border-b border-border/60 flex-none">
            <DrawerTitle className="text-base font-semibold text-foreground">
              Insert Block
            </DrawerTitle>
            <DrawerDescription className="sr-only">
              Select a block to insert into this connection line.
            </DrawerDescription>
            <DrawerClose asChild>
              <Button
                variant="ghost"
                size="icon"
                className="size-8 rounded-full text-muted-foreground hover:text-foreground"
              >
                <XIcon className="size-4" />
                <span className="sr-only">Close</span>
              </Button>
            </DrawerClose>
          </div>

          <div className="px-4 py-2 border-b border-border/40 flex-none">
            {headerControls}
          </div>

          <ScrollArea className="flex-1 min-h-0">
            {blockList}
          </ScrollArea>
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-xl max-h-[85vh] h-[640px] flex flex-col p-0 gap-0 overflow-hidden select-none">
        <DialogHeader className="p-4 pb-3 border-b border-border/60 flex-none">
          <DialogTitle className="text-base font-semibold text-foreground">
            Insert Block
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Select a block to insert into this connection line.
          </DialogDescription>
          <div className="pt-2">{headerControls}</div>
        </DialogHeader>

        <ScrollArea className="flex-1 min-h-0">
          {blockList}
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

function BlockItem({
  values,
  onClick,
}: {
  values: NodeValues;
  onClick: () => void;
}) {
  return (
    <div
      onClick={onClick}
      className="p-2 hover:bg-muted rounded-md relative select-none cursor-pointer transition-colors"
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
        <div className="overflow-hidden min-w-0 flex-1">
          <div className="flex items-center gap-2 mb-1">
            <div className="font-medium text-foreground leading-5 truncate text-sm">
              {values.defaultTitle}
            </div>
          </div>
          <div className="text-xs text-muted-foreground line-clamp-2">
            {values.defaultDescription}
          </div>
        </div>
      </div>
    </div>
  );
}
