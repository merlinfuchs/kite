import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";

const content: LandingPageContent = {
  title: "Free BotGhost Alternative",
  description:
    "Looking for a BotGhost alternative? Kite is a free, open source Discord bot maker with custom status, up to 100 servers and no branding on the free plan.",
  heading: "A Free BotGhost Alternative",
  intro:
    "Kite and BotGhost both let you make a Discord bot without code. Kite's free plan has higher limits, lets you set your own status and doesn't add any branding to your bot. Kite is also open source.",
  sections: [
    {
      heading: "Why Use Kite Instead of BotGhost",
      points: [
        {
          title: "Your Own Status for Free",
          description:
            "On the free plan of BotGhost, the status of your bot says Made with BotGhost. Changing it needs Premium. With Kite you set your own status for free and there is no branding.",
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
        {
          title: "Build With AI",
          description:
            "Describe what a command should do and the flow AI adds and connects the blocks for you.",
        },
        {
          title: "Open Source",
          description:
            "The code of Kite is on GitHub. You can check how your bot is run, or host Kite yourself.",
        },
      ],
    },
    {
      heading: "When BotGhost Is the Better Choice",
      body: "BotGhost has many ready-made modules, like economy, giveaways and YouTube or Twitch notifications. Kite has fewer of these. You can build many of them with flows, but it takes more time. If you want a lot of features without building them, BotGhost is faster to set up.",
    },
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
        feature: "Custom status (free)",
        kite: "Yes",
        competitor: "No, shows Made with BotGhost",
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
    ],
  },
  faqs: [
    {
      question: "Can I move my bot from BotGhost to Kite?",
      answer:
        "Yes. Add the token of your existing Discord application to Kite and build your commands again. Your bot keeps its name, avatar and servers. Turn it off in BotGhost first, so the two don't run at the same time.",
    },
    {
      question: "Is Kite really free?",
      answer:
        "Yes. Your bot is hosted 24/7 for free. Premium plans give you higher limits, like more servers and commands, but you don't need one to run a bot.",
    },
    {
      question: "Do I need to know how to code?",
      answer:
        "No. You build commands by connecting blocks in a visual editor. You can also ask the flow AI to build them for you.",
    },
  ],
};

export default function BotGhostAlternativePage() {
  return <LandingPage content={content} />;
}
