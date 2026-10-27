import {
  creatorPath,
  Template,
  templateCategories,
} from "@/lib/flow/templates";
import { discordEventType } from "@/lib/flow/eventTypes";
import {
  ArrowLeftIcon,
  BadgeCheckIcon,
  FlagIcon,
  SatelliteDishIcon,
  ShieldAlertIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link, { LinkProps } from "next/link";
import { Separator } from "../ui/separator";
import { cn } from "@/lib/utils";
import { TemplateContents, TemplateInstallCount } from "./TemplateListEntry";
import {
  FlowItemCard,
  MessageItemCard,
  PreviewItemsProvider,
} from "./ItemPreviewCards";
import BundleReportDialog from "./BundleReportDialog";
import dynamic from "next/dynamic";

// Estimating credits needs the block definitions, which would make the page
// a lot bigger.
const TemplateCreditUsage = dynamic(() => import("./TemplateCreditUsage"), {
  ssr: false,
});
// It needs the block definitions too, and only bundles of users have it.
const BundleCapabilities = dynamic(() => import("./BundleCapabilities"), {
  ssr: false,
});

// TemplateDetails is the page of a template, with the card for adding it next
// to it.
export function TemplateDetails({
  template,
  backHref,
  backOnMobile,
  installs,
  aside,
}: {
  template: Template;
  // Bundles of others have no list to go back to.
  backHref?: LinkProps["href"];
  // The app has its own back button on mobile.
  backOnMobile?: boolean;
  installs?: number;
  aside: React.ReactNode;
}) {
  const category = templateCategories.find(
    (c) => c.value === template.category
  );

  return (
    <div>
      {backHref && (
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
      )}

      {template.author.kind === "user" && !template.listed && (
        <UserContentNotice madeBy={template.author.name} />
      )}

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
            {template.author.kind === "kite" ? (
              <span className="flex items-center gap-1">
                <BadgeCheckIcon className="h-4 w-4 text-primary" />
                Made by Kite
              </span>
            ) : template.author.kind === "user" ? (
              <span className="flex items-center gap-1.5">
                Made by
                {/* Only listed templates are on the page of their creator. */}
                {template.listed ? (
                  <Link
                    href={creatorPath(template.author.discordId)}
                    className="flex items-center gap-1.5 text-foreground hover:underline"
                  >
                    <TemplateAuthorName author={template.author} />
                  </Link>
                ) : (
                  <span
                    className="flex items-center gap-1.5 text-foreground"
                    title={`@${template.author.username} on Discord`}
                  >
                    <TemplateAuthorName author={template.author} />
                  </span>
                )}
              </span>
            ) : (
              <span>Made by you</span>
            )}
            {category && <span>{category.label}</span>}
            <TemplateContents template={template} />
            <TemplateInstallCount installs={installs} />
            {template.author.kind === "user" && (
              <BundleReportDialog bundleId={template.id} name={template.name}>
                <button className="flex items-center gap-1 hover:text-foreground">
                  <FlagIcon className="h-3.5 w-3.5" />
                  Report
                </button>
              </BundleReportDialog>
            )}
          </div>
        </div>
      </div>

      <Separator className="my-8" />

      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_340px] gap-8 items-start">
        <div className="space-y-8 min-w-0">
          {!!template.about.length && (
            <section>
              <h2 className="font-semibold mb-3">How it works</h2>
              <div className="space-y-3 text-sm text-muted-foreground max-w-2xl">
                {template.about.map((paragraph, i) => (
                  <p key={i}>{paragraph}</p>
                ))}
              </div>
            </section>
          )}

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

          {template.author.kind !== "kite" && (
            <BundleCapabilities template={template} />
          )}

          <section>
            <h2 className="font-semibold mb-1">What&apos;s inside</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Open one to see its blocks.
            </p>
            <PreviewItemsProvider
              variables={template.variables}
              messages={template.messages}
            >
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
                    name={
                      discordEventType(listener.type)?.label ?? listener.type
                    }
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
            </PreviewItemsProvider>
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

function TemplateAuthorName({
  author,
}: {
  author: { name: string; avatarUrl: string };
}) {
  return (
    <>
      <img src={author.avatarUrl} alt="" className="h-4 w-4 rounded-full" />
      {author.name}
    </>
  );
}

// UserContentNotice says that a bundle is made by a user. Kite only reviews
// the ones listed in the marketplace, so it can't vouch for the others.
function UserContentNotice({ madeBy }: { madeBy: string }) {
  return (
    <div className="flex gap-3 rounded-lg border border-amber-500/50 bg-amber-500/10 p-4 mb-8">
      <ShieldAlertIcon className="h-5 w-5 flex-none mt-0.5 text-amber-500" />
      <div className="text-sm">
        <div className="font-semibold mb-1">Made by {madeBy}, not by Kite</div>
        <p className="text-muted-foreground">
          This template is user content. Kite only reviews templates before
          listing them in the marketplace, and this one isn&apos;t listed, so
          Kite can&apos;t tell you whether it works or is safe, and {madeBy}{" "}
          alone is responsible for it. It runs with your bot and your credits
          once you add it, so read what it does first, and report it if it
          breaks the rules.
        </p>
      </div>
    </div>
  );
}
