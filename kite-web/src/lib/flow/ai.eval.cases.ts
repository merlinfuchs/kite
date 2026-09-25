import { FlowContextType } from "./context";
import { FlowData, NodeData } from "./dataSchema";
import {
  getAITemplate,
  getModerationTemplate,
  getWelcomerTemplate,
  prepareTemplateFlow,
} from "./templates";

// What should happen with a prompt: the flow is changed, the user is asked
// for something that's missing, or the question is answered.
export type EvalRoute = "build" | "clarify" | "answer";

export interface EvalCase {
  name: string;
  context: FlowContextType;
  flow: () => FlowData;
  prompt: string;
  // Any of these is fine.
  route: EvalRoute[];
  // Block types the flow must have after building. "a|b" means either.
  types?: string[];
  // Whether a block's output for a button or select menu must be used.
  componentBranch?: boolean;
}

const channelId = "123456789012345678";
const roleId = "987654321098765432";

function entry(type: string, data: NodeData): () => FlowData {
  return () => ({
    nodes: [{ id: "entry", type, position: { x: 0, y: 0 }, data }],
    edges: [],
  });
}

const command = (name: string) =>
  entry("entry_command", { name, description: `The ${name} command` });
const event = (eventType: string) =>
  entry("entry_event", { event_type: eventType, description: "An event" });
const schedule = entry("entry_event", {
  event_type: "cron",
  event_schedule_cron: "0 * * * *",
  description: "A schedule",
});
const button = entry("entry_component_button", {});

const template = (flow: Parameters<typeof prepareTemplateFlow>[0]) => () =>
  prepareTemplateFlow(flow);
const banFlow = template(getModerationTemplate().commands[0].flowSource({}));
const welcomeFlow = template(
  getWelcomerTemplate().eventListeners[0].flowSource({ channel_id: channelId })
);
const aiFlow = template(
  getAITemplate().eventListeners[0].flowSource({
    system_prompt: "You are a friendly assistant.",
  })
);

export const evalCases: EvalCase[] = [
  // Commands
  {
    name: "ping",
    context: "command",
    flow: command("ping"),
    prompt: "Reply with pong",
    route: ["build"],
    types: ["action_response_create"],
  },
  {
    name: "greet user",
    context: "command",
    flow: command("hello"),
    prompt: "make it say hello to the user who used the command",
    route: ["build"],
    types: ["action_response_create"],
  },
  {
    name: "ban with dm",
    context: "command",
    flow: command("ban"),
    prompt:
      "Ban the user from the argument and send them a DM with the reason before",
    route: ["build"],
    types: [
      "option_command_argument",
      "action_private_message_create",
      "action_member_ban",
    ],
  },
  {
    name: "role from argument",
    context: "command",
    flow: command("giverole"),
    prompt: "Give the user the role I pick in an argument",
    route: ["build"],
    types: ["option_command_argument", "action_member_role_add"],
  },
  {
    name: "dice",
    context: "command",
    flow: command("dice"),
    prompt: "Roll a dice from 1 to 6 and send the result",
    route: ["build"],
    types: ["action_random_generate", "action_response_create"],
  },
  {
    name: "admins only",
    context: "command",
    flow: command("secret"),
    prompt: "Only admins should be able to use this and it says 'hi admin'",
    route: ["build"],
    types: ["option_command_permissions", "action_response_create"],
  },
  {
    name: "log without channel",
    context: "command",
    flow: command("report"),
    prompt: "log every time someone uses this command in #mod-logs",
    route: ["clarify"],
  },
  {
    name: "yes no buttons",
    context: "command",
    flow: command("confirm"),
    prompt:
      "Send a message with two buttons Yes and No, and when they click yes reply Great!",
    route: ["build"],
    types: ["action_response_create"],
    componentBranch: true,
  },
  {
    name: "modal color",
    context: "command",
    flow: command("color"),
    prompt:
      "Ask the user with a popup what their favorite color is and reply with it",
    route: ["build"],
    types: ["suspend_response_modal", "action_response_create"],
  },
  {
    name: "usage counter",
    context: "command",
    flow: command("count"),
    // Counting needs a stored variable, which is created outside the flow.
    prompt: "count how many times this command was used and show the number",
    route: ["clarify", "answer"],
  },
  {
    name: "german random number",
    context: "command",
    flow: command("zahl"),
    prompt: "Mach dass der Befehl eine zufällige Zahl zwischen 1 und 100 sagt",
    route: ["build"],
    types: ["action_random_generate", "action_response_create"],
  },
  {
    name: "vague cool",
    context: "command",
    flow: command("cool"),
    prompt: "make it cool",
    route: ["clarify"],
  },
  {
    name: "ticket system",
    context: "command",
    flow: command("ticket"),
    prompt: "i want a ticket system",
    route: ["clarify", "build"],
  },
  {
    name: "timeout",
    context: "command",
    flow: command("mute"),
    prompt: "timeout the user from the argument for 10 minutes",
    route: ["build"],
    types: ["option_command_argument", "action_member_timeout"],
  },
  {
    name: "cat api",
    context: "command",
    flow: command("cat"),
    prompt:
      "Send a random cat picture from the api https://api.thecatapi.com/v1/images/search",
    route: ["build"],
    types: ["action_http_request", "action_response_create"],
  },
  {
    name: "ask ai",
    context: "command",
    flow: command("ask"),
    prompt:
      "Ask ChatGPT the question from the argument and reply with the answer",
    route: ["build"],
    types: ["option_command_argument", "action_ai_chat_completion"],
  },
  {
    name: "select poll",
    context: "command",
    flow: command("poll"),
    prompt:
      "make a poll with a select menu with pizza, pasta and burger and reply with what they picked",
    route: ["build"],
    types: ["action_response_create"],
    componentBranch: true,
  },
  {
    name: "question one channel",
    context: "command",
    flow: command("hello"),
    prompt: "How can I make the bot only answer in one channel?",
    route: ["answer"],
  },
  {
    name: "question placeholders",
    context: "command",
    flow: command("hello"),
    prompt: "whats a placeholder and how do i use them",
    route: ["answer"],
  },

  // Changing the ban command of the moderation template
  {
    name: "ban explain",
    context: "command",
    flow: banFlow,
    prompt: "What does this command do?",
    route: ["answer"],
  },
  {
    name: "ban log reason",
    context: "command",
    flow: banFlow,
    prompt: `also log the ban and the reason to channel ${channelId}`,
    route: ["build"],
    types: ["action_member_ban", "action_message_create"],
  },
  {
    name: "ban ephemeral",
    context: "command",
    flow: banFlow,
    prompt: "make the reply visible to everyone, not just the moderator",
    route: ["build"],
    types: ["action_member_ban", "action_response_create"],
  },
  {
    name: "ban protect admins",
    context: "command",
    flow: banFlow,
    prompt: `if the user has the role ${roleId}, dont ban them and say you cant`,
    route: ["build"],
    types: [
      "control_condition_role|control_condition_user",
      "action_member_ban",
    ],
  },
  {
    name: "ban error handling",
    context: "command",
    flow: banFlow,
    prompt: "if the ban fails tell the user it didn't work",
    route: ["build"],
    types: ["control_error_handler"],
  },

  // Discord events
  {
    name: "reply hello",
    context: "event_discord",
    flow: event("message_create"),
    prompt: "When someone says hello, reply hi",
    route: ["build"],
    types: [
      "control_condition_compare|option_event_filter",
      "action_message_create",
    ],
  },
  {
    name: "delete invites",
    context: "event_discord",
    flow: event("message_create"),
    prompt: "Delete messages that contain a discord invite link",
    route: ["build"],
    types: [
      "control_condition_compare|option_event_filter",
      "action_message_delete",
    ],
  },
  {
    name: "react in channel",
    context: "event_discord",
    flow: event("message_create"),
    prompt: `react with 👍 to every message in channel ${channelId}`,
    route: ["build"],
    types: [
      "control_condition_channel|option_event_filter",
      "action_message_reaction_create",
    ],
  },
  {
    name: "welcome without channel",
    context: "event_discord",
    flow: event("guild_member_add"),
    prompt: "Welcome new members in #welcome",
    route: ["clarify"],
  },
  {
    name: "join role",
    context: "event_discord",
    flow: event("guild_member_add"),
    prompt: `Give new members the role ${roleId}`,
    route: ["build"],
    types: ["action_member_role_add"],
  },
  {
    name: "join dm rules",
    context: "event_discord",
    flow: event("guild_member_add"),
    prompt: "DM new members a welcome message with the server rules",
    route: ["build", "clarify"],
    types: ["action_private_message_create"],
  },
  {
    name: "log deletes",
    context: "event_discord",
    flow: event("message_delete"),
    prompt: `log deleted messages to channel ${channelId}`,
    route: ["build"],
    types: ["action_message_create"],
  },
  {
    name: "german good night",
    context: "event_discord",
    flow: event("message_create"),
    prompt: "Wenn jemand 'gute nacht' schreibt, antworte mit 'Schlaf gut!'",
    route: ["build"],
    types: [
      "control_condition_compare|option_event_filter",
      "action_message_create",
    ],
  },
  {
    name: "welcome add role",
    context: "event_discord",
    flow: welcomeFlow,
    prompt: `also give them the role ${roleId}`,
    route: ["build"],
    types: ["action_member_role_add", "action_message_create"],
  },
  {
    name: "welcome question",
    context: "event_discord",
    flow: welcomeFlow,
    prompt: "why isnt my welcome message showing up??",
    route: ["answer"],
  },
  {
    name: "ai one channel",
    context: "event_discord",
    flow: aiFlow,
    prompt: `make it only answer in channel ${channelId}`,
    route: ["build"],
    types: [
      "control_condition_channel|option_event_filter",
      "action_ai_chat_completion",
    ],
  },
  {
    name: "ai memory question",
    context: "event_discord",
    flow: aiFlow,
    prompt: "Can the bot remember earlier messages?",
    route: ["answer"],
  },

  // Schedules
  {
    name: "good morning",
    context: "event_schedule",
    flow: schedule,
    prompt: `Every day at 9am UTC send good morning in channel ${channelId}`,
    route: ["build"],
    types: ["action_message_create"],
  },
  {
    name: "status rotation",
    context: "event_schedule",
    flow: schedule,
    prompt: "Every hour change the bot status to a random game",
    route: ["build"],
    types: ["action_status_set"],
  },
  {
    name: "vague reminder",
    context: "event_schedule",
    flow: schedule,
    prompt: "post a daily reminder",
    route: ["clarify"],
  },

  // Buttons and select menus
  {
    name: "button role",
    context: "component_button",
    flow: button,
    prompt: `When clicked, give the user the role ${roleId} and say done`,
    route: ["build"],
    types: ["action_member_role_add", "action_response_create"],
  },
  {
    name: "button counter",
    context: "component_button",
    flow: button,
    prompt: "reply with how many times the button was clicked",
    route: ["clarify", "answer"],
  },
  {
    name: "button feedback modal",
    context: "component_button",
    flow: button,
    prompt: `open a modal asking for feedback and send it to channel ${channelId}`,
    route: ["build"],
    types: ["suspend_response_modal", "action_message_create"],
  },
  {
    name: "select reply",
    context: "component_select_menu",
    flow: button,
    prompt: "Reply with the option the user picked",
    route: ["build"],
    types: ["action_response_create"],
  },
  {
    name: "select role",
    context: "component_select_menu",
    flow: button,
    prompt:
      "the option values are role ids, give the user the role they picked",
    route: ["build"],
    types: ["action_member_role_add"],
  },
  {
    name: "button question",
    context: "component_button",
    flow: button,
    prompt: "how do i know who clicked the button?",
    route: ["answer"],
  },
];
