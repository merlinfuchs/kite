import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  ClipboardPasteIcon,
  CopyIcon,
  CopyPlusIcon,
  Maximize2Icon,
  ListTreeIcon,
  Redo2Icon,
  ScissorsIcon,
  Trash2Icon,
  Undo2Icon,
} from "lucide-react";
import { ReactNode } from "react";
import { createPortal } from "react-dom";

export type FlowContextMenuTarget =
  | { type: "selection"; ids: string[]; canCopy: boolean; canDelete: boolean }
  | { type: "edge"; edgeId: string }
  | { type: "pane" };

export interface FlowContextMenuState {
  target: FlowContextMenuTarget;
  x: number;
  y: number;
}

interface Props {
  menu: FlowContextMenuState | null;
  onClose: () => void;
  canUndo: boolean;
  canRedo: boolean;
  onCopy: (ids: string[]) => void;
  onCut: (ids: string[]) => void;
  onDuplicate: (ids: string[]) => void;
  onDelete: (ids: string[]) => void;
  onDeleteEdge: (edgeId: string) => void;
  onPaste: (position: { x: number; y: number }) => void;
  onUndo: () => void;
  onRedo: () => void;
  onFormat: () => void;
  onFitView: () => void;
}

export default function FlowContextMenu({
  menu,
  onClose,
  canUndo,
  canRedo,
  onCopy,
  onCut,
  onDuplicate,
  onDelete,
  onDeleteEdge,
  onPaste,
  onUndo,
  onRedo,
  onFormat,
  onFitView,
}: Props) {
  if (!menu) return null;

  const { target, x, y } = menu;

  return (
    <DropdownMenu
      // Opening it somewhere else starts over instead of moving the open menu.
      key={`${x},${y}`}
      open
      onOpenChange={(open) => !open && onClose()}
      modal={false}
    >
      {/* In the body, so a transformed parent like a dialog doesn't move it */}
      {createPortal(
        <DropdownMenuTrigger asChild>
          <div className="fixed size-0" style={{ left: x, top: y }} />
        </DropdownMenuTrigger>,
        document.body
      )}
      <DropdownMenuContent
        align="start"
        className="w-52"
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        {target.type === "selection" && (
          <>
            <Item
              icon={<CopyIcon />}
              shortcut={shortcut("C")}
              disabled={!target.canCopy}
              onSelect={() => onCopy(target.ids)}
            >
              Copy
            </Item>
            <Item
              icon={<ScissorsIcon />}
              shortcut={shortcut("X")}
              disabled={!target.canCopy}
              onSelect={() => onCut(target.ids)}
            >
              Cut
            </Item>
            <Item
              icon={<CopyPlusIcon />}
              disabled={!target.canCopy}
              onSelect={() => onDuplicate(target.ids)}
            >
              Duplicate
            </Item>
            <DropdownMenuSeparator />
            <Item
              icon={<Trash2Icon />}
              shortcut={isMac() ? "⌫" : "Del"}
              disabled={!target.canDelete}
              onSelect={() => onDelete(target.ids)}
              destructive
            >
              Delete
            </Item>
          </>
        )}

        {target.type === "edge" && (
          <Item
            icon={<Trash2Icon />}
            onSelect={() => onDeleteEdge(target.edgeId)}
            destructive
          >
            Delete connection
          </Item>
        )}

        {target.type === "pane" && (
          <>
            <Item
              icon={<ClipboardPasteIcon />}
              shortcut={shortcut("V")}
              onSelect={() => onPaste({ x, y })}
            >
              Paste
            </Item>
            <DropdownMenuSeparator />
            <Item
              icon={<Undo2Icon />}
              shortcut={shortcut("Z")}
              disabled={!canUndo}
              onSelect={onUndo}
            >
              Undo
            </Item>
            <Item
              icon={<Redo2Icon />}
              shortcut={isMac() ? "⇧⌘Z" : "Ctrl+Y"}
              disabled={!canRedo}
              onSelect={onRedo}
            >
              Redo
            </Item>
            <DropdownMenuSeparator />
            <Item icon={<ListTreeIcon />} onSelect={onFormat}>
              Auto arrange layout
            </Item>
            <Item icon={<Maximize2Icon />} onSelect={onFitView}>
              Fit to view
            </Item>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function Item({
  icon,
  shortcut,
  disabled,
  destructive,
  onSelect,
  children,
}: {
  icon: ReactNode;
  shortcut?: string;
  disabled?: boolean;
  destructive?: boolean;
  onSelect: () => void;
  children: ReactNode;
}) {
  return (
    <DropdownMenuItem
      disabled={disabled}
      onSelect={onSelect}
      className={
        destructive
          ? "gap-2 text-destructive focus:text-destructive [&>svg]:size-4"
          : "gap-2 [&>svg]:size-4 [&>svg]:text-muted-foreground"
      }
    >
      {icon}
      {children}
      {shortcut && <DropdownMenuShortcut>{shortcut}</DropdownMenuShortcut>}
    </DropdownMenuItem>
  );
}

function isMac() {
  return /Mac|iPhone|iPad/.test(navigator.platform);
}

function shortcut(key: string) {
  return isMac() ? `⌘${key}` : `Ctrl+${key}`;
}
