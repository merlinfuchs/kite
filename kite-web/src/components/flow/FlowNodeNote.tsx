import { NodeProps } from "@/lib/flow/dataSchema";
import { FlowReadOnlyContext } from "@/lib/flow/context";
import {
  defaultNoteTitle,
  getNoteTitle,
  noteTitleMaxLength,
  getNoteColor,
  listSeparator,
  separateLists,
  getNoteSize,
  minNoteSize,
  maxNoteSize,
  noteColors,
  NoteColor,
  noteMaxLength,
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
import SimpleMarkdown from "simple-markdown";
import { renderMarkdown } from "@/lib/markdown/render";
import {
  KeyboardEvent,
  memo,
  ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

// Notes support the markdown people know from Discord.
const {
  newline,
  paragraph,
  list,
  strong,
  em,
  u,
  del,
  inlineCode,
  codeBlock,
  fence,
  blockQuote,
  heading,
  link,
  autolink,
  url,
  br,
  escape,
  text,
} = SimpleMarkdown.defaultRules;
const parse = SimpleMarkdown.parserFor({
  newline,
  paragraph,
  list,
  heading,
  codeBlock,
  fence,
  blockQuote,
  strong,
  em,
  u,
  del,
  inlineCode,
  link,
  autolink,
  url,
  br,
  escape,
  text,
});

function NoteLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer nofollow"
      className="nodrag underline underline-offset-2"
      onClick={(e) => e.stopPropagation()}
    >
      {children}
    </a>
  );
}

function NoteMarkdown({ text }: { text: string }) {
  const content = useMemo(() => {
    // Like on Discord, headings and lists don't need a blank line around
    // them, which the parser expects.
    const blocks = separateLists(text.trim())
      .replace(/^(#{1,3} .+)\n(?!\n)/gm, "$1\n\n")
      .replace(
        /^(?!\s*(?:[-*]|\d+\.)\s)(.+)\n(?=\s*(?:[-*]|\d+\.)\s)/gm,
        "$1\n\n"
      );
    const ast = parse(blocks + "\n\n", { inline: false }).filter(
      (node) =>
        !(
          node.type === "paragraph" &&
          node.content.length === 1 &&
          node.content[0].content === listSeparator
        )
    );
    return renderMarkdown(ast, { link: NoteLink });
  }, [text]);
  return (
    <div className="space-y-1.5 break-words [&_.paragraph]:whitespace-pre-line [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:list-disc [&_ul]:pl-5 [&_h1]:text-lg [&_h1]:font-bold [&_h2]:text-base [&_h2]:font-bold [&_h3]:font-semibold [&_blockquote]:border-l-2 [&_blockquote]:border-foreground/30 [&_blockquote]:pl-2 [&_code]:rounded [&_code]:bg-background/60 [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs [&_pre]:whitespace-pre-wrap [&_pre]:rounded [&_pre]:bg-background/60 [&_pre]:p-2">
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

  const startEditingTitle = useCallback(() => {
    if (readOnly) return;
    setTitleDraft(title);
    setEditingTitle(true);
  }, [readOnly, title]);

  const finishEditingTitle = useCallback(() => {
    setEditingTitle(false);
    const value = titleDraft.trim().slice(0, noteTitleMaxLength);
    if (value !== title) {
      updateNodeData(id, { note_title: value || defaultNoteTitle });
    }
  }, [id, title, titleDraft, updateNodeData]);

  const onTitleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      // Typing in the title mustn't move or delete the note.
      e.stopPropagation();
      if (e.key === "Enter") {
        e.preventDefault();
        finishEditingTitle();
      } else if (e.key === "Escape") {
        e.preventDefault();
        setEditingTitle(false);
      }
    },
    [finishEditingTitle]
  );
  const color = getNoteColor(data);
  const size = getNoteSize(data);

  // New notes start out being written.
  const [editing, setEditing] = useState(
    () => !readOnly && !text && !!selected
  );
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [overflowing, setOverflowing] = useState(false);

  // Long notes scroll instead of zooming the canvas, others let it zoom.
  useEffect(() => {
    const el = contentRef.current;
    if (!el) return;
    const check = () => setOverflowing(el.scrollHeight > el.clientHeight + 1);
    check();
    const observer = new ResizeObserver(check);
    observer.observe(el);
    return () => observer.disconnect();
  }, [text, editing]);

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

  const startEditing = useCallback(() => {
    if (!readOnly) setEditing(true);
  }, [readOnly]);

  const onKeyDown = useCallback((e: KeyboardEvent<HTMLTextAreaElement>) => {
    // Typing in the note mustn't move or delete it.
    e.stopPropagation();
    if (e.key === "Escape" || (e.key === "Enter" && (e.ctrlKey || e.metaKey))) {
      e.preventDefault();
      setEditing(false);
    }
  }, []);

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

      {!readOnly && (
        <NodeToolbar
          isVisible={selected}
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
          width: width ?? size.width,
          height: height ?? size.height,
          borderColor: selected ? color : `${color}80`,
          backgroundImage: `linear-gradient(${color}24, ${color}24)`,
        }}
        onDoubleClick={startEditing}
      >
        <div
          className="flex flex-none items-center gap-1.5 px-2.5 py-1 text-xs font-semibold"
          style={{ backgroundColor: `${color}40`, color }}
        >
          <StickyNoteIcon className="size-3.5 flex-none" />
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
