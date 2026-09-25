import { Edge, Node } from "@xyflow/react";
import { APIResponse } from "../api/response";
import {
  FlowAIChatMessage,
  FlowAIChatRequest,
  FlowAIChatResponse,
  FlowAICheckField,
  FlowAICheckRequest,
  FlowAICheckResponse,
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
// when they arrive. Problems its edits caused are sent back to be fixed in
// repairs, which don't count as prompts, until the API allows no more.
export async function runFlowAIPrompt({
  context,
  messages,
  getFlow,
  applyFlow,
  send,
  signal,
}: {
  context: FlowContextType;
  // The chat so far, ending with the user's new message.
  messages: FlowAIChatMessage[];
  getFlow: () => Flow;
  applyFlow: (flow: Flow, changedNodeIds: string[]) => void;
  send: (req: FlowAIChatRequest) => Promise<APIResponse<FlowAIChatResponse>>;
  // Stops the prompt, e.g. when the editor is closed.
  signal?: AbortSignal;
}): Promise<FlowAIResult> {
  // The changed blocks are selected to highlight them, so only the blocks the
  // user selected before are sent as selected.
  const original = getFlow();
  const selectedIds = original.nodes.filter((n) => n.selected).map((n) => n.id);

  const request = async (flow: Flow, req: Partial<FlowAIChatRequest>) => {
    signal?.throwIfAborted();
    const res = await send({
      flow: serializeFlow(flow.nodes, flow.edges, context, selectedIds),
      messages: toRequestMessages(messages),
      repair_prompt_id: "",
      issues: [],
      ...req,
    });
    signal?.throwIfAborted();
    if (!res.success) {
      throw new FlowAIError(res.error.message, res.error.code);
    }
    return res.data;
  };
  const getErrors = (issues: FlowIssue[]) =>
    new Set(issues.filter((i) => i.severity === "error").map(describeIssue));

  let res = await request(original, {});
  const { prompt_id: promptId, message } = res;
  const changed = new Set<string>();
  let result: FlowAIResult = {
    message,
    repairs: 0,
    issues: [],
    changedNodeIds: [],
  };
  let applied: Flow | undefined;
  // The errors the AI's edits caused, rather than the user.
  let caused = new Set<string>();

  for (let repairs = 0; ; repairs++) {
    // The flow may have been edited while the AI was working.
    let flow = getFlow();
    if (applied && wasReverted(applied, flow, result.changedNodeIds)) {
      // The user undid or changed the edits, so they aren't fixed anymore.
      return { ...result, issues: [] };
    }

    const before = getErrors(validateFlow(flow.nodes, flow.edges, context));
    let after = before;
    if (res.edits.length > 0) {
      const edited = applyFlowEdits(flow, res.edits as FlowEdit[], context);
      for (const id of getChangedNodeIds(flow.nodes, edited.nodes)) {
        changed.add(id);
      }
      flow = applied = { nodes: edited.nodes, edges: edited.edges };
      after = getErrors(edited.issues);
    }
    caused = new Set([...after].filter((i) => !before.has(i) || caused.has(i)));

    // Blocks added and removed again by a repair are left out.
    const ids = new Set(flow.nodes.map((n) => n.id));
    result = {
      message,
      repairs,
      issues: [...res.issues, ...caused],
      changedNodeIds: [...changed].filter((id) => ids.has(id)),
    };
    if (res.edits.length > 0) applyFlow(flow, result.changedNodeIds);
    if (result.issues.length === 0) return result;

    try {
      // The editor may not have rendered the applied flow yet, so it's sent
      // as applied.
      res = await request(flow, {
        messages: toRequestMessages([
          ...messages,
          { role: "assistant", content: message },
        ]),
        repair_prompt_id: promptId,
        issues: result.issues.slice(0, maxIssues),
      });
    } catch (err) {
      if (err instanceof FlowAIError && err.code === "repair_limit") {
        return result;
      }
      signal?.throwIfAborted();
      return { ...result, issues: [...result.issues, (err as Error).message] };
    }
  }
}

// wasReverted reports whether blocks the AI changed are gone or have other
// settings than it gave them.
function wasReverted(applied: Flow, current: Flow, changedNodeIds: string[]) {
  const data = new Map(current.nodes.map((n) => [n.id, n.data]));
  const appliedData = new Map(applied.nodes.map((n) => [n.id, n.data]));
  return changedNodeIds.some((id) => data.get(id) !== appliedData.get(id));
}

// Asks a cheaper model whether the first prompt of a chat has what the AI
// needs. Returns null if the check fails, so the prompt is sent as it is.
export async function checkFlowAIPrompt({
  context,
  prompt,
  flow,
  send,
}: {
  context: FlowContextType;
  prompt: string;
  flow: Flow;
  send: (req: FlowAICheckRequest) => Promise<APIResponse<FlowAICheckResponse>>;
}): Promise<FlowAICheckResponse | null> {
  const selectedIds = flow.nodes.filter((n) => n.selected).map((n) => n.id);
  try {
    const res = await send({
      flow: serializeFlow(flow.nodes, flow.edges, context, selectedIds),
      prompt,
    });
    return res.success ? res.data : null;
  } catch {
    return null;
  }
}

// Adds the values the user filled in for a checked prompt, leaving out empty
// ones.
export function composeCheckedPrompt(
  prompt: string,
  fields: FlowAICheckField[],
  values: string[]
) {
  const details = fields
    .map((f, i) => [f.label, values[i]?.trim()])
    .filter(([, value]) => value)
    .map(([label, value]) => `- ${label}: ${value}`);
  return details.length > 0 ? `${prompt}\n\n${details.join("\n")}` : prompt;
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
