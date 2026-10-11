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
  title: "Free BotGhost Alternative",
  description:
    "Looking for a BotGhost alternative? Kite is a free, open source Discord bot maker with custom status, up to 100 servers and no branding on the free plan.",
  heading: "A Free BotGhost Alternative",
  intro:
    "Kite and BotGhost both let you make a Discord bot without code. Kite's free plan has higher limits, lets you set your own status and doesn't add any branding to your bot. Kite is also open source.",
  showcase: {
    heading: "Build Commands by Connecting Blocks",
    body: "This is the /ban command from the moderation template. Each block is one step, and the lines between them set the order.",
    flow: banFlow,
    trigger: "Alex used /ban",
    reply: "Banned the user.",
  },
  sections: [
    {
      heading: "Why Use Kite Instead of BotGhost",
      points: [
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
    },
    {
      heading: "When BotGhost Is the Better Choice",
      body: "BotGhost has many ready-made modules, like economy, giveaways and YouTube or Twitch notifications. Kite has fewer of these. You can build many of them with blocks, but it takes more time. If you want a lot of features without building them, BotGhost is faster to set up.",
    },
    switchSection("BotGhost"),
  ],
  comparison: {
    competitor: "BotGhost",
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
  },
  faqs: [freeFaq, noCodeFaq],
};

export default function BotGhostAlternativePage() {
  return <LandingPage content={content} />;
}
