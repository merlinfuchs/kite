import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { kiteTools } from "@/tools/registry";
import { ChevronRightIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";
import ToolCard, { ToolIcon } from "./ToolCard";

export default function ToolLayout({
  toolId,
  title,
  description,
  children,
  article,
  jsonLd,
}: {
  // The id of the tool in the registry, for its icon and the related tools.
  toolId: string;
  title: string;
  description: string;
  children: ReactNode;
  // Explains the tool below it.
  article?: ReactNode;
  jsonLd?: object[];
}) {
  const tool = kiteTools.find((t) => t.id === toolId)!;
  const otherTools = kiteTools.filter((t) => t.id !== toolId);

  return (
    <HomeLayout title={title} description={description} jsonLd={jsonLd}>
      <section className="border-b bg-muted/30">
        <div className="container max-w-5xl py-10 md:py-14">
          <nav className="flex items-center gap-1 text-sm text-muted-foreground mb-6">
            <Link href="/tools" className="hover:text-foreground">
              Tools
            </Link>
            <ChevronRightIcon className="w-4 h-4" />
            <span className="text-foreground">{title}</span>
          </nav>
          <div className="flex gap-5 items-start">
            <ToolIcon tool={tool} className="w-14 h-14 hidden sm:flex" />
            <div>
              <h1 className="text-3xl md:text-4xl font-bold">{title}</h1>
              <p className="text-lg text-muted-foreground mt-2 max-w-2xl">
                {description}
              </p>
            </div>
          </div>
        </div>
      </section>

      <div className="container max-w-5xl py-10">
        <div className="rounded-xl border bg-card shadow-sm p-5 md:p-8">
          {children}
        </div>

        {article && <div className="max-w-3xl">{article}</div>}

        <section className="mt-16">
          <h2 className="text-2xl font-bold mb-6">More Discord Tools</h2>
          <div className="grid md:grid-cols-3 gap-4">
            {otherTools.map((t) => (
              <ToolCard key={t.id} tool={t} />
            ))}
          </div>
        </section>
      </div>
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
