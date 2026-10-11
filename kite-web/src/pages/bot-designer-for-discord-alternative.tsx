import ComparisonPage, {
  ComparisonPageContent,
} from "@/components/landing/ComparisonPage";
import {
  aiBuilderPoint,
  freeFaq,
  openSourcePoint,
} from "@/components/landing/copy";

const content: ComparisonPageContent = {
  title: "Free Bot Designer for Discord (BDFD) Alternative",
  description:
    "Looking for a Bot Designer for Discord alternative? With Kite you build your Discord bot from visual blocks instead of BDScript. Free, open source and hosted 24/7.",
  competitor: "BDFD",
  heading: "A Free Bot Designer for Discord Alternative",
  intro:
    "In Bot Designer for Discord (BDFD) you write your commands in BDScript, its own scripting language. In Kite you build them by connecting blocks in a visual editor, so you don't need to learn a language. Kite is free and open source.",
  chooseKite: [
    "You'd rather connect blocks than learn a scripting language",
    "You want to see every step of a command at once",
    "You build your bot on a computer",
  ],
  chooseCompetitor: [
    "You build your bot on your phone",
    "You already know BDScript",
  ],
  rows: [
    {
      feature: "How you build commands",
      kite: "Connect blocks",
      competitor: "Write BDScript",
    },
    {
      feature: "Build with AI",
      kite: "Yes",
      competitor: "Yes",
    },
    {
      feature: "Web editor",
      kite: "Yes",
      competitor: "Yes",
    },
    {
      feature: "Phone app",
      kite: "No",
      competitor: "Android and iOS",
    },
    {
      feature: "Hosting",
      kite: "Included, 24/7",
      competitor: "Included",
    },
  ],
  reasons: [
    {
      title: "Blocks Instead of Script",
      description:
        "You don't need to learn BDScript functions or syntax. You drag in blocks like Ban member or Send message and connect them.",
    },
    {
      title: "See How Your Command Works",
      description:
        "Each command is shown as connected blocks, so you can see every step and every condition.",
    },
    aiBuilderPoint,
    openSourcePoint,
  ],
  faqs: [
    freeFaq,
    {
      question: "Can I import my BDScript code?",
      answer:
        "No. Kite doesn't use BDScript. You build your commands again with blocks, or describe them to Kite's AI.",
    },
  ],
};

export default function BotDesignerForDiscordAlternativePage() {
  return <ComparisonPage content={content} />;
}
