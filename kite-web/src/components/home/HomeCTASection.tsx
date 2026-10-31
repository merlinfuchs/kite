import orangeLogo from "@/assets/logo/orange@1024.png";
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
            <Link href="/apps">Get Started for Free</Link>
          </Button>
        </div>
      </section>
    );
  }

  return (
    <section id="cta" className="container pb-16 sm:pb-24">
      <div className="border-t pt-16 sm:pt-24 text-center">
        <img src={orangeLogo.src} alt="" className="w-12 h-12 mx-auto" />
        <h2 className="text-3xl md:text-4xl font-bold mt-6">
          Make Your Own Discord Bot
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-xl mx-auto">
          Free to start, no credit card. Your bot is online in a few minutes.
        </p>
        <div className="flex flex-col sm:flex-row justify-center gap-4 mt-8">
          <Button asChild size="lg">
            <Link href="/apps">Get Started for Free</Link>
          </Button>
          <Button asChild variant="outline" size="lg">
            <a href={env.NEXT_PUBLIC_DISCORD_LINK} target="_blank">
              Join the Discord
            </a>
          </Button>
        </div>
      </div>
    </section>
  );
}
