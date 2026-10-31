import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuShortcut,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";
import {
  BetweenHorizontalStartIcon,
  ClipboardPasteIcon,
  CopyIcon,
  CopyPlusIcon,
  Maximize2Icon,
  ListTreeIcon,
  Redo2Icon,
  ScissorsIcon,
  StickyNoteIcon,
  Trash2Icon,
  Undo2Icon,
} from "lucide-react";
import { ReactNode, useRef, useState } from "react";

export type FlowContextMenuTarget =
  | { type: "selection"; ids: string[]; canCopy: boolean; canDelete: boolean }
  | { type: "edge"; edgeId: string; canInsert: boolean }
  | { type: "pane"; position: { x: number; y: number } };

export type FlowContextMenuPoint = { x: number; y: number };

interface Props {
  // Called when the menu opens on the element at the point. Returns what the
  // menu is for, or null if it has no menu, which keeps it closed.
  onOpen: (
    el: Element,
    point: FlowContextMenuPoint
  ) => FlowContextMenuTarget | null;
  // The flow, which is the trigger of the menu.
  children: ReactNode;
  canUndo: boolean;
  canRedo: boolean;
  onCopy: (ids: string[]) => void;
  onCut: (ids: string[]) => void;
  onDuplicate: (ids: string[]) => void;
  onDelete: (ids: string[]) => void;
  onDeleteEdge: (edgeId: string) => void;
  onInsertIntoEdge: (edgeId: string) => void;
  onPaste: (position: { x: number; y: number }) => void;
  onUndo: () => void;
  onRedo: () => void;
  onFormat: () => void;
  onFitView: () => void;
  onAddNote: (position: { x: number; y: number }) => void;
}

export default function FlowContextMenu({
  onOpen,
  children,
  canUndo,
  canRedo,
  onCopy,
  onCut,
  onDuplicate,
  onDelete,
  onDeleteEdge,
  onInsertIntoEdge,
  onPaste,
  onUndo,
  onRedo,
  onFormat,
  onFitView,
  onAddNote,
}: Props) {
  const [open, setOpen] = useState(false);
  const [target, setTarget] = useState<FlowContextMenuTarget | null>(null);
  // Where the last right click or touch was, which the menu opens for.
  const pressed = useRef<{
    el: Element;
    point: FlowContextMenuPoint;
    touch: boolean;
  } | null>(null);
  const openedByTouch = useRef(false);

  const onOpenChange = (open: boolean) => {
    if (!open) return setOpen(false);
    const target =
      pressed.current && onOpen(pressed.current.el, pressed.current.point);
    // Also closes a menu that's open somewhere else.
    if (!target) return setOpen(false);
    setTarget(target);
    setOpen(true);
    openedByTouch.current = pressed.current!.touch;
  };

  return (
    // Not modal, so right clicking somewhere else in a dialog reopens it there.
    <ContextMenu modal={false} open={open} onOpenChange={onOpenChange}>
      <ContextMenuTrigger
        asChild
        // Touch screens open the menu with a long press, which iOS sends no
        // contextmenu event for.
        onPointerDownCapture={(e) => {
          pressed.current = {
            el: e.target as Element,
            point: { x: e.clientX, y: e.clientY },
            touch: e.pointerType !== "mouse",
          };
        }}
        // Chromium sends one for long presses too, as a pointer event.
        onContextMenuCapture={(e) => {
          const native = e.nativeEvent as Partial<PointerEvent>;
          pressed.current = {
            el: e.target as Element,
            point: { x: e.clientX, y: e.clientY },
            touch: !!native.pointerType && native.pointerType !== "mouse",
          };
        }}
        // Lifting the finger after a long press would click what it was on,
        // which moves the focus there and closes the menu. Captured, since
        // blocks stop touch events from bubbling up.
        onTouchEndCapture={(e) => {
          if (openedByTouch.current) e.preventDefault();
          openedByTouch.current = false;
        }}
        onTouchCancelCapture={() => (openedByTouch.current = false)}
      >
        {children}
      </ContextMenuTrigger>
      <ContextMenuContent
        className="w-52"
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        {target?.type === "selection" && (
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
            <ContextMenuSeparator />
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

        {target?.type === "edge" && (
          <>
            <Item
              icon={<BetweenHorizontalStartIcon />}
              disabled={!target.canInsert}
              onSelect={() => onInsertIntoEdge(target.edgeId)}
            >
              Insert block here
            </Item>
            <ContextMenuSeparator />
            <Item
              icon={<Trash2Icon />}
              onSelect={() => onDeleteEdge(target.edgeId)}
              destructive
            >
              Delete connection
            </Item>
          </>
        )}

        {target?.type === "pane" && (
          <>
            <Item
              icon={<ClipboardPasteIcon />}
              shortcut={shortcut("V")}
              onSelect={() => onPaste(target.position)}
            >
              Paste
            </Item>
            <ContextMenuSeparator />
            <Item
              icon={<StickyNoteIcon />}
              onSelect={() => onAddNote(target.position)}
            >
              Add note
            </Item>
            <ContextMenuSeparator />
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
            <ContextMenuSeparator />
            <Item icon={<ListTreeIcon />} onSelect={onFormat}>
              Auto arrange layout
            </Item>
            <Item icon={<Maximize2Icon />} onSelect={onFitView}>
              Fit to view
            </Item>
          </>
        )}
      </ContextMenuContent>
    </ContextMenu>
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
    <ContextMenuItem
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
      {shortcut && <ContextMenuShortcut>{shortcut}</ContextMenuShortcut>}
    </ContextMenuItem>
  );
}

function isMac() {
  return /Mac|iPhone|iPad/.test(navigator.platform);
}

function shortcut(key: string) {
  return isMac() ? `⌘${key}` : `Ctrl+${key}`;
}
