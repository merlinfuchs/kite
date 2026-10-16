import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { openSourcePoint } from "@/components/landing/copy";
import { helloFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Free 24/7 Discord Bot Hosting",
  description:
    "Host your Discord bot 24/7 for free with Kite. Build the bot in a visual editor and Kite runs it for you. No server, no code and no credit card needed.",
  heading: "Free 24/7 Discord Bot Hosting",
  intro:
    "Kite runs your Discord bot on its servers 24/7, for free. You build the bot in Kite's editor and add your bot token. You don't need a VPS or a computer that stays on.",
  secondaryAction: {
    label: "Read the Guide",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/getting-started`,
  },
  showcase: {
    flow: helloFlow,
    command: "hello",
    reply: "Hello @Alex!",
  },
  sections: [
    {
      heading: "What the Free Plan Includes",
      points: [
        {
          title: "24/7 Uptime",
          description:
            "Your bot stays connected to Discord at all times. If the connection drops, Kite reconnects it for you.",
        },
        {
          title: "10,000 Credits per Month",
          description:
            "Most actions, like sending a message, cost 1 credit. Conditions and loops are free. This is enough for most servers.",
        },
        {
          title: "Up to 100 Servers",
          description:
            "You can add your bot to up to 100 servers. Premium plans raise this limit if you need more.",
        },
        {
          title: "Your Own Name and Avatar",
          description:
            "Your bot is your own Discord application. It shows your name, avatar and status, not Kite's.",
        },
        {
          title: "No Credit Card",
          description:
            "You sign in with your Discord account. There is no trial that runs out.",
        },
        openSourcePoint,
      ],
    },
    {
      heading: "Can Kite Host My discord.js or discord.py Bot?",
      body: "No. Kite only runs bots that are built in Kite. If your bot is written in JavaScript or Python, you need a server or a hosting service that runs your code. If your bot mostly responds to commands and events, building it again in Kite is often faster than keeping a server running.",
    },
    {
      heading: "How It Works",
      points: [
        {
          title: "1. Create a Discord App",
          description:
            "Create an application in the Discord Developer Portal and copy its bot token.",
        },
        {
          title: "2. Add It to Kite",
          description:
            "Sign in to Kite with Discord, click on Create app and paste the token.",
        },
        {
          title: "3. Build Your Commands",
          description:
            "Create commands and event listeners in the editor. They go live as soon as you save them.",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "Is the hosting really free?",
      answer:
        "Yes. Every bot gets free 24/7 hosting with 10,000 credits per month. Premium plans give you more credits and higher limits, but you don't need one to keep your bot online.",
    },
    {
      question: "Will my bot stay online when I close my computer?",
      answer:
        "Yes. Your bot runs on Kite's servers, not on your computer. It stays online when your computer is off.",
    },
    {
      question: "Can I host a bot that I wrote in code?",
      answer:
        "No. Kite runs bots built with its visual editor. It can't run discord.js, discord.py or other code.",
    },
    {
      question: "What happens when my bot runs out of credits?",
      answer:
        "Kite stops the bot until the next month starts or until you have more credits. Then you start it again in the settings of your app.",
    },
  ],
};

export default function DiscordBotHostingPage() {
  return <LandingPage content={content} />;
}
