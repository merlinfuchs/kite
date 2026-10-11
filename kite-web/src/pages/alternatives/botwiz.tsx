import ComparisonPage, {
  ComparisonPageContent,
} from "@/components/landing/ComparisonPage";
import {
  aiBuilderPoint,
  freeFaq,
  openSourcePoint,
} from "@/components/landing/copy";

const content: ComparisonPageContent = {
  title: "Free BotWiz Alternative",
  description:
    "Looking for a BotWiz alternative? Kite is a free, open source Discord bot maker with a visual editor and 24/7 hosting. No trial and no credit card.",
  competitor: "BotWiz",
  heading: "A Free BotWiz Alternative",
  intro:
    "Kite and BotWiz both let you build a Discord bot with blocks instead of code. BotWiz has a free trial and then needs a subscription. Kite has a free plan that doesn't run out, and Kite is open source.",
  chooseKite: [
    "You want a free plan with no time limit",
    "You want up to 10 bots without paying",
    "You want a tool that is open source",
  ],
  chooseCompetitor: [
    "You want to host a bot you wrote in Node.js or Python",
    "You need more than 25 commands per bot and are fine with paying",
  ],
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
  reasons: [
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
  faqs: [
    freeFaq,
    {
      question: "Can Kite host my discord.js or discord.py bot?",
      answer: "No. Kite only runs bots built with its visual editor.",
    },
  ],
};

export default function BotWizAlternativePage() {
  return <ComparisonPage content={content} />;
}
