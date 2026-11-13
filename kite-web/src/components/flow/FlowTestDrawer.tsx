import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
} from "@/components/ui/drawer";
import { FlowTestState, FlowTestTarget } from "@/lib/flow/flowTest";
import { XIcon } from "lucide-react";
import { Dispatch, SetStateAction } from "react";
import FlowTestPanel from "./FlowTestPanel";

interface Props {
  target: FlowTestTarget;
  state: FlowTestState;
  onStateChange: Dispatch<SetStateAction<FlowTestState>>;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onShowNode: (id: string) => void;
}

export default function FlowTestDrawer({
  target,
  state,
  onStateChange,
  open,
  onOpenChange,
  onShowNode,
}: Props) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="max-h-[85dvh] h-[85dvh] flex flex-col after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-background">
        <div className="flex items-center justify-between px-4 py-3 border-b border-border/60 flex-none">
          <DrawerTitle className="text-base font-semibold text-foreground">
            Test
          </DrawerTitle>
          <DrawerDescription className="sr-only">
            Run the flow as if someone used it, without sending or changing
            anything.
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
        <div className="flex-1 min-h-0 overflow-hidden">
          <FlowTestPanel
            target={target}
            state={state}
            onStateChange={onStateChange}
            hideHeader
            onShowNode={onShowNode}
          />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
