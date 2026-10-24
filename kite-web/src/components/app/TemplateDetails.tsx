import { Template, templateCategories } from "@/lib/flow/templates";
import { discordEventType } from "@/lib/flow/eventTypes";
import {
  ArrowLeftIcon,
  BadgeCheckIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link, { LinkProps } from "next/link";
import { Separator } from "../ui/separator";
import { cn } from "@/lib/utils";
import { TemplateContents } from "./TemplateListEntry";
import { FlowItemCard, MessageItemCard } from "./ItemPreviewCards";
import dynamic from "next/dynamic";

// Estimating credits needs the block definitions, which would make the page
// a lot bigger.
const TemplateCreditUsage = dynamic(() => import("./TemplateCreditUsage"), {
  ssr: false,
});

// TemplateDetails is the page of a template, with the card for adding it next
// to it.
export function TemplateDetails({
  template,
  backHref,
  backOnMobile,
  aside,
}: {
  template: Template;
  backHref: LinkProps["href"];
  // The app has its own back button on mobile.
  backOnMobile?: boolean;
  aside: React.ReactNode;
}) {
  const category = templateCategories.find(
    (c) => c.value === template.category
  );

  return (
    <div>
      <Link
        href={backHref}
        className={cn(
          "items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground mb-6",
          backOnMobile ? "inline-flex" : "hidden md:inline-flex"
        )}
      >
        <ArrowLeftIcon className="h-4 w-4" />
        All templates
      </Link>

      <div className="flex gap-5 items-start">
        <div className="h-14 w-14 md:h-16 md:w-16 bg-primary/15 flex-none rounded-xl flex items-center justify-center">
          <template.icon className="h-7 w-7 md:h-8 md:w-8 text-primary" />
        </div>
        <div className="min-w-0">
          <h1 className="text-xl font-semibold md:text-2xl mb-1">
            {template.name}
          </h1>
          <p className="text-muted-foreground text-sm md:text-base max-w-2xl">
            {template.description}
          </p>
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 mt-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <BadgeCheckIcon className="h-4 w-4 text-primary" />
              Made by Kite
            </span>
            {category && <span>{category.label}</span>}
            <TemplateContents template={template} />
          </div>
        </div>
      </div>

      <Separator className="my-8" />

      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_340px] gap-8 items-start">
        <div className="space-y-8 min-w-0">
          <section>
            <h2 className="font-semibold mb-3">How it works</h2>
            <div className="space-y-3 text-sm text-muted-foreground max-w-2xl">
              {template.about.map((paragraph, i) => (
                <p key={i}>{paragraph}</p>
              ))}
            </div>
          </section>

          {!!template.setup?.length && (
            <section>
              <h2 className="font-semibold mb-3">Before you start</h2>
              <ol className="space-y-2">
                {template.setup.map((step, i) => (
                  <li key={i} className="flex gap-3 text-sm">
                    <span className="h-5 w-5 flex-none rounded-full bg-muted text-xs font-medium flex items-center justify-center">
                      {i + 1}
                    </span>
                    {step}
                  </li>
                ))}
              </ol>
            </section>
          )}

          <section>
            <h2 className="font-semibold mb-1">What&apos;s inside</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Open one to see its blocks.
            </p>
            <div className="space-y-3">
              {template.commands.map((command) => (
                <FlowItemCard
                  key={command.name}
                  icon={SlashSquareIcon}
                  name={`/${command.name}`}
                  description={command.description}
                  context="command"
                  flow={command.flowSource}
                />
              ))}
              {template.eventListeners.map((listener, i) => (
                <FlowItemCard
                  key={i}
                  icon={SatelliteDishIcon}
                  name={discordEventType(listener.type)?.label ?? listener.type}
                  description={listener.description}
                  context="event_discord"
                  flow={listener.flowSource}
                />
              ))}
              {template.messages.map((message) => (
                <MessageItemCard
                  key={message.key}
                  name={message.name}
                  description={message.description}
                  data={message.data}
                  flowSources={message.flowSources}
                />
              ))}
            </div>
          </section>

          <TemplateCreditUsage template={template} />
        </div>

        {/* First on small screens, so adding doesn't need scrolling past every item. */}
        <div className="order-first lg:order-none lg:sticky lg:top-6">
          {aside}
        </div>
      </div>
    </div>
  );
}
