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
    "Looking for an inventor.bot alternative? Kite is a free, open source Discord bot maker. Make up to 10 bots for free, each in up to 100 servers, with no inactivity rule.",
  competitor: "inventor.bot",
  lastChecked: "October 2026",
  heading: "A Free inventor.bot Alternative",
  intro:
    "Kite and inventor.bot (formerly inventor.gg) both let you make a Discord bot with a visual editor. Kite's free plan has higher limits, no ad in your bot's status and keeps your bot online even on quiet days. Kite is also open source.",
  chooseKite: [
    "You want more than 2 bots for free",
    "Your bot should be in more than 15 servers without paying",
    "You want your own status without paying",
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
      feature: "Servers per bot (free)",
      kite: "100",
      competitor: "15",
    },
    {
      feature: "Commands and events (free)",
      kite: "25 commands and 10 event listeners",
      competitor: "15 combined",
    },
    {
      feature: "Branding on your bot (free)",
      kite: 'Default "Powered by Kite.onl" status, free to change',
      competitor: "invt.bot ad in the status, removing it needs Premium",
    },
    {
      feature: "Stays online when unused (free)",
      kite: "Yes",
      competitor: "Up to 14 days without an event",
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
      kite: "50+ templates in the marketplace",
      competitor: "Plugins",
    },
  ],
  reasons: [
    {
      title: "Up to 10 Free Bots",
      description:
        "inventor.bot includes 2 free bots and has paid plans for more. With Kite you can make up to 10 bots for free.",
    },
    {
      title: "Higher Free Limits",
      description:
        "inventor.bot's free plan allows 15 servers per bot and 15 commands and events combined. Kite's free plan allows 100 servers, 25 commands and 10 event listeners.",
    },
    {
      title: "No Ad in Your Status",
      description:
        "Free inventor.bot bots show an invt.bot ad as their status, and free bots go offline after 14 days without an event. Kite lets you set your own status for free and keeps your bot online even on quiet days.",
    },
    aiBuilderPoint,
    {
      title: "Templates You Can Change",
      description:
        "Add templates for moderation, welcome messages, counting, a starboard or an AI chatbot, and change every block of them.",
    },
    openSourcePoint,
  ],
  faqs: [freeFaq, noCodeFaq],
};

export default function InventorAlternativePage() {
  return <ComparisonPage content={content} />;
}
