import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { noCodeFaq } from "@/components/landing/copy";
import { erlcStatusFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Make a Free ER:LC Discord Bot",
  description:
    "Make your own ER:LC Discord bot for free with Kite. Show your private server's status and players, run in-game commands from Discord and log game events, without writing code.",
  heading: "Make Your Own ER:LC Discord Bot",
  intro:
    "Connect your Emergency Response: Liberty County private server to a Discord bot with your own name and avatar. Start with the ER:LC template, then build the commands your community needs. You don't need to write code.",
  secondaryAction: {
    label: "Read the Guide",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/erlc`,
  },
  showcase: {
    flow: erlcStatusFlow,
    command: "status",
    reply: "32 players are in the server. Join with libertyrp!",
  },
  sections: [
    {
      heading: "Start With the ER:LC Template",
      body: "Enable ER:LC with your server key and import the template. Your bot gets three slash commands.",
      points: [
        {
          title: "/status",
          description:
            "Show the server's name, how many players are in it and its join code.",
        },
        {
          title: "/players",
          description: "List everyone in the server with their team.",
        },
        {
          title: "/announce",
          description:
            "Show a message to everyone in the game. Only staff with Manage Server can use it.",
        },
      ],
    },
    {
      heading: "Build the Bot Your Server Needs",
      points: [
        {
          title: "Staff Commands",
          description:
            "Kick players, send hints or run any other in-game command from Discord, only for your staff.",
        },
        {
          title: "Live Server Info",
          description:
            "Count the police online, show the queue or list the staff in the server.",
        },
        {
          title: "Logs in Discord",
          description:
            "Post join logs, kill logs, command logs and mod calls to your staff channels.",
        },
        {
          title: "In-Game Commands",
          description:
            "React to emergency calls and to chat messages starting with ; in the game, through ER:LC's event webhooks.",
        },
        {
          title: "Your Own Branding",
          description:
            "Give the bot your server's name and avatar, and send replies as embeds in your colors.",
        },
        {
          title: "Let AI Build It",
          description:
            "Describe the command you want, like a /police command that lists the officers on duty, and Kite's AI builds it.",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "Is the ER:LC bot free?",
      answer:
        "Yes. You can build and host your ER:LC bot on Kite for free. Getting the server's status or running a command costs 1 of your 10,000 free credits per month.",
    },
    {
      question: "What do I need to connect my server?",
      answer:
        "The server key from your private server's settings. You enter it under Integrations in your app, where Kite stores it encrypted and only sends it to ER:LC. To run commands, the server owner also authorizes Kite once on ER:LC's page.",
    },
    {
      question: "Can I use it next to other ER:LC bots?",
      answer:
        "Yes. Your Kite bot is its own Discord bot, so it runs next to any other bot in your server. Use it for the commands the others don't have.",
    },
    {
      question: "Is Kite made by ER:LC?",
      answer:
        "No. Kite is a third-party app that uses the ER:LC API and isn't affiliated with ER:LC or Police Roleplay Community. For help with your Kite bot, ask in Kite's Discord server.",
    },
    noCodeFaq,
  ],
};

export default function ERLCDiscordBotPage() {
  return <LandingPage content={content} />;
}
