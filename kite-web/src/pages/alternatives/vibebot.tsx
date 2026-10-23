import ComparisonPage, {
  ComparisonPageContent,
} from "@/components/landing/ComparisonPage";
import { freeFaq, noCodeFaq, openSourcePoint } from "@/components/landing/copy";

const content: ComparisonPageContent = {
  title: "Free VibeBot Alternative",
  description:
    "Looking for a VibeBot alternative? Kite is a free, open source Discord bot maker. Get your own bot with AI help and 24/7 hosting. No trial and no credit card.",
  competitor: "VibeBot",
  lastChecked: "October 2026",
  heading: "A Free VibeBot Alternative",
  intro:
    "Kite and VibeBot both let you make a Discord bot by describing what it should do. VibeBot's AI writes code for your bot, while Kite's AI builds blocks you can see and change. On VibeBot your own bot costs from $5 a month after a 3-day trial. On Kite it's free.",
  chooseKite: [
    "You want your own bot without paying",
    "You don't want to add a card to get started",
    "You want to see and change every step your bot takes",
    "You want a tool that is open source",
  ],
  chooseCompetitor: [
    "You want AI to write real code for your bot",
    "You want a music player or an AI chat persona ready to use",
    "You want to be able to export your bot's source code",
  ],
  rows: [
    {
      feature: "Your own bot",
      kite: "Free",
      competitor: "From $5/month after a 3-day trial",
    },
    {
      feature: "Free plan",
      kite: "Your own bot, name and avatar",
      competitor: "A shared bot named VibeBot",
    },
    {
      feature: "Card needed to start",
      kite: "No",
      competitor: "Yes, for the trial",
    },
    {
      feature: "How you build",
      kite: "Blocks in a visual editor, with AI help",
      competitor: "AI writes the code, plus templates",
    },
    {
      feature: "AI builder (free)",
      kite: "Yes, with a monthly limit",
      competitor: "No, Pro only ($10/month)",
    },
    {
      feature: "Bots",
      kite: "Up to 10 free",
      competitor: "1, extra bots $5/month each",
    },
    {
      feature: "Music player",
      kite: "No",
      competitor: "Yes, on Pro",
    },
    {
      feature: "Export your bot's code",
      kite: "No",
      competitor: "Paid add-on",
    },
  ],
  reasons: [
    {
      title: "Your Own Bot for Free",
      description:
        "VibeBot's free option is a shared bot that always has the VibeBot name and avatar. Your own bot needs Lite ($5/month) or Pro ($10/month) after the trial. On Kite your bot has your own name and avatar for free.",
    },
    {
      title: "See Every Step",
      description:
        "When Kite's AI builds a command, you get connected blocks you can read and change. You don't need to read code to know what your bot does.",
    },
    {
      title: "Up to 10 Free Bots",
      description:
        "VibeBot's plans include 1 bot and each extra bot costs $5 a month. You can make up to 10 bots on Kite without paying.",
    },
    openSourcePoint,
  ],
  faqs: [freeFaq, noCodeFaq],
};

export default function VibeBotAlternativePage() {
  return <ComparisonPage content={content} />;
}
