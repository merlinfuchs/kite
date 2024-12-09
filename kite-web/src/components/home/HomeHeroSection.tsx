import { buttonVariants, Button } from "@/components/ui/button";
import { CodeIcon } from "lucide-react";
import FlowExample from "../flow/FlowExample";
import env from "@/lib/env/client";
import Link from "next/link";

export default function HomeHeroSection() {
  return (
    <section className="container grid lg:grid-cols-2 place-items-center py-20 md:py-32 gap-10">
      <div className="text-center lg:text-start space-y-6">
        <main className="text-5xl md:text-6xl font-bold">
          <h2 className="inline">
            Make{" "}
            <span className="inline bg-gradient-to-r from-[#58d1f2] to-[#5865F2] text-transparent bg-clip-text">
              Discord Bots
            </span>
          </h2>{" "}
          with{" "}
          <h1 className="inline">
            <span className="inline bg-gradient-to-r from-[#f9ad15] to-primary text-transparent bg-clip-text">
              Kite
            </span>
          </h1>
        </main>

        <p className="text-xl text-muted-foreground md:w-10/12 mx-auto lg:mx-0">
          Make your own Discord Bot with Kite for free without a single line of
          code. With support for slash commands, buttons, events, and more.
        </p>

        <div className="space-y-4 md:space-y-0 md:space-x-4">
          <Button className="w-full md:w-1/3" asChild>
            <Link href="/apps">Get Started</Link>
          </Button>

          <a
            rel="noreferrer noopener"
            href={env.NEXT_PUBLIC_GITHUB_LINK}
            target="_blank"
            className={`w-full md:w-1/3 ${buttonVariants({
              variant: "outline",
            })}`}
          >
            Github Repository
            <CodeIcon className="ml-2 w-5 h-5" />
          </a>
        </div>
      </div>

      {/* Hero cards sections */}
      {/*<div className="z-10">
        <HomeHeroCards />
      </div>*/}
      <div className="hidden lg:block w-[700px] h-[550px] z-10">
        <FlowExample />
      </div>

      {/* Shadow effect */}
      <div className="shadow"></div>
    </section>
  );
}
