import { Template, templateCategories } from "@/lib/flow/templates";
import { discordEventType } from "@/lib/flow/eventTypes";
import { FlowContextType } from "@/lib/flow/context";
import {
  ArrowLeftIcon,
  MailIcon,
  BadgeCheckIcon,
  ChevronDownIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link, { LinkProps } from "next/link";
import { useMemo, useState } from "react";
import { Card } from "../ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "../ui/collapsible";
import { Separator } from "../ui/separator";
import { cn, formatNumber } from "@/lib/utils";
import {
  formatFlowCredits,
  getFlowCredits,
  getTemplateCredits,
  TemplateFlowCredits,
} from "@/lib/flow/credits";
import env from "@/lib/env/client";
import { Message, messageSchema } from "@/lib/message/schema";
import { useTheme } from "next-themes";
import { TemplateContents } from "./TemplateListEntry";
import dynamic from "next/dynamic";

// The previews pull in the flow editor and the message renderer. They only
// show when an item is opened, so they're loaded then.
const TemplateFlowPreview = dynamic(() => import("./TemplateFlowPreview"), {
  ssr: false,
});
const MessagePreview = dynamic(() => import("../message/MessagePreview"), {
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
              {template.messages.map((message) => (
                <TemplateMessageItem key={message.key} message={message} />
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

function TemplateItemCard({
  icon: Icon,
  name,
  description,
  meta,
  children,
}: {
  icon: React.ComponentType<{ className?: string }>;
  name: string;
  description: string;
  meta: string;
  children: React.ReactNode;
}) {
  const [open, setOpen] = useState(false);

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
            {meta}
          </span>
          <ChevronDownIcon
            className={cn(
              "h-4 w-4 text-muted-foreground flex-none transition-transform",
              open && "rotate-180"
            )}
          />
        </CollapsibleTrigger>
        <CollapsibleContent className="px-4 pb-4">
          {children}
        </CollapsibleContent>
      </Card>
    </Collapsible>
  );
}

function TemplateItem({
  icon,
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
  // The preview shows the flow without the values of the settings.
  const { nodes, edges } = useMemo(() => flow({}), [flow]);
  const credits = useMemo(() => getFlowCredits(nodes, edges), [nodes, edges]);

  return (
    <TemplateItemCard
      icon={icon}
      name={name}
      description={description}
      meta={`${nodes.length} blocks · ${formatFlowCredits(credits)}`}
    >
      <TemplateFlowPreview nodes={nodes} edges={edges} context={context} />
    </TemplateItemCard>
  );
}

function TemplateMessageItem({
  message,
}: {
  message: Template["messages"][number];
}) {
  // The theme that's shown, which theme isn't when it follows the system.
  const { resolvedTheme } = useTheme();
  const data = useMemo(() => {
    const res = messageSchema.safeParse(message.data({}));
    return res.success ? res.data : null;
  }, [message]);
  const flows = useMemo(
    () => Object.entries(message.flowSources({})),
    [message]
  );

  return (
    <TemplateItemCard
      icon={MailIcon}
      name={message.name}
      description={message.description ?? "A message template."}
      meta={`${flows.length} ${flows.length === 1 ? "button" : "buttons"}`}
    >
      <div className="space-y-4">
        {data && (
          // The preview has no padding of its own, and its margins would
          // leave gaps above and below its background.
          <div className="rounded-md border overflow-hidden [&_.discord-messages]:flow-root [&_.discord-messages]:px-4">
            <MessagePreview msg={data} lightTheme={resolvedTheme === "light"} />
          </div>
        )}
        {flows.map(([id, flow]) => (
          <div key={id}>
            <div className="text-sm font-medium mb-2">
              When {componentLabel(data, id) ?? "a button"} is clicked
            </div>
            <TemplateFlowPreview
              nodes={flow.nodes}
              edges={flow.edges}
              context="component_button"
            />
          </div>
        ))}
      </div>
    </TemplateItemCard>
  );
}

// componentLabel is the label of the button or select menu that runs a flow.
function componentLabel(data: Message | null, flowSourceId: string) {
  let label: string | undefined;
  const walk = (value: unknown) => {
    if (label || !value || typeof value !== "object") return;
    const component = value as Record<string, unknown>;
    if (component.flow_source_id === flowSourceId) {
      label = String(component.label ?? component.placeholder ?? "");
      return;
    }
    Object.values(component).forEach(walk);
  };
  walk(data);
  return label ? `"${label}"` : undefined;
}

// Runs a month the credits of each flow are shown for. Small screens only
// have room for the middle one.
const creditVolumes = [100, 500, 2000];
const volumeClassName = (i: number) => (i === 1 ? "" : "hidden sm:table-cell");

// TemplateCreditUsage shows what each flow of a template costs per run and
// at a few levels of activity, so people know before adding it.
function TemplateCreditUsage({ template }: { template: Template }) {
  const rows = useMemo(() => getTemplateCredits(template), [template]);

  return (
    <section>
      <h2 className="font-semibold mb-1">Credit usage</h2>
      <p className="text-sm text-muted-foreground mb-4 max-w-2xl">
        The most credits each part costs when it runs. Reading data, conditions
        and loops are free, and the free plan includes 10,000 credits a month.{" "}
        <Link
          href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system`}
          target="_blank"
          className="underline hover:text-foreground"
        >
          About credits
        </Link>
      </p>
      <Card className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b text-left text-muted-foreground">
              <th className="font-medium px-4 py-2.5">Part</th>
              <th className="font-medium px-4 py-2.5 text-right whitespace-nowrap">
                Per run
              </th>
              {creditVolumes.map((v, i) => (
                <th
                  key={v}
                  className={cn(
                    "font-medium px-4 py-2.5 text-right whitespace-nowrap",
                    volumeClassName(i)
                  )}
                >
                  {formatNumber(v)} runs
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.key} className="border-b last:border-0">
                <td className="px-4 py-2.5">
                  <div className="font-medium">
                    {creditRowName(template, row)}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {creditRowTrigger(row)}
                  </div>
                </td>
                <td className="px-4 py-2.5 text-right whitespace-nowrap">
                  {row.credits.max}
                </td>
                {creditVolumes.map((v, i) => (
                  <td
                    key={v}
                    className={cn(
                      "px-4 py-2.5 text-right whitespace-nowrap text-muted-foreground",
                      volumeClassName(i)
                    )}
                  >
                    {formatNumber(row.credits.max * v)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </section>
  );
}

function creditRowName(template: Template, row: TemplateFlowCredits) {
  switch (row.kind) {
    case "command":
      return `/${row.name}`;
    case "event":
      return discordEventType(row.eventType)?.label ?? row.eventType;
    case "button": {
      const message = template.messages.find((m) => m.key === row.messageKey);
      const res = messageSchema.safeParse(message?.data({}));
      return `${
        componentLabel(res.success ? res.data : null, row.flowSourceId) ?? "A"
      } button`;
    }
  }
}

function creditRowTrigger(row: TemplateFlowCredits) {
  switch (row.kind) {
    case "command":
      return "Someone uses the command";
    case "event":
      return (
        discordEventType(row.eventType)?.description ?? "The event happens"
      );
    case "button":
      return "Someone clicks the button";
  }
}
