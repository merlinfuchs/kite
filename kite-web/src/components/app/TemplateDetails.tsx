import TemplateIcon from "./TemplateIcon";
import {
  creatorPath,
  Template,
  TemplateCategory,
  templateCategories,
  relatedTemplates,
  templateCategory,
} from "@/lib/flow/templates";
import { discordEventType } from "@/lib/flow/eventTypes";
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  BadgeCheckIcon,
  FlagIcon,
  SatelliteDishIcon,
  ShieldAlertIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link, { LinkProps } from "next/link";
import { Separator } from "../ui/separator";
import { cn } from "@/lib/utils";
import {
  TemplateContents,
  TemplateInstallCount,
  TemplateListEntry,
} from "./TemplateListEntry";
import {
  FlowItemCard,
  MessageItemCard,
  PreviewItemsProvider,
} from "./ItemPreviewCards";
import BundleReportDialog from "./BundleReportDialog";
import TemplateDiscordPreview from "./TemplateDiscordPreview";
import dynamic from "next/dynamic";
import { communityTemplateFaq } from "@/lib/flow/templateFaq";
import { templatePermissionFaq } from "@/lib/flow/templatePermissions";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "../ui/accordion";

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
  aside,
  related,
}: {
  template: Template;
  // Bundles of others have no list to go back to.
  backHref?: LinkProps["href"];
  // The app has its own back button on mobile.
  backOnMobile?: boolean;
  aside: React.ReactNode;
  // Other templates to look at, see relatedTemplates.
  related?: {
    templates: Template[];
    href: (template: Template) => LinkProps["href"];
    // The page of the category, which only the public marketplace has.
    categoryHref?: (category: TemplateCategory) => LinkProps["href"];
  };
}) {
  const category = templateCategories.find(
    (c) => c.value === template.category
  );
  // The questions every listed template users made has go after its own.
  const baseFaq =
    template.author.kind === "user" && template.listed
      ? [...template.faq, ...communityTemplateFaq(template)]
      : template.faq;

  const permFaq = templatePermissionFaq(template);
  let faq = baseFaq;
  if (permFaq) {
    const existingIndex = baseFaq.findIndex((q) =>
      q.question.toLowerCase().includes("permission")
    );
    if (existingIndex !== -1) {
      faq = [
        ...baseFaq.slice(0, existingIndex),
        permFaq,
        ...baseFaq.slice(existingIndex + 1),
      ];
    } else {
      faq = [...baseFaq, permFaq];
    }
  }

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
        <div
          className={cn(
            "h-14 w-14 md:h-16 md:w-16 flex-none rounded-xl flex items-center justify-center",
            templateCategory(template).tile
          )}
        >
          <TemplateIcon
            name={template.icon}
            className={cn(
              "h-7 w-7 md:h-8 md:w-8",
              templateCategory(template).icon
            )}
          />
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
            <TemplateInstallCount installs={template.installs} />
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
          <section
            className={cn(
              "rounded-xl border bg-gradient-to-br to-transparent p-4 md:p-6",
              templateCategory(template).banner
            )}
          >
            <div className="max-w-md">
              <TemplateDiscordPreview template={template} rows={4} />
            </div>
          </section>

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
              Open one to see its blocks and what it costs.
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
            <TemplateCreditUsage template={template} />
          </section>

          {faq.length > 0 && (
            <section>
              <h2 className="font-semibold mb-1">Questions</h2>
              <Accordion type="single" collapsible>
                {faq.map((q, i) => (
                  <AccordionItem key={i} value={String(i)}>
                    <AccordionTrigger className="text-sm text-left">
                      {q.question}
                    </AccordionTrigger>
                    <AccordionContent className="text-sm text-muted-foreground whitespace-pre-wrap">
                      {q.answer}
                    </AccordionContent>
                  </AccordionItem>
                ))}
              </Accordion>
            </section>
          )}
        </div>

        {/* First on small screens, so adding doesn't need scrolling past every item. */}
        <div className="order-first lg:order-none lg:sticky lg:top-6">
          {aside}
        </div>
      </div>

      {related && <RelatedTemplates template={template} {...related} />}
    </div>
  );
}

function RelatedTemplates({
  template,
  templates,
  href,
  categoryHref,
}: {
  template: Template;
  templates: Template[];
  href: (template: Template) => LinkProps["href"];
  categoryHref?: (category: TemplateCategory) => LinkProps["href"];
}) {
  const { byCreator, similar } = relatedTemplates(template, templates);
  const category = templateCategory(template);

  const row = (title: string, list: Template[], more?: LinkProps["href"]) =>
    !!list.length && (
      <section>
        <div className="flex items-center justify-between gap-4 mb-4">
          <h2 className="font-semibold">{title}</h2>
          {more && (
            <Link
              href={more}
              className="flex items-center gap-1 text-sm text-primary hover:underline"
            >
              See all
              <ArrowRightIcon className="h-4 w-4" />
            </Link>
          )}
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {list.map((t) => (
            <TemplateListEntry key={t.id} template={t} href={href(t)} />
          ))}
        </div>
      </section>
    );

  if (!byCreator.length && !similar.length) return null;
  return (
    <>
      <Separator className="my-10" />
      <div className="space-y-10">
        {template.author.kind === "user" &&
          row(`More by ${template.author.name}`, byCreator)}
        {row(
          `More in ${category.label}`,
          similar,
          categoryHref?.(category.value)
        )}
      </div>
    </>
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
