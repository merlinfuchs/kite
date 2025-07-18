import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import env from "@/lib/env/client";
import { ChevronDown } from "lucide-react";

interface FAQProps {
  question: string;
  answer: string;
}

const FAQList: FAQProps[] = [
  {
    question: "What is Kite?",
    answer:
      "Kite is an open source platform for building and hosting Discord bots without the need to write a single line of code. It's powered by an advanced no-code editor and free to use for everyone.",
  },
  {
    question: "Is Kite free to use?",
    answer:
      "Yes. Kite is open source and free to use for everyone. Your bots will be hosted 24/7 and all the necessary features are free to use!",
  },
  {
    question: "Can I customize the name and avatar of my Discord bot?",
    answer:
      "Yes. With Kite you create the bot yourself so you can customize the name and avatar however you like. You can even customize the bot's status and activity for free, without any limitations!",
  },
  {
    question: "Does Kite support slash commands and other Discord features?",
    answer:
      "Kite currently supports slash commands, message components, and event listeners. You can respond to slash commands and message components, and run actions based on events. We are working on adding more features!",
  },
  {
    question: "How many servers can I add my bot to?",
    answer:
      "You can add your bot to up to 100 servers. This limit may change in the future.",
  },
  {
    question: "Can I create multiple bots?",
    answer:
      "Yes. You can create up to 10 bots with Kite. This limit may change in the future.",
  },
];

export default function HomeFAQSection() {
  return (
    <section id="faq" className="container py-24 sm:py-32">
      <h2 className="text-3xl md:text-4xl font-bold mb-4">
        Frequently Asked{" "}
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Questions
        </span>
      </h2>

      <Accordion type="single" collapsible className="w-full AccordionRoot">
        {FAQList.map(({ question, answer }: FAQProps) => (
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
