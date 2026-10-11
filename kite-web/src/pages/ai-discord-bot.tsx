import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { aiBuilderPoint, noCodeFaq } from "@/components/landing/copy";
import { askAiFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Make a Free AI Discord Bot",
  description:
    "Make your own AI Discord bot for free with Kite. Answer questions with /ask, reply when someone mentions the bot and search the web, without writing code.",
  heading: "Make Your Own AI Discord Bot",
  intro:
    "Build an AI chatbot for your Discord server with your own name, avatar and personality. Start with the Ask AI template, or describe a command and let Kite's AI build it. You don't need to write code or get an API key.",
  secondaryAction: {
    label: "Read the AI Guide",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/flow-ai`,
  },
  showcase: {
    flow: askAiFlow,
    command: "ask",
    reply:
      "A Discord bot is an app that runs in your server. It can answer commands, post messages and manage roles.",
  },
  sections: [
    {
      heading: "Start With the Ask AI Template",
      body: "The template adds a command and an event listener to your bot. You give the AI a personality when you import it.",
      points: [
        {
          title: "/ask",
          description:
            "Members ask a question with the /ask command and the bot answers.",
        },
        {
          title: "Reply When Mentioned",
          description:
            "When someone mentions your bot in a message, it answers in the channel.",
        },
        {
          title: "Your Own Personality",
          description:
            "Tell the AI how to answer, for example as a helpful support agent or as a character from your community.",
        },
      ],
    },
    {
      heading: "More You Can Build With AI",
      points: [
        {
          title: "Search the Web",
          description:
            "The Search the Web block finds current information, so your bot can answer questions about recent events.",
        },
        {
          title: "Choose the Model",
          description:
            "Pick Fast for quick and cheap answers, or Balanced and Smartest for harder questions.",
        },
        {
          title: "Combine AI With Actions",
          description:
            "Use the AI's answer in other blocks, for example to post it in another channel or save it in a variable.",
        },
        aiBuilderPoint,
      ],
    },
  ],
  faqs: [
    {
      question: "Do I need an OpenAI API key?",
      answer:
        "No. The AI blocks work without your own API key. They use the credits of your app instead.",
    },
    {
      question: "How many credits does the AI use?",
      answer:
        "An answer from the Ask AI block costs 5 credits with the Fast model, 20 with Balanced and 100 with Smartest. A web search costs 25 credits with the Fast model. The free plan includes 10,000 credits per month.",
    },
    {
      question: "Can I change how the AI answers?",
      answer:
        "Yes. You set a personality when you import the template, and you can change the instructions in the Ask AI block at any time.",
    },
    noCodeFaq,
  ],
};

export default function AIDiscordBotPage() {
  return <LandingPage content={content} />;
}
