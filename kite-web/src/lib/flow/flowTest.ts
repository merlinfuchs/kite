import { Message, messageSchema } from "../message/schema";
import {
  FlowTestEffect,
  FlowTestMessage,
  FlowTestResponse,
} from "../types/wire.gen";

// What a test of the flow in the editor runs. Drafts are tested as they are
// in the editor, without saving them.
export type FlowTestTarget =
  | { kind: "command"; id: string }
  | { kind: "event_listener"; id: string }
  | {
      kind: "message";
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
  result: FlowTestResponse | null;
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
  result: null,
  error: null,
};

export function flowTestPath(target: FlowTestTarget) {
  switch (target.kind) {
    case "command":
      return `commands/${target.id}`;
    case "event_listener":
      return `event-listeners/${target.id}`;
    case "message":
      return `messages/${target.id}`;
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

// The message an effect would send or edit to, for the preview. It's in
// Discord's format, which is close enough to the editor's to show it. Null if
// it has no message or it can't be shown.
export function flowTestEffectMessage(effect: FlowTestEffect): Message | null {
  const data = effect.data?.message;
  if (!data || typeof data !== "object") return null;

  const res = messageSchema.safeParse({
    ...(toEditorFormat(data) as object),
    username: effect.data?.name || undefined,
    avatar_url: effect.data?.avatar_url || undefined,
    attachments: [],
  });
  return res.success ? res.data : null;
}

// Leaves out nulls, which the editor leaves out instead, and makes the custom
// IDs of buttons and select menus their flow_source_id.
function toEditorFormat(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(toEditorFormat);
  if (!value || typeof value !== "object") return value;

  const res: Record<string, unknown> = {};
  for (const [key, v] of Object.entries(value)) {
    if (key === "custom_id") {
      res.flow_source_id = v;
    } else if (v !== null) {
      res[key] = toEditorFormat(v);
    }
  }
  return res;
}
