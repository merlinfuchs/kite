import { describe, expect, it } from "vitest";
import { applyFlowEdits } from "./edits";
import { copyFlowNodes, pasteFlowNodes } from "./clipboard";
import { getFlowLayout } from "./layout";
import {
  createNoteNode,
  getNoteColor,
  getNoteTitle,
  separateLists,
  noteColors,
  parseNoteMarkdown,
} from "./notes";
import { serializeFlow } from "./serialize";
import { referencedParameters } from "./parameters";
import { testEdge, testNode } from "./testUtils";
import { validateFlow } from "./validate";

const entry = testNode("entry", "entry_command", {
  name: "test",
  description: "Test",
});
const log = testNode("log", "action_log", {
  log_level: "info",
  log_message: "hi",
});
const note = {
  ...testNode("note_abc", "note", {
    note_text: "Logs **hi**",
    note_color: "blue",
  }),
  width: 300,
  height: 120,
};
const edges = [testEdge("entry", "log")];

describe("notes", () => {
  it("aren't validated as blocks", () => {
    expect(validateFlow([entry, log, note], edges, "command")).toEqual(
      validateFlow([entry, log], edges, "command")
    );
  });

  it("are listed apart from the blocks for the AI", () => {
    const text = serializeFlow([entry, log, note], edges, "command");
    expect(text).not.toContain("- note_abc: note");
    expect(text).toContain('- note_abc "Note": "Logs **hi**"');
  });

  it("can't be edited by the AI", () => {
    const res = applyFlowEdits(
      { nodes: [entry, log, note], edges },
      [{ op: "connect", source: "log", target: "note_abc" }],
      "command"
    );
    expect(res.edges).toEqual(edges);
    expect(res.issues.map((i) => i.message).join()).toContain("is a note");
  });

  it("fall back to defaults for invalid settings", () => {
    expect(getNoteColor({ note_color: "nope" })).toBe(noteColors.yellow);
  });

  it("only move on auto arrange when blocks cover them", () => {
    const far = { ...note, id: "note_far", position: { x: 5000, y: 5000 } };
    const laid = getFlowLayout([entry, log, note, far], edges);
    const byId = new Map(laid.map((n) => [n.id, n]));

    expect(byId.get("note_abc")!.position.y).toBe(0);
    expect(byId.get("note_abc")!.position.x).toBeGreaterThan(
      byId.get("log")!.position.x
    );
    expect(byId.get("note_far")!.position).toEqual({ x: 5000, y: 5000 });
  });

  it("keep their size when copied", () => {
    const clipboard = copyFlowNodes([note], [], ["note_abc"])!;
    const [[pasted]] = pasteFlowNodes(clipboard, null);
    expect(pasted).toMatchObject({ type: "note", width: 300, height: 120 });
  });

  it("get an ID that says they're notes", () => {
    const n = createNoteNode({ x: 0, y: 0 });
    expect(n.id).toMatch(/^note_[a-z0-9]{3}$/);
    expect(n.type).toBe("note");
  });

  it("don't use the parameters they mention", () => {
    const mention = testNode("note_p", "note", {
      note_text: "Set {{params.log_channel}} first",
    });
    expect(referencedParameters({ nodes: [entry, mention] })).toEqual([]);
  });
});

describe("note markdown", () => {
  it("separates lists of different kinds", () => {
    expect(separateLists("- a\n- b\n  - c\n\n1. one\n2. two")).toBe(
      "- a\n- b\n  - c\n\n\n​\n\n1. one\n2. two"
    );
    expect(separateLists("1. one\n2. two")).toBe("1. one\n2. two");
    expect(separateLists("- a\ntext\n1. one")).toBe("- a\ntext\n1. one");
  });

  it("works like on Discord without blank lines", () => {
    const types = (text: string) => parseNoteMarkdown(text).map((n) => n.type);
    expect(types("Intro\n# Title\nmore")).toEqual([
      "paragraph",
      "heading",
      "paragraph",
    ]);
    expect(types("a\n> quote\nb")).toEqual([
      "paragraph",
      "blockQuote",
      "paragraph",
    ]);
    expect(types("- a\n- b\nnext")).toEqual(["list", "paragraph"]);
    expect(types("1. one\n- a")).toEqual(["list", "list"]);
  });

  it("keeps code blocks as they are", () => {
    const code = "foo\n- a\n1. b\n# x\ny";
    const [, block] = parseNoteMarkdown("Code:\n```\n" + code + "\n```\nafter");
    expect(block).toMatchObject({ type: "codeBlock", content: code });
  });

  it("gives notes a title", () => {
    expect(getNoteTitle({})).toBe("Note");
    expect(getNoteTitle({ note_title: "  Setup  " })).toBe("Setup");
  });
});
