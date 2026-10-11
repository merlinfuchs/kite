import LandingPage, {
  aiBuilderPoint,
  freeFaq,
  noCodeFaq,
  openSourcePoint,
  switchSection,
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { banFlow } from "@/components/landing/flows";

const content: LandingPageContent = {
  title: "Free inventor.bot Alternative",
  description:
    "Looking for an inventor.bot alternative? Kite is a free, open source Discord bot maker. Make up to 10 bots for free and build them in a visual editor.",
  heading: "A Free inventor.bot Alternative",
  intro:
    "Kite and inventor.bot (formerly inventor.gg) both let you make a Discord bot with a visual editor. With Kite you can make up to 10 bots for free, and Kite is open source.",
  showcase: {
    heading: "Build Commands by Connecting Blocks",
    body: "This is the /ban command from the moderation template. Each block is one step, and the lines between them set the order.",
    flow: banFlow,
    trigger: "Alex used /ban",
    reply: "Banned the user.",
  },
  sections: [
    {
      heading: "Why Use Kite Instead of inventor.bot",
      points: [
        {
          title: "Up to 10 Free Bots",
          description:
            "inventor.bot includes 2 free bots and has paid plans for more. With Kite you can make up to 10 bots for free.",
        },
        openSourcePoint,
        aiBuilderPoint,
        {
          title: "Templates and Plugins",
          description:
            "Import templates for moderation, welcome messages and an AI chatbot, or turn on the Counting and Starboard plugins.",
        },
      ],
    },
    {
      heading: "When inventor.bot Is the Better Choice",
      body: "inventor.bot lets you build and fix your bot with an AI agent of your choice, and it has a built-in database. If you already work with an AI agent and want it to edit your bot, inventor.bot may suit you better.",
    },
    switchSection("inventor.bot"),
  ],
  comparison: {
    competitor: "inventor.bot",
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
  },
  faqs: [freeFaq, noCodeFaq],
};

export default function InventorAlternativePage() {
  return <LandingPage content={content} />;
}
