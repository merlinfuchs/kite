import { itemVersionPathSegments } from "../api/queries";
import { parsePreviewMessage } from "../message/documentConvert";
import { Message } from "../message/schema";
import {
  FlowData,
  TestNote,
  TestNoteAICheckNotAsked,
  TestNoteAIChoiceNotAsked,
  TestNoteMocked,
  TestNoteNotRun,
  TestNoteSignatureNotChecked,
  TestNoteWaitSkipped,
} from "../types/flow.gen";
import { getBlockDefinition } from "../blocks";
import { ComponentData, ComponentTypeStringSelect } from "../types/message.gen";
import {
  FlowTestEffect,
  FlowTestMessage,
  FlowTestMock,
  FlowTestRequest,
  FlowTestResponse,
} from "../types/wire.gen";
import type { TestNodeStatus } from "./context";
import { isSelectMenu } from "./resume";
import { findFlowSourceComponent } from "./templates";

// What a test of the flow in the editor runs. Drafts are tested as they are
// in the editor, without saving them.
export type FlowTestTarget =
  | { type: "command"; id: string }
  | { type: "event_listener"; id: string }
  | {
      type: "message";
      id: string;
      // The button or select menu whose flow is tested.
      flowSourceId: string;
      // The message as it is in the editor.
      getMessage: () => FlowTestMessage;
    };

// The form and result of the test panel, kept by the editor so they stay
// while the panel is closed.
export type FlowTestState = {
  guildId: string | null;
  channelId: string | null;
  userId: string;
  // Values of command arguments by the ID of their block.
  options: Record<string, string>;
  // The user or message of a User or Message command.
  targetId: string;
  content: string;
  emoji: string;
  webhookBody: string;
  // Options picked in a select menu, or the IDs for menus of users, roles or
  // channels.
  values: string[];
  valueIds: string;
  // What blocks do instead of running, by the ID of the block.
  mocks: Record<string, FlowTestMock>;
  // Values of stored variables the test starts with.
  variables: FlowTestVariableInput[];
  // The last test and its result. Clicks and submits run it again with the
  // same seed and one more action, see TestRandom in the service.
  last: { request: FlowTestRequest; result: FlowTestResponse } | null;
  error: string | null;
};

export const initialFlowTestState: FlowTestState = {
  guildId: null,
  channelId: null,
  userId: "",
  options: {},
  targetId: "",
  content: "",
  emoji: "",
  webhookBody: "",
  values: [],
  valueIds: "",
  mocks: {},
  variables: [],
  last: null,
  error: null,
};

export function flowTestPath(target: FlowTestTarget) {
  return `${itemVersionPathSegments[target.type]}/${target.id}`;
}

// A seed for a new test, which JSON keeps exactly.
export function newFlowTestSeed() {
  return Math.floor(Math.random() * Number.MAX_SAFE_INTEGER) + 1;
}

// How each block ran in a test. A block that ran more than once, like in a
// loop, shows the worst.
export function flowTestNodeStatus(
  result: FlowTestResponse | undefined
): Record<string, TestNodeStatus> {
  const rank = { ok: 0, skipped: 1, warning: 2, error: 3 };
  const res: Record<string, TestNodeStatus> = {};
  for (const execution of result?.executions ?? []) {
    if (execution.other_flow) continue;
    for (const step of execution.steps) {
      const status: TestNodeStatus = step.error
        ? "error"
        : step.effects?.some((e) => e.warnings?.length)
        ? "warning"
        : step.notes?.some((n) => n.code === TestNoteNotRun)
        ? "skipped"
        : "ok";
      const current = res[step.node_id];
      if (!current || rank[status] > rank[current]) {
        res[step.node_id] = status;
      }
    }
  }
  return res;
}

// The IDs in a text like "123, 456".
export function splitIds(text: string) {
  return text
    .split(",")
    .map((v) => v.trim())
    .filter(Boolean);
}

// The select menu with the flow, or null if the flow is of a button. Only
// string selects have options, the others pick users, roles or channels.
export function findTestSelectMenu(
  data: unknown,
  flowSourceId: string
): ComponentData | null {
  const c = findFlowSourceComponent(data, flowSourceId) as
    | ComponentData
    | undefined;
  return c && isSelectMenu(c) ? c : null;
}

const messageEvents = [
  "message_create",
  "direct_message_create",
  "message_update",
  "message_delete",
];
const reactionEvents = ["message_reaction_add", "message_reaction_remove"];

// What the form asks for, which depends on the flow.
export function flowTestInputs(eventType: string) {
  return {
    message: messageEvents.includes(eventType),
    reaction: reactionEvents.includes(eventType),
    webhook: eventType === "webhook",
    // Scheduled and webhook runs have no user, server or channel.
    place: eventType !== "webhook" && eventType !== "schedule_cron",
  };
}

export type FlowTestVariableInput = {
  variableId: string;
  scope: string;
  // Like on the page of the variable: string, number, bool or json.
  type: string;
  // Empty for no value.
  value: string;
};

// The result of a block's test result as entered. Blocks that return text,
// like Ask AI, get the text. Others, like Get Roblox User, return more and
// take JSON.
export function mockTestResult(nodeType: string, text: string): unknown {
  return blockReturnsText(nodeType) ? text : jsonOrText(text);
}

export function blockReturnsText(nodeType: string) {
  return !!getBlockDefinition(nodeType)?.result?.schema?.safeParse("").success;
}

export function jsonOrText(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

// The notes of the service are written for MCP agents, which mock blocks.
// The editor calls that setting a test result.
const noteTexts: Record<string, string> = {
  [TestNoteNotRun]:
    "Didn't run. Tests don't send requests outside of Discord. Set a test result to test what comes after it.",
  [TestNoteAIChoiceNotAsked]:
    "AI isn't asked in tests, so no option was picked. Set a test result to pick one.",
  [TestNoteAICheckNotAsked]:
    "AI isn't asked in tests, so the answer was no. Set a test result to change it.",
  [TestNoteSignatureNotChecked]:
    "The signature isn't checked in tests, so the request counts as valid. Set a test result to test an invalid one.",
  [TestNoteMocked]: "Used the test result you set.",
  [TestNoteWaitSkipped]: "Tests don't wait, so the flow went on right away.",
};

// The notes of blocks that can be given a test result.
const mockableNotes = new Set<string>([
  TestNoteNotRun,
  TestNoteAIChoiceNotAsked,
  TestNoteAICheckNotAsked,
  TestNoteSignatureNotChecked,
  TestNoteMocked,
]);

export function flowTestNoteText(note: TestNote) {
  return Object.hasOwn(noteTexts, note.code) ? noteTexts[note.code] : note.text;
}

// Whether a step can be given a test result: blocks that didn't run because
// they call other services, asked AI or checked a webhook, and the ones that
// have one.
export function canSetTestResult(step: { notes?: TestNote[] }) {
  return !!step.notes?.some((n) => mockableNotes.has(n.code));
}

export type FlowTestArgument = { name: string; type?: string };

// The request of a new test from the form.
export function buildFlowTestRequest(
  state: FlowTestState,
  {
    target,
    flow,
    guildId,
    eventType,
    commandType,
    args,
  }: {
    target: FlowTestTarget;
    flow: FlowData;
    guildId: string | null;
    eventType: string;
    commandType: string;
    // The arguments of a command by the ID of their block.
    args: Record<string, FlowTestArgument>;
  }
): FlowTestRequest {
  const inputs = flowTestInputs(eventType);
  const req: FlowTestRequest = {
    user_id: state.userId.trim() || undefined,
    guild_id: guildId ?? undefined,
    channel_id: inputs.place ? state.channelId ?? undefined : undefined,
  };

  if (target.type === "message") {
    // The flow of the dialog is only in the message once it's closed.
    const message = target.getMessage();
    req.message = {
      ...message,
      flow_sources: { ...message.flow_sources, [target.flowSourceId]: flow },
    };
    req.flow_source_id = target.flowSourceId;
    const selectMenu = findTestSelectMenu(message.data, target.flowSourceId);
    if (selectMenu) {
      req.values =
        selectMenu.type === ComponentTypeStringSelect
          ? state.values
          : splitIds(state.valueIds);
    }
  } else {
    req.flow_source = flow;
  }

  if (target.type === "command") {
    if (commandType === "chat_input") {
      req.options = {};
      for (const [id, arg] of Object.entries(args)) {
        const value = state.options[id]?.trim();
        if (arg.name && value) {
          req.options[arg.name] = optionValue(arg, value);
        }
      }
    } else if (state.targetId.trim()) {
      req.options = { target: state.targetId.trim() };
    }
  }

  if (Object.keys(state.mocks).length) req.mocks = state.mocks;
  const variables = state.variables.filter((v) => v.variableId);
  if (variables.length) {
    // The service reads the text like the page of the variable does, and
    // rejects invalid values.
    req.variables = variables.map((v) => ({
      variable_id: v.variableId,
      scope: v.scope.trim() || undefined,
      ...(v.value.trim() ? { type: v.type, value: v.value } : { value: null }),
    }));
  }

  if (inputs.message && state.content) {
    req.event = { content: state.content };
  } else if (inputs.reaction && state.emoji.trim()) {
    req.event = { emoji: { name: state.emoji.trim() } };
  } else if (inputs.webhook) {
    req.webhook = { body: state.webhookBody };
  }

  return req;
}

// Throws for numbers Discord wouldn't send.
function optionValue(arg: FlowTestArgument, value: string): unknown {
  switch (arg.type) {
    case "integer":
      if (!/^-?\d+$/.test(value)) {
        throw new Error(`${arg.name} must be a whole number.`);
      }
      return Number(value);
    case "number":
      if (!Number.isFinite(Number(value))) {
        throw new Error(`${arg.name} must be a number.`);
      }
      return Number(value);
    case "boolean":
      return value === "true";
    default:
      return value;
  }
}

const effectTitles: Record<string, string> = {
  add_reaction: "Would add a reaction",
  create_channel: "Would create a channel",
  create_followup: "Would send a follow-up message",
  create_forum_post: "Would create a forum post",
  create_message: "Would send a message",
  create_poll: "Would create a poll",
  create_thread: "Would create a thread",
  delete_followup: "Would delete a follow-up message",
  delete_response: "Would delete the response",
  delete_variable: "Would delete a stored value",
  discord_api_request: "Would send a Discord API request",
  edit_channel: "Would edit a channel",
  edit_followup: "Would edit a follow-up message",
  edit_member: "Would edit a member",
  edit_message: "Would edit a message",
  edit_response: "Would edit the response",
  edit_server_profile: "Would change the bot's server profile",
  forward_message: "Would forward a message",
  leave_server: "Would leave the server",
  log: "Would write a log",
  play_audio: "Would play audio",
  respond: "Would respond",
  respond_webhook: "Would answer the webhook request",
  set_variable: "Would change a stored value",
  stop_audio: "Would stop audio",
  update_status: "Would change the bot's status",
  update_voice_state: "Would join or move in voice",
};

const respondTitles: Record<string, string> = {
  message: "Would respond with a message",
  update_message: "Would update the message",
  modal: "Would show a modal",
  defer: "Would show that the bot is thinking",
  defer_update: "Would acknowledge the interaction",
};

export function flowTestEffectTitle(effect: FlowTestEffect) {
  if (effect.action === "respond") {
    const title = respondTitles[effect.data?.type];
    if (title) return title;
  }
  return effectTitles[effect.action] ?? "Would make a change in Discord";
}

// The message an effect would send or edit to, for the preview, or null if
// it has none. It's in Discord's format, which the editor reads like its own
// once custom IDs are flow_source_id.
export function flowTestEffectMessage(effect: FlowTestEffect): Message | null {
  const data = effect.data?.message;
  if (!data || typeof data !== "object") return null;

  const msg = parsePreviewMessage({
    ...data,
    components: withFlowSourceIds(data.components),
  });
  if (!msg) return null;
  return {
    ...msg,
    username: effect.data?.name || undefined,
    avatar_url: effect.data?.avatar_url || undefined,
  };
}

function withFlowSourceIds(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(withFlowSourceIds);
  if (!value || typeof value !== "object") return value;
  return Object.fromEntries(
    Object.entries(value).map(([key, v]) => [
      key === "custom_id" ? "flow_source_id" : key,
      withFlowSourceIds(v),
    ])
  );
}
