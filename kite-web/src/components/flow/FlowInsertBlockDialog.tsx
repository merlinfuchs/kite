import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { createContext } from "react";
import FlowBlockPicker from "./FlowBlockPicker";

// Opens the dialog for inserting a block into a connection. Flow previews
// don't provide it, so their connections have no insert button.
export const InsertIntoEdgeContext = createContext<
  ((edgeId: string) => void) | null
>(null);

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onPick: (type: string) => void;
}

export default function FlowInsertBlockDialog({
  open,
  onOpenChange,
  onPick,
}: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="h-[80dvh] max-w-lg flex flex-col gap-0 p-0 overflow-hidden">
        <DialogHeader className="p-4 pb-3 border-b border-border/60">
          <DialogTitle>Insert block</DialogTitle>
          <DialogDescription>
            The block is added into the connection, between the two blocks.
          </DialogDescription>
        </DialogHeader>
        <FlowBlockPicker
          categories={["action", "control_flow"]}
          onPick={(type) => {
            onPick(type);
            onOpenChange(false);
          }}
        />
      </DialogContent>
    </Dialog>
  );
}
