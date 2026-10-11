import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";

const content: LandingPageContent = {
  title: "Free Bot Designer for Discord (BDFD) Alternative",
  description:
    "Looking for a Bot Designer for Discord alternative? With Kite you build your Discord bot from visual blocks instead of BDScript. Free, open source and hosted 24/7.",
  heading: "A Free Bot Designer for Discord Alternative",
  intro:
    "In Bot Designer for Discord (BDFD) you write your commands in BDScript, its own scripting language. In Kite you build them by connecting blocks in a visual editor, so you don't need to learn a language. Kite is free and open source.",
  sections: [
    {
      heading: "Why Use Kite Instead of BDFD",
      points: [
        {
          title: "Blocks Instead of Script",
          description:
            "You don't need to learn BDScript functions or syntax. You drag in blocks like Ban member or Send message and connect them.",
        },
        {
          title: "See How Your Command Works",
          description:
            "Each command is a diagram of connected blocks, so you can see every step and every condition at a glance.",
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
      heading: "When BDFD Is the Better Choice",
      body: "BDFD has apps for Android and iOS, so you can build your bot on your phone. Kite runs in the browser, and its editor works best on a computer. If you only have a phone, BDFD is easier to use.",
    },
  ],
  comparison: {
    competitor: "BDFD",
    rows: [
      {
        feature: "How you build commands",
        kite: "Connect blocks",
        competitor: "Write BDScript",
      },
      {
        feature: "Build with AI",
        kite: "Yes",
        competitor: "Yes",
      },
      {
        feature: "Phone app",
        kite: "No, browser only",
        competitor: "Android and iOS",
      },
    ],
  },
  faqs: [
    {
      question: "Can I move my bot from BDFD to Kite?",
      answer:
        "Yes. Add the token of your existing Discord application to Kite and build your commands again. Your bot keeps its name, avatar and servers. Turn it off in BDFD first, so the two don't run at the same time.",
    },
    {
      question: "Can I import my BDScript code?",
      answer:
        "No. Kite doesn't use BDScript. You build your commands again with blocks, or describe them to the flow AI.",
    },
    {
      question: "Is Kite really free?",
      answer:
        "Yes. Your bot is hosted 24/7 for free. Premium plans give you higher limits, like more servers and commands, but you don't need one to run a bot.",
    },
  ],
};

export default function BotDesignerForDiscordAlternativePage() {
  return <LandingPage content={content} />;
}
