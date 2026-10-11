import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ChevronLeftIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";

export default function ToolLayout({
  children,
  title,
  description,
  className,
  jsonLd,
}: {
  children: ReactNode;
  title: string;
  description: string;
  className?: string;
  jsonLd?: object[];
}) {
  return (
    <HomeLayout title={title} description={description} jsonLd={jsonLd}>
      <div className={cn("py-10 px-5 max-w-4xl mx-auto", className)}>
        <div className="mb-14">
          <Button variant="outline" asChild>
            <Link href="/tools">
              <ChevronLeftIcon className="h-5 w-5 mr-2 -ml-1" />
              Other Tools
            </Link>
          </Button>
        </div>
        <div className="flex flex-col space-y-2 mb-5">
          <h1 className="text-3xl font-semibold leading-none tracking-tight">
            {title}
          </h1>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
        <div>{children}</div>

        <div className="mt-16 rounded-lg border bg-muted/50 p-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="text-lg font-semibold">
              Make your own Discord bot
            </div>
            <div className="text-muted-foreground">
              Build commands from blocks with Kite. It&apos;s free and hosted
              24/7.
            </div>
          </div>
          <Button asChild className="flex-none">
            <Link href="/guides/how-to-make-a-discord-bot">Get Started</Link>
          </Button>
        </div>
      </div>
      <HomeFooter />
    </HomeLayout>
  );
}
