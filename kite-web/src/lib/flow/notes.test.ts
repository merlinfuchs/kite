import { describe, expect, it } from "vitest";
import { applyFlowEdits } from "./edits";
import {
  createNoteNode,
  getNoteColor,
  getNoteSize,
  getNoteTitle,
  separateLists,
  maxNoteSize,
  noteColors,
  storeNoteSizes,
  withNoteSizes,
} from "./notes";
import { serializeFlow } from "./serialize";
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
const note = testNode("note_abc", "note", {
  note_text: "Logs **hi**",
  note_color: "blue",
  note_width: 300,
  note_height: 120,
});
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

  it("keep their size in their data", () => {
    const [loaded] = withNoteSizes([note]);
    expect(loaded).toMatchObject({ width: 300, height: 120 });

    const [saved] = storeNoteSizes([{ ...loaded, width: 410.6, height: 99 }]);
    expect(saved.data).toMatchObject({ note_width: 411, note_height: 99 });
    expect(storeNoteSizes([entry])[0]).toBe(entry);
  });

  it("fall back to defaults for invalid settings", () => {
    expect(getNoteColor({ note_color: "nope" })).toBe(noteColors.yellow);
    expect(getNoteSize({ note_width: "x", note_height: 99999 })).toEqual({
      width: 260,
      height: maxNoteSize.height,
    });
  });

  it("get an ID that isn't taken", () => {
    const n = createNoteNode({ x: 0, y: 0 }, ["note_abc"]);
    expect(n.id).toMatch(/^note_[a-z0-9]{3}$/);
    expect(n.id).not.toBe("note_abc");
    expect(n.type).toBe("note");
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

  it("gives notes a title", () => {
    expect(getNoteTitle({})).toBe("Note");
    expect(getNoteTitle({ note_title: "  Setup  " })).toBe("Setup");
  });
});
