import { Button } from "@/components/ui/button";
import { banFlow } from "@/components/landing/flows";
import ShowcaseVisual from "@/components/landing/ShowcaseVisual";
import env from "@/lib/env/client";
import { CodeIcon } from "lucide-react";
import Link from "next/link";

export default function HomeHeroSection() {
  return (
    <section className="container grid grid-cols-1 lg:grid-cols-2 gap-4 lg:gap-12 items-center py-16 md:py-24">
      <div className="text-center lg:text-start space-y-6 z-10">
        <h1 className="text-5xl md:text-6xl font-bold">
          Make{" "}
          <span className="inline bg-gradient-to-r from-[#58d1f2] to-[#5865F2] text-transparent bg-clip-text">
            Discord Bots
          </span>{" "}
          with{" "}
          <span className="inline bg-gradient-to-r from-[#f9ad15] to-primary text-transparent bg-clip-text">
            Kite
          </span>
        </h1>

        <p className="text-lg md:text-xl text-muted-foreground md:w-10/12 mx-auto lg:mx-0">
          Kite is a free Discord bot maker. Build your own bot with slash
          commands, buttons, events and AI in a visual editor, without writing
          code. We host it for you 24/7.
        </p>

        <div className="flex flex-col md:flex-row justify-center lg:justify-start gap-4">
          <Button className="md:w-48" asChild>
            <Link href="/apps">Get Started for Free</Link>
          </Button>
          <Button variant="outline" className="md:w-48" asChild>
            <a href="#how-it-works">See How It Works</a>
          </Button>
        </div>

        <a
          href={env.NEXT_PUBLIC_GITHUB_LINK}
          target="_blank"
          rel="noreferrer noopener"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
        >
          <CodeIcon className="w-4 h-4" />
          Open source on GitHub
        </a>
      </div>

      <div className="z-10">
        <ShowcaseVisual
          showcase={{
            flow: banFlow,
            command: "ban",
            reply: "Banned the user.",
          }}
          hero
        />
      </div>

      {/* Shadow effect */}
      <div className="shadow"></div>
    </section>
  );
}
