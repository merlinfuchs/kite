import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import env from "@/lib/env/client";
import { ChevronDown } from "lucide-react";

export interface FAQProps {
  question: string;
  answer: string;
}

export const FAQList: FAQProps[] = [
  {
    question: "What is Kite?",
    answer:
      "Kite is a free, open source Discord bot maker. You build commands and event listeners in a visual editor by connecting blocks, and Kite hosts your bot 24/7. You don't need to write any code.",
  },
  {
    question: "Is Kite free to use?",
    answer:
      "Yes. Every bot starts on the free plan, which includes 24/7 hosting, slash commands, event listeners and custom status. Premium plans raise the limits, like the number of servers and commands, but you don't need them to run a bot.",
  },
  {
    question: "Do I need to know how to code?",
    answer:
      'No. You build your bot by connecting blocks like "Ban member", "Send message" or "Ask AI". You can also describe a command and Kite\'s AI builds it for you, or import a template for moderation, welcome messages or an AI chatbot.',
  },
  {
    question: "Can I customize the name and avatar of my Discord bot?",
    answer:
      "Yes. Your bot is your own Discord application, so you choose its name and avatar. You can also set its status for free.",
  },
  {
    question: "What can my bot do?",
    answer:
      "Kite supports slash commands, buttons, select menus, modals and event listeners for things like new members, messages, reactions and server boosts. Your flows can manage members, roles, channels and messages, store data in variables, send requests to other APIs and use AI.",
  },
  {
    question: "What are credits?",
    answer:
      "Most actions your bot takes, like sending a message or banning a member, cost 1 credit. Conditions and loops are free, and AI blocks cost more. The free plan includes 10,000 credits per month, which is enough for most servers. If your bot runs out, Kite stops it until the next month starts or you upgrade, and you start it again in its settings.",
  },
  {
    question: "How many servers can I add my bot to?",
    answer:
      "On the free plan your bot can be in up to 100 servers. Premium plans raise this limit.",
  },
  {
    question: "Can I create multiple bots?",
    answer: "Yes. You can create up to 10 bots with Kite.",
  },
];

export default function HomeFAQSection({
  faqs = FAQList,
}: {
  faqs?: FAQProps[];
}) {
  return (
    <section id="faq" className="scroll-mt-16 container py-16 sm:py-24">
      <h2 className="text-3xl md:text-4xl font-bold mb-4">
        Frequently Asked{" "}
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Questions
        </span>
      </h2>

      <Accordion type="single" collapsible className="w-full AccordionRoot">
        {faqs.map(({ question, answer }: FAQProps) => (
          <FAQItem key={question} question={question} answer={answer} />
        ))}
      </Accordion>

      <h3 className="font-medium mt-4">
        Still have questions?{" "}
        <a
          rel="noreferrer noopener"
          href={env.NEXT_PUBLIC_DISCORD_LINK}
          target="_blank"
          className="text-primary transition-all border-primary hover:border-b-2"
        >
          Join the Discord server
        </a>
      </h3>
    </section>
  );
}

function FAQItem({ question, answer }: FAQProps) {
  return (
    <details className="text-left border-b [&_svg]:open:-rotate-180">
      <summary className="flex flex-1 items-center justify-between py-4 font-medium transition-all hover:underline cursor-pointer">
        <div>{question}</div>
        <ChevronDown className="h-4 w-4 shrink-0 transition-transform duration-200" />
      </summary>

      <div className="overflow-hidden text-sm transition-all">
        <p className="pb-4 pt-0">{answer}</p>
      </div>
    </details>
  );
}
