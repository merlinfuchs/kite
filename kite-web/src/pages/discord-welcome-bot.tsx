import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { noCodeFaq } from "@/components/landing/copy";
import { welcomeFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Make a Free Discord Welcome Bot",
  description:
    "Make your own Discord welcome bot for free with Kite. Greet new members, give them a role and send them a direct message, without writing code.",
  heading: "Make Your Own Discord Welcome Bot",
  intro:
    "Build a welcome bot for free, with your own name and avatar. Start with the welcome template, then change the message or add more steps. You don't need to write code.",
  secondaryAction: {
    label: "View Templates",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/templates`,
  },
  showcase: {
    flow: welcomeFlow,
    event: "Alex joined the server.",
    reply: "Welcome to the server, @Alex!",
  },
  sections: [
    {
      heading: "What Your Welcome Bot Can Do",
      points: [
        {
          title: "Welcome Messages",
          description:
            "Post a message in a channel when someone joins. Mention the new member and show the member count of your server.",
        },
        {
          title: "Embeds and Buttons",
          description:
            "Design the welcome message in the message editor, with embeds, images and buttons.",
        },
        {
          title: "Auto Roles",
          description:
            "Give new members a role as soon as they join, for example a Member or Unverified role.",
        },
        {
          title: "Direct Messages",
          description:
            "Send new members a private message with your rules or links, if they allow messages from the server.",
        },
        {
          title: "Goodbye Messages",
          description: "Post a message when a member leaves the server.",
        },
        {
          title: "Verification",
          description:
            "Add a button that gives members access to the server when they click it.",
        },
      ],
    },
    {
      heading: "How to Set It Up",
      points: [
        {
          title: "1. Create Your Bot",
          description:
            "Create a Discord application, add its token to Kite and invite the bot to your server.",
        },
        {
          title: "2. Import the Template",
          description:
            "Open Templates in your app, choose Welcomer and enter the ID of your welcome channel.",
        },
        {
          title: "3. Change the Message",
          description:
            "Open the event listener and change the text, or add blocks like Add role.",
        },
      ],
    },
  ],
  faqs: [
    {
      question: "Is the welcome bot free?",
      answer:
        "Yes. You can build and host your welcome bot on Kite for free. Each welcome message costs 1 of your 10,000 free credits per month.",
    },
    {
      question: "Why doesn't my bot see new members?",
      answer:
        "Discord only tells bots about new members if the Server Members Intent is turned on. Turn it on in the Bot section of your app in the Discord Developer Portal.",
    },
    {
      question: "Can I send the welcome message as a direct message?",
      answer:
        "Yes. Use the Send direct message block. Some members don't allow messages from server members, so the message won't reach everyone.",
    },
    noCodeFaq,
  ],
};

export default function DiscordWelcomeBotPage() {
  return <LandingPage content={content} />;
}
