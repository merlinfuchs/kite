import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { freeFaq, openSourcePoint } from "@/components/landing/copy";
import { banFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "AI Discord Bot Maker: Describe Your Bot, AI Builds It",
  description:
    "Make a Discord bot by describing it. Kite's AI builds the commands in the editor, or connect Claude or ChatGPT to build the whole bot. Free hosting, no code.",
  heading: "Describe Your Discord Bot, AI Builds It",
  intro:
    "Tell Kite's AI what a command should do and it adds the blocks for you. Claude and ChatGPT can build whole bots when you connect them to Kite. Whatever they build runs on Kite 24/7, and you can open every flow and change it yourself.",
  secondaryAction: {
    label: "Read the AI Guide",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/flow-ai`,
  },
  showcase: {
    flow: banFlow,
    command: "ban",
    reply: "Banned @Sam: Spamming invites",
  },
  sections: [
    {
      heading: "Build With the AI You Already Use",
      body: "Kite's own AI works in the editor without any setup. Claude, ChatGPT and coding agents like Claude Code, Codex and Cursor connect through Kite's MCP server and use their own model.",
      points: [
        {
          title: "Ask AI in the Editor",
          description:
            "Open a command, click on Ask AI and describe what it should do. The AI adds and connects the blocks, and you check the changes before saving.",
          href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/flow-ai`,
        },
        {
          title: "Claude",
          description:
            "Add Kite as a connector, even on Claude's free plan, and ask for commands, events and messages in a chat.",
          href: "/guides/make-a-discord-bot-with-claude",
        },
        {
          title: "ChatGPT",
          description:
            "Add Kite as a custom MCP server and describe your bot instead of copying code from the chat.",
          href: "/guides/make-a-discord-bot-with-chatgpt",
        },
      ],
    },
    {
      heading: "Kite Runs What the AI Builds",
      body: "Many AI bot makers write code that you then have to host. Here the AI builds flows of blocks, and Kite runs them on its own servers.",
      points: [
        {
          title: "Hosted 24/7",
          description:
            "Your bot runs on Kite's servers for free, with your own name and avatar. Nothing to deploy or keep online.",
        },
        {
          title: "Checked Before Saving",
          description:
            "Kite checks every flow the AI builds, like missing settings or blocks in the wrong place, before it goes live.",
        },
        {
          title: "Change It Yourself",
          description:
            "Open what the AI built in the editor and change a message or add a block, without reading code.",
        },
        {
          title: "Undo Any Change",
          description:
            "Every change is kept in the history of the command, so you can restore an earlier version.",
        },
        {
          title: "Fix Errors From the Logs",
          description:
            "When a command fails, the logs show which block failed and why. Claude and ChatGPT can read them and fix it.",
        },
        openSourcePoint,
      ],
    },
    {
      heading: "Want Your Bot to Use AI Too?",
      body: "Your bot can use AI while it runs, too, like answering questions in your server.",
      points: [
        {
          title: "AI Discord Bot",
          description:
            "Add an /ask command and replies when someone mentions the bot, with your own personality and no API key.",
          href: "/ai-discord-bot",
        },
        {
          title: "AI Decisions",
          description:
            "Let AI decide what your bot does, like whether a message is spam or which team a question goes to, for 1 credit each.",
        },
        {
          title: "Templates",
          description:
            "Start from verification, modmail or giveaways in the Marketplace and ask the AI to change them for your server.",
          href: "/marketplace",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "Can AI make a Discord bot for me?",
      answer:
        "Yes. Describe what the bot should do and Kite's AI builds the command in the editor. Claude and ChatGPT can build whole bots with several commands and events when you connect them to Kite.",
    },
    {
      question: "Is the AI Discord bot maker free?",
      answer:
        "Yes. Kite hosts your bot for free. The free plan includes a few Ask AI prompts per bot each month, and connecting Claude or ChatGPT doesn't use them, because they use their own model.",
    },
    {
      question: "Do I need to know how to code?",
      answer:
        "No. The AI builds flows of blocks like Ban member or Send message, and you can read and change them in the editor without code.",
    },
    {
      question: "What's the difference between Ask AI and Claude or ChatGPT?",
      answer:
        "Ask AI changes the flow you have open in the editor. Claude and ChatGPT connect through Kite's MCP server and can create several commands, events and messages in one go, look at your logs and add templates.",
    },
    freeFaq,
  ],
};

export default function AIDiscordBotMakerPage() {
  return <LandingPage content={content} />;
}
