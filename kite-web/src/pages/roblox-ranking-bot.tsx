import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { noCodeFaq } from "@/components/landing/copy";
import { robloxRankFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Make a Free Roblox Ranking Discord Bot",
  description:
    "Make your own Roblox ranking bot for Discord with Kite. Rank group members, show their rank and accept join requests from Discord, without a cookie and without writing code.",
  heading: "Make Your Own Roblox Ranking Bot",
  intro:
    "Connect your Roblox group to a Discord bot with your own name and avatar. Staff rank members with a slash command, and members verified with Bloxlink let themselves into the group. You don't need to write code or give the bot your cookie.",
  secondaryAction: {
    label: "Read the Guide",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/roblox-ranking`,
  },
  showcase: {
    flow: robloxRankFlow,
    command: "setrank",
    reply: "Ranked builderman to Moderator!",
  },
  sections: [
    {
      heading: "Start With the Roblox Ranking Template",
      body: "Enable Bloxlink and Roblox Open Cloud with their keys, enter your group's ID and import the template. Your bot gets three slash commands.",
      points: [
        {
          title: "/setrank",
          description:
            "Change a member's role to the one you name or its rank number. Only staff with Manage Roles can use it.",
        },
        {
          title: "/getrank",
          description: "Show anyone's role and rank in the group.",
        },
        {
          title: "/joingroup",
          description:
            "Accept a verified member's request to join the group, so you don't have to do it on Roblox.",
        },
      ],
    },
    {
      heading: "Build the Bot Your Group Needs",
      points: [
        {
          title: "Ranks for Milestones",
          description:
            "Rank members when they reach a goal you track in Kite, like a number of messages or events attended.",
        },
        {
          title: "Discord Roles From Ranks",
          description:
            "Give members the Discord role that matches their rank in the group.",
        },
        {
          title: "Several Roles per Member",
          description:
            "Give or take single roles and keep the others, for groups that use more than one role per member.",
        },
        {
          title: "Rank Logs",
          description:
            "Post every rank change with who made it to a staff channel.",
        },
        {
          title: "Your Own Branding",
          description:
            "Give the bot your group's name and avatar, and send replies as embeds in your colors.",
        },
        {
          title: "Let AI Build It",
          description:
            "Describe the command you want, like a /verify command that gives verified members the Member role in the group, and Kite's AI builds it.",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "Is the ranking bot free?",
      answer:
        "Yes. You can build and host your ranking bot on Kite for free. Looking up a member with Bloxlink and changing their role costs 1 credit each, out of 10,000 free credits per month.",
    },
    {
      question: "Do I need to give the bot my Roblox cookie?",
      answer:
        "No. Kite uses an Open Cloud API key, which you create on Roblox's Creator Hub with only the group permissions. Kite stores it encrypted and only sends it to Roblox. Roblox only lets the key give roles below the highest role of the account it belongs to.",
    },
    {
      question: "Why do members need Bloxlink?",
      answer:
        "Roblox doesn't know which Discord account belongs to whom. Bloxlink does, once members verified with it, so the bot asks Bloxlink for their Roblox account.",
    },
    {
      question: "Is Kite made by Roblox?",
      answer:
        "No. Kite is a third-party app that uses Roblox's and Bloxlink's APIs and isn't affiliated with either. For help with your Kite bot, ask in Kite's Discord server.",
    },
    noCodeFaq,
  ],
};

export default function RobloxRankingBotPage() {
  return <LandingPage content={content} />;
}
