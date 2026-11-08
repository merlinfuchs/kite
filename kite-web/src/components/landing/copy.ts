import { FAQProps } from "@/components/home/HomeFAQSections";

// Copy that several landing pages share.
export const openSourcePoint = {
  title: "Open source",
  description:
    "Kite's code is on GitHub. You can see how your bot runs, or host Kite yourself.",
};

export const aiBuilderPoint = {
  title: "Let AI build it",
  description:
    "Describe what a command should do and Flow AI, the AI assistant in the editor, adds and connects the blocks for you.",
};

export const freeFaq: FAQProps = {
  question: "Is Kite really free?",
  answer:
    "Yes. Your bot is hosted 24/7 for free, even with no activity. Paid plans give you higher limits, like more servers and commands, but you don't need one to run a bot.",
};

export const noCodeFaq: FAQProps = {
  question: "Do I need to know how to code?",
  answer:
    "No. You build commands by connecting blocks in a visual editor. You can also describe a command to Flow AI and it builds it for you.",
};
