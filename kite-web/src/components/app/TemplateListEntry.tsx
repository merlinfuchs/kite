import { Card } from "@/components/ui/card";
import { Template, templateCategories } from "@/lib/flow/templates";
import {
  BotIcon,
  CircleCheckIcon,
  MailIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
} from "lucide-react";
import Link, { LinkProps } from "next/link";
import { formatNumber } from "@/lib/utils";

export function TemplateListEntry({
  template,
  href,
  added,
  installs,
}: {
  template: Template;
  href: LinkProps["href"];
  added?: boolean;
  installs?: number;
}) {
  const category = templateCategories.find(
    (c) => c.value === template.category
  );

  return (
    <Link href={href} className="group">
      <Card className="h-full flex flex-col p-5 transition-colors group-hover:border-primary/60 group-hover:bg-muted/30">
        <div className="flex items-start justify-between gap-3 mb-4">
          <div className="h-11 w-11 bg-primary/15 flex-none rounded-lg flex items-center justify-center">
            <template.icon className="w-6 h-6 text-primary" />
          </div>
          {added ? (
            <span className="flex items-center gap-1 text-xs text-green-600 dark:text-green-500">
              <CircleCheckIcon className="h-3.5 w-3.5" />
              Added
            </span>
          ) : (
            category && (
              <span className="text-xs text-muted-foreground">
                {category.label}
              </span>
            )
          )}
        </div>
        <div className="font-semibold mb-1">{template.name}</div>
        <div className="text-sm text-muted-foreground line-clamp-2 mb-4">
          {template.description}
        </div>
        <div className="mt-auto flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
          {template.author.kind === "you" && <span>Made by you</span>}
          {template.author.kind === "user" && (
            <span className="flex items-center gap-1.5">
              <img
                src={template.author.avatarUrl}
                alt=""
                className="h-3.5 w-3.5 rounded-full"
              />
              {template.author.name}
            </span>
          )}
          <TemplateContents template={template} />
          <TemplateInstallCount installs={installs} />
        </div>
      </Card>
    </Link>
  );
}

export function TemplateContents({ template }: { template: Template }) {
  const commands = template.commands.length;
  const eventListeners = template.eventListeners.length;
  const messages = template.messages.length;

  return (
    <>
      {!!commands && (
        <span className="flex items-center gap-1">
          <SlashSquareIcon className="h-3.5 w-3.5" />
          {commands} {commands === 1 ? "command" : "commands"}
        </span>
      )}
      {!!eventListeners && (
        <span className="flex items-center gap-1">
          <SatelliteDishIcon className="h-3.5 w-3.5" />
          {eventListeners}{" "}
          {eventListeners === 1 ? "event listener" : "event listeners"}
        </span>
      )}
      {!!messages && (
        <span className="flex items-center gap-1">
          <MailIcon className="h-3.5 w-3.5" />
          {messages} {messages === 1 ? "message" : "messages"}
        </span>
      )}
    </>
  );
}

// Fewer installs look like nobody uses a template, so they aren't shown.
const minShownInstalls = 10;

export function TemplateInstallCount({ installs }: { installs?: number }) {
  if (!installs || installs < minShownInstalls) {
    return null;
  }

  return (
    <span className="flex items-center gap-1">
      <BotIcon className="h-3.5 w-3.5" />
      Added to {formatNumber(installs)} bots
    </span>
  );
}
