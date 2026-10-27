import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { welcomeFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Discord MCP Server for Claude Code and Cursor",
  description:
    "Connect Claude Code or Cursor to Kite's MCP server and let AI build the commands and events of your Discord bot. Kite hosts the bot for you.",
  heading: "Build Your Discord Bot From Claude Code or Cursor",
  intro:
    "Kite is an MCP server for Discord bots. Connect your AI client, describe what your bot should do, and it builds the commands, event listeners and messages in your Kite app. Kite runs the bot, so there's nothing to deploy or host.",
  secondaryAction: {
    label: "Read the MCP Guide",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/mcp`,
  },
  showcase: {
    flow: welcomeFlow,
    prompt:
      "When someone joins, give them the Member role and welcome them in #welcome.",
    event: "Alex joined the server.",
    reply: "Welcome to the server, @Alex!",
  },
  sections: [
    {
      heading: "Set It Up in Two Minutes",
      points: [
        {
          title: "Create an API Token",
          description:
            "Choose whether the token can change your apps or only read them, and which apps it can use.",
        },
        {
          title: "Connect Your Client",
          description:
            "Add Kite to Claude Code with one command, or to Cursor and other clients that connect to remote MCP servers.",
        },
        {
          title: "Describe Your Bot",
          description:
            "Ask for a command or an event in plain words. The client builds the flow, Kite checks it, and it goes live in your server.",
        },
      ],
    },
    {
      heading: "What Your AI Client Can Do",
      body: "The client builds the same flows you would build in Kite's editor, so you can open them there to see and change what it did.",
      points: [
        {
          title: "Build and Change Flows",
          description:
            "Create commands, event listeners and message templates with buttons and select menus, or change the ones you already have.",
        },
        {
          title: "Fix What's Broken",
          description:
            "Ask why a command fails. The client reads your app's logs, finds the block that failed and fixes it.",
        },
        {
          title: "Use Your Real Channels and Roles",
          description:
            "The client sees the channels and roles of your servers, so the blocks it adds point at the right ones.",
        },
        {
          title: "Manage Your Server",
          description:
            "With Discord API access, the client can also read members, send a message or create channels and roles as your bot.",
        },
        {
          title: "Undo Any Change",
          description:
            "Every change is kept in the history of the command, event listener or message, so you can restore an earlier version.",
        },
        {
          title: "Your Own AI Model",
          description:
            "The client runs on its own model, so building with it doesn't use the prompts of Kite's flow AI.",
        },
      ],
    },
    {
      heading: "You Decide What It Can Touch",
      points: [
        {
          title: "Read-Only Tokens",
          description:
            "Give a token read access only, and the client can look at your apps and logs without changing anything.",
        },
        {
          title: "Only Some Apps",
          description:
            "Limit a token to the apps you pick. The client doesn't see your other bots.",
        },
        {
          title: "Every Change Logged",
          description:
            "Requests to Discord are logged in your app's logs and the server's audit log with the name of the token. Mass bans, pruning and leaving a server are blocked.",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "What is an MCP server?",
      answer:
        "MCP, the Model Context Protocol, is a standard way for AI clients to use tools of other apps. Kite's MCP server gives clients tools to read and build your Discord bot, like listing your commands or adding a block to a flow.",
    },
    {
      question: "Which AI clients work with Kite?",
      answer:
        "Claude Code, Cursor and other clients that connect to remote MCP servers with a token in a header. Connectors in claude.ai, the Claude app and ChatGPT need a sign-in instead, which Kite doesn't support yet.",
    },
    {
      question: "Does it cost anything?",
      answer:
        "No. The MCP server works on every plan and doesn't use credits or flow AI prompts. The flows it builds use credits when they run, like any other flow.",
    },
    {
      question: "Is it safe to let an AI change my bot?",
      answer:
        "You choose what each token can do, and every change can be undone from the history. Kite tells the client to only delete or disable things you asked for. Keep the token private, because anyone who has it can control your bot.",
    },
    {
      question: "Do I need to host the bot myself?",
      answer:
        "No. Kite runs your bot 24/7. The AI client only builds and changes it, so the bot keeps running when you close the client.",
    },
  ],
};

export default function DiscordMCPServerPage() {
  return <LandingPage content={content} />;
}
