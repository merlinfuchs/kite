import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
} from "@/components/ui/drawer";
import { XIcon } from "lucide-react";
import FlowBlockPicker from "./FlowBlockPicker";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export default function FlowAddBlockDrawer({ open, onOpenChange }: Props) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="max-h-[85dvh] h-[85dvh] flex flex-col after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-background">
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

        <FlowBlockPicker
          categories={["action", "control_flow", "option"]}
          onNodeSelect={() => onOpenChange(false)}
        />
      </DrawerContent>
    </Drawer>
  );
}
