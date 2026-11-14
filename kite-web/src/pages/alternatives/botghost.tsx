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
  title: "Free BotGhost alternative",
  description:
    "Looking for a BotGhost alternative? Kite is a free, open source Discord bot maker with up to 100 servers, 15 commands, and up to 10 bots online on the free plan, and no inactivity rule.",
  competitor: "BotGhost",
  lastChecked: "November 2026",
  heading: "A free BotGhost alternative",
  intro:
    "Kite and BotGhost both let you make a Discord bot by connecting blocks. Kite's free plan has higher limits, keeps several bots online, and lets you replace Kite's status with your own. Kite is also open source.",
  chooseKite: [
    "You need more than 5 servers or 3 commands without paying",
    "You want more than one bot online on the free plan",
    "You want your own status without paying to remove branding",
    "Your bot should stay online even when nobody uses it for a few days",
    "You want to change how a ready-made feature works without paying",
    "You want a cheaper paid plan",
    "You want a tool that is open source",
  ],
  chooseCompetitor: [
    "You want tickets, giveaways, reaction roles, or other modules ready to switch on, and their settings are enough",
    "You'd rather turn on modules than build commands",
    "You want YouTube, Twitch, or Reddit notifications and are fine paying for them",
  ],
  importNote: {
    text: "Already built commands in BotGhost? Kite can import some of your custom commands and custom events, so you don't have to build them all again.",
    guide: "botghost-import",
  },
  rows: [
    {
      feature: "Servers per bot (free)",
      kite: "100",
      competitor: "5",
    },
    {
      feature: "Custom commands (free)",
      kite: "15",
      competitor: "3",
    },
    {
      feature: "Event listeners (free)",
      kite: "10",
      competitor: "3",
    },
    {
      feature: "Bots online at once (free)",
      kite: "Up to 10",
      competitor: "1",
    },
    {
      feature: "Custom status (free)",
      kite: "Yes",
      competitor: 'No, shows "Made with BotGhost"',
    },
    {
      feature: "Branding on your bot (free)",
      kite: 'Default "Powered by Kite.onl" status, free to change',
      competitor: '"Made with BotGhost" status, removing it needs Premium',
    },
    {
      feature: "Stays online when unused (free)",
      kite: "Yes",
      competitor: "Switched off after 2 days without activity",
    },
    {
      feature: "Paid plan for one bot",
      kite: "From $2.99 a month",
      competitor: "$12.50 a month, or $75 a year",
    },
    {
      feature: "Ready-made features",
      kite: "100+ templates, including leveling and economy",
      competitor:
        "35+ modules and a marketplace. 11 of them are free and don't count toward the limits.",
    },
    {
      feature: "Change a ready-made feature's blocks (free)",
      kite: "Yes, a template's commands are yours",
      competitor:
        "No. You can look at a module's blocks, but changing them needs Premium.",
    },
    {
      feature: "Open source",
      kite: "Yes",
      competitor: "No",
    },
  ],
  reasons: [
    {
      title: "Your own status for free",
      description:
        'On BotGhost\'s free plan, your bot\'s status says "Made with BotGhost" and changing it needs Premium. Kite shows "Powered by Kite.onl" until you set your own status, and that\'s free.',
    },
    {
      title: "Higher free limits",
      description:
        "BotGhost's free plan allows 5 servers, 3 custom commands, and 3 custom events per bot, with one bot online at a time. Kite's free plan allows 100 servers, 15 commands, and 10 event listeners per bot, and up to 10 bots online.",
    },
    {
      title: "Change any template for free",
      description:
        "On BotGhost's free plan you can look at a module's blocks, but changing them needs Premium. A Kite template adds normal commands to your bot, and you can change every block for free. Those commands count toward your 15 commands and 10 event listeners.",
    },
    {
      title: "No inactivity rule",
      description:
        "BotGhost switches a free bot off after 2 days without activity, until you start it again. Kite keeps your bot online whether it's used or not.",
    },
    {
      title: "Cheaper when you need more",
      description:
        "BotGhost Premium costs $12.50 a month for one bot, or $75 a year. Kite Plus costs $2.99 a month and Kite Premium $5.99 a month, each for one bot.",
    },
    aiBuilderPoint,
    openSourcePoint,
  ],
  faqs: [
    freeFaq,
    noCodeFaq,
    {
      question: "Can I move my BotGhost commands to Kite?",
      answer:
        "Some of them. Kite's import reads the custom commands and custom events of your BotGhost bot and adds them to your app, where you can change them like anything you built yourself. Custom variables, forms, loops, music, and tickets have to be rebuilt, and the import copies what your bot does, not the bot itself.",
    },
  ],
};

export default function BotGhostAlternativePage() {
  return <ComparisonPage content={content} />;
}
