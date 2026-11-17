import { banFlow } from "@/components/landing/flows";
import ShowcaseVisual from "@/components/landing/ShowcaseVisual";
import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import { CodeIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useRef } from "react";

export default function HomeHeroSection() {
  const stageRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = stageRef.current;
    if (!el) return;

    if (!window.matchMedia("(prefers-reduced-motion: no-preference)").matches) {
      el.style.transform = "";
      return;
    }

    let rafId = 0;
    const update = () => {
      rafId = 0;
      const p = Math.min(1, Math.max(0, window.scrollY / 520));
      el.style.transform = `translate3d(0, ${-p * 18}px, 0) scale(${
        1 - p * 0.03
      })`;
    };

    const onScroll = () => {
      if (!rafId) rafId = window.requestAnimationFrame(update);
    };

    update();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => {
      window.removeEventListener("scroll", onScroll);
      if (rafId) window.cancelAnimationFrame(rafId);
    };
  }, []);

  return (
    <section className="relative lg:min-h-[calc(100dvh-3.5rem)] lg:flex lg:items-end">
      <div className="container relative z-10 grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-12 items-center py-8 md:py-10 lg:pt-6 lg:pb-16">
        <div className="text-center lg:text-start lg:self-end space-y-6">
          <h1 className="text-5xl md:text-6xl font-bold">
            Make{" "}
            <span className="inline bg-gradient-to-r from-[#58d1f2] to-[#5865F2] text-transparent bg-clip-text">
              Discord bots
            </span>{" "}
            with{" "}
            <span className="inline bg-gradient-to-r from-[#f9ad15] to-primary text-transparent bg-clip-text">
              Kite
            </span>
          </h1>

          <p className="text-lg md:text-xl text-muted-foreground md:w-10/12 mx-auto lg:mx-0">
            Kite is a free Discord bot maker. Build your own bot with slash
            commands, buttons, events and AI in a visual editor, without writing
            code. We host it for you 24/7, even on quiet days.
          </p>

          <div className="flex flex-col md:flex-row justify-center lg:justify-start gap-4">
            <Button className="md:w-48" asChild>
              <Link href="/apps">Get started for free</Link>
            </Button>
            <Button variant="outline" className="md:w-48" asChild>
              <a href="#how-it-works">See how it works</a>
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

        <div
          ref={stageRef}
          className="will-change-transform [&>div]:!h-[clamp(340px,46dvh,420px)] lg:[&>div]:!h-[clamp(420px,calc(100dvh-10rem),620px)]"
        >
          <ShowcaseVisual
            showcase={{
              flow: banFlow,
              command: "ban",
              reply: "Banned the user.",
            }}
            hero
          />
        </div>
      </div>

      {/* Shadow effect */}
      <div aria-hidden="true" className="shadow" />
    </section>
  );
}
