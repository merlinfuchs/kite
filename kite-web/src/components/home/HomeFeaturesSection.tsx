import {
  BlocksIcon,
  CodeIcon,
  DatabaseZapIcon,
  GlobeIcon,
  LayoutTemplateIcon,
  MessageSquareIcon,
  SparklesIcon,
  UsersIcon,
} from "lucide-react";

const features = [
  {
    icon: DatabaseZapIcon,
    title: "24/7 Hosting",
    description:
      "Your bot stays online without a server or a computer that stays on.",
  },
  {
    icon: LayoutTemplateIcon,
    title: "Templates",
    description:
      "Import moderation, welcome and AI chatbot commands, then change them.",
  },
  {
    icon: BlocksIcon,
    title: "Plugins",
    description:
      "Add counting or a starboard to your server without building anything.",
  },
  {
    icon: MessageSquareIcon,
    title: "Message Editor",
    description:
      "Design messages with embeds, buttons and select menus, and see a live preview.",
  },
  {
    icon: SparklesIcon,
    title: "AI Blocks",
    description:
      "Let your bot answer questions and search the web. No API key needed.",
  },
  {
    icon: GlobeIcon,
    title: "APIs and Integrations",
    description:
      "Send requests to any HTTP API, or use the built-in blocks for Roblox.",
  },
  {
    icon: UsersIcon,
    title: "Collaboration",
    description: "Invite others to build and manage the bot with you.",
  },
  {
    icon: CodeIcon,
    title: "Open Source",
    description:
      "Kite's code is on GitHub. See how it works, or host it yourself.",
  },
];

export default function HomeFeaturesSection() {
  return (
    <section id="features" className="scroll-mt-16 container py-16 sm:py-24">
      <h2 className="text-3xl md:text-4xl font-bold">
        Everything Your{" "}
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Bot Needs
        </span>
      </h2>
      <p className="text-lg md:text-xl text-muted-foreground mt-4 max-w-2xl">
        Slash commands, buttons, events and more, all in one place and free to
        start.
      </p>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-x-8 gap-y-6 sm:gap-y-10 mt-8 md:mt-12">
        {features.map((feature) => (
          <div key={feature.title} className="flex sm:block gap-4">
            <div className="flex-none w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
              <feature.icon className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-lg font-semibold sm:mt-4">{feature.title}</h3>
              <p className="text-muted-foreground mt-1">
                {feature.description}
              </p>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
