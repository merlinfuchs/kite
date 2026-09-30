import { Edge, Node, XYPosition } from "@xyflow/react";
import { humanId } from "human-id";
import { useMemo } from "react";
import { ZodSchema } from "zod";
import { Features } from "../types/wire.gen";
import { getUniqueId } from "../utils";
import {
  blockDataFields,
  blockDataSchema,
  blockDefinitions,
  getBlockDefinition,
  hasFieldSettings,
} from "../blocks";
import {
  actionColor,
  controlColor,
  entryColor,
  optionColor,
  suspendColor,
} from "../blocks/colors";
import { BlockDefinition } from "../blocks/types";
import { FlowContextType } from "./context";
import { getComponentHandleIds } from "./resume";
import { NodeData } from "./dataSchema";

export {
  actionColor,
  controlColor,
  entryColor,
  errorColor,
  optionColor,
  primaryColor,
  suspendColor,
} from "../blocks/colors";

export interface NodeValues {
  color: string;
  icon: string;
  defaultTitle: string;
  defaultDescription: string;
  dataSchema?: ZodSchema;
  dataFields: string[];
  resultSchema?: ZodSchema;
  // IDs of the source handles edges can be drawn from. Defaults to
  // ["default"]. Owned children are connected with fixed edges instead.
  // Message blocks also get one per button or select menu in their message.
  outputs?: string[];
  // Flow types the block can be used in. Blocks listed in the block explorer
  // otherwise take them from their section.
  contexts?: FlowContextType[];
  fixed?: boolean;
  creditsCost?: number | ((data: NodeData) => number);
  // The block fails when the app doesn't have this feature
  premiumFeature?: keyof Features;
}

const kindColors: Record<string, string> = {
  entry: entryColor,
  action: actionColor,
  control: controlColor,
  option: optionColor,
  suspend: suspendColor,
};

export const nodeTypes: Record<string, NodeValues> = Object.fromEntries(
  blockDefinitions.map((block) => [block.type, toNodeValues(block)])
);

function toNodeValues(block: BlockDefinition): NodeValues {
  // Blocks without a schema of their own get one from their fields.
  const withFields = hasFieldSettings(block) ? block : undefined;
  const schema =
    typeof block.schema === "function" ? block.schema() : block.schema;

  return {
    color: block.color ?? kindColors[block.type.split("_")[0]],
    icon: block.icon,
    defaultTitle: block.title,
    defaultDescription: block.description,
    dataSchema: withFields ? blockDataSchema(withFields) : schema,
    dataFields: withFields ? blockDataFields(withFields) : block.inputs ?? [],
    resultSchema: block.result?.schema,
    outputs: block.outputs,
    contexts: block.contexts,
    fixed: block.fixed,
    creditsCost: block.credits,
    premiumFeature: block.premium_feature,
  };
}

const unknownNodeType: NodeValues = {
  color: "#ff0000",
  icon: "circle-help",
  defaultTitle: "Unknown",
  defaultDescription: "Unknown node type.",
  dataFields: [],
};

export function isKnownNodeType(nodeType: string) {
  return Object.hasOwn(nodeTypes, nodeType);
}

export function getNodeValues(nodeType: string): NodeValues {
  return isKnownNodeType(nodeType) ? nodeTypes[nodeType] : unknownNodeType;
}

export function getNodeCreditsCost(
  values: NodeValues,
  data: NodeData
): number | undefined {
  return typeof values.creditsCost === "function"
    ? values.creditsCost(data)
    : values.creditsCost;
}

// The IDs of the outputs edges can start from. Message blocks also get one per
// button or select menu in their message.
export function getNodeOutputs(node: { type?: string; data: NodeData }) {
  return [
    ...(getNodeValues(node.type!).outputs ?? ["default"]),
    ...getComponentHandleIds(node.data.message_data?.components ?? []),
  ];
}

export function getNodeTitle(node: { type?: string; data: NodeData }) {
  return node.data.custom_label || getNodeValues(node.type!).defaultTitle;
}

// The blocks an owner is created with and connected to, e.g. the items and
// else branch of a condition.
export function getOwnedChildTypes(type: string) {
  return getBlockDefinition(type)?.owns ?? [];
}

// Returns the IDs of the given blocks plus the blocks they own, e.g. the
// branches of a condition, which are deleted, copied and removed with them.
export function withOwnedNodes(
  ids: string[],
  nodes: Node<NodeData>[],
  edges: Edge[]
): Set<string> {
  const types = new Map(nodes.map((n) => [n.id, n.type!]));
  const targets = new Map<string, string[]>();
  for (const edge of edges) {
    if (!targets.has(edge.source)) targets.set(edge.source, []);
    targets.get(edge.source)!.push(edge.target);
  }
  const res = new Set<string>();

  const add = (id: string) => {
    if (!types.has(id) || res.has(id)) return;
    res.add(id);

    const owned = getOwnedChildTypes(types.get(id)!);
    (targets.get(id) ?? [])
      .filter((target) => owned.includes(types.get(target)!))
      .forEach(add);
  };
  ids.forEach(add);

  return res;
}

// Returns the blocks the editor deletes when the given ones are deleted: the
// blocks they own go with them, while fixed blocks, e.g. the else branch of a
// condition, are only deleted together with the block they belong to.
export function getDeletedNodeIds(
  ids: string[],
  nodes: Node<NodeData>[],
  edges: Edge[]
) {
  const types = new Map(nodes.map((n) => [n.id, n.type!]));
  return withOwnedNodes(
    ids.filter((id) => !getNodeValues(types.get(id) ?? "").fixed),
    nodes,
    edges
  );
}

let ownerTypes: Map<string, string[]> | undefined;

// The types of the blocks that own blocks of the given type, e.g. the four
// condition types for the else branch.
export function getOwnerTypes(type: string) {
  if (!ownerTypes) {
    ownerTypes = new Map();
    for (const owner of Object.keys(nodeTypes)) {
      for (const owned of getOwnedChildTypes(owner)) {
        ownerTypes.set(owned, [...(ownerTypes.get(owned) ?? []), owner]);
      }
    }
  }
  return ownerTypes.get(type) ?? [];
}

// No handle and "default" both mean a block's default output.
export function normalizeHandle(handle?: string | null) {
  return handle && handle !== "default" ? handle : null;
}

// Edges to the blocks a block owns are fixed, the rest can be deleted in the
// editor like hand-drawn ones. Blocks that name their outputs, like the error
// handler, render their default output with the ID "default", so edges have
// to name it too.
export function createEdge(
  source: Node<NodeData>,
  target: Node<NodeData>,
  handle?: string | null
): Edge {
  const owned = getOwnedChildTypes(source.type!).includes(target.type!);
  const namesDefault = getNodeValues(source.type!).outputs?.includes("default");
  return {
    id: getEdgeId(),
    source: source.id,
    target: target.id,
    sourceHandle: normalizeHandle(handle) ?? (namesDefault ? "default" : null),
    type: owned ? "fixed" : "delete_button",
  };
}

// Options connect into the entry of commands and event listeners, nothing else
// connects into an entry, and nothing connects into an option.
export function canConnect(sourceType: string, targetType: string) {
  if (sourceType.startsWith("option_")) {
    return targetType === "entry_command" || targetType === "entry_event";
  }
  return !targetType.startsWith("entry_") && !targetType.startsWith("option_");
}

export function useNodeValues(nodeType: string): NodeValues {
  return useMemo(() => getNodeValues(nodeType), [nodeType]);
}

// Where blocks created together with an owner go, relative to it: the first
// to the right, like the else branch of a condition, the second to the left.
const ownedOffsets = [
  { x: 200, y: 200 },
  { x: -150, y: 200 },
];

export function createNode(
  type: string,
  position: XYPosition,
  props?: Partial<Node<NodeData>>
): [Node<NodeData>[], Edge[]] {
  const id = getNodeId();

  const nodes: Node<NodeData>[] = [
    {
      id,
      type,
      position,
      data: {},
      ...props,
    },
  ];
  const edges: Edge[] = [];

  // TODO?: connect option types to entry automatically?

  getOwnedChildTypes(type).forEach((ownedType, i) => {
    const offset = ownedOffsets[i] ?? ownedOffsets[ownedOffsets.length - 1];
    const [ownedNodes, ownedEdges] = createNode(ownedType, {
      x: position.x + offset.x,
      y: position.y + offset.y,
    });

    nodes.push(...ownedNodes);
    edges.push({
      id: getEdgeId(),
      source: id,
      target: ownedNodes[0].id,
      type: "fixed",
    });
    edges.push(...ownedEdges);
  });

  return [nodes, edges];
}

export function getNodeId(): string {
  // This gives us a pool size of 75000
  // There is a small chance of collision, but reactflow handles it gracefully
  return humanId({
    separator: "",
    capitalize: false,
    addAdverb: false,
    adjectiveCount: 0,
  });
}

export function getEdgeId(): string {
  return getUniqueId().toString();
}
