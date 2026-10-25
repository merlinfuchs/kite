import { isGroupNode } from "./groups";
import { isNoteNode } from "./notes";

// Notes and groups are on the canvas to explain and organize a flow. They're
// saved with its blocks, but aren't blocks: nothing connects to them, the
// service never runs them and the flow AI doesn't edit them.
export function isAnnotationNode(node: { type?: string }) {
  return isNoteNode(node) || isGroupNode(node);
}
