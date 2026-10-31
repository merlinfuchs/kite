import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
} from "@/components/ui/drawer";
import { XIcon } from "lucide-react";
import FlowParameterList from "./FlowParameterList";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export default function FlowParametersDrawer({ open, onOpenChange }: Props) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="max-h-[85dvh] h-[85dvh] flex flex-col after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-background">
        <div className="flex items-center justify-between px-4 py-3 border-b border-border/60 flex-none">
          <DrawerTitle className="text-base font-semibold text-foreground">
            Parameters
          </DrawerTitle>
          <DrawerDescription className="sr-only">
            Set the parameters of the flow.
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
          <FlowParameterList hideTitle />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
