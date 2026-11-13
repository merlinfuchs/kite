import { Edge, Node } from "@xyflow/react";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { FlowEditResult, layoutAddedNodes } from "./edits";
import { getFlowLayout, isFlowLaidOut } from "./layout";
import {
  createEdge,
  createNode,
  getNodeId,
  getNodeOutputs,
  getNodeValues,
  getOwnedChildTypes,
  isKnownNodeType,
  normalizeHandle,
} from "./nodes";
import { getNoteTitle, isNoteNode, noteMaxLength } from "./notes";
import { compact } from "./serialize";
import { FlowIssue, getConnectionIssue, validateFlow } from "./validate";

// The flow AI can write a flow as code instead of edits: one call per block,
// like action_response_create({ message_data: { content: "Hi" } }), where the
// nesting is the connections. A list of blocks runs one after the other, and
// a block's other outputs and branches are lists in its second argument:
//
//   entry_command({ $id: "entry", name: "ban", description: "Ban someone" }, [
//     option_command_argument({ name: "member", type: "user" }),
//     control_if({ condition_base_value: "{{...}}" }, {
//       yes: [action_member_ban({ ... })],
//       no: [action_response_create({ ... })],
//     }),
//   ])
//
// Existing blocks keep their ID as $id, so the whole flow is written again
// and matched to the current one. The code is read, not run.

// Writes the flow as code. Blocks the user selected are marked with a
// comment, and notes are listed as comments at the end.
export function printFlowCode(
  flowNodes: Node<NodeData>[],
  edges: Edge[],
  context: FlowContextType,
  selectedIds: string[] = []
) {
  const nodes = flowNodes.filter((n) => !isNoteNode(n));
  const notes = flowNodes.filter(isNoteNode);
  const byId = new Map(nodes.map((n) => [n.id, n]));
  const selected = new Set(selectedIds);
  const visited = new Set<string>();

  const outgoing = (id: string) => edges.filter((e) => e.source === id);
  const owned = (node: Node<NodeData>) => {
    const types = getOwnedChildTypes(node.type!);
    return outgoing(node.id)
      .map((e) => byId.get(e.target))
      .filter((n): n is Node<NodeData> => !!n && types.includes(n.type!));
  };

  const call = (node: Node<NodeData>, indent: string, outputs: string) => {
    visited.add(node.id);
    const data = compact(node.data) as Record<string, unknown> | undefined;
    const settings = printValue({ $id: node.id, ...data }, indent);
    const mark = selected.has(node.id) ? "/* selected */ " : "";
    return `${mark}${node.type}(${settings}${outputs ? `, ${outputs}` : ""})`;
  };

  // The blocks after one output, as a list that continues with the blocks
  // after the default output of its last block.
  const chain = (targets: string[], indent: string): string => {
    const inner = indent + "  ";
    const lines: string[] = [];
    let next = targets;
    while (next.length === 1) {
      const node = byId.get(next[0]);
      if (!node || visited.has(node.id)) {
        lines.push(`link(${JSON.stringify(next[0])})`);
        next = [];
        break;
      }
      const [text, rest] = block(node, inner);
      lines.push(text);
      next = rest;
    }
    // Several blocks after the same output run one after the other, each
    // with the blocks after it. Blocks only fork at the start of a list.
    if (next.length > 1) {
      return `[\n${next
        .map((t) => `${inner}${chain([t], inner)},\n`)
        .join("")}${indent}]`;
    }
    if (lines.length === 0) return "[]";
    return `[\n${lines.map((l) => `${inner}${l},\n`).join("")}${indent}]`;
  };

  // A block and the blocks after its default output, which continue the list
  // it's in.
  // The block that runs after each of the given outputs, once the blocks in
  // them are done, like after both outputs of an If, which is written after
  // the block in the list it's in.
  const joined = (sources: [string, string | null][]) => {
    if (sources.length < 2) return;
    const lasts = sources.map(([id, handle]) => {
      const targets = outgoing(id)
        .filter((e) => normalizeHandle(e.sourceHandle) === handle)
        .map((e) => e.target);
      return targets[targets.length - 1];
    });
    const after = lasts[0];
    if (!after || visited.has(after) || lasts.some((l) => l !== after)) return;
    const incoming = edges.filter((e) => e.target === after).length;
    return incoming === sources.length ? after : undefined;
  };

  const block = (node: Node<NodeData>, indent: string): [string, string[]] => {
    visited.add(node.id);
    const inner = indent + "  ";
    const types = getOwnedChildTypes(node.type!);
    if (types.length > 0) {
      const children = owned(node);
      const after = types.some((t) => t.includes("_item_"))
        ? joined(children.map((c) => [c.id, null]))
        : undefined;
      const targetsOf = (n: Node<NodeData>) =>
        defaultTargets(n).filter((t) => t !== after);
      const parts: string[] = [];
      const items = children.filter((c) => !getNodeValues(c.type!).fixed);
      if (items.length > 0 || types.some((t) => t.includes("_item_"))) {
        const branches = items.map(
          (item) =>
            `${inner}  ${call(
              item,
              inner + "  ",
              chain(targetsOf(item), inner + "  ")
            )},\n`
        );
        parts.push(`${inner}branches: [\n${branches.join("")}${inner}],\n`);
      }
      for (const child of children.filter(
        (c) => getNodeValues(c.type!).fixed
      )) {
        visited.add(child.id);
        const name = child.type!.split("_").pop();
        parts.push(`${inner}${name}: ${chain(targetsOf(child), inner)},\n`);
      }
      return [
        call(node, indent, `{\n${parts.join("")}${indent}}`),
        after ? [after] : [],
      ];
    }

    const handles = new Map<string, string[]>();
    for (const e of outgoing(node.id)) {
      const handle = normalizeHandle(e.sourceHandle) ?? "default";
      handles.set(handle, [...(handles.get(handle) ?? []), e.target]);
    }
    const outputs = getNodeOutputs(node).filter(
      (o) => !o.startsWith("component_")
    );
    const after = outputs.includes("default")
      ? undefined
      : joined(outputs.map((o) => [node.id, o]));
    if (after) {
      for (const [handle, targets] of handles) {
        handles.set(
          handle,
          targets.filter((t) => t !== after)
        );
      }
    }
    const defaults = after ? [after] : handles.get("default") ?? [];
    handles.delete("default");
    // Several blocks after the default output are listed in the block, as a
    // list of lists, rather than continuing the list it's in.
    const fork = defaults.length > 1;
    if (handles.size === 0) {
      return fork
        ? [call(node, indent, chain(defaults, indent)), []]
        : [call(node, indent, ""), defaults];
    }

    const parts = [...handles].map(
      ([handle, targets]) =>
        `${inner}${printKey(handle)}: ${chain(targets, inner)},\n`
    );
    if (fork) parts.unshift(`${inner}default: ${chain(defaults, inner)},\n`);
    return [
      call(node, indent, `{\n${parts.join("")}${indent}}`),
      fork ? [] : defaults,
    ];
  };

  const defaultTargets = (node: Node<NodeData>) =>
    outgoing(node.id)
      .filter((e) => !normalizeHandle(e.sourceHandle))
      .map((e) => e.target);

  const parts = [`// Flow type: ${context}`];
  const entry = nodes.find((n) => n.type?.startsWith("entry_"));
  if (entry) {
    visited.add(entry.id);
    const options = nodes.filter(
      (n) =>
        n.type?.startsWith("option_") &&
        edges.some((e) => e.source === n.id && e.target === entry.id)
    );
    const optionLines = options.map((o) => `  ${call(o, "  ", "")},\n`);
    const body = chain(defaultTargets(entry), "");
    const list =
      body === "[]"
        ? optionLines.length > 0
          ? `[\n${optionLines.join("")}]`
          : "[]"
        : `[\n${optionLines.join("")}${body.slice(2)}`;
    parts.push(`${call(entry, "", list)}`);
  }

  // Blocks that aren't connected to the entry, starting with the ones
  // nothing runs before.
  const unvisited = () => nodes.filter((n) => !visited.has(n.id));
  const isOwned = (n: Node<NodeData>) =>
    edges.some(
      (e) =>
        e.target === n.id &&
        getOwnedChildTypes(byId.get(e.source)?.type ?? "").includes(n.type!)
    );
  while (unvisited().length > 0) {
    const roots = unvisited().filter(
      (n) => !isOwned(n) && !edges.some((e) => e.target === n.id)
    );
    const root =
      roots[0] ?? unvisited().find((n) => !isOwned(n)) ?? unvisited()[0];
    parts.push(`// Not connected to the entry:\n${chain([root.id], "")}`);
  }

  for (const note of notes) {
    const text = String(note.data?.note_text ?? "")
      .trim()
      .slice(0, noteMaxLength);
    if (text) {
      parts.push(
        `// Note ${JSON.stringify(getNoteTitle(note.data))}: ${JSON.stringify(
          text
        )}`
      );
    }
  }
  return parts.join("\n\n");
}

const identifierRegex = /^[A-Za-z_$][A-Za-z0-9_$]*$/;

function printKey(key: string) {
  return identifierRegex.test(key) ? key : JSON.stringify(key);
}

// Writes settings on one line, unless they are long.
function printValue(value: unknown, indent: string): string {
  const flat = printFlat(value);
  if (flat.length + indent.length <= 100) return flat;
  const inner = indent + "  ";
  if (Array.isArray(value)) {
    return `[\n${value
      .map((v) => `${inner}${printValue(v, inner)},\n`)
      .join("")}${indent}]`;
  }
  if (value && typeof value === "object") {
    return `{\n${Object.entries(value)
      .map(([k, v]) => `${inner}${printKey(k)}: ${printValue(v, inner)},\n`)
      .join("")}${indent}}`;
  }
  return flat;
}

function printFlat(value: unknown): string {
  if (Array.isArray(value)) return `[${value.map(printFlat).join(", ")}]`;
  if (value && typeof value === "object") {
    const entries = Object.entries(value);
    if (entries.length === 0) return "{}";
    return `{ ${entries
      .map(([k, v]) => `${printKey(k)}: ${printFlat(v)}`)
      .join(", ")} }`;
  }
  return JSON.stringify(value ?? null);
}

// Parsing

type Value =
  | { kind: "literal"; value: string | number | boolean | null; line: number }
  | { kind: "array"; items: Value[]; line: number }
  | { kind: "object"; entries: [string, Value][]; line: number }
  | { kind: "call"; name: string; args: Value[]; line: number };

class CodeError extends Error {
  constructor(message: string, public line?: number) {
    super(line ? `Line ${line}: ${message}` : message);
  }
}

// Reads the code as a list of expressions, like JavaScript object literals
// and calls, with comments, trailing commas and any quotes.
export function parseFlowCode(code: string): Value[] {
  // The AI sometimes wraps the code in a Markdown code block.
  const src = code
    .trim()
    .replace(/^```[a-z]*\n/, "")
    .replace(/\n?```$/, "");
  let i = 0;
  let line = 1;

  const fail = (message: string): never => {
    throw new CodeError(message, line);
  };
  const skip = () => {
    while (i < src.length) {
      const c = src[i];
      if (c === "\n") {
        line++;
        i++;
      } else if (/\s/.test(c)) {
        i++;
      } else if (src.startsWith("//", i)) {
        while (i < src.length && src[i] !== "\n") i++;
      } else if (src.startsWith("/*", i)) {
        const end = src.indexOf("*/", i + 2);
        if (end < 0) fail("A comment isn't closed with */.");
        for (; i < end + 2; i++) if (src[i] === "\n") line++;
      } else {
        break;
      }
    }
  };
  const near = () => JSON.stringify(src.slice(i, i + 20));
  const expect = (c: string, what: string) => {
    skip();
    if (src[i] !== c) fail(`Expected ${what}, found ${near()}.`);
    i++;
  };

  const readString = (): string => {
    const quote = src[i++];
    let res = "";
    while (i < src.length && src[i] !== quote) {
      let c = src[i++];
      if (c === "\n") {
        if (quote !== "`") fail("A string isn't closed on its line.");
        line++;
      }
      if (c === "\\") {
        c = src[i++];
        if (c === "u") {
          const hex = src.startsWith("{", i)
            ? src.slice(i + 1, src.indexOf("}", i))
            : src.slice(i, i + 4);
          if (!/^[0-9a-fA-F]{1,6}$/.test(hex)) fail("Invalid \\u escape.");
          i += src[i] === "{" ? hex.length + 2 : 4;
          c = String.fromCodePoint(parseInt(hex, 16));
        } else if (c === "\n") {
          line++;
          c = "";
        } else {
          c =
            { n: "\n", t: "\t", r: "\r", b: "\b", f: "\f", v: "\v", 0: "\0" }[
              c
            ] ?? c;
        }
      }
      res += c;
    }
    if (i >= src.length) fail("A string isn't closed.");
    i++;
    return res;
  };

  // The closing brackets of the lists, objects and calls being read,
  // innermost last. The AI sometimes leaves out closing brackets, so one of
  // an outer one, or the end of the code, closes the inner ones too.
  const open: string[] = [];
  const closesOuter = () =>
    i >= src.length || open.slice(0, -1).includes(src[i]);
  const readItems = (close: string, readItem: () => void) => {
    open.push(close);
    for (;;) {
      skip();
      if (src[i] === close) {
        i++;
        break;
      }
      if (closesOuter()) break;
      readItem();
      skip();
      if (src[i] === ",") {
        i++;
      } else if (src[i] === close) {
        i++;
        break;
      } else if (!closesOuter()) {
        fail(`Expected , or ${close}, found ${near()}.`);
      }
    }
    open.pop();
  };

  const readValue = (): Value => {
    skip();
    const start = line;
    const c = src[i];
    if (c === undefined) fail("The code ends too early.");
    if (c === '"' || c === "'" || c === "`") {
      return { kind: "literal", value: readString(), line: start };
    }
    if (c === "[") {
      i++;
      const items: Value[] = [];
      readItems("]", () => items.push(readValue()));
      return { kind: "array", items, line: start };
    }
    if (c === "{") {
      i++;
      const entries: [string, Value][] = [];
      readItems("}", () => {
        let key: string;
        if (src[i] === '"' || src[i] === "'" || src[i] === "`") {
          key = readString();
        } else {
          const m = /^[A-Za-z0-9_$]+/.exec(src.slice(i));
          if (!m) fail(`Expected a key, found ${near()}.`);
          key = m![0];
          i += key.length;
        }
        expect(":", `: after the key '${key}'`);
        if (entries.some(([k]) => k === key)) {
          fail(`The key '${key}' is set twice.`);
        }
        entries.push([key, readValue()]);
      });
      return { kind: "object", entries, line: start };
    }
    const num = /^-?(\d+\.?\d*|\.\d+)([eE][-+]?\d+)?/.exec(src.slice(i));
    if (num) {
      i += num[0].length;
      return { kind: "literal", value: Number(num[0]), line: start };
    }
    const name = /^[A-Za-z_$][A-Za-z0-9_$]*/.exec(src.slice(i))?.[0];
    if (!name) fail(`Unexpected ${near()}.`);
    i += name!.length;
    const literals: Record<string, boolean | null> = {
      true: true,
      false: false,
      null: null,
      undefined: null,
    };
    if (Object.hasOwn(literals, name!)) {
      return { kind: "literal", value: literals[name!], line: start };
    }
    skip();
    if (src[i] !== "(") {
      fail(
        `'${name}' isn't a block call. Text needs quotes, and blocks are written like ${name}({ ... }).`
      );
    }
    i++;
    const args: Value[] = [];
    readItems(")", () => args.push(readValue()));
    return { kind: "call", name: name!, args, line: start };
  };

  const res: Value[] = [];
  for (;;) {
    skip();
    while (src[i] === ";") {
      i++;
      skip();
    }
    if (i >= src.length) break;
    res.push(readValue());
  }
  return res;
}

function toPlain(value: Value): unknown {
  switch (value.kind) {
    case "literal":
      return value.value;
    case "array":
      return value.items.map(toPlain);
    case "object":
      return Object.fromEntries(value.entries.map(([k, v]) => [k, toPlain(v)]));
    case "call":
      throw new CodeError(
        `Settings can't contain blocks like ${value.name}(). Put blocks in the list after the settings.`,
        value.line
      );
  }
}

// Applying

// A block of the code.
interface CodeBlock {
  type: string;
  data: NodeData;
  id?: string;
  line: number;
  // The blocks after each output, or for conditions and loops after each of
  // the blocks they own, by the last part of its type, like "else".
  outputs: Map<string, Target[]>;
  // The branches of a condition.
  items: CodeBlock[];
}

type Target = CodeBlock | { link: string; line: number };

// Applies the code to the flow, keeping the blocks whose $id it uses and
// removing the others, and validates the result like applyFlowEdits.
export function applyFlowCode(
  flow: { nodes: Node<NodeData>[]; edges: Edge[] },
  code: string,
  context: FlowContextType
): FlowEditResult {
  const issues: FlowIssue[] = [];
  const unchanged = () => ({
    nodes: flow.nodes,
    edges: flow.edges,
    refs: {},
    issues: [...issues, ...validateFlow(flow.nodes, flow.edges, context)],
  });
  const report = (err: unknown) => {
    if (!(err instanceof CodeError)) throw err;
    issues.push({ severity: "error", message: err.message });
  };

  let blocks: CodeBlock[];
  try {
    blocks = readBlocks(parseFlowCode(code), flow.nodes);
  } catch (err) {
    report(err);
    return unchanged();
  }

  const existing = new Map(
    flow.nodes.filter((n) => !isNoteNode(n)).map((n) => [n.id, n])
  );
  const nodes: Node<NodeData>[] = [];
  const nodeOf = new Map<CodeBlock, Node<NodeData>>();
  const added = new Set<string>();
  const taken = new Set(flow.nodes.map((n) => n.id));
  const used = new Set<string>();
  const labels = new Map<string, CodeBlock>();

  // Blocks keep their node if their $id is an existing block of the same
  // type, and new blocks get their $id as ID if it's free.
  for (const b of blocks) {
    try {
      // A block written twice, like in two branches, is copied.
      if (b.id !== undefined && labels.has(b.id)) b.id = undefined;
      if (b.id !== undefined) labels.set(b.id, b);
      const node = b.id !== undefined ? existing.get(b.id) : undefined;
      if (node) {
        if (node.type !== b.type) {
          throw new CodeError(
            `The block '${b.id}' is a ${node.type}, not a ${b.type}. Leave out $id to add a new block instead.`,
            b.line
          );
        }
        used.add(node.id);
        const before = compact(node.data) ?? {};
        nodeOf.set(
          b,
          deepEqual(before, b.data) ? node : { ...node, data: b.data }
        );
        continue;
      }
      let id = b.id && /^[A-Za-z0-9_]{1,40}$/.test(b.id) ? b.id : undefined;
      if (!id || taken.has(id)) {
        do id = getNodeId(b.type);
        while (taken.has(id));
      }
      taken.add(id);
      added.add(id);
      nodeOf.set(b, {
        id,
        type: b.type,
        position: { x: 0, y: 0 },
        data: b.data,
      });
    } catch (err) {
      report(err);
    }
  }
  if (issues.length > 0) return unchanged();
  nodes.push(...nodeOf.values());

  // The blocks conditions and loops own without a call in the code, like
  // their else branch, are kept or created.
  const edges: Edge[] = [];
  const fixedOf = new Map<CodeBlock, Map<string, Node<NodeData>>>();
  for (const b of blocks) {
    const types = getOwnedChildTypes(b.type);
    if (types.length === 0) continue;
    const owner = nodeOf.get(b)!;
    const fixed = new Map<string, Node<NodeData>>();
    const fixedTypes = types.filter((t) => getNodeValues(t).fixed);
    const current = flow.edges
      .filter((e) => e.source === owner.id)
      .map((e) => existing.get(e.target))
      .filter((n): n is Node<NodeData> => !!n);
    let created: Node<NodeData>[] | undefined;
    for (const t of fixedTypes) {
      let child = current.find((n) => n.type === t && !used.has(n.id));
      if (!child) {
        created ??= createNode(b.type, { x: 0, y: 0 })[0];
        child = { ...created.find((n) => n.type === t)! };
        while (taken.has(child.id)) child.id = getNodeId(t);
        taken.add(child.id);
        added.add(child.id);
      }
      used.add(child.id);
      nodes.push(child);
      fixed.set(t.split("_").pop()!, child);
    }
    fixedOf.set(b, fixed);
    for (const item of b.items) {
      edges.push(createEdge(owner, nodeOf.get(item)!));
    }
    for (const child of fixed.values()) edges.push(createEdge(owner, child));
  }

  const resolve = (t: Target) => {
    if ("link" in t) {
      const target = labels.get(t.link);
      if (!target) {
        throw new CodeError(
          `link('${t.link}') refers to no block of the code.`,
          t.line
        );
      }
      return nodeOf.get(target)!;
    }
    return nodeOf.get(t)!;
  };
  const connect = (source: Node<NodeData>, target: Target, handle: string) => {
    try {
      const node = resolve(target);
      const issue = getConnectionIssue(source, node, handle, nodes, edges);
      if (issue) throw new CodeError(issue, target.line);
      edges.push(createEdge(source, node, handle));
    } catch (err) {
      report(err);
    }
  };

  const entry = blocks.find((b) => b.type.startsWith("entry_"));
  for (const b of blocks) {
    const node = nodeOf.get(b)!;
    if (b.type.startsWith("option_") && entry) {
      edges.push(createEdge(node, nodeOf.get(entry)!));
    }
    const fixed = fixedOf.get(b);
    for (const [handle, targets] of b.outputs) {
      const source = fixed?.get(handle) ?? node;
      for (const t of targets) {
        connect(source, t, fixed?.has(handle) ? "default" : handle);
      }
    }
  }

  // The AI sometimes uses the temporary_name of a new block in result()
  // instead of its $id, which only works with its ID.
  const ids = new Set(nodes.map((n) => n.id));
  const names = new Map<string, string>();
  for (const b of blocks) {
    const name = b.data.temporary_name;
    const node = nodeOf.get(b)!;
    if (typeof name === "string" && added.has(node.id) && !ids.has(name)) {
      names.set(name, node.id);
    }
  }
  if (names.size > 0) {
    for (const [i, n] of nodes.entries()) {
      let changed = false;
      const data = mapStrings(n.data, (s) =>
        s.replace(resultRefRegex, (match, quote: string, name: string) => {
          const id = names.get(name);
          if (!id) return match;
          changed = true;
          return `result(${quote}${id}${quote})`;
        })
      ) as NodeData;
      if (changed) nodes[i] = { ...n, data };
    }
  }

  // Unchanged connections keep their edge.
  const key = (e: Edge) =>
    `${e.source}:${normalizeHandle(e.sourceHandle)}>${e.target}`;
  const previous = new Map(flow.edges.map((e) => [key(e), e]));
  const seen = new Set<string>();
  const finalEdges = edges
    .filter((e) => {
      const k = key(e);
      if (seen.has(k)) return false;
      seen.add(k);
      return true;
    })
    .map((e) => previous.get(key(e)) ?? e);

  // Notes stay as they are.
  const notes = flow.nodes.filter(isNoteNode);
  let finalNodes = [...nodes, ...notes];
  const noteIds = new Set(notes.map((n) => n.id));
  finalEdges.push(
    ...flow.edges.filter((e) => noteIds.has(e.source) || noteIds.has(e.target))
  );

  finalNodes = isFlowLaidOut(flow.nodes, flow.edges)
    ? getFlowLayout(finalNodes, finalEdges)
    : layoutAddedNodes(finalNodes, finalEdges, added);

  return {
    nodes: finalNodes,
    edges: finalEdges,
    refs: {},
    issues: [...issues, ...validateFlow(finalNodes, finalEdges, context)],
  };
}

// Turns the parsed code into blocks. The first expression is the entry
// block, others are lists of blocks that aren't connected to it.
function readBlocks(values: Value[], current: Node<NodeData>[]): CodeBlock[] {
  const blocks: CodeBlock[] = [];
  const options = new Set<CodeBlock>();
  const currentEntry = current.find((n) => n.type?.startsWith("entry_"));

  const readBlock = (value: Value, allowEntry = false): CodeBlock => {
    if (value.kind !== "call") {
      throw new CodeError(
        "Expected a block, like action_response_create({ ... }).",
        value.line
      );
    }
    if (value.name === "link") {
      throw new CodeError(
        "link() can only be the last block of a list.",
        value.line
      );
    }
    if (!isKnownNodeType(value.name)) {
      throw new CodeError(`Unknown block type '${value.name}'.`, value.line);
    }
    if (value.name.startsWith("entry_") && !allowEntry) {
      throw new CodeError(
        "A flow has one entry block, and it comes first. Blocks that run when a button or select menu of a message this flow sends is used go in the component_<id> output of the message block.",
        value.line
      );
    }
    let [settings, outputs]: (Value | undefined)[] = value.args;
    if (
      settings &&
      (settings.kind === "array" || settings.kind === "call") &&
      !outputs
    ) {
      outputs = settings;
      settings = undefined;
    }
    // Conditions are sometimes written with their branches as a list, and
    // the else branch as a third argument.
    const isCondition = getOwnedChildTypes(value.name).some(
      (t) => !getNodeValues(t).fixed
    );
    const extra = value.args[2];
    if (isCondition && outputs?.kind === "array" && value.args.length <= 3) {
      const entries: [string, Value][] = [["branches", outputs]];
      if (extra?.kind === "array") entries.push(["else", extra]);
      if (extra?.kind === "object") entries.push(...extra.entries);
      outputs = { kind: "object", entries, line: outputs.line };
    } else if (value.args.length > 2) {
      throw new CodeError(
        `${value.name}() takes its settings and the blocks after it, not ${value.args.length} arguments.`,
        value.line
      );
    }
    // Options in the list after the entry are connected to it, rather than
    // run after the block before them.
    if (allowEntry && outputs?.kind === "array") {
      outputs = {
        ...outputs,
        items: outputs.items.filter((v) => {
          if (v.kind !== "call" || !v.name.startsWith("option_")) return true;
          options.add(readBlock(v));
          return false;
        }),
      };
    }
    if (settings && settings.kind !== "object") {
      throw new CodeError(
        `The settings of ${value.name}() must be an object, like { name: "value" }.`,
        settings.line
      );
    }
    const data = (settings ? toPlain(settings) : {}) as Record<string, unknown>;
    const id = data.$id;
    delete data.$id;
    if (id !== undefined && typeof id !== "string") {
      throw new CodeError("$id must be a string.", value.line);
    }
    const block: CodeBlock = {
      type: value.name,
      data: withoutNulls(data) as NodeData,
      id,
      line: value.line,
      outputs: new Map(),
      items: [],
    };
    blocks.push(block);

    const types = getOwnedChildTypes(block.type);
    const itemType = types.find((t) => !getNodeValues(t).fixed);
    const fixedNames = types
      .filter((t) => getNodeValues(t).fixed)
      .map((t) => t.split("_").pop()!);
    const names =
      types.length > 0
        ? [...(itemType ? ["branches"] : []), ...fixedNames]
        : getOutputNames(block);

    if (!outputs) return block;
    if (outputs.kind === "array") {
      if (types.length > 0) {
        throw new CodeError(
          `${
            block.type
          }() has no outputs of its own. Put the blocks after it in ${names.join(
            ", "
          )}, like ${block.type}({ ... }, { ${names[0]}: [...] }).`,
          outputs.line
        );
      }
      // A list after a block like an If is what runs if it's true.
      const handle =
        !names.includes("default") && names.includes("yes") ? "yes" : "default";
      addOutput(block, handle, readList(outputs));
      return block;
    }
    if (outputs.kind !== "object") {
      throw new CodeError(
        `The second argument of ${block.type}() must be a list of the blocks after it, or an object with a list for each output.`,
        outputs.line
      );
    }
    for (const [name, list] of outputs.entries) {
      if (name === "branches" && itemType) {
        if (list.kind !== "array") {
          throw new CodeError("branches must be a list.", list.line);
        }
        for (const v of list.items) {
          // The else branch is sometimes written as the last branch.
          if (v.kind === "call" && v.name.endsWith("_item_else")) {
            const elseBlock = readBlock(v);
            blocks.splice(blocks.indexOf(elseBlock), 1);
            addOutput(block, "else", elseBlock.outputs.get("default") ?? []);
            continue;
          }
          const item = readBlock(v);
          if (item.type !== itemType) {
            throw new CodeError(
              `The branches of ${block.type}() are ${itemType}() blocks, not ${item.type}().`,
              v.line
            );
          }
          block.items.push(item);
        }
        continue;
      }
      if (!names.includes(name)) {
        throw new CodeError(
          `${block.type}() has no output '${name}', only ${names
            .map((n) => `'${n}'`)
            .join(", ")}.`,
          list.line
        );
      }
      addOutput(block, name, readList(list));
    }
    return block;
  };

  // A list of blocks that run one after the other, or a list of such lists
  // that each start after the same output.
  const readList = (value: Value): Target[] => {
    if (value.kind !== "array") {
      throw new CodeError(
        "Expected a list of blocks, like [action_response_create({ ... })].",
        value.line
      );
    }
    if (
      value.items.length > 0 &&
      value.items.every((v) => v.kind === "array")
    ) {
      return value.items.flatMap(readList);
    }
    let first: Target | undefined;
    let last: CodeBlock | undefined;
    value.items.forEach((v, i) => {
      let target: Target;
      if (v.kind === "call" && v.name === "link") {
        const [arg] = v.args;
        if (
          arg?.kind !== "literal" ||
          typeof arg.value !== "string" ||
          i !== value.items.length - 1
        ) {
          throw new CodeError(
            `link() takes the $id of a block, like link("abc"), and must be the last item of a list.`,
            v.line
          );
        }
        target = { link: arg.value, line: v.line };
      } else {
        target = readBlock(v);
      }
      if (!last) {
        first = target;
      } else {
        connectNext(last, target, v.line);
      }
      if (!("link" in target)) last = target;
    });
    return first ? [first] : [];
  };

  const [entryValue, ...rest] = values;
  if (!entryValue) throw new CodeError("The code is empty.");
  const entry = readBlock(entryValue, true);
  if (!entry.type.startsWith("entry_")) {
    throw new CodeError(
      `The code must start with the entry block${
        currentEntry ? `, ${currentEntry.type}()` : ""
      }.`,
      entryValue.line
    );
  }
  if (currentEntry && entry.type !== currentEntry.type) {
    throw new CodeError(
      `The entry block of this flow is ${currentEntry.type}() and can't be changed.`,
      entryValue.line
    );
  }
  for (const v of rest) {
    if (v.kind === "call" && v.name.startsWith("entry_")) {
      throw new CodeError(
        "A flow has one entry block. Other commands and event listeners are separate flows.",
        v.line
      );
    }
    if (v.kind === "array") readList(v);
    else readBlock(v);
  }
  for (const b of blocks) {
    if (b !== entry && b.type.startsWith("option_") && !options.has(b)) {
      throw new CodeError(
        `${b.type}() is an option of the entry block, so it goes in the list after the entry.`,
        b.line
      );
    }
  }
  return blocks;
}

// Connects the next block of a list after a block: after its default output,
// the end of a loop, or else after each of its outputs or branches, like
// both outputs of an If, so it runs after whichever one ran, once the blocks
// in it are done.
function connectNext(block: CodeBlock, target: Target, line: number) {
  const types = getOwnedChildTypes(block.type);
  if (types.includes("control_loop_end")) {
    addOutput(block, "end", [target]);
    return;
  }
  if (types.length > 0) {
    for (const item of block.items) addOutput(item, "default", [target]);
    addOutput(block, "else", [target]);
    return;
  }
  const names = getOutputNames(block).filter(
    (n) => !n.startsWith("component_")
  );
  if (names.length === 0) {
    throw new CodeError(
      `${block.type}() has no outputs, so nothing can come after it.`,
      line
    );
  }
  for (const name of names.includes("default") ? ["default"] : names) {
    addOutput(block, name, [target]);
  }
}

function getOutputNames(block: CodeBlock) {
  return getNodeOutputs({ type: block.type, data: block.data });
}

function addOutput(block: CodeBlock, handle: string, targets: Target[]) {
  block.outputs.set(handle, [...(block.outputs.get(handle) ?? []), ...targets]);
}

const resultRefRegex = /\bresult\(\s*(['"])([A-Za-z0-9_]+)\1\s*\)/g;

function mapStrings(value: unknown, fn: (s: string) => string): unknown {
  if (typeof value === "string") return fn(value);
  if (Array.isArray(value)) return value.map((v) => mapStrings(v, fn));
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value).map(([k, v]) => [k, mapStrings(v, fn)])
    );
  }
  return value;
}

// Leaves out null and empty settings, which the AI writes for settings it
// doesn't use, and the editor treats as unset.
function withoutNulls(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(withoutNulls);
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value)
        .filter(([, v]) => v !== null && v !== "")
        .map(([k, v]) => [k, withoutNulls(v)])
    );
  }
  return value;
}

function deepEqual(a: unknown, b: unknown): boolean {
  return JSON.stringify(sortKeys(a)) === JSON.stringify(sortKeys(b));
}

function sortKeys(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sortKeys);
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([k, v]) => [k, sortKeys(v)])
    );
  }
  return value;
}
