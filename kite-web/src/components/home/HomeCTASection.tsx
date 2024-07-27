import { Button } from "@/components/ui/button";
import Link from "next/link";

export default function HomeCTASection() {
  return (
    <section id="cta" className="bg-muted/50 py-16 mt-24 sm:mt-32">
      <div className="container lg:grid lg:grid-cols-2 place-items-center">
        <div className="lg:col-start-1">
          <h2 className="text-3xl md:text-4xl font-bold ">
            Create
            <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
              {" "}
              your own Discord Bot{" "}
            </span>
            for free right now
          </h2>
          <p className="text-muted-foreground text-xl mt-4 mb-8 lg:mb-0">
            Kite lets you create your own Discord bot without writing a single
            line of code. With support for slash commands, buttons, and more.
          </p>
        </div>

        <div className="space-y-4 lg:col-start-2">
          <Button className="w-full md:mr-4 md:w-auto" asChild>
            <Link href="/apps">Get started</Link>
          </Button>
          <Button variant="outline" className="w-full md:w-auto" asChild>
            <a href="/docs">Documentation</a>
          </Button>
        </div>
      </div>
    </section>
  );
}
