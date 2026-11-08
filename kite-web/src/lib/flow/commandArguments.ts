import { Edge, Node } from "@xyflow/react";
import { NodeData } from "./dataSchema";

// The argument blocks of a command in the order they are shown in Discord,
// like commandArgumentNodes in the service. Required arguments always come
// first, then arguments follow command_argument_order of the command.
export function getCommandArguments(
  entry: Node<NodeData>,
  nodes: Node<NodeData>[],
  edges: Edge[]
): Node<NodeData>[] {
  const required: Node<NodeData>[] = [];
  const optional: Node<NodeData>[] = [];

  for (const edge of edges) {
    // The service only reads connections into a block's single input
    if (edge.target !== entry.id || edge.targetHandle != null) continue;

    const node = nodes.find((n) => n.id === edge.source);
    if (!node || node.type !== "option_command_argument") continue;

    if (node.data.command_argument_required) {
      // Without an order the last connected required argument is first
      required.unshift(node);
    } else {
      optional.push(node);
    }
  }

  const order = entry.data.command_argument_order ?? [];
  const rank = (node: Node<NodeData>) => {
    const i = order.indexOf(node.id);
    return i === -1 ? order.length : i;
  };
  const byRank = (a: Node<NodeData>, b: Node<NodeData>) => rank(a) - rank(b);

  return [...required.sort(byRank), ...optional.sort(byRank)];
}

// Moves an argument up or down among the arguments it can be swapped with,
// and returns the new command_argument_order. Returns null when it can't move
// further, as required arguments can't go below optional ones.
export function moveCommandArgument(
  args: Node<NodeData>[],
  index: number,
  direction: -1 | 1
): string[] | null {
  const arg = args[index];
  const other = args[index + direction];
  if (!arg || !other) return null;

  if (
    !!arg.data.command_argument_required !==
    !!other.data.command_argument_required
  ) {
    return null;
  }

  const order = args.map((a) => a.id);
  order[index] = other.id;
  order[index + direction] = arg.id;
  return order;
}
