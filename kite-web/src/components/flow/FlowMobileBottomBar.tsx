import { Button } from "@/components/ui/button";
import { getNodeValues } from "@/lib/flow/nodes";
import { useReactFlow } from "@xyflow/react";
import { cn } from "@/lib/utils";
import {
  ListTreeIcon,
  Maximize2Icon,
  MessageSquareWarningIcon,
  PencilIcon,
  PlusIcon,
  Redo2Icon,
  Undo2Icon,
  XIcon,
} from "lucide-react";

interface Props {
  selectedNodeId: string | null;
  hidden?: boolean;
  onAddBlock: () => void;
  onEditNode: () => void;
  onDeselectNode: () => void;
  onOpenLogs: () => void;
  onFitView: () => void;
  onUndo?: () => void;
  onRedo?: () => void;
  onFormat?: () => void;
  canUndo?: boolean;
  canRedo?: boolean;
  logCount?: number;
}

export default function FlowMobileBottomBar({
  selectedNodeId,
  hidden,
  onAddBlock,
  onEditNode,
  onDeselectNode,
  onOpenLogs,
  onFitView,
  onUndo,
  onRedo,
  onFormat,
  canUndo = false,
  canRedo = false,
  logCount = 0,
}: Props) {
  const { getNode } = useReactFlow();
  const selectedNode = selectedNodeId ? getNode(selectedNodeId) : null;
  const nodeValues = selectedNode?.type
    ? getNodeValues(selectedNode.type)
    : null;
  const nodeTitle = nodeValues?.defaultTitle ?? "Block";

  return (
    <div
      className={cn(
        "fixed bottom-0 inset-x-0 z-30 md:hidden bg-background/95 backdrop-blur-md border-t border-border px-3 py-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] flex items-center justify-between gap-2 shadow-lg select-none",
        hidden && "hidden"
      )}
    >
      <div className="flex items-center gap-1.5 min-w-0">
        <Button
          size="sm"
          className="gap-1.5 h-9 rounded-full px-3 flex-none font-medium shadow-sm"
          onClick={onAddBlock}
        >
          <PlusIcon className="size-4" />
          <span className={selectedNodeId ? "hidden sm:inline" : ""}>
            Add Block
          </span>
        </Button>

        {selectedNodeId && (
          <div className="flex items-center gap-1 bg-muted/80 rounded-full pl-2.5 pr-1 h-9 border border-primary/40 min-w-0 max-w-[140px] sm:max-w-[190px]">
            <button
              type="button"
              className="flex items-center gap-1.5 text-xs font-medium text-foreground truncate"
              onClick={onEditNode}
              title={`Edit ${nodeTitle}`}
            >
              <PencilIcon className="size-3 text-primary flex-none" />
              <span className="truncate">{nodeTitle}</span>
            </button>
            <button
              type="button"
              className="p-1 text-muted-foreground hover:text-foreground rounded-full flex-none"
              onClick={onDeselectNode}
              title="Deselect block"
            >
              <XIcon className="size-3.5" />
            </button>
          </div>
        )}
      </div>

      <div className="flex items-center gap-0.5 bg-muted/60 p-0.5 rounded-full border border-border/50 flex-none">
        <Button
          variant="ghost"
          size="icon"
          className="size-8 rounded-full disabled:opacity-30"
          onClick={onUndo}
          disabled={!canUndo}
          title="Undo"
        >
          <Undo2Icon className="size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-8 rounded-full disabled:opacity-30"
          onClick={onRedo}
          disabled={!canRedo}
          title="Redo"
        >
          <Redo2Icon className="size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-8 rounded-full"
          onClick={onFitView}
          title="Fit to view"
        >
          <Maximize2Icon className="size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-8 rounded-full"
          onClick={onFormat}
          title="Auto arrange layout"
        >
          <ListTreeIcon className="size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-8 rounded-full relative"
          onClick={onOpenLogs}
          title="Logs"
        >
          <MessageSquareWarningIcon className="size-4" />
          {logCount > 0 && (
            <span className="absolute -top-1 -right-1 size-3.5 bg-primary text-primary-foreground text-[9px] font-bold rounded-full flex items-center justify-center">
              {logCount > 9 ? "9+" : logCount}
            </span>
          )}
        </Button>
      </div>
    </div>
  );
}
