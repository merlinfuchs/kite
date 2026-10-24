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
  title: "Free SCNX Alternative",
  description:
    "Looking for an SCNX alternative? Kite is a free, open source Discord bot maker. One bot can join up to 100 servers, with no ad check-ins and no branding.",
  competitor: "SCNX",
  lastChecked: "October 2026",
  heading: "A Free SCNX Alternative",
  intro:
    "Kite and SCNX both give you your own Discord bot without code. SCNX is built around ready-made modules, while in Kite you build your own commands from blocks. On Kite's free plan, one bot can join up to 100 servers and you never need to watch an ad to keep it online.",
  chooseKite: [
    "Your bot should work on more than one server",
    "You need more than 10 custom commands without paying",
    "You don't want to watch an ad every week to keep your bot online",
    "You want no branding on your bot without paying",
  ],
  chooseCompetitor: [
    "You want dozens of ready-made modules like leveling, economy and social media notifications",
    "You'd rather turn on modules than build commands",
    "You want a ticket and modmail support bot and are fine with paying",
  ],
  rows: [
    {
      feature: "Servers per bot",
      kite: "100 free, more with Premium",
      competitor: "1, the server it was made for",
    },
    {
      feature: "Custom commands (free)",
      kite: "25",
      competitor: "10 triggers",
    },
    {
      feature: "Keeping your bot online (free)",
      kite: "Always online",
      competitor: "Watch an ad every 7 days",
    },
    {
      feature: "Branding on your bot (free)",
      kite: "None",
      competitor: '"Powered by scnx.xyz" in embeds',
    },
    {
      feature: "Custom status (free)",
      kite: "Yes",
      competitor: "Yes",
    },
    {
      feature: "Ready-made modules",
      kite: "A few templates",
      competitor: "Many",
    },
    {
      feature: "Open source",
      kite: "Yes",
      competitor: "Bot core is source-available",
    },
  ],
  reasons: [
    {
      title: "One Bot for Many Servers",
      description:
        "An SCNX bot only works on the server it was made for. A Kite bot can join up to 100 servers on the free plan.",
    },
    {
      title: "No Ad Check-Ins",
      description:
        "To keep a free SCNX bot online, you watch a short ad every 7 days. Kite keeps your bot online without any check-ins.",
    },
    {
      title: "No Branding",
      description:
        "SCNX adds \"Powered by scnx.xyz\" to your bot's embeds, and removing all SCNX branding needs the Professional plan. Kite doesn't add any branding to your bot.",
    },
    aiBuilderPoint,
    openSourcePoint,
  ],
  faqs: [freeFaq, noCodeFaq],
};

export default function ScnxAlternativePage() {
  return <ComparisonPage content={content} />;
}
