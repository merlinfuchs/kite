import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { noCodeFaq } from "@/components/landing/copy";
import { banFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Make a Free Discord Moderation Bot",
  description:
    "Make your own Discord moderation bot for free with Kite. Get /ban, /kick and /mute commands from a template and change them without writing code.",
  heading: "Make Your Own Discord Moderation Bot",
  intro:
    "Build a moderation bot with your own name and avatar for free. Start with the moderation template, then change the commands so they work the way your server needs. You don't need to write code.",
  secondaryAction: {
    label: "View Templates",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/templates`,
  },
  showcase: {
    flow: banFlow,
    command: "ban",
    reply: "Banned the user.",
  },
  sections: [
    {
      heading: "Start With the Moderation Template",
      body: "The template adds four slash commands to your bot. Only members with the right Discord permissions can use them.",
      points: [
        {
          title: "/ban",
          description: "Ban a member from the server, with a reason.",
        },
        {
          title: "/unban",
          description: "Unban a user, with a reason.",
        },
        {
          title: "/kick",
          description: "Kick a member from the server, with a reason.",
        },
        {
          title: "/mute",
          description:
            "Time out a member so they can't send messages for a while.",
        },
      ],
    },
    {
      heading: "Change It to Fit Your Server",
      points: [
        {
          title: "Log Every Action",
          description:
            "Send a message to a log channel every time a moderator bans, kicks or mutes someone.",
        },
        {
          title: "Clear Messages",
          description:
            "Add a /purge command that deletes up to 100 messages in a channel at once.",
        },
        {
          title: "Filter Messages",
          description:
            "Listen for new messages, check them with conditions and delete the ones you don't allow.",
        },
        {
          title: "Manage Roles",
          description:
            "Give or remove roles with a command, for example to add a warning role.",
        },
        {
          title: "Server Logs",
          description:
            "Post a message when members join or leave, or when messages are edited or deleted.",
        },
        {
          title: "Let AI Build It",
          description:
            "Describe the command you want, like a /warn command that counts warnings, and Kite's AI builds it.",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "Is the moderation bot free?",
      answer:
        "Yes. You can build and host your moderation bot on Kite for free. Most actions, like banning a member, cost 1 of your 10,000 free credits per month.",
    },
    {
      question: "Can normal members use the moderation commands?",
      answer:
        "No. The template only lets members use the commands if they have the matching Discord permission, like Ban Members. You can change the permissions for each command.",
    },
    {
      question: "Does Kite have automod?",
      answer:
        "Kite has no automod you can turn on. You can build simple filters with a message event listener and conditions, for example to delete messages that contain certain words.",
    },
    noCodeFaq,
  ],
};

export default function DiscordModerationBotPage() {
  return <LandingPage content={content} />;
}
