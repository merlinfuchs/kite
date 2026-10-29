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
  title: "Free BotGhost Alternative",
  description:
    "Looking for a BotGhost alternative? Kite is a free, open source Discord bot maker with a free custom status, up to 100 servers and no inactivity rule.",
  competitor: "BotGhost",
  lastChecked: "October 2026",
  heading: "A Free BotGhost Alternative",
  intro:
    "Kite and BotGhost both let you make a Discord bot without code. Kite's free plan has higher limits, lets you replace Kite's status with your own for free and keeps your bot online even when nobody uses it. Kite is also open source.",
  chooseKite: [
    "You need more than 5 servers or 3 commands without paying",
    "You want your own status without paying to remove branding",
    "Your bot should stay online even when nobody uses it for a few days",
    "You want a tool that is open source",
  ],
  chooseCompetitor: [
    "You want economy, giveaways or YouTube and Twitch notifications ready to use",
    "You'd rather turn on modules than build commands",
  ],
  rows: [
    {
      feature: "Servers per bot (free)",
      kite: "100",
      competitor: "5",
    },
    {
      feature: "Custom commands (free)",
      kite: "25",
      competitor: "3",
    },
    {
      feature: "Event listeners (free)",
      kite: "10",
      competitor: "3",
    },
    {
      feature: "Custom status (free)",
      kite: "Yes",
      competitor: 'No, shows "Made with BotGhost"',
    },
    {
      feature: "Branding on your bot (free)",
      kite: '"Powered by Kite.onl" status until you set your own',
      competitor: '"Made with BotGhost" status, removing it needs Premium',
    },
    {
      feature: "Stays online when unused (free)",
      kite: "Yes",
      competitor: "Needs a command used every 48 hours",
    },
    {
      feature: "Ready-made modules",
      kite: "A few templates",
      competitor: "Many",
    },
    {
      feature: "Open source",
      kite: "Yes",
      competitor: "No",
    },
  ],
  reasons: [
    {
      title: "Your Own Status for Free",
      description:
        'On BotGhost\'s free plan, your bot\'s status says "Made with BotGhost" and changing it needs Premium. Kite shows "Powered by Kite.onl" until you set your own status, and that\'s free.',
    },
    {
      title: "Higher Free Limits",
      description:
        "BotGhost's free plan allows 5 servers, 3 custom commands and 3 custom events per bot. Kite's free plan allows 100 servers, 25 commands and 10 event listeners.",
    },
    {
      title: "No Inactivity Rule",
      description:
        "Free BotGhost bots only stay online if a command is used at least every 48 hours. Kite keeps your bot online whether it's used or not.",
    },
    aiBuilderPoint,
    openSourcePoint,
  ],
  faqs: [freeFaq, noCodeFaq],
};

export default function BotGhostAlternativePage() {
  return <ComparisonPage content={content} />;
}
