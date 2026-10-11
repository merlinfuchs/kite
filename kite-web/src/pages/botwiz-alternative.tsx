import LandingPage, {
  aiBuilderPoint,
  freeFaq,
  openSourcePoint,
  switchSection,
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { banFlow } from "@/components/landing/flows";

const content: LandingPageContent = {
  title: "Free BotWiz Alternative",
  description:
    "Looking for a BotWiz alternative? Kite is a free, open source Discord bot maker with a visual editor and 24/7 hosting. No trial and no credit card.",
  heading: "A Free BotWiz Alternative",
  intro:
    "Kite and BotWiz both let you build a Discord bot with blocks instead of code. BotWiz has a free trial and then needs a subscription. Kite has a free plan that doesn't run out, and Kite is open source.",
  showcase: {
    heading: "Build Commands by Connecting Blocks",
    body: "This is the /ban command from the moderation template. Each block is one step, and the lines between them set the order.",
    flow: banFlow,
    trigger: "Alex used /ban",
    reply: "Banned the user.",
  },
  sections: [
    {
      heading: "Why Use Kite Instead of BotWiz",
      points: [
        {
          title: "Free, Not a Trial",
          description:
            "BotWiz gives your first bot 7 days for free. After that your bot goes offline until you pay. Kite's free plan keeps your bot online with no time limit.",
        },
        {
          title: "Up to 10 Free Bots",
          description:
            "You can make up to 10 bots on Kite without paying. Each one gets 24/7 hosting.",
        },
        aiBuilderPoint,
        openSourcePoint,
      ],
    },
    {
      heading: "When BotWiz Is the Better Choice",
      body: "BotWiz can also host bots you wrote yourself in Node.js or Python. Kite only runs bots built in Kite. If you have a discord.js or discord.py bot that you want to keep, BotWiz Hosting can run it.",
    },
    switchSection("BotWiz"),
  ],
  comparison: {
    competitor: "BotWiz",
    rows: [
      {
        feature: "Free plan",
        kite: "Yes, no time limit",
        competitor: "7-day trial",
      },
      {
        feature: "Visual editor",
        kite: "Yes",
        competitor: "Yes",
      },
      {
        feature: "Commands per bot",
        kite: "25 free, more with Premium",
        competitor: "100 on paid plans",
      },
      {
        feature: "Servers per bot",
        kite: "100 free, more with Premium",
        competitor: "Unlimited on paid plans",
      },
      {
        feature: "Host your own code",
        kite: "No",
        competitor: "Node.js and Python",
      },
    ],
  },
  faqs: [
    freeFaq,
    {
      question: "Can Kite host my discord.js or discord.py bot?",
      answer: "No. Kite only runs bots built with its visual editor.",
    },
  ],
};

export default function BotWizAlternativePage() {
  return <LandingPage content={content} />;
}
