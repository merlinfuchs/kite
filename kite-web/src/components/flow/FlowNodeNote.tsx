import { NodeProps } from "@/lib/flow/dataSchema";
import { FlowReadOnlyContext } from "@/lib/flow/context";
import {
  defaultNoteTitle,
  getNoteTitle,
  noteTitleMaxLength,
  getNoteColor,
  defaultNoteSize,
  minNoteSize,
  maxNoteSize,
  noteColors,
  NoteColor,
  noteMaxLength,
  parseNoteMarkdown,
} from "@/lib/flow/notes";
import { cn } from "@/lib/utils";
import {
  NodeResizer,
  NodeToolbar,
  Position,
  useReactFlow,
} from "@xyflow/react";
import {
  CheckIcon,
  PencilIcon,
  StickyNoteIcon,
  TrashIcon,
  TypeIcon,
} from "lucide-react";
import { renderMarkdown } from "@/lib/markdown/render";
import {
  KeyboardEvent,
  memo,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";

function NoteMarkdown({ text }: { text: string }) {
  const content = useMemo(
    () => renderMarkdown(parseNoteMarkdown(text)),
    [text]
  );
  return (
    <div className="space-y-1.5 break-words [&_.paragraph]:whitespace-pre-line [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:list-disc [&_ul]:pl-5 [&_h1]:text-lg [&_h1]:font-bold [&_h2]:text-base [&_h2]:font-bold [&_h3]:font-semibold [&_blockquote]:border-l-2 [&_blockquote]:border-foreground/30 [&_blockquote]:pl-2 [&_code]:rounded [&_code]:bg-background/60 [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs [&_pre]:whitespace-pre-wrap [&_pre]:rounded [&_pre]:bg-background/60 [&_pre]:p-2 [&_pre_code]:bg-transparent [&_pre_code]:p-0">
      {content}
    </div>
  );
}

export default memo(function FlowNodeNote({
  id,
  data,
  selected,
  width,
  height,
}: NodeProps) {
  const readOnly = useContext(FlowReadOnlyContext);
  const { updateNodeData, deleteElements } = useReactFlow();

  const text = (data.note_text as string | undefined) ?? "";
  const title = getNoteTitle(data);
  const [editingTitle, setEditingTitle] = useState(false);
  const [titleDraft, setTitleDraft] = useState(title);
  const titleInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editingTitle) titleInputRef.current?.select();
  }, [editingTitle]);

  function startEditingTitle() {
    if (readOnly) return;
    setTitleDraft(title);
    setEditingTitle(true);
  }

  function finishEditingTitle() {
    setEditingTitle(false);
    if (titleDraft.trim() !== title) {
      updateNodeData(id, { note_title: titleDraft.trim() });
    }
  }

  function onTitleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    // Typing in the title mustn't move or delete the note.
    e.stopPropagation();
    if (e.key === "Enter") {
      e.preventDefault();
      finishEditingTitle();
    } else if (e.key === "Escape") {
      e.preventDefault();
      setEditingTitle(false);
    }
  }

  const color = getNoteColor(data);

  // New notes start out being written.
  const [editing, setEditing] = useState(
    () => !readOnly && !text && !!selected
  );
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [overflowing, setOverflowing] = useState(false);

  // Long notes scroll instead of zooming the canvas, others let it zoom.
  useLayoutEffect(() => {
    const el = contentRef.current;
    if (el) setOverflowing(el.scrollHeight > el.clientHeight + 1);
  }, [text, editing, width, height]);

  useEffect(() => {
    if (!editing) return;
    const el = textareaRef.current;
    if (!el) return;
    el.focus();
    el.setSelectionRange(el.value.length, el.value.length);
  }, [editing]);

  // Deselecting the note, e.g. by clicking the canvas, finishes writing it.
  useEffect(() => {
    if (!selected) setEditing(false);
  }, [selected]);

  function onKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    // Typing in the note mustn't move or delete it.
    e.stopPropagation();
    if (e.key === "Escape" || (e.key === "Enter" && (e.ctrlKey || e.metaKey))) {
      e.preventDefault();
      setEditing(false);
    }
  }

  return (
    <>
      {!readOnly && (
        <NodeResizer
          isVisible={selected}
          minWidth={minNoteSize.width}
          minHeight={minNoteSize.height}
          maxWidth={maxNoteSize.width}
          maxHeight={maxNoteSize.height}
          color={color}
          handleClassName="!size-2.5 !rounded-sm !border-2 !bg-background"
        />
      )}

      {/* Only mounted when selected, as the toolbar follows the viewport. */}
      {!readOnly && selected && (
        <NodeToolbar
          isVisible
          position={Position.Top}
          offset={8}
          className="flex items-center gap-1 rounded-md border bg-background p-1 shadow-md"
          // Keeps the note being written focused while the toolbar is used.
          onMouseDown={(e) => e.preventDefault()}
        >
          {(Object.keys(noteColors) as NoteColor[]).map((key) => (
            <button
              key={key}
              type="button"
              title={key[0].toUpperCase() + key.slice(1)}
              className={cn(
                "size-5 rounded-full border-2 transition-transform hover:scale-110",
                noteColors[key] === color
                  ? "border-foreground"
                  : "border-transparent"
              )}
              style={{ backgroundColor: noteColors[key] }}
              onClick={() => updateNodeData(id, { note_color: key })}
            />
          ))}
          <div className="mx-1 h-5 w-px bg-border" />
          <button
            type="button"
            title="Rename note"
            className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
            onClick={startEditingTitle}
          >
            <TypeIcon className="size-4" />
          </button>
          <button
            type="button"
            title={editing ? "Done" : "Edit note"}
            className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
            onClick={() => setEditing((e) => !e)}
          >
            {editing ? (
              <CheckIcon className="size-4" />
            ) : (
              <PencilIcon className="size-4" />
            )}
          </button>
          <button
            type="button"
            title="Delete note"
            className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-red-500"
            onClick={() => deleteElements({ nodes: [{ id }] })}
          >
            <TrashIcon className="size-4" />
          </button>
        </NodeToolbar>
      )}

      <div
        className={cn(
          "flex h-full w-full flex-col overflow-hidden rounded-md border-2 bg-muted shadow-md",
          !readOnly && "cursor-grab"
        )}
        style={{
          width: width ?? defaultNoteSize.width,
          height: height ?? defaultNoteSize.height,
          borderColor: selected ? color : `${color}80`,
          backgroundImage: `linear-gradient(${color}24, ${color}24)`,
        }}
        onDoubleClick={() => !readOnly && setEditing(true)}
      >
        <div
          className="flex flex-none items-center gap-1.5 px-2.5 py-1 text-xs font-semibold"
          style={{ backgroundColor: `${color}40` }}
        >
          <StickyNoteIcon className="size-3.5 flex-none" style={{ color }} />
          {editingTitle ? (
            <input
              ref={titleInputRef}
              value={titleDraft}
              maxLength={noteTitleMaxLength}
              placeholder={defaultNoteTitle}
              className="nodrag nopan min-w-0 flex-auto rounded bg-background/80 px-1 outline-none ring-1"
              style={{ ["--tw-ring-color" as string]: color }}
              onChange={(e) => setTitleDraft(e.target.value)}
              onKeyDown={onTitleKeyDown}
              onBlur={finishEditingTitle}
              onContextMenu={(e) => e.stopPropagation()}
            />
          ) : (
            <div
              className="min-w-0 truncate"
              title={readOnly ? title : "Double-click to rename"}
              onDoubleClick={(e) => {
                // Renaming isn't writing the note.
                e.stopPropagation();
                startEditingTitle();
              }}
            >
              {title}
            </div>
          )}
        </div>

        {editing ? (
          <textarea
            ref={textareaRef}
            value={text}
            maxLength={noteMaxLength}
            placeholder="Write a note… Markdown like **bold**, lists and links works."
            className="nodrag nowheel nopan min-h-0 flex-auto resize-none bg-transparent px-2.5 py-2 text-sm text-foreground outline-none placeholder:text-muted-foreground"
            onChange={(e) => updateNodeData(id, { note_text: e.target.value })}
            onKeyDown={onKeyDown}
            onBlur={() => setEditing(false)}
            // The browser's menu, to paste or fix spelling, not the flow's.
            onContextMenu={(e) => e.stopPropagation()}
          />
        ) : (
          <div
            ref={contentRef}
            className={cn(
              "min-h-0 flex-auto overflow-y-auto px-2.5 py-2 text-sm text-foreground",
              overflowing && "nowheel"
            )}
          >
            {text ? (
              <NoteMarkdown text={text} />
            ) : (
              <div className="italic text-muted-foreground">
                {readOnly ? "Empty note" : "Double-click to write a note"}
              </div>
            )}
          </div>
        )}
      </div>
    </>
  );
});
