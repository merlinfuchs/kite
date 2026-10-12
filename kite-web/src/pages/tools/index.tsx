import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import ToolCard from "@/tools/common/components/ToolCard";
import { externalTools, kiteTools } from "@/tools/registry";

export default function ToolsPage() {
  return (
    <HomeLayout
      title="Free Discord Tools"
      description="Free tools for Discord: calculate bot permissions, generate timestamps, decode IDs, inspect webhooks and create messages with embeds."
    >
      <section className="border-b bg-muted/30">
        <div className="container max-w-5xl py-14 md:py-20 text-center">
          <h1 className="text-4xl md:text-5xl font-bold">Free Discord Tools</h1>
          <p className="text-xl text-muted-foreground mt-4 max-w-2xl mx-auto">
            Small tools for everyone who runs a Discord server or builds a bot.
            They run in your browser and don&apos;t need an account.
          </p>
        </div>
      </section>

      <div className="container max-w-5xl py-12 space-y-14">
        <section>
          <h2 className="text-2xl font-bold mb-6">Kite Tools</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {kiteTools.map((tool) => (
              <ToolCard key={tool.id} tool={tool} />
            ))}
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold">More Tools</h2>
          <p className="text-muted-foreground mt-1 mb-6">
            Tools on message.style and dis.wtf, which open in a new tab.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {externalTools.map((tool) => (
              <ToolCard key={tool.id} tool={tool} />
            ))}
          </div>
        </section>
      </div>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
