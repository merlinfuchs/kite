import { Edge, Node } from "@xyflow/react";
import { APIResponse } from "../api/response";
import {
  FlowAIChatMessage,
  FlowAIChatRequest,
  FlowAIChatResponse,
  FlowAIUsage,
} from "../types/wire.gen";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { applyFlowEdits, FlowEdit } from "./edits";
import { serializeFlow } from "./serialize";
import { validateFlow } from "./validate";

// The limits of the API.
const maxMessages = 20;
const maxMessageLength = 4000;
const maxIssues = 50;
const maxRepairs = 2;

export class FlowAIError extends Error {
  constructor(message: string, public code: string) {
    super(message);
  }
}

export interface FlowAIResult {
  nodes: Node<NodeData>[];
  edges: Edge[];
  message: string;
  // Whether the AI changed the flow, rather than only answering.
  edited: boolean;
  // How many rounds of problems the AI fixed in its own changes.
  repairs: number;
  // Problems the AI couldn't fix, or why it couldn't.
  issues: string[];
  changedNodeIds: string[];
  usage: FlowAIUsage;
}

// Asks the AI to change the flow and applies its edits. Problems the editor
// finds with them, that the flow didn't have before, are sent back to be fixed
// in up to two repairs, which don't count as prompts.
export async function runFlowAIPrompt({
  flow,
  context,
  selectedIds,
  messages,
  send,
}: {
  flow: { nodes: Node<NodeData>[]; edges: Edge[] };
  context: FlowContextType;
  selectedIds: string[];
  // The chat so far, ending with the user's new message.
  messages: FlowAIChatMessage[];
  send: (req: FlowAIChatRequest) => Promise<APIResponse<FlowAIChatResponse>>;
}): Promise<FlowAIResult> {
  let current = { nodes: flow.nodes, edges: flow.edges };
  const request = async (req: Partial<FlowAIChatRequest>) => {
    const res = await send({
      flow_type: context,
      flow: serializeFlow(current.nodes, current.edges, context, selectedIds),
      messages: toRequestMessages(messages, maxMessages),
      repair_prompt_id: "",
      issues: [],
      ...req,
    });
    if (!res.success) {
      throw new FlowAIError(res.error.message, res.error.code);
    }
    return res.data;
  };

  let res = await request({});
  const { prompt_id: promptId, message } = res;
  let edited = false;
  const known = new Set(
    validateFlow(flow.nodes, flow.edges, context).map((i) => i.message)
  );

  for (let repairs = 0; ; repairs++) {
    const applied = applyFlowEdits(current, res.edits as FlowEdit[], context);
    current = { nodes: applied.nodes, edges: applied.edges };
    edited ||= res.edits.length > 0;
    const issues = [
      ...res.issues,
      ...applied.issues
        .filter((i) => i.severity === "error" && !known.has(i.message))
        .map((i) => i.message),
    ];

    const result = {
      ...current,
      message,
      edited,
      repairs,
      issues,
      changedNodeIds: getChangedNodeIds(flow.nodes, current.nodes),
      usage: res.usage,
    };
    if (issues.length === 0 || repairs === maxRepairs) return result;

    try {
      res = await request({
        messages: toRequestMessages(
          [...messages, { role: "assistant", content: message }],
          maxMessages
        ),
        repair_prompt_id: promptId,
        issues: issues.slice(0, maxIssues),
      });
    } catch (err) {
      return { ...result, issues: [...issues, (err as Error).message] };
    }
  }
}

function toRequestMessages(messages: FlowAIChatMessage[], limit: number) {
  return messages.slice(-limit).map((m) => ({
    ...m,
    // The API needs content, but the AI can answer with edits alone.
    content: m.content.slice(0, maxMessageLength) || "Done.",
  }));
}

// The blocks that were added or whose settings changed, to highlight them.
export function getChangedNodeIds(
  before: Node<NodeData>[],
  after: Node<NodeData>[]
) {
  const previous = new Map(before.map((n) => [n.id, n.data]));
  return after
    .filter((n) => !previous.has(n.id) || previous.get(n.id) !== n.data)
    .map((n) => n.id);
}
