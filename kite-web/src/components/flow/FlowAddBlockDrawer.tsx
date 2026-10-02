import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { NodeCategory } from "@/lib/flow/categories";
import {
  BoxIcon,
  GitCompareIcon,
  TextCursorInputIcon,
  XIcon,
} from "lucide-react";
import { useState } from "react";
import FlowNodeExplorer from "./FlowNodeExplorer";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export default function FlowAddBlockDrawer({ open, onOpenChange }: Props) {
  const [category, setCategory] = useState<NodeCategory>("action");

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="max-h-[85dvh] h-[85dvh] flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-border/60 flex-none">
          <DrawerTitle className="text-base font-semibold text-foreground">
            Add Block
          </DrawerTitle>
          <DrawerDescription className="sr-only">
            Choose a block to add to your flow.
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
          <Tabs
            value={category}
            onValueChange={(val) => setCategory(val as NodeCategory)}
          >
            <TabsList className="grid w-full grid-cols-3 bg-muted/60 p-1 rounded-lg">
              <TabsTrigger value="action" className="gap-1.5 text-xs">
                <BoxIcon className="size-3.5" />
                <span>Action</span>
              </TabsTrigger>
              <TabsTrigger value="control_flow" className="gap-1.5 text-xs">
                <GitCompareIcon className="size-3.5" />
                <span>Control</span>
              </TabsTrigger>
              <TabsTrigger value="option" className="gap-1.5 text-xs">
                <TextCursorInputIcon className="size-3.5" />
                <span>Option</span>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <div className="flex-1 min-h-0 overflow-hidden">
          <FlowNodeExplorer
            category={category}
            hideHeader
            onNodeSelect={() => onOpenChange(false)}
          />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
