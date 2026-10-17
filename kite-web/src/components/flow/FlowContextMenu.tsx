import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  CheckSquareIcon,
  ClipboardPasteIcon,
  CopyIcon,
  CopyPlusIcon,
  FocusIcon,
  GitForkIcon,
  LayersIcon,
  ListTreeIcon,
  PencilIcon,
  PlusCircleIcon,
  Redo2Icon,
  ScanIcon,
  ScissorsIcon,
  SquareIcon,
  Trash2Icon,
  Undo2Icon,
  UnlinkIcon,
  ZoomInIcon,
} from "lucide-react";

export interface FlowContextMenuState {
  open: boolean;
  x: number;
  y: number;
  kind: "node" | "pane" | "edge";
  edgeId?: string;
  canInsertIntoEdge?: boolean;
  nodeId?: string;
}

interface Props {
  state: FlowContextMenuState;
  onClose: () => void;
  topLevelCount: number;
  actionableCount: number;
  canDisconnect: boolean;
  hasClipboard: boolean;
  canUndo: boolean;
  canRedo: boolean;
  onCopy: () => void;
  onCut: () => void;
  onDuplicate: () => void;
  onDuplicateWithConnections: () => void;
  onRename: () => void;
  onDisconnect: () => void;
  onDelete: () => void;
  onPaste: (screenPos: { x: number; y: number }) => void;
  onUndo: () => void;
  onRedo: () => void;
  onFormat: () => void;
  onCenterFlow: () => void;
  onResetZoom: () => void;
  onSelectAllInView: () => void;
  onSelectAllInFlow: () => void;
  onClearSelection: () => void;
  onDeleteEdge?: (edgeId: string) => void;
  onOpenInsertDialog?: (edgeId: string) => void;
}

export default function FlowContextMenu({
  state,
  onClose,
  topLevelCount,
  actionableCount,
  canDisconnect,
  hasClipboard,
  canUndo,
  canRedo,
  onCopy,
  onCut,
  onDuplicate,
  onDuplicateWithConnections,
  onRename,
  onDisconnect,
  onDelete,
  onPaste,
  onUndo,
  onRedo,
  onFormat,
  onCenterFlow,
  onResetZoom,
  onSelectAllInView,
  onSelectAllInFlow,
  onClearSelection,
  onDeleteEdge,
  onOpenInsertDialog,
}: Props) {
  const isMulti = topLevelCount > 1;
  const actionSuffix = isMulti
    ? ` (${actionableCount > 0 ? actionableCount : topLevelCount})`
    : "";
  const totalSuffix = isMulti ? ` (${topLevelCount})` : "";
  const disableActionable = actionableCount === 0;

  return (
    <DropdownMenu
      open={state.open}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      modal={false}
    >
      <DropdownMenuTrigger asChild>
        <div
          className="fixed w-0 h-0 pointer-events-none"
          style={{ left: state.x, top: state.y }}
          aria-hidden
        />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="start"
        side="bottom"
        sideOffset={2}
        className="w-56 select-none"
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        {state.kind === "node" ? (
          <>
            {actionableCount === 1 && (
              <DropdownMenuItem
                onClick={() => {
                  onRename();
                  onClose();
                }}
                className="gap-2 cursor-pointer"
              >
                <PencilIcon className="size-4 text-muted-foreground" />
                <span>Rename</span>
                <DropdownMenuShortcut>F2</DropdownMenuShortcut>
              </DropdownMenuItem>
            )}

            <DropdownMenuItem
              disabled={disableActionable}
              onClick={() => {
                onCopy();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <CopyIcon className="size-4 text-muted-foreground" />
              <span>Copy{actionSuffix}</span>
              <DropdownMenuShortcut>Ctrl+C</DropdownMenuShortcut>
            </DropdownMenuItem>

            <DropdownMenuItem
              disabled={disableActionable}
              onClick={() => {
                onCut();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <ScissorsIcon className="size-4 text-muted-foreground" />
              <span>Cut{actionSuffix}</span>
              <DropdownMenuShortcut>Ctrl+X</DropdownMenuShortcut>
            </DropdownMenuItem>

            <DropdownMenuItem
              disabled={disableActionable}
              onClick={() => {
                onDuplicate();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <CopyPlusIcon className="size-4 text-muted-foreground" />
              <span>Duplicate{actionSuffix}</span>
              <DropdownMenuShortcut>Ctrl+D</DropdownMenuShortcut>
            </DropdownMenuItem>

            <DropdownMenuItem
              disabled={disableActionable}
              onClick={() => {
                onDuplicateWithConnections();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <GitForkIcon className="size-4 text-muted-foreground" />
              <span>Duplicate with connections{actionSuffix}</span>
            </DropdownMenuItem>

            <DropdownMenuItem
              disabled={!canDisconnect}
              onClick={() => {
                onDisconnect();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <UnlinkIcon className="size-4 text-muted-foreground" />
              <span>Disconnect{totalSuffix}</span>
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem
              disabled={disableActionable}
              onClick={() => {
                onDelete();
                onClose();
              }}
              className="gap-2 cursor-pointer text-destructive focus:text-destructive"
            >
              <Trash2Icon className="size-4" />
              <span>Delete{actionSuffix}</span>
              <DropdownMenuShortcut>Del</DropdownMenuShortcut>
            </DropdownMenuItem>
          </>
        ) : state.kind === "edge" ? (
          <>
            <DropdownMenuItem
              disabled={state.canInsertIntoEdge === false}
              onClick={() => {
                if (state.edgeId) {
                  onOpenInsertDialog?.(state.edgeId);
                }
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <PlusCircleIcon className="size-4 text-muted-foreground" />
              <span>Insert block here...</span>
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem
              onClick={() => {
                if (state.edgeId) {
                  onDeleteEdge?.(state.edgeId);
                }
                onClose();
              }}
              className="gap-2 cursor-pointer text-destructive focus:text-destructive"
            >
              <Trash2Icon className="size-4" />
              <span>Delete connection</span>
            </DropdownMenuItem>
          </>
        ) : (
          <>
            <DropdownMenuItem
              disabled={!hasClipboard}
              onClick={() => {
                onPaste({ x: state.x, y: state.y });
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <ClipboardPasteIcon className="size-4 text-muted-foreground" />
              <span>Paste</span>
              <DropdownMenuShortcut>Ctrl+V</DropdownMenuShortcut>
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuSub>
              <DropdownMenuSubTrigger className="gap-2 cursor-pointer">
                <CheckSquareIcon className="size-4 text-muted-foreground" />
                <span>Select all</span>
              </DropdownMenuSubTrigger>
              <DropdownMenuSubContent className="w-48 select-none">
                <DropdownMenuItem
                  onClick={() => {
                    onSelectAllInView();
                    onClose();
                  }}
                  className="gap-2 cursor-pointer"
                >
                  <ScanIcon className="size-4 text-muted-foreground" />
                  <span>Select all in view</span>
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => {
                    onSelectAllInFlow();
                    onClose();
                  }}
                  className="gap-2 cursor-pointer"
                >
                  <LayersIcon className="size-4 text-muted-foreground" />
                  <span>Select all in flow</span>
                </DropdownMenuItem>
              </DropdownMenuSubContent>
            </DropdownMenuSub>

            <DropdownMenuItem
              disabled={topLevelCount === 0}
              onClick={() => {
                onClearSelection();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <SquareIcon className="size-4 text-muted-foreground" />
              <span>Clear selection</span>
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem
              disabled={!canUndo}
              onClick={() => {
                onUndo();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <Undo2Icon className="size-4 text-muted-foreground" />
              <span>Undo</span>
              <DropdownMenuShortcut>Ctrl+Z</DropdownMenuShortcut>
            </DropdownMenuItem>

            <DropdownMenuItem
              disabled={!canRedo}
              onClick={() => {
                onRedo();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <Redo2Icon className="size-4 text-muted-foreground" />
              <span>Redo</span>
              <DropdownMenuShortcut>Ctrl+Y</DropdownMenuShortcut>
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem
              onClick={() => {
                onFormat();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <ListTreeIcon className="size-4 text-muted-foreground" />
              <span>Auto arrange layout</span>
            </DropdownMenuItem>

            <DropdownMenuItem
              onClick={() => {
                onResetZoom();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <ZoomInIcon className="size-4 text-muted-foreground" />
              <span>Reset zoom (100%)</span>
            </DropdownMenuItem>

            <DropdownMenuItem
              onClick={() => {
                onCenterFlow();
                onClose();
              }}
              className="gap-2 cursor-pointer"
            >
              <FocusIcon className="size-4 text-muted-foreground" />
              <span>Center flow</span>
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
