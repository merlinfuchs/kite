import {
  BookOpenIcon,
  CircleHelpIcon,
  LucideIcon,
  ScaleIcon,
  SparklesIcon,
  WorkflowIcon,
} from "lucide-react";
import { externalTools, kiteTools, Tool } from "@/tools/registry";

export interface NavLink {
  title: string;
  description: string;
  href: string;
  icon: LucideIcon;
}

// Links shown under "Product" in the navbar and the mobile menu.
export const productLinks: NavLink[] = [
  {
    title: "Features",
    description: "Everything Kite can do for your Discord bot.",
    href: "/#features",
    icon: SparklesIcon,
  },
  {
    title: "How It Works",
    description: "See how you make a Discord bot without writing code.",
    href: "/#how-it-works",
    icon: WorkflowIcon,
  },
  {
    title: "Guide",
    description: "Step by step: make your first Discord bot.",
    href: "/guides/how-to-make-a-discord-bot",
    icon: BookOpenIcon,
  },
  {
    title: "Compare",
    description: "How Kite stacks up against other bot makers.",
    href: "/alternatives",
    icon: ScaleIcon,
  },
  {
    title: "FAQ",
    description: "Answers to common questions about Kite.",
    href: "/#faq",
    icon: CircleHelpIcon,
  },
];

// Kite's own tools and the message creator. All tools are listed on /tools.
export const navTools: Tool[] = [
  ...kiteTools,
  externalTools.find((t) => t.id === "message-creator")!,
];
