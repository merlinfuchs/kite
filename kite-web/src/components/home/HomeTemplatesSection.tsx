import { FeaturedTemplateEntry } from "@/components/app/TemplateListEntry";
import { Button } from "@/components/ui/button";
import { templatePath, templateFromBundle } from "@/lib/flow/templates";
import { Bundle } from "@/lib/types/wire.gen";
import { ArrowRightIcon, StoreIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";

// HomeTemplatesSection shows a few templates of the marketplace, picked by
// the page like the featured ones there, so visitors see what they can start
// from. It's left out when the page was built without them.
export default function HomeTemplatesSection({
  bundles,
  total,
}: {
  bundles: Bundle[] | null;
  // How many templates the marketplace has.
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
            Add a ready-made bot to your server in one click, then change
            anything in the editor.
          </p>
        </div>
        <Button asChild variant="outline" className="flex-none hidden md:flex">
          <Link href="/marketplace">
            <StoreIcon className="mr-2 h-4 w-4" />
            Browse all {total} templates
          </Link>
        </Button>
      </div>

      {/* A row to swipe on phones, a grid from md up. */}
      <div className="mt-8 md:mt-10 -mx-4 px-4 sm:mx-0 sm:px-0 flex gap-4 overflow-x-auto snap-x snap-mandatory pb-2 [scrollbar-width:none] md:grid md:grid-cols-3 md:overflow-visible">
        {templates.map((template) => (
          <div
            key={template.id}
            className="flex w-[85%] flex-none snap-center md:w-auto [&>a]:w-full"
          >
            <FeaturedTemplateEntry
              template={template}
              href={templatePath(template.id)}
            />
          </div>
        ))}
      </div>

      <Button asChild variant="outline" className="mt-6 w-full md:hidden">
        <Link href="/marketplace">
          Browse all {total} templates
          <ArrowRightIcon className="ml-2 h-4 w-4" />
        </Link>
      </Button>
    </section>
  );
}
