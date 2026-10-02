import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
} from "@/components/ui/drawer";
import { LogEntry } from "@/lib/types/wire.gen";
import { XIcon } from "lucide-react";
import FlowLogList from "./FlowLogList";

interface Props {
  logs?: LogEntry[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export default function FlowLogsDrawer({ logs, open, onOpenChange }: Props) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="max-h-[85dvh] h-[85dvh] flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-border/60 flex-none">
          <DrawerTitle className="text-base font-semibold text-foreground">
            Logs
          </DrawerTitle>
          <DrawerDescription className="sr-only">
            View flow execution logs.
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
          <FlowLogList logs={logs} hideHeader />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
