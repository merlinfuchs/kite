import { Template, templateCategories } from "@/lib/flow/templates";
import { discordEventType } from "@/lib/flow/eventTypes";
import { FlowContextType } from "@/lib/flow/context";
import { useAppId } from "@/lib/hooks/params";
import {
  ArrowLeftIcon,
  BadgeCheckIcon,
  ChevronDownIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";
import { Card } from "../ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "../ui/collapsible";
import { Separator } from "../ui/separator";
import { cn } from "@/lib/utils";
import TemplateFlowPreview from "./TemplateFlowPreview";
import { TemplateContents } from "./TemplateListEntry";
import { TemplateInstallCard } from "./TemplateInstallCard";

export function TemplateDetails({ template }: { template: Template }) {
  const appId = useAppId();
  const category = templateCategories.find(
    (c) => c.value === template.category
  );

  return (
    <div>
      <Link
        href={{ pathname: "/apps/[appId]/templates", query: { appId } }}
        className="hidden md:inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground mb-6"
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
                <TemplateItem
                  key={command.name}
                  icon={SlashSquareIcon}
                  name={`/${command.name}`}
                  description={command.description}
                  context="command"
                  flow={command.flowSource}
                />
              ))}
              {template.eventListeners.map((listener, i) => (
                <TemplateItem
                  key={i}
                  icon={SatelliteDishIcon}
                  name={discordEventType(listener.type)?.label ?? listener.type}
                  description={listener.description}
                  context="event_discord"
                  flow={listener.flowSource}
                />
              ))}
            </div>
          </section>
        </div>

        {/* First on small screens, so adding doesn't need scrolling past every item. */}
        <div className="order-first lg:order-none lg:sticky lg:top-6">
          <TemplateInstallCard template={template} />
        </div>
      </div>
    </div>
  );
}

function TemplateItem({
  icon: Icon,
  name,
  description,
  context,
  flow,
}: {
  icon: React.ComponentType<{ className?: string }>;
  name: string;
  description: string;
  context: FlowContextType;
  flow: Template["commands"][number]["flowSource"];
}) {
  const [open, setOpen] = useState(false);

  // The preview shows the flow without the values of the settings.
  const { nodes, edges } = useMemo(() => flow({}), [flow]);

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <Card className="overflow-hidden">
        <CollapsibleTrigger className="w-full flex items-center gap-3 p-4 text-left hover:bg-muted/40 transition-colors">
          <Icon className="h-5 w-5 text-muted-foreground flex-none" />
          <div className="min-w-0 flex-1">
            <div className="font-medium">{name}</div>
            <div className="text-sm text-muted-foreground">{description}</div>
          </div>
          <span className="text-xs text-muted-foreground flex-none hidden sm:block">
            {nodes.length} blocks
          </span>
          <ChevronDownIcon
            className={cn(
              "h-4 w-4 text-muted-foreground flex-none transition-transform",
              open && "rotate-180"
            )}
          />
        </CollapsibleTrigger>
        <CollapsibleContent className="px-4 pb-4">
          <TemplateFlowPreview nodes={nodes} edges={edges} context={context} />
        </CollapsibleContent>
      </Card>
    </Collapsible>
  );
}
