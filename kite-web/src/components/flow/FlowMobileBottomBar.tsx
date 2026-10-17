import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import {
  CheckIcon,
  ChevronUpIcon,
  ClipboardPasteIcon,
  CopyIcon,
  CopyPlusIcon,
  FocusIcon,
  ListTreeIcon,
  MessageSquareWarningIcon,
  MoveIcon,
  PencilIcon,
  PlusIcon,
  Redo2Icon,
  ScissorsIcon,
  Trash2Icon,
  Undo2Icon,
  UnlinkIcon,
  XIcon,
} from "lucide-react";

interface Props {
  topLevelCount: number;
  actionableCount: number;
  canDisconnect: boolean;
  hasClipboard: boolean;
  isSelectMode: boolean;
  isMoveMode: boolean;
  hidden?: boolean;
  onAddBlock: () => void;
  onEditNode: () => void;
  onDeselectNode: () => void;
  onStartMove: () => void;
  onStopMove: () => void;
  onCopy: () => void;
  onCut: () => void;
  onDuplicate: () => void;
  onDisconnect: () => void;
  onDelete: () => void;
  onPaste: () => void;
  onOpenLogs: () => void;
  onCenterFlow?: () => void;
  onUndo?: () => void;
  onRedo?: () => void;
  onFormat?: () => void;
  canUndo?: boolean;
  canRedo?: boolean;
  logCount?: number;
}

export default function FlowMobileBottomBar({
  topLevelCount,
  actionableCount,
  canDisconnect,
  hasClipboard,
  isSelectMode,
  isMoveMode,
  hidden,
  onAddBlock,
  onEditNode,
  onDeselectNode,
  onStartMove,
  onStopMove,
  onCopy,
  onCut,
  onDuplicate,
  onDisconnect,
  onDelete,
  onPaste,
  onOpenLogs,
  onCenterFlow,
  onUndo,
  onRedo,
  onFormat,
  canUndo = false,
  canRedo = false,
  logCount = 0,
}: Props) {
  const showChip = (isSelectMode || isMoveMode) && topLevelCount > 0;
  const isMulti = topLevelCount > 1;
  const actionSuffix = isMulti
    ? ` (${actionableCount > 0 ? actionableCount : topLevelCount})`
    : "";
  const totalSuffix = isMulti ? ` (${topLevelCount})` : "";
  const disableActionable = actionableCount === 0;

  return (
    <div
      className={cn(
        "fixed bottom-0 inset-x-0 z-30 md:hidden bg-background/95 backdrop-blur-md border-t border-border px-2.5 sm:px-3 py-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] flex items-center justify-between gap-1.5 shadow-lg select-none",
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
          <span className={showChip ? "hidden sm:inline" : ""}>Add Block</span>
        </Button>

        {showChip && isMoveMode && (
          <div className="flex items-center gap-1 bg-primary text-primary-foreground rounded-full pl-2.5 pr-1 h-9 shadow-sm min-w-0">
            <button
              type="button"
              className="flex items-center gap-1.5 text-xs font-medium truncate pr-1"
              onClick={onStopMove}
              title="Finish moving blocks"
            >
              <MoveIcon className="size-3.5 flex-none" />
              <span className="truncate">Moving ({topLevelCount})</span>
            </button>
            <button
              type="button"
              className="p-1 bg-primary-foreground/20 hover:bg-primary-foreground/30 rounded-full flex-none"
              onClick={onStopMove}
              title="Done moving"
            >
              <CheckIcon className="size-3.5" />
            </button>
          </div>
        )}

        {showChip && !isMoveMode && (
          <div className="flex items-center gap-0.5 bg-muted/80 rounded-full pl-2.5 pr-1 h-9 border border-primary/40 min-w-0">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className="flex items-center gap-1 text-xs font-medium text-foreground truncate py-1 outline-none"
                  title="Selected block options"
                >
                  <span className="truncate">Selected ({topLevelCount})</span>
                  <ChevronUpIcon className="size-3.5 text-muted-foreground flex-none" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                side="top"
                align="start"
                sideOffset={8}
                className="w-48 select-none"
              >
                {topLevelCount === 1 && (
                  <>
                    <DropdownMenuItem
                      onClick={onEditNode}
                      className="gap-2 cursor-pointer"
                    >
                      <PencilIcon className="size-4 text-muted-foreground" />
                      <span>Edit settings</span>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                  </>
                )}

                <DropdownMenuItem
                  onClick={onStartMove}
                  className="gap-2 cursor-pointer"
                >
                  <MoveIcon className="size-4 text-muted-foreground" />
                  <span>Move{totalSuffix}</span>
                </DropdownMenuItem>

                <DropdownMenuItem
                  disabled={disableActionable}
                  onClick={onCopy}
                  className="gap-2 cursor-pointer"
                >
                  <CopyIcon className="size-4 text-muted-foreground" />
                  <span>Copy{actionSuffix}</span>
                </DropdownMenuItem>

                <DropdownMenuItem
                  disabled={disableActionable}
                  onClick={onCut}
                  className="gap-2 cursor-pointer"
                >
                  <ScissorsIcon className="size-4 text-muted-foreground" />
                  <span>Cut{actionSuffix}</span>
                </DropdownMenuItem>

                <DropdownMenuItem
                  disabled={disableActionable}
                  onClick={onDuplicate}
                  className="gap-2 cursor-pointer"
                >
                  <CopyPlusIcon className="size-4 text-muted-foreground" />
                  <span>Duplicate{actionSuffix}</span>
                </DropdownMenuItem>

                <DropdownMenuItem
                  disabled={!canDisconnect}
                  onClick={onDisconnect}
                  className="gap-2 cursor-pointer"
                >
                  <UnlinkIcon className="size-4 text-muted-foreground" />
                  <span>Disconnect{totalSuffix}</span>
                </DropdownMenuItem>

                <DropdownMenuSeparator />

                <DropdownMenuItem
                  disabled={disableActionable}
                  onClick={onDelete}
                  className="gap-2 cursor-pointer text-destructive focus:text-destructive"
                >
                  <Trash2Icon className="size-4" />
                  <span>Delete{actionSuffix}</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            <button
              type="button"
              className="p-1 text-muted-foreground hover:text-foreground rounded-full flex-none"
              onClick={onDeselectNode}
              title="Clear selection"
            >
              <XIcon className="size-3.5" />
            </button>
          </div>
        )}
      </div>

      <div className="flex items-center gap-0 sm:gap-0.5 bg-muted/60 p-0.5 rounded-full border border-border/50 flex-none">
        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30 text-muted-foreground hover:text-destructive"
          onClick={onDelete}
          disabled={disableActionable}
          title="Delete selected"
        >
          <Trash2Icon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30"
          onClick={onPaste}
          disabled={!hasClipboard}
          title="Paste"
        >
          <ClipboardPasteIcon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30"
          onClick={onUndo}
          disabled={!canUndo}
          title="Undo"
        >
          <Undo2Icon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30"
          onClick={onRedo}
          disabled={!canRedo}
          title="Redo"
        >
          <Redo2Icon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full"
          onClick={onCenterFlow}
          title="Center flow"
        >
          <FocusIcon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full"
          onClick={onFormat}
          title="Auto arrange layout"
        >
          <ListTreeIcon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full relative"
          onClick={onOpenLogs}
          title="Logs"
        >
          <MessageSquareWarningIcon className="size-3.5 sm:size-4" />
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
