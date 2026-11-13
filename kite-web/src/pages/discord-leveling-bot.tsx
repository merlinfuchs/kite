import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { noCodeFaq } from "@/components/landing/copy";
import { levelingFlow } from "@/components/landing/flows";

const content: LandingPageContent = {
  title: "Make a free Discord leveling bot",
  description:
    "Make your own Discord leveling bot for free with Kite. Members earn XP for chatting, level up, and get reward roles, with /rank and /leaderboard.",
  heading: "Make your own Discord leveling bot",
  intro:
    "Members earn XP for chatting and level up, under your bot's own name and avatar. Start with the Leveling template, then change the XP, the messages, or the rewards. You don't need to write code.",
  secondaryAction: {
    label: "View the template",
    href: "/marketplace/leveling",
  },
  showcase: {
    flow: levelingFlow,
    event: "Alex sent a message.",
    reply: "🎉 @Alex reached level 3!",
  },
  sections: [
    {
      heading: "What your leveling bot can do",
      points: [
        {
          title: "XP for chatting",
          description:
            "Members get 15 to 25 XP for a message, at most once a minute, so spamming doesn't help.",
        },
        {
          title: "Levels that take longer",
          description:
            "Level 1 needs 100 XP, level 5 needs 2,500, and level 10 needs 10,000.",
        },
        {
          title: "Level-up messages",
          description:
            "The bot congratulates members in a channel you pick, or in the channel they chatted in.",
        },
        {
          title: "/rank and /leaderboard",
          description:
            "Members see their level and the XP they need for the next one. The leaderboard shows 10 members per page.",
        },
        {
          title: "Reward roles",
          description:
            "Members get a role you pick when they reach a level you choose.",
        },
        {
          title: "Coins with the Economy template",
          description:
            "Add /daily, /work, /pay, and /richest with the Economy template. It works next to Leveling.",
          href: "/marketplace/economy",
        },
      ],
    },
    {
      heading: "How to set it up",
      points: [
        {
          title: "1. Create your bot",
          description:
            "Create a Discord application, add its token to Kite, and invite the bot to your server.",
        },
        {
          title: "2. Add the template",
          description:
            "Open the Marketplace in your app and choose Leveling. Pick a level-up channel and a reward role if you want them.",
        },
        {
          title: "3. Change what you like",
          description:
            "Open the event listener to change how much XP a message gives or what the level-up message says.",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "Is the leveling bot free?",
      answer:
        "Yes. A message that earns XP costs 2 of your 10,000 free credits per month. Messages within a minute of a member's last XP cost none.",
    },
    {
      question: "Does the bot need the Message Content Intent?",
      answer: "No. It counts messages without reading them.",
    },
    {
      question: "Can I give roles at more than one level?",
      answer:
        "Yes. The template has one reward role. Copy its reward role blocks in the event listener once for each other level.",
    },
    {
      question: "How do I reset everyone's XP?",
      answer:
        "Open Stored variables in the sidebar and delete the values of the xp variable.",
    },
    noCodeFaq,
  ],
};

export default function DiscordLevelingBotPage() {
  return <LandingPage content={content} />;
}
