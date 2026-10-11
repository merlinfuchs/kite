import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";

const content: LandingPageContent = {
  title: "Free inventor.bot Alternative",
  description:
    "Looking for an inventor.bot alternative? Kite is a free, open source Discord bot maker. Make up to 10 bots for free and build them in a visual editor.",
  heading: "A Free inventor.bot Alternative",
  intro:
    "Kite and inventor.bot (formerly inventor.gg) both let you make a Discord bot with a visual editor. With Kite you can make up to 10 bots for free, and Kite is open source.",
  sections: [
    {
      heading: "Why Use Kite Instead of inventor.bot",
      points: [
        {
          title: "Up to 10 Free Bots",
          description:
            "inventor.bot includes 2 free bots and has paid plans for more. With Kite you can make up to 10 bots for free.",
        },
        {
          title: "Open Source",
          description:
            "The code of Kite is on GitHub. You can check how your bot is run, or host Kite yourself.",
        },
        {
          title: "Build With AI",
          description:
            "Describe what a command should do and the flow AI adds and connects the blocks for you. It checks the flow after every change.",
        },
        {
          title: "Templates and Plugins",
          description:
            "Import templates for moderation, welcome messages and an AI chatbot, or turn on the Counting and Starboard plugins.",
        },
      ],
    },
    {
      heading: "When inventor.bot Is the Better Choice",
      body: "inventor.bot lets you build and fix your bot with an outside AI agent, and it has a built-in database. If you already work with an AI agent and want it to edit your bot, inventor.bot may suit you better.",
    },
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
        kite: "Built-in flow AI",
        competitor: "Your own AI agent",
      },
    ],
  },
  faqs: [
    {
      question: "Can I move my bot from inventor.bot to Kite?",
      answer:
        "Yes. Add the token of your existing Discord application to Kite and build your commands again. Your bot keeps its name, avatar and servers. Turn it off in inventor.bot first, so the two don't run at the same time.",
    },
    {
      question: "Is Kite really free?",
      answer:
        "Yes. Your bots are hosted 24/7 for free. Premium plans give you higher limits, like more servers and commands, but you don't need one to run a bot.",
    },
    {
      question: "Do I need to know how to code?",
      answer:
        "No. You build commands by connecting blocks in a visual editor. You can also ask the flow AI to build them for you.",
    },
  ],
};

export default function InventorAlternativePage() {
  return <LandingPage content={content} />;
}
