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
  title: "Free inventor.bot Alternative",
  description:
    "Looking for an inventor.bot alternative? Kite is a free, open source Discord bot maker. Make up to 10 bots for free and build them in a visual editor.",
  competitor: "inventor.bot",
  heading: "A Free inventor.bot Alternative",
  intro:
    "Kite and inventor.bot (formerly inventor.gg) both let you make a Discord bot with a visual editor. With Kite you can make up to 10 bots for free, and Kite is open source.",
  chooseKite: [
    "You want more than 2 bots for free",
    "You want AI built into the editor",
    "You want a tool that is open source",
  ],
  chooseCompetitor: [
    "You want to build and fix your bot with an AI agent of your choice",
    "You want a built-in database",
  ],
  rows: [
    {
      feature: "Free bots",
      kite: "10",
      competitor: "2",
    },
    {
      feature: "Visual editor",
      kite: "Yes",
      competitor: "Yes",
    },
    {
      feature: "Build with AI",
      kite: "Built into the editor",
      competitor: "With your own AI agent",
    },
    {
      feature: "Store data",
      kite: "Stored variables",
      competitor: "Built-in database",
    },
    {
      feature: "Logs for debugging",
      kite: "Yes",
      competitor: "Yes",
    },
    {
      feature: "Ready-made features",
      kite: "Templates and plugins",
      competitor: "Plugins",
    },
  ],
  reasons: [
    {
      title: "Up to 10 Free Bots",
      description:
        "inventor.bot includes 2 free bots and has paid plans for more. With Kite you can make up to 10 bots for free.",
    },
    aiBuilderPoint,
    {
      title: "Templates and Plugins",
      description:
        "Import templates for moderation, welcome messages and an AI chatbot, or turn on the Counting and Starboard plugins.",
    },
    openSourcePoint,
  ],
  faqs: [freeFaq, noCodeFaq],
};

export default function InventorAlternativePage() {
  return <ComparisonPage content={content} />;
}
