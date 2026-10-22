import { ArrowRightIcon } from "lucide-react";
import Link from "next/link";

const steps = [
  {
    title: "Create a Discord App",
    description:
      "Create an application in the Discord Developer Portal and copy its bot token. It takes two minutes.",
  },
  {
    title: "Add It to Kite",
    description:
      "Sign in with Discord, paste the token and invite the bot to your server. It's online right away.",
  },
  {
    title: "Build Your Commands",
    description:
      "Connect blocks in the editor, import a template or describe a command to Kite's AI. Save it and it shows up in Discord.",
  },
];

export default function HomeHowItWorksSection() {
  return (
    <section
      id="how-it-works"
      className="container py-16 sm:py-24 scroll-mt-16"
    >
      <h2 className="text-3xl md:text-4xl font-bold text-center">
        How It{" "}
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Works
        </span>
      </h2>
      <ol className="relative grid md:grid-cols-3 gap-8 mt-12 max-w-5xl mx-auto">
        {/* Connects the numbers of the steps. */}
        <div className="hidden md:block absolute top-6 left-[16.7%] right-[16.7%] border-t-2 border-dashed" />
        {steps.map((step, i) => (
          <li key={step.title} className="text-center">
            <div className="relative w-12 h-12 mx-auto rounded-full bg-primary text-primary-foreground flex items-center justify-center text-lg font-bold">
              {i + 1}
            </div>
            <h3 className="text-xl font-semibold mt-4">{step.title}</h3>
            <p className="text-muted-foreground mt-2">{step.description}</p>
          </li>
        ))}
      </ol>
      <div className="text-center mt-10">
        <Link
          href="/guides/how-to-make-a-discord-bot"
          className="inline-flex items-center gap-2 text-primary font-medium hover:underline"
        >
          Read the step-by-step guide
          <ArrowRightIcon className="w-4 h-4" />
        </Link>
      </div>
    </section>
  );
}
