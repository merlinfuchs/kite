import { TemplateList } from "@/components/app/TemplateList";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";

export default function TemplatesPage() {
  return (
    <HomeLayout
      title="Discord Bot Templates"
      description="Free templates for Discord bots made with Kite: moderation commands, welcome messages, verification, modmail, an AI assistant and ER:LC commands. Add one to your bot and change anything in the editor."
    >
      <section className="border-b bg-muted/30">
        <div className="container max-w-5xl py-14 md:py-20 text-center">
          <h1 className="text-4xl md:text-5xl font-bold">
            Discord Bot Templates
          </h1>
          <p className="text-xl text-muted-foreground mt-4 max-w-2xl mx-auto">
            Ready-made commands, event listeners and messages for your bot. Add
            one in Kite and change it like anything you built yourself.
          </p>
        </div>
      </section>

      <div className="container max-w-5xl py-12">
        <TemplateList href={(t) => `/templates/${t.id}`} />
      </div>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
