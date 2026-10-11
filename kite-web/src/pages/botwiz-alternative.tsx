import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";

const content: LandingPageContent = {
  title: "Free BotWiz Alternative",
  description:
    "Looking for a BotWiz alternative? Kite is a free, open source Discord bot maker with a visual editor and 24/7 hosting. No trial and no credit card.",
  heading: "A Free BotWiz Alternative",
  intro:
    "Kite and BotWiz both let you build a Discord bot with blocks instead of code. BotWiz has a free trial and then needs a subscription. Kite has a free plan that doesn't run out, and Kite is open source.",
  sections: [
    {
      heading: "Why Use Kite Instead of BotWiz",
      points: [
        {
          title: "Free, Not a Trial",
          description:
            "BotWiz gives your first bot 7 days for free. After that your bot goes offline until you pay. Kite's free plan keeps your bot online with no time limit.",
        },
        {
          title: "Up to 10 Free Bots",
          description:
            "You can make up to 10 bots on Kite without paying. Each one gets 24/7 hosting.",
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
      heading: "When BotWiz Is the Better Choice",
      body: "BotWiz can also host bots you wrote yourself in Node.js or Python. Kite only runs bots built in Kite. If you have a discord.js or discord.py bot that you want to keep, BotWiz Hosting can run it.",
    },
  ],
  comparison: {
    competitor: "BotWiz",
    rows: [
      {
        feature: "Free plan",
        kite: "Yes, no time limit",
        competitor: "7-day trial",
      },
      {
        feature: "Visual editor",
        kite: "Yes",
        competitor: "Yes",
      },
      {
        feature: "Host your own code",
        kite: "No",
        competitor: "Yes",
      },
    ],
  },
  faqs: [
    {
      question: "Can I move my bot from BotWiz to Kite?",
      answer:
        "Yes. Add the token of your existing Discord application to Kite and build your commands again. Your bot keeps its name, avatar and servers. Turn it off in BotWiz first, so the two don't run at the same time.",
    },
    {
      question: "Is Kite really free?",
      answer:
        "Yes. Your bots are hosted 24/7 for free. Premium plans give you higher limits, like more servers and commands, but you don't need one to run a bot.",
    },
    {
      question: "Can Kite host my discord.js or discord.py bot?",
      answer: "No. Kite only runs bots built with its visual editor.",
    },
  ],
};

export default function BotWizAlternativePage() {
  return <LandingPage content={content} />;
}
