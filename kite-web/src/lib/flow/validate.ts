import { Edge, Node } from "@xyflow/react";
import { isNodeTypeAvailable } from "./categories";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import {
  canConnect,
  getNodeOutputs,
  getNodeTitle,
  getNodeValues,
  getOwnedChildTypes,
  isKnownNodeType,
  nodeTypes,
} from "./nodes";
import {
  getAvailablePlaceholders,
  getProvidedPlaceholders,
  walkDownstream,
  walkUpstream,
} from "./placeholders";

export interface FlowIssue {
  severity: "error" | "warning";
  message: string;
  nodeId?: string;
  edgeId?: string;
}

// Owned blocks, e.g. the items and else branch of a condition, are connected
// to their owner and to nothing else.
const ownedTypes = new Map(
  Object.keys(nodeTypes).map((type) => [type, getOwnedChildTypes(type)])
);
const ownerTypes = new Map<string, string[]>();
for (const [owner, owned] of ownedTypes) {
  for (const type of owned) {
    ownerTypes.set(type, [...(ownerTypes.get(type) ?? []), owner]);
  }
}

// Checks what the editor and the service expect of a flow, beyond what each
// block's settings schema covers: one entry, valid connections, complete
// conditions and loops, and placeholders that point at something that exists.
export function validateFlow(
  flowNodes: Node<NodeData>[],
  edges: Edge[],
  context: FlowContextType
): FlowIssue[] {
  // Generated flows can leave out the data of blocks without settings.
  const nodes = flowNodes.map((n) => (n.data ? n : { ...n, data: {} }));
  const issues: FlowIssue[] = [];
  const report = (
    severity: FlowIssue["severity"],
    message: string,
    ref: { nodeId?: string; edgeId?: string } = {}
  ) => issues.push({ severity, message, ...ref });

  const nodeIds = new Set(nodes.map((n) => n.id));
  const knownNodes = nodes.filter((node) => {
    if (isKnownNodeType(node.type!)) return true;
    report("error", `Unknown block type '${node.type}'.`, { nodeId: node.id });
    return false;
  });
  const knownById = new Map(knownNodes.map((n) => [n.id, n]));

  const entries = knownNodes.filter((n) => n.type!.startsWith("entry_"));
  if (entries.length !== 1) {
    report(
      "error",
      `The flow needs exactly one entry block, but has ${entries.length}.`
    );
  }

  for (const node of knownNodes) {
    if (!isNodeTypeAvailable(node.type!, context)) {
      report(
        "error",
        `'${getNodeTitle(node)}' can't be used in this kind of flow.`,
        { nodeId: node.id }
      );
    }

    const res = getNodeValues(node.type!).dataSchema?.safeParse(node.data);
    for (const issue of res?.error?.issues ?? []) {
      const path = issue.path.join(".");
      report(
        "error",
        path
          ? `'${getNodeTitle(node)}' setting '${path}': ${issue.message}`
          : `'${getNodeTitle(node)}': ${issue.message}`,
        { nodeId: node.id }
      );
    }
  }

  const connections = new Set<string>();
  for (const edge of edges) {
    const connection = `${edge.source}:${edge.sourceHandle || "default"}:${
      edge.target
    }`;
    if (connections.has(connection)) {
      report(
        "warning",
        "Two connections join the same blocks, so the second block runs twice.",
        { edgeId: edge.id }
      );
    }
    connections.add(connection);

    if (!nodeIds.has(edge.source) || !nodeIds.has(edge.target)) {
      report("error", "Connection to a block that doesn't exist.", {
        edgeId: edge.id,
      });
      continue;
    }

    // The service only reads connections into a block's single input.
    if (edge.targetHandle != null) {
      report(
        "error",
        `Connections can't target the input '${edge.targetHandle}'.`,
        { edgeId: edge.id }
      );
    }

    // Blocks of unknown types were reported above.
    const source = knownById.get(edge.source);
    const target = knownById.get(edge.target);
    if (!source || !target) continue;

    const handle = edge.sourceHandle || "default";

    if (!canConnect(source.type!, target.type!)) {
      report(
        "error",
        source.type!.startsWith("option_")
          ? `'${getNodeTitle(
              source
            )}' can only be connected to the entry block.`
          : target.type!.startsWith("option_")
          ? `Nothing can be connected into '${getNodeTitle(target)}'.`
          : `Only options can be connected into '${getNodeTitle(target)}'.`,
        { edgeId: edge.id }
      );
      continue;
    }
    if (source.type!.startsWith("option_")) continue;
    if (ownedTypes.get(source.type!)!.includes(target.type!)) {
      if (handle !== "default") {
        report(
          "error",
          `'${getNodeTitle(
            target
          )}' must be connected to the default output of '${getNodeTitle(
            source
          )}'.`,
          { edgeId: edge.id }
        );
      }
      continue;
    }

    if (ownerTypes.has(target.type!)) {
      report(
        "error",
        `'${getNodeTitle(
          target
        )}' can only be connected to the block it belongs to.`,
        { edgeId: edge.id }
      );
      continue;
    }

    const outputs = getNodeOutputs(source);
    if (outputs.length === 0) {
      report(
        "error",
        `'${getNodeTitle(
          source
        )}' has no outputs. Connect blocks to its branches instead.`,
        { edgeId: edge.id }
      );
    } else if (!outputs.includes(handle)) {
      report("error", `'${getNodeTitle(source)}' has no output '${handle}'.`, {
        edgeId: edge.id,
      });
    }
  }

  for (const node of knownNodes) {
    const children = edges
      .filter((e) => e.source === node.id)
      .map((e) => knownById.get(e.target));

    // Items can be added and removed, the else branch and the loop's
    // each and end blocks can't.
    for (const type of ownedTypes.get(node.type!)!) {
      if (!getNodeValues(type).fixed) continue;

      const count = children.filter((c) => c?.type === type).length;
      if (count !== 1) {
        report(
          // The service runs a condition without an else branch fine.
          type === "control_condition_item_else" && count === 0
            ? "warning"
            : "error",
          `'${getNodeTitle(node)}' needs exactly one '${
            getNodeValues(type).defaultTitle
          }' block, but has ${count}.`,
          { nodeId: node.id }
        );
      }
    }

    const owners = ownerTypes.get(node.type!);
    if (owners) {
      const count = edges.filter(
        (e) =>
          e.target === node.id &&
          owners.includes(knownById.get(e.source)?.type ?? "")
      ).length;
      if (count !== 1) {
        report(
          "error",
          `'${getNodeTitle(
            node
          )}' must belong to exactly one of these blocks: ${owners
            .map((t) => `'${getNodeValues(t).defaultTitle}'`)
            .join(", ")}.`,
          { nodeId: node.id }
        );
      }
    }

    // Checked on the blocks leading up to it, like the service does, as
    // blocks after the loop run later but aren't inside it.
    if (
      node.type === "control_loop_exit" &&
      !walkUpstream([node.id], edges).some(
        (id) => knownById.get(id)?.type === "control_loop_each"
      )
    ) {
      report("error", `'${getNodeTitle(node)}' only works inside a loop.`, {
        nodeId: node.id,
      });
    }
  }

  const entryIds = entries.map((n) => n.id);
  const reachable = new Set([...entryIds, ...walkDownstream(entryIds, edges)]);
  for (const node of knownNodes) {
    if (node.type!.startsWith("option_")) {
      if (!edges.some((e) => e.source === node.id)) {
        report(
          "warning",
          `'${getNodeTitle(
            node
          )}' isn't connected to the entry block, so it has no effect.`,
          { nodeId: node.id }
        );
      }
    } else if (!reachable.has(node.id)) {
      report(
        "warning",
        `'${getNodeTitle(
          node
        )}' never runs, as it isn't connected to the entry block.`,
        { nodeId: node.id }
      );
    }
  }

  const providedAnywhere = new Set(
    knownNodes.flatMap((n) => getProvidedPlaceholders(n).map((p) => p.value))
  );
  for (const node of knownNodes) {
    const refs = findReferences(node);
    if (refs.length === 0) continue;

    const available = new Set(
      getAvailablePlaceholders(node.id, nodes, edges, context).flatMap((g) =>
        g.placeholders.map((p) => p.value)
      )
    );
    for (const { fn, name } of refs) {
      const placeholder = `${fn}('${name}')`;
      if (available.has(placeholder)) continue;

      // Blocks with the same parent run one after another, in an order the
      // editor doesn't show, so a block beside this one may have set it.
      if (providedAnywhere.has(placeholder)) {
        report(
          "warning",
          `'${getNodeTitle(
            node
          )}' uses ${placeholder}, which is set by a block that doesn't always run before it, so it may be empty.`,
          { nodeId: node.id }
        );
      } else {
        report(
          "error",
          `'${getNodeTitle(node)}' uses ${placeholder}, but ${referenceHints[
            fn
          ](name)}`,
          { nodeId: node.id }
        );
      }
    }
  }

  return issues;
}

const referenceHints: Record<string, (name: string) => string> = {
  arg: (name) => `the command has no argument named '${name}'.`,
  result: (id) => `block '${id}' doesn't run before it or has no result.`,
  var: (name) =>
    `no block before it stores a temporary variable named '${name}'.`,
  input: (id) => `no modal before it has an input with the identifier '${id}'.`,
};

// Finds the placeholders that point at other parts of the flow, e.g.
// {{arg('user')}}. nodes.id and node('id') are older forms of result('id').
function findReferences(node: Node<NodeData>) {
  const expressions: string[] = [];
  const visit = (value: unknown, key?: string) => {
    if (typeof value === "string") {
      // The expression block takes an expression without the brackets.
      if (node.type === "action_expression_evaluate" && key === "expression") {
        expressions.push(value);
      }
      for (const match of value.matchAll(/\{\{([\s\S]*?)\}\}/g)) {
        expressions.push(match[1]);
      }
    } else if (Array.isArray(value)) {
      value.forEach((v) => visit(v));
    } else if (value && typeof value === "object") {
      Object.entries(value).forEach(([k, v]) => visit(v, k));
    }
  };
  visit(node.data);

  const res: { fn: string; name: string }[] = [];
  for (const expression of expressions) {
    for (const match of expression.matchAll(
      /\b(arg|input|result|var|node)\(\s*(['"])(.*?)\2\s*\)/g
    )) {
      res.push({
        fn: match[1] === "node" ? "result" : match[1],
        name: match[3],
      });
    }
    for (const match of expression.matchAll(/\bnodes\.([A-Za-z0-9_]+)/g)) {
      res.push({ fn: "result", name: match[1] });
    }
  }
  return res;
}
