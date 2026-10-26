import ComparisonPage, {
  ComparisonPageContent,
} from "@/components/landing/ComparisonPage";
import {
  aiBuilderPoint,
  freeFaq,
  noCodeFaq,
  openSourcePoint,
} from "@/components/landing/copy";

const content: ComparisonPageContent = {
  title: "Free FromFlow Alternative",
  description:
    "Looking for a FromFlow alternative? Kite is a free, open source Discord bot maker with a visual editor, AI help and 24/7 hosting. No trial and no credit card.",
  competitor: "FromFlow",
  lastChecked: "October 2026",
  heading: "A Free FromFlow Alternative",
  intro:
    "Kite and FromFlow both let you build a Discord bot in a visual editor, with an AI assistant that can draft commands for you. FromFlow has a 3-day trial and then costs $4.99 a month. Kite has a free plan that doesn't run out, and Kite is open source.",
  chooseKite: [
    "You want a free plan with no time limit",
    "You don't want to add a card to get started",
    "You want a tool that is open source",
  ],
  chooseCompetitor: [
    "You want ready-made apps like tickets, leveling, giveaways and music",
    "You'd rather pay once for lifetime access",
    "You need more than 25 commands per bot and are fine with paying",
  ],
  rows: [
    {
      feature: "Free plan",
      kite: "Yes, no time limit",
      competitor: "3-day trial",
    },
    {
      feature: "Card needed to start",
      kite: "No",
      competitor: "Yes, $1 card check",
    },
    {
      feature: "Visual editor",
      kite: "Yes",
      competitor: "Yes",
    },
    {
      feature: "Build with AI",
      kite: "Yes",
      competitor: "Yes",
    },
    {
      feature: "Commands per bot",
      kite: "25 free, more with Premium",
      competitor: "Unlimited on Pro",
    },
    {
      feature: "Ready-made features",
      kite: "A few templates",
      competitor: "Apps for tickets, leveling, music and more",
    },
  ],
  reasons: [
    {
      title: "Free, Not a Trial",
      description:
        "FromFlow's trial lasts 3 days and starts with a $1 card check. After that, Pro costs $4.99 a month. Kite's free plan keeps your bot online with no time limit.",
    },
    {
      title: "Up to 10 Free Bots",
      description:
        "You can make up to 10 bots on Kite without paying. Each one gets 24/7 hosting.",
    },
    aiBuilderPoint,
    openSourcePoint,
  ],
  faqs: [freeFaq, noCodeFaq],
};

export default function FromFlowAlternativePage() {
  return <ComparisonPage content={content} />;
}
