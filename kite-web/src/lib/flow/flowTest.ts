import { itemVersionPathSegments } from "../api/queries";
import { parsePreviewMessage } from "../message/documentConvert";
import { Message } from "../message/schema";
import { FlowData } from "../types/flow.gen";
import { ComponentData, ComponentTypeStringSelect } from "../types/message.gen";
import {
  FlowTestEffect,
  FlowTestMessage,
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
  const rank = { ok: 0, warning: 1, error: 2 };
  const res: Record<string, TestNodeStatus> = {};
  for (const execution of result?.executions ?? []) {
    for (const step of execution.steps) {
      const status: TestNodeStatus = step.error
        ? "error"
        : step.effects?.some((e) => e.warnings?.length)
        ? "warning"
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
          req.options[arg.name] = optionValue(arg.type, value);
        }
      }
    } else if (state.targetId.trim()) {
      req.options = { target: state.targetId.trim() };
    }
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

function optionValue(type: string | undefined, value: string): unknown {
  switch (type) {
    case "integer":
    case "number":
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
  return (
    effectTitles[effect.action] ?? "Would " + effect.action.replaceAll("_", " ")
  );
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
