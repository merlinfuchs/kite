import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import env from "@/lib/env/client";

interface FAQProps {
  question: string;
  answer: string;
  value: string;
}

const FAQList: FAQProps[] = [
  {
    question: "Is Kite free to use?",
    answer: "Yes. Kite is open source and free to use for everyone.",
    value: "item-1",
  },
  {
    question: "Can I customize the name and avatar of my Discord bot?",
    answer:
      "Yes. With Kite you create the bot yourself so you can customize the name and avatar however you like.",
    value: "item-2",
  },
  {
    question: "Does Kite support slash commands and other Discord features?",
    answer:
      "Kite currently supports slash commands, responding to them, and a few more actions. We are working on adding more features.",
    value: "item-3",
  },
  {
    question: "How many servers can I add my bot to?",
    answer:
      "You can add your bot to up to 100 servers. This limit may change in the future.",
    value: "item-4",
  },
  {
    question: "Can I create multiple bots?",
    answer:
      "Yes. You can create multiple bots with Kite. There is currently no limit to the number of bots you can create.",
    value: "item-5",
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
        {FAQList.map(({ question, answer, value }: FAQProps) => (
          <AccordionItem key={value} value={value}>
            <AccordionTrigger className="text-left">
              {question}
            </AccordionTrigger>

            <AccordionContent>{answer}</AccordionContent>
          </AccordionItem>
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
