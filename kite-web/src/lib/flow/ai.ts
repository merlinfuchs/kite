import { Edge, Node } from "@xyflow/react";
import { APIResponse } from "../api/response";
import {
  FlowAIChatMessage,
  FlowAIChatRequest,
  FlowAIChatResponse,
} from "../types/wire.gen";
import { FlowContextType } from "./context";
import { NodeData } from "./dataSchema";
import { applyFlowEdits, FlowEdit } from "./edits";
import { serializeFlow } from "./serialize";
import { FlowIssue, validateFlow } from "./validate";

// The limits of the API.
const maxMessages = 20;
const maxMessageLength = 4000;
const maxIssues = 50;

interface Flow {
  nodes: Node<NodeData>[];
  edges: Edge[];
}

export class FlowAIError extends Error {
  constructor(message: string, public code: string) {
    super(message);
  }
}

export interface FlowAIResult {
  message: string;
  // How many rounds of problems the AI fixed in its own changes.
  repairs: number;
  // Problems the AI couldn't fix, or why it couldn't.
  issues: string[];
  // The blocks that were added or whose settings changed.
  changedNodeIds: string[];
}

// Asks the AI to change the flow and applies its edits to the flow as it is
// when they arrive. Problems the editor finds with them, that the flow didn't
// have before, are sent back to be fixed in repairs, which don't count as
// prompts, until the API allows no more.
export async function runFlowAIPrompt({
  context,
  messages,
  getFlow,
  applyFlow,
  send,
}: {
  context: FlowContextType;
  // The chat so far, ending with the user's new message.
  messages: FlowAIChatMessage[];
  getFlow: () => Flow;
  applyFlow: (flow: Flow, changedNodeIds: string[]) => void;
  send: (req: FlowAIChatRequest) => Promise<APIResponse<FlowAIChatResponse>>;
}): Promise<FlowAIResult> {
  const request = async (flow: Flow, req: Partial<FlowAIChatRequest>) => {
    const selectedIds = flow.nodes.filter((n) => n.selected).map((n) => n.id);
    const res = await send({
      flow: serializeFlow(flow.nodes, flow.edges, context, selectedIds),
      messages: toRequestMessages(messages),
      repair_prompt_id: "",
      issues: [],
      ...req,
    });
    if (!res.success) {
      throw new FlowAIError(res.error.message, res.error.code);
    }
    return res.data;
  };

  const original = getFlow();
  let known: Set<string> | undefined;
  let res = await request(original, {});
  const { prompt_id: promptId, message } = res;
  const changed = new Set<string>();
  let changedNodeIds: string[] = [];

  for (let repairs = 0; ; repairs++) {
    // The flow may have been edited while the AI was working.
    let flow = getFlow();
    let issues = res.issues;
    if (res.edits.length > 0) {
      const applied = applyFlowEdits(flow, res.edits as FlowEdit[], context);
      for (const id of getChangedNodeIds(flow.nodes, applied.nodes)) {
        changed.add(id);
      }
      flow = { nodes: applied.nodes, edges: applied.edges };
      // Blocks added and removed again by a repair are left out.
      const ids = new Set(flow.nodes.map((n) => n.id));
      changedNodeIds = [...changed].filter((id) => ids.has(id));
      applyFlow(flow, changedNodeIds);

      const errors = applied.issues
        .filter((i) => i.severity === "error")
        .map(describeIssue);
      if (errors.length > 0) {
        known ??= new Set(
          validateFlow(original.nodes, original.edges, context).map(
            describeIssue
          )
        );
        issues = [...issues, ...errors.filter((i) => !known!.has(i))];
      }
    }

    const result = { message, repairs, issues, changedNodeIds };
    if (issues.length === 0) return result;

    try {
      // The editor may not have rendered the applied flow yet, so it's sent
      // as applied.
      res = await request(flow, {
        messages: toRequestMessages([
          ...messages,
          { role: "assistant", content: message },
        ]),
        repair_prompt_id: promptId,
        issues: issues.slice(0, maxIssues),
      });
    } catch (err) {
      if (err instanceof FlowAIError && err.code === "repair_limit") {
        return result;
      }
      return { ...result, issues: [...issues, (err as Error).message] };
    }
  }
}

function toRequestMessages(messages: FlowAIChatMessage[]) {
  return messages
    .slice(-maxMessages)
    .map((m) => ({ ...m, content: m.content.slice(0, maxMessageLength) }));
}

function getChangedNodeIds(before: Node<NodeData>[], after: Node<NodeData>[]) {
  const previous = new Map(before.map((n) => [n.id, n.data]));
  return after
    .filter((n) => !previous.has(n.id) || previous.get(n.id) !== n.data)
    .map((n) => n.id);
}

// Blocks of the same type have the same title, so the ID tells the AI which
// one is meant.
function describeIssue(issue: FlowIssue) {
  const id = issue.nodeId ?? issue.edgeId;
  return id ? `${id}: ${issue.message}` : issue.message;
}
