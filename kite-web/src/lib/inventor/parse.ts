// inventor.bot has no export, so what's read here is what its dashboard sends
// for each component, copied out by the helper in exporter.ts. The format isn't
// documented and can change, so everything is read loosely: fields that are
// missing or have another type are skipped instead of failing the import.

export interface InventorBlock {
  // The kind of block, like send_advanced_message.
  type: string;
  // The label the user gave the block, if any.
  tag: string;
  deactivated: boolean;
  // The fields of the block by their internal ID. A field has a plain value
  // and/or blocks inside it, like the parts of a message.
  fields: Record<string, { value: string; blocks: InventorBlock[] }>;
}

export interface InventorItem {
  // The component's name, which is the command's name for slash commands.
  name: string;
  // The kind of trigger, like slash_command.
  trigger: string;
  triggerFields: Record<string, string>;
  blocks: InventorBlock[];
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function str(value: unknown) {
  return typeof value === "string" ? value : "";
}

function readBlocks(blockList: unknown, depth: number): InventorBlock[] {
  const blocks = isRecord(blockList) ? blockList.blocks : null;
  if (!Array.isArray(blocks) || depth > 12) return [];
  return blocks.filter(isRecord).map((b) => {
    const fields: InventorBlock["fields"] = {};
    for (const f of Array.isArray(b.fields) ? b.fields.filter(isRecord) : []) {
      const id = str(f.internal_id);
      if (!id) continue;
      fields[id] = {
        value: str(f.value),
        blocks: readBlocks(f.block_list, depth + 1),
      };
    }
    return {
      type: str(b.internal_id),
      tag: str(b.tag),
      deactivated: b.deactivated === true,
      fields,
    };
  });
}

// parseInventorExport reads what the user pasted or uploaded. It accepts the
// file of the helper, and a single component as inventor.bot has it.
export function parseInventorExport(text: string): InventorItem[] {
  let data: unknown;
  try {
    data = JSON.parse(text.trim());
  } catch {
    throw new Error(
      "This isn't the text the inventor.bot helper copies. Copy it again and paste all of it."
    );
  }

  const raw =
    isRecord(data) && Array.isArray(data.items)
      ? data.items
      : isRecord(data) && isRecord(data.blob)
      ? [data]
      : [];

  return raw.filter(isRecord).flatMap((entry) => {
    const blob = entry.blob;
    // Only flows have blocks, other kinds of components are skipped.
    if (!isRecord(blob) || !isRecord(blob.trigger)) return [];
    if (entry.type !== undefined && entry.type !== "flow") return [];
    const name = str(entry.name).trim();
    if (!name) return [];

    const triggerFields: Record<string, string> = {};
    for (const f of Array.isArray(blob.trigger.fields)
      ? blob.trigger.fields.filter(isRecord)
      : []) {
      triggerFields[str(f.internal_id)] = str(f.value);
    }
    return [
      {
        name,
        trigger: str(blob.trigger.internal_id),
        triggerFields,
        blocks: readBlocks(blob.block_list, 0),
      },
    ];
  });
}
