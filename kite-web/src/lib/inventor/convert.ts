import { Node } from "@xyflow/react";
import { NodeData } from "../flow/dataSchema";
import { applyFlowEdits, FlowEdit } from "../flow/edits";
import { newCommandFlow } from "../flow/newFlows";
import { getNodeId } from "../flow/nodes";
import { defaultNoteSize, noteMaxLength, noteNodeType } from "../flow/notes";
import { FlowData } from "../types/flow.gen";
import type { ConvertedItem } from "../botghost/convert";
import { InventorBlock, InventorItem } from "./parse";

// Only the blocks listed in `handled` below are converted. inventor.bot's
// format isn't documented, so a block that isn't known is left out and listed
// on the canvas instead of guessed at.

// The part of a message that holds its text.
const textPartType = "advmsgcmp_content_text";

// Names for the blocks that aren't imported, as a person would read them.
function blockName(type: string) {
  return (
    type.replace(/^advmsgcmp_/, "message ").replace(/_/g, " ") || "unknown"
  );
}

// fieldText is the text a block holds. inventor.bot keeps what's typed into a
// block in its fields, so any plain value in them counts.
function fieldText(block: InventorBlock) {
  return Object.values(block.fields)
    .map((f) => f.value)
    .filter((v) => v.trim() !== "")
    .join("\n");
}

function commandName(name: string) {
  return (
    name
      .toLowerCase()
      .replace(/^\//, "")
      .split(/\s+/)
      .map((word) => word.replace(/[^\p{L}\p{N}_-]/gu, "").slice(0, 32))
      .filter(Boolean)
      .slice(0, 3)
      .join(" ") || "imported"
  );
}

class Converter {
  edits: FlowEdit[] = [];
  notes: string[] = [];
  skipped = new Set<string>();
  private count = 0;

  note(text: string) {
    if (!this.notes.includes(text)) this.notes.push(text);
  }

  private add(type: string, data: NodeData, after: string) {
    const ref = `$b${++this.count}`;
    this.edits.push({ op: "add_node", ref, type, data, after });
    return ref;
  }

  blocks(blocks: InventorBlock[], start: string) {
    let after = start;
    for (const block of blocks) {
      if (block.deactivated) {
        this.note("Blocks that were turned off in inventor.bot were left out.");
        continue;
      }
      after = this.block(block, after);
    }
    return after;
  }

  private block(block: InventorBlock, after: string) {
    switch (block.type) {
      case "send_advanced_message":
        return this.message(block, after);
      default:
        this.skipped.add(blockName(block.type));
        return after;
    }
  }

  // An advanced message is made of parts, of which the text is the one that's
  // known so far.
  private message(block: InventorBlock, after: string) {
    const parts = block.fields.message_components?.blocks ?? [];
    const content = parts
      .filter((p) => p.type === textPartType && !p.deactivated)
      .map(fieldText)
      .filter(Boolean)
      .join("\n")
      .slice(0, 2000);
    for (const part of parts) {
      if (part.type !== textPartType) {
        this.skipped.add(`Message part: ${blockName(part.type)}`);
      }
    }
    if (!content) {
      this.note("A message without text wasn't imported.");
      return after;
    }
    return this.add(
      "action_response_create",
      { message_data: { content } },
      after
    );
  }
}

// The note on the canvas that lists what to check, right of the blocks.
function importNote(notes: string[], nodes: Node<NodeData>[]) {
  const right = Math.max(...nodes.map((n) => n.position.x));
  const top = Math.min(...nodes.map((n) => n.position.y));
  return {
    id: getNodeId(noteNodeType),
    type: noteNodeType,
    position: { x: right + 500, y: top },
    width: defaultNoteSize.width + 140,
    height: defaultNoteSize.height + 140,
    data: {
      note_title: "Imported from inventor.bot",
      note_text: notes
        .map((n) => `- ${n}`)
        .join("\n")
        .slice(0, noteMaxLength),
    },
  };
}

// convertInventorItem turns an inventor.bot component into a Kite command, as
// close to what it did as Kite's blocks allow.
export function convertInventorItem(item: InventorItem): ConvertedItem {
  const name = commandName(item.name);
  const base = {
    kind: "command" as const,
    sourceName: item.name,
    name,
    description: "Imported from inventor.bot",
    eventType: "",
  };

  if (item.trigger !== "slash_command") {
    return {
      ...base,
      status: "unsupported",
      notes: [
        item.trigger
          ? `Only slash commands can be imported so far, this one starts from "${item.trigger.replace(
              /_/g,
              " "
            )}".`
          : "It has no trigger.",
      ],
      flow: null,
    };
  }

  const converter = new Converter();
  converter.blocks(item.blocks, "$entry");

  if (converter.edits.length === 0) {
    return {
      ...base,
      status: "unsupported",
      notes: [
        item.blocks.length === 0
          ? "It has no blocks."
          : converter.skipped.size > 0
          ? `None of its blocks have a match in Kite yet: ${[
              ...converter.skipped,
            ].join(", ")}.`
          : converter.notes.join(" "),
      ],
      flow: null,
    };
  }

  if (converter.skipped.size > 0) {
    converter.note(
      `These inventor.bot blocks have no match in Kite yet and were left out: ${[
        ...converter.skipped,
      ].join(", ")}.`
    );
  }

  const start = newCommandFlow(name, base.description);
  const result = applyFlowEdits(
    { nodes: start.nodes as Node<NodeData>[], edges: [] },
    converter.edits,
    "command"
  );
  for (const issue of result.issues) {
    if (issue.severity === "error") {
      converter.note(
        `Kite found a problem to fix in the editor: ${issue.message}`
      );
    }
  }

  const nodes = result.nodes.map((n) => ({
    id: n.id,
    type: n.type,
    position: n.position,
    data: n.data,
  }));
  if (converter.notes.length > 0) {
    nodes.push(importNote(converter.notes, result.nodes));
  }

  return {
    ...base,
    status: converter.notes.length > 0 ? "review" : "ready",
    notes: converter.notes,
    flow: {
      nodes,
      edges: result.edges.map((e) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        ...(e.sourceHandle ? { sourceHandle: e.sourceHandle } : {}),
        ...(e.targetHandle ? { targetHandle: e.targetHandle } : {}),
        ...(e.type ? { type: e.type } : {}),
      })),
    } as FlowData,
  };
}

export function convertInventorItems(items: InventorItem[]) {
  return items.map(convertInventorItem);
}
