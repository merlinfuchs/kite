import logo from "@/assets/logo/white@1024.png";
import orangeLogo from "@/assets/logo/orange@1024.png";
import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import { CheckIcon } from "lucide-react";
import Link from "next/link";

const highlights = [
  "No credit card",
  "Hosted 24/7",
  "Your own status",
  "Open source",
];

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
      <div className="relative overflow-hidden rounded-2xl bg-primary px-6 py-14 md:px-16 text-center text-primary-foreground">
        {/* Soft circles for some depth behind the text. */}
        <div className="absolute -top-24 -left-24 w-72 h-72 rounded-full bg-white/10 blur-2xl" />
        <div className="absolute -bottom-32 -right-16 w-96 h-96 rounded-full bg-black/10 blur-2xl" />

        <div className="relative">
          <img src={logo.src} alt="" className="w-14 h-14 mx-auto" />
          <h2 className="text-3xl md:text-5xl font-bold mt-6">
            Make Your Own Discord Bot Today
          </h2>
          <p className="text-lg md:text-xl opacity-90 mt-4 max-w-2xl mx-auto">
            Build commands by connecting blocks, or describe them and let AI
            build them. Kite is free to start and hosts your bot 24/7.
          </p>

          <div className="flex flex-col sm:flex-row justify-center gap-4 mt-8">
            <Button asChild variant="secondary" size="lg">
              <Link href="/apps">Get Started for Free</Link>
            </Button>
            <Button
              asChild
              variant="outline"
              size="lg"
              className="bg-transparent border-white/40 text-primary-foreground hover:bg-white/10 hover:text-primary-foreground"
            >
              <a href={env.NEXT_PUBLIC_DISCORD_LINK} target="_blank">
                Join the Discord
              </a>
            </Button>
          </div>

          <div className="flex flex-wrap justify-center gap-x-6 gap-y-2 mt-8 text-sm opacity-90">
            {highlights.map((highlight) => (
              <span key={highlight} className="flex items-center gap-1.5">
                <CheckIcon className="w-4 h-4" />
                {highlight}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
