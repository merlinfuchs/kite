import { Node, XYPosition } from "@xyflow/react";
import SimpleMarkdown from "simple-markdown";
import { NodeData } from "./dataSchema";
import { getNodeId } from "./nodes";
import { addListBreaks } from "../markdown/render";
import { FlowNodeTypeNote } from "../types/flow.gen";

// Notes are text on the canvas that explains a flow to whoever opens it, like
// someone it was shared with. They're kept with the blocks so they're saved,
// exported and shared with the flow, but they aren't blocks: nothing connects
// to them and the service never reaches them when it runs the flow. Their
// size is the width and height of the node.
export const noteNodeType = FlowNodeTypeNote;

export const noteColors = {
  yellow: "#eab308",
  blue: "#3b82f6",
  green: "#22c55e",
  pink: "#ec4899",
  purple: "#8b5cf6",
  gray: "#71717a",
} as const;

export type NoteColor = keyof typeof noteColors;

export const defaultNoteColor: NoteColor = "yellow";

export const noteMaxLength = 2000;

export const defaultNoteTitle = "Note";
export const noteTitleMaxLength = 100;

export const defaultNoteSize = { width: 260, height: 160 };
export const minNoteSize = { width: 160, height: 80 };
export const maxNoteSize = { width: 1200, height: 1200 };

export function isNoteNode(node: { type?: string }) {
  return node.type === noteNodeType;
}

export function getNoteColor(data: NodeData): string {
  const color = data.note_color as string | undefined;
  return color && Object.hasOwn(noteColors, color)
    ? noteColors[color as NoteColor]
    : noteColors[defaultNoteColor];
}

export function getNoteTitle(data: NodeData): string {
  const title = data.note_title;
  return typeof title === "string" && title.trim()
    ? title.trim().slice(0, noteTitleMaxLength)
    : defaultNoteTitle;
}

export function createNoteNode(position: XYPosition): Node<NodeData> {
  return {
    id: getNodeId(noteNodeType),
    type: noteNodeType,
    position,
    ...defaultNoteSize,
    selected: true,
    data: {},
  };
}

// Notes support the markdown people know from Discord.
const {
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

export function parseNoteMarkdown(text: string) {
  return parse(addBlockBreaks(text.trim()) + "\n\n", {
    inline: false,
  }).filter(
    (node) =>
      !(
        node.type === "paragraph" &&
        node.content.length === 1 &&
        node.content[0].content === listSeparator
      )
  );
}

// Like on Discord, headings, quotes and lists don't need blank lines around
// them, which the parser expects. Code blocks are kept as they are.
function addBlockBreaks(text: string) {
  return text
    .split(/(```[\s\S]*?```)/)
    .map((part, i) =>
      i % 2
        ? `\n\n${part}\n\n`
        : addListBreaks(separateLists(part))
            .replace(/^(.+)\n(?=#{1,3} )/gm, "$1\n\n")
            .replace(/^(#{1,3} .+)\n(?!\n)/gm, "$1\n\n")
            .replace(/^(?!> )(.+)\n(?=> )/gm, "$1\n\n")
            .replace(/^(> .*)\n(?!> |\n)/gm, "$1\n\n")
            .replace(
              /^(\s*(?:[-*+]|\d+\.)\s.*)\n(?!\s|(?:[-*+]|\d+\.)\s)/gm,
              "$1\n\n"
            )
    )
    .join("");
}

// Separates lists of different kinds in markdown, see separateLists.
const listSeparator = "\u200b";

// The parser continues a list when an item of the other kind follows, so a
// numbered list after a bulleted one would be bulleted too. Lists of
// different kinds are separated by a paragraph that isn't rendered.
export function separateLists(text: string) {
  const out: string[] = [];
  let kind: "bullet" | "number" | null = null;
  for (const line of text.split("\n")) {
    const item = /^(\s*)(?:([-*+])|\d+\.)\s/.exec(line);
    if (item && item[1] === "") {
      const lineKind = item[2] ? "bullet" : "number";
      if (kind && kind !== lineKind) out.push("", listSeparator, "");
      kind = lineKind;
    } else if (line.trim() !== "" && !/^\s/.test(line)) {
      kind = null;
    }
    out.push(line);
  }
  return out.join("\n");
}
