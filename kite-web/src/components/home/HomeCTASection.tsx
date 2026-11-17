import orangeLogo from "@/assets/logo/orange@1024.png";
import orangeLogoSvg from "@/assets/logo/orange@scalable.svg";
import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import Link from "next/link";

// light is a smaller version for pages where the content should keep the
// attention, like the marketplace.
export default function HomeCTASection({ light }: { light?: boolean }) {
  if (light) {
    return (
      <section id="cta" className="container max-w-5xl pb-16">
        <div className="rounded-xl border bg-muted/30 p-6 flex flex-col md:flex-row md:items-center gap-5">
          <img src={orangeLogo.src} alt="" className="w-10 h-10 flex-none" />
          <div className="flex-1">
            <h2 className="text-lg font-semibold">
              Make your own Discord bot with Kite
            </h2>
            <p className="text-sm text-muted-foreground mt-1">
              Add templates or build commands with blocks and AI. Free to start,
              hosted 24/7.
            </p>
          </div>
          <Button asChild className="flex-none">
            <Link href="/apps">Get started for free</Link>
          </Button>
        </div>
      </section>
    );
  }

  return (
    <section id="cta" className="container pb-16 sm:pb-24">
      <div className="relative overflow-hidden rounded-3xl border bg-muted/30 px-6 py-14 sm:px-12 sm:py-20 text-center">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute -top-24 left-1/2 -translate-x-1/2 h-56 w-80 rounded-full bg-primary/15 blur-[100px]"
        />
        <div className="relative z-10">
          <img src={orangeLogoSvg.src} alt="" className="w-12 h-12 mx-auto" />
          <h2 className="text-3xl md:text-4xl font-bold mt-6">
            Make your own Discord bot
          </h2>
          <p className="text-lg text-muted-foreground mt-4 max-w-xl mx-auto">
            Free to start, no credit card. Your bot is online in a few minutes.
          </p>
          <div className="flex flex-col sm:flex-row justify-center gap-4 mt-8">
            <Button asChild size="lg">
              <Link href="/apps">Get started for free</Link>
            </Button>
            <Button asChild variant="outline" size="lg">
              <a
                href={env.NEXT_PUBLIC_DISCORD_LINK}
                target="_blank"
                rel="noreferrer noopener"
              >
                Join the Discord server
              </a>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
