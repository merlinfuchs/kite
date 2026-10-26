import { Node, XYPosition } from "@xyflow/react";
import { NodeData } from "./dataSchema";

// Notes are text on the canvas that explains a flow to whoever opens it, like
// someone it was shared with. They're kept with the blocks so they're saved,
// exported and shared with the flow, but they aren't blocks: nothing connects
// to them and the service never reaches them when it runs the flow.
export const noteNodeType = "note";

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

const noteIdChars = "abcdefghijklmnopqrstuvwxyz0123456789";

export function isNoteNode(node: { type?: string }) {
  return node.type === noteNodeType;
}

export function getNoteColor(data: NodeData): string {
  const color = data.note_color as string | undefined;
  return color && Object.hasOwn(noteColors, color)
    ? noteColors[color as NoteColor]
    : noteColors[defaultNoteColor];
}

export function getNoteTitle(data: NodeData = {}): string {
  const title = data.note_title;
  return typeof title === "string" && title.trim()
    ? title.trim().slice(0, noteTitleMaxLength)
    : defaultNoteTitle;
}

// The service only keeps the data and position of a block, so a note's size
// is stored in its data. In the editor react-flow keeps it as the width and
// height of the node, which resizing changes and undo restores.
export function getNoteSize(data: NodeData) {
  return {
    width: clampSize(data.note_width, defaultNoteSize.width, "width"),
    height: clampSize(data.note_height, defaultNoteSize.height, "height"),
  };
}

function clampSize(
  value: unknown,
  fallback: number,
  dimension: "width" | "height"
) {
  const n = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(n) || n <= 0) return fallback;
  return Math.round(
    Math.min(maxNoteSize[dimension], Math.max(minNoteSize[dimension], n))
  );
}

// withNoteSizes gives notes loaded from a saved flow, or pasted, their stored
// size. Notes that already have one in the editor keep it.
export function withNoteSizes<N extends Node<NodeData>>(nodes: N[]): N[] {
  return nodes.map((node) =>
    isNoteNode(node) && !(node.width && node.height)
      ? { ...node, ...getNoteSize(node.data ?? {}) }
      : node
  );
}

// storeNoteSizes writes the size notes have in the editor into their data, so
// it's saved and copied with them.
export function storeNoteSizes<N extends Node<NodeData>>(nodes: N[]): N[] {
  return nodes.map((node) => {
    if (!isNoteNode(node) || !node.width || !node.height) return node;
    const size = getNoteSize({
      ...node.data,
      note_width: node.width,
      note_height: node.height,
    });
    if (
      node.data.note_width === size.width &&
      node.data.note_height === size.height
    ) {
      return node;
    }
    return {
      ...node,
      data: { ...node.data, note_width: size.width, note_height: size.height },
    };
  });
}

export function createNoteNode(
  position: XYPosition,
  existingIds: Iterable<string> = []
): Node<NodeData> {
  const taken = new Set(existingIds);
  let id: string;
  do {
    let suffix = "";
    for (let i = 0; i < 3; i++) {
      suffix += noteIdChars[Math.floor(Math.random() * noteIdChars.length)];
    }
    id = `note_${suffix}`;
  } while (taken.has(id));

  return {
    id,
    type: noteNodeType,
    position,
    ...defaultNoteSize,
    selected: true,
    data: {
      note_text: "",
      note_color: defaultNoteColor,
      note_width: defaultNoteSize.width,
      note_height: defaultNoteSize.height,
    },
  };
}

// Separates lists of different kinds in markdown, see separateLists.
export const listSeparator = "\u200b";

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
