import { Edge, Node } from "@xyflow/react";
import { z } from "zod";
import { FlowContextType } from "../flow/context";
import { NodeData } from "../flow/dataSchema";
import { applyFlowEdits, FlowEdit } from "../flow/edits";
import { isSelectMenu } from "../flow/resume";
import { findFlowSourceComponent } from "../flow/templates";
import { describeIssue, FlowIssue, validateFlow } from "../flow/validate";
import { FlowData } from "../types/flow.gen";
import { ButtonStyleLink, ComponentData } from "../types/message.gen";

const settings = z.record(z.string(), z.unknown());
const block = z
  .string()
  .describe(
    "The ID of a block, $entry for the entry block, or the ref of a block added before in the same call."
  );

export const flowEditSchema = z.discriminatedUnion("op", [
  z.object({
    op: z.literal("add_node"),
    ref: z.string().describe('Names the block for later edits, like "$ban".'),
    type: z.string().describe("The block type."),
    data: settings.optional().describe("The settings of the block."),
    after: block.optional(),
    handle: z.string().optional(),
    before: block.optional(),
    items: z
      .array(settings)
      .optional()
      .describe("The settings of each branch of a condition, without else."),
  }),
  z.object({ op: z.literal("update_node"), id: block, data: settings }),
  z.object({
    op: z.literal("remove_node"),
    id: block,
    reconnect: z.boolean().optional(),
  }),
  z.object({
    op: z.literal("connect"),
    source: block,
    target: block,
    handle: z.string().optional(),
  }),
  z.object({
    op: z.literal("disconnect"),
    source: block,
    target: block,
    handle: z.string().optional(),
  }),
]);

// Applies edits to a flow, unless they leave it with errors it didn't have
// before. Those are thrown, so the client can fix them and send all edits
// again. Existing errors don't block saving, as the client may not be able to
// fix them.
export function editFlow(
  flow: FlowData,
  edits: z.infer<typeof flowEditSchema>[],
  context: FlowContextType
) {
  const nodes = flow.nodes as unknown as Node<NodeData>[];
  const edges = flow.edges as unknown as Edge[];

  const before = new Set(
    validateFlow(nodes, edges, context)
      .filter((i) => i.severity === "error")
      .map(describeIssue)
  );
  const res = applyFlowEdits({ nodes, edges }, edits as FlowEdit[], context);

  const errors = res.issues
    .filter((i) => i.severity === "error")
    .map(describeIssue)
    .filter((issue) => !before.has(issue));
  if (errors.length) {
    throw new Error(
      `Nothing was saved, the edits have problems:\n${list(
        errors
      )}\nFix them and send all edits again.`
    );
  }

  return {
    flow: {
      ...flow,
      nodes: res.nodes as unknown as FlowData["nodes"],
      edges: res.edges as unknown as FlowData["edges"],
    },
    refs: res.refs,
    warnings: res.issues.filter((i) => i.severity === "warning"),
  };
}

export function describeWarnings(warnings: FlowIssue[]) {
  return warnings.length
    ? `Warnings:\n${list(warnings.map(describeIssue))}`
    : "";
}

function list(items: string[]) {
  return items.map((item) => `- ${item}`).join("\n");
}

// The type of flow a button or select menu of a message template runs, by its
// flow_source_id, or undefined if the message has no such component that runs
// one.
export function componentFlowContext(
  data: unknown,
  flowSourceId: string
): FlowContextType | undefined {
  const component = findFlowSourceComponent(data, flowSourceId);
  if (!component || component.style === ButtonStyleLink) return undefined;
  return isSelectMenu(component as ComponentData)
    ? "component_select_menu"
    : "component_button";
}
