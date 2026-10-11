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
    "Looking for a BotGhost alternative? Kite is a free, open source Discord bot maker with custom status, up to 100 servers and no branding on the free plan.",
  competitor: "BotGhost",
  heading: "A Free BotGhost Alternative",
  intro:
    "Kite and BotGhost both let you make a Discord bot without code. Kite's free plan has higher limits, lets you set your own status and doesn't add any branding to your bot. Kite is also open source.",
  chooseKite: [
    "You need more than 5 servers or 8 commands without paying",
    "You want your own status and no branding on your bot",
    "Your bot should stay online even when nobody uses it for a week",
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
      competitor: "8",
    },
    {
      feature: "Event listeners (free)",
      kite: "5",
      competitor: "3",
    },
    {
      feature: "Custom status (free)",
      kite: "Yes",
      competitor: 'No, shows "Made with BotGhost"',
    },
    {
      feature: "Branding on your bot (free)",
      kite: "None",
      competitor: '"Made with BotGhost" status',
    },
    {
      feature: "Stays online when unused (free)",
      kite: "Yes",
      competitor: "Needs an interaction every 7 days",
    },
    {
      feature: "Ready-made modules",
      kite: "A few templates and plugins",
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
        "On BotGhost's free plan, your bot's status says \"Made with BotGhost\" and changing it needs Premium. With Kite you set your own status for free and there is no branding.",
    },
    {
      title: "Higher Free Limits",
      description:
        "BotGhost's free plan allows 5 servers, 8 custom commands and 3 custom events per bot. Kite's free plan allows 100 servers, 25 commands and 5 event listeners.",
    },
    {
      title: "No Inactivity Rule",
      description:
        "Free BotGhost bots only stay online if they get an interaction at least every 7 days. Kite keeps your bot online whether it's used or not.",
    },
    aiBuilderPoint,
    openSourcePoint,
  ],
  faqs: [freeFaq, noCodeFaq],
};

export default function BotGhostAlternativePage() {
  return <ComparisonPage content={content} />;
}
