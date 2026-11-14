import LZString from "lz-string";

// BotGhost has no export, so what's read here is the data its dashboard
// holds for a bot, copied out by the helper in exporter.ts. Its format isn't
// documented and can change, so everything is read loosely: fields that are
// missing or have another type are skipped instead of failing the import.

export type BotGhostAction = Record<string, unknown> & { type?: string };

export interface BotGhostOption {
  type?: string;
  name?: string;
  description?: string;
  required?: boolean;
  choices?: unknown[];
}

export interface BotGhostItem {
  kind: "command" | "event";
  name: string;
  description: string;
  // The event of a custom event, like guildMemberAdd.
  eventType: string;
  options: BotGhostOption[];
  actions: BotGhostAction[];
  raw: Record<string, unknown>;
}

export interface BotGhostExport {
  commands: BotGhostItem[];
  events: BotGhostItem[];
}

// A node or a connection of the tree BotGhost's builder shows.
type TreeElement = Record<string, unknown>;

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function str(value: unknown) {
  return typeof value === "string" ? value : "";
}

// parseBotGhostExport reads what the user pasted or uploaded. It accepts the
// file of the helper, and any other JSON that has BotGhost commands or
// custom events somewhere in it, like a single command.
export function parseBotGhostExport(text: string): BotGhostExport {
  let data: unknown;
  try {
    data = JSON.parse(text.trim());
  } catch {
    throw new Error(
      "This isn't the text the BotGhost helper copies. Copy it again and paste all of it."
    );
  }

  const found: BotGhostItem[] = [];
  const seen = new Set<string>();
  const add = (raw: Record<string, unknown>, kind: BotGhostItem["kind"]) => {
    const item = readItem(raw, kind);
    if (!item) return;
    const key = `${kind}:${item.name}:${JSON.stringify(item.actions)}`;
    if (seen.has(key)) return;
    seen.add(key);
    found.push(item);
  };

  const visit = (value: unknown, inEvents: boolean, depth: number) => {
    if (depth > 12) return;
    if (Array.isArray(value)) {
      for (const child of value) visit(child, inEvents, depth + 1);
      return;
    }
    if (!isRecord(value)) return;

    if (isItem(value)) {
      add(value, inEvents || isEvent(value) ? "event" : "command");
      return;
    }
    for (const [key, child] of Object.entries(value)) {
      visit(child, inEvents || /event/i.test(key), depth + 1);
    }
  };
  visit(data, false, 0);

  return {
    commands: found.filter((i) => i.kind === "command"),
    events: found.filter((i) => i.kind === "event"),
  };
}

function isItem(value: Record<string, unknown>) {
  return (
    typeof value.name === "string" &&
    (Array.isArray(value.actions) || typeof value.compressed_tree === "string")
  );
}

// Custom events have the event they run on as their type and no description
// of their own.
function isEvent(value: Record<string, unknown>) {
  return (
    typeof value.type === "string" &&
    /^[a-z]+[A-Z]/.test(value.type) &&
    typeof value.description !== "string"
  );
}

function readItem(
  raw: Record<string, unknown>,
  kind: BotGhostItem["kind"]
): BotGhostItem | null {
  const name = str(raw.name).trim();
  if (!name) return null;

  let actions = Array.isArray(raw.actions) ? raw.actions.filter(isRecord) : [];
  let options = Array.isArray(raw.options) ? raw.options.filter(isRecord) : [];

  // Commands only saved by the newer builder keep their blocks as its tree.
  if (actions.length === 0 && typeof raw.compressed_tree === "string") {
    const tree = readTree(raw.compressed_tree);
    if (tree) {
      actions = treeActions(tree);
      if (options.length === 0) options = treeOptions(tree);
    }
  }

  return {
    kind,
    name,
    description: str(raw.description).trim(),
    eventType: kind === "event" ? str(raw.type) : "",
    options,
    actions,
    raw,
  };
}

function readTree(compressed: string): TreeElement[] | null {
  try {
    const tree: unknown = JSON.parse(
      LZString.decompressFromBase64(compressed) || ""
    );
    return Array.isArray(tree) ? tree.filter(isRecord) : null;
  } catch {
    return null;
  }
}

const isEdge = (e: TreeElement) => "source" in e && "target" in e;

function nodeData(node: TreeElement): Record<string, unknown> {
  const data = isRecord(node.data) ? node.data.data : undefined;
  return isRecord(data) ? data : {};
}

function treeOptions(tree: TreeElement[]) {
  return tree
    .filter((e) => !isEdge(e) && e.type === "option")
    .map((e) => nodeData(e));
}

// treeActions turns the tree of the builder into the nested list of actions
// that older commands have, which is what gets converted.
export function treeActions(tree: TreeElement[]): BotGhostAction[] {
  const nodes = new Map(
    tree.filter((e) => !isEdge(e)).map((e) => [str(e.id), e])
  );
  const edges = tree.filter(isEdge);
  const visited = new Set<string>();

  const children = (id: string, types: string[]) =>
    edges
      .filter((e) => e.source === id && str(e.sourceHandle) !== "error_actions")
      .map((e) => nodes.get(str(e.target)))
      .filter((n): n is TreeElement => !!n && types.includes(str(n.type)));

  const chain = (id: string): BotGhostAction[] =>
    children(id, ["action", "condition"]).flatMap((node) => {
      const nodeId = str(node.id);
      if (visited.has(nodeId)) return [];
      visited.add(nodeId);

      if (node.type === "condition") {
        const branches = children(nodeId, ["conditionChild"]);
        const options = isRecord(node.data) ? node.data.node_options : null;
        const action: BotGhostAction = {
          ...nodeData(node),
          type: "condition",
          condition_type:
            (isRecord(options) && str(options.condition_type)) ||
            str(nodeData(node).condition_type),
          conditions: branches
            .filter((b) => nodeData(b).type !== "else")
            .map((b) => ({ ...nodeData(b), actions: chain(str(b.id)) })),
          else: {
            actions: branches
              .filter((b) => nodeData(b).type === "else")
              .flatMap((b) => chain(str(b.id))),
          },
        };
        return [action];
      }

      const action: BotGhostAction = { ...nodeData(node) };
      const buttons = children(nodeId, ["button", "advMessageButton"]);
      if (buttons.length > 0) {
        action.buttons = buttons.map((b) => ({
          ...nodeData(b),
          actions: chain(str(b.id)),
        }));
      }
      const options = children(nodeId, ["selectMenuOption"]);
      if (options.length > 0) {
        action.options = options.map((o) => ({
          ...nodeData(o),
          actions: chain(str(o.id)),
        }));
      }
      return [action, ...chain(nodeId)];
    });

  const root = nodes.has("root")
    ? "root"
    : str(tree.find((e) => e.type === "root")?.id);
  return chain(root);
}
