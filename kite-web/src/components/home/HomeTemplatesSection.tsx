import { FeaturedTemplateEntry } from "@/components/app/TemplateListEntry";
import { Button } from "@/components/ui/button";
import { templateFromBundle, templatePath } from "@/lib/flow/templates";
import { Bundle } from "@/lib/types/wire.gen";
import { ArrowRightIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

// HomeTemplatesSection shows the most added templates of the Marketplace.
// It's left out when the page was built without the API.
export default function HomeTemplatesSection({
  bundles,
  total,
}: {
  bundles: Bundle[] | null;
  // How many templates the Marketplace has.
  total: number;
}) {
  const templates = useMemo(
    () => (bundles ?? []).flatMap((b) => templateFromBundle(b) ?? []),
    [bundles]
  );

  if (templates.length === 0) return null;

  return (
    <section id="templates" className="container pb-16 sm:pb-24">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h2 className="text-3xl md:text-4xl font-bold">
            Start from a template
          </h2>
          <p className="text-lg md:text-xl text-muted-foreground mt-4 max-w-2xl">
            These are the templates added most. Each one is copied into your
            bot, so you can change any block in the editor.
          </p>
        </div>
        <Button asChild variant="outline" className="flex-none">
          <Link href="/marketplace">
            Browse all {total} templates
            <ArrowRightIcon className="ml-2 h-4 w-4" />
          </Link>
        </Button>
      </div>

      {/* A row to swipe on phones, so six cards don't fill the screen. */}
      <div className="mt-8 md:mt-10 -mx-8 px-8 md:mx-0 md:px-0 flex gap-4 overflow-x-auto snap-x snap-mandatory pb-2 [scrollbar-width:none] md:grid md:grid-cols-2 lg:grid-cols-3 md:overflow-visible">
        {templates.map((template) => (
          <div
            key={template.id}
            className="flex w-[85%] max-w-sm flex-none snap-center md:w-auto [&>a]:w-full"
          >
            <FeaturedTemplateEntry
              template={template}
              href={templatePath(template.id)}
            />
          </div>
        ))}
      </div>
    </section>
  );
}
