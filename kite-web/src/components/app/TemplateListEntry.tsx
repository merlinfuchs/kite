import logo from "@/assets/logo/orange@1024.png";
import TemplateDiscordPreview from "./TemplateDiscordPreview";
import TemplateIcon from "./TemplateIcon";
import { Card } from "@/components/ui/card";
import { discordEventType, eventSourceName } from "@/lib/flow/eventTypes";
import {
  minShownStars,
  Template,
  templateCategory,
} from "@/lib/flow/templates";
import {
  BotIcon,
  CircleCheckIcon,
  MailIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
  SparklesIcon,
  StarIcon,
} from "lucide-react";
import Link, { LinkProps } from "next/link";
import { cn, formatNumber } from "@/lib/utils";

// TemplateListEntry links to the page of the template, or calls onClick
// instead, like to show it in a dialog.
export function TemplateListEntry({
  template,
  href,
  onClick,
  added,
}: {
  template: Template;
  added?: boolean;
} & (
  | { href: LinkProps["href"]; onClick?: never }
  | { href?: never; onClick: () => void }
)) {
  const card = (
    <Card className="h-full flex flex-col gap-3 p-4 transition-colors group-hover:border-primary/60 group-hover:bg-muted/30">
      <TemplateEntryHead template={template} added={added} />
      <div className="text-sm text-muted-foreground line-clamp-2">
        {template.description}
      </div>
      <TemplateEntryLine template={template} />
    </Card>
  );
  return href ? (
    <Link href={href} className="group">
      {card}
    </Link>
  ) : (
    <button type="button" className="group w-full text-left" onClick={onClick}>
      {card}
    </button>
  );
}

// FeaturedTemplateEntry is a bigger card for the templates the marketplace
// shows first, with what members see of it in Discord.
export function FeaturedTemplateEntry({
  template,
  href,
  added,
}: {
  template: Template;
  href: LinkProps["href"];
  added?: boolean;
}) {
  const category = templateCategory(template);

  return (
    <Link href={href} className="group">
      <Card className="h-full overflow-hidden flex flex-col transition-colors group-hover:border-primary/60">
        <div
          className={cn(
            "relative h-48 bg-gradient-to-br to-transparent px-4 pt-4 overflow-hidden",
            category.banner
          )}
        >
          <div className="transition-transform duration-300 group-hover:-translate-y-1">
            <TemplateDiscordPreview template={template} />
          </div>
        </div>
        <div className="flex-1 flex flex-col gap-3 p-4 border-t">
          <TemplateEntryHead template={template} added={added} />
          <div className="text-sm text-muted-foreground line-clamp-2">
            {template.description}
          </div>
        </div>
      </Card>
    </Link>
  );
}

function TemplateEntryHead({
  template,
  added,
}: {
  template: Template;
  added?: boolean;
}) {
  const category = templateCategory(template);
  const stars = template.stars ?? 0;

  return (
    <div className="flex items-center gap-3 min-w-0">
      <div
        className={cn(
          "h-10 w-10 flex-none rounded-lg flex items-center justify-center",
          category.tile
        )}
      >
        <TemplateIcon
          name={template.icon}
          className={cn("h-5 w-5", category.icon)}
        />
      </div>
      <div className="min-w-0 flex-1">
        <div className="font-semibold truncate">{template.name}</div>
        <TemplateAuthor template={template} />
      </div>
      {added ? (
        <span className="flex-none flex items-center gap-1 text-xs text-green-600 dark:text-green-500">
          <CircleCheckIcon className="h-3.5 w-3.5" />
          Added
        </span>
      ) : (
        <div className="flex-none flex items-center gap-2.5 text-xs text-muted-foreground">
          {stars >= minShownStars && (
            <span
              className="flex items-center gap-1"
              title={`Starred by ${formatNumber(stars)} users`}
            >
              <StarIcon className="h-3.5 w-3.5" />
              {formatNumber(stars)}
            </span>
          )}
          {!!template.installs && (
            <span
              className="flex items-center gap-1"
              title={`Added to ${formatNumber(template.installs)} bots`}
            >
              <BotIcon className="h-3.5 w-3.5" />
              {formatNumber(template.installs)}
            </span>
          )}
        </div>
      )}
      {template.aiUse && (
        <span
          className="flex-none flex items-center gap-1 text-xs text-amber-700 dark:text-amber-400"
          title={
            template.aiUse === "unknown"
              ? "The creator hasn't disclosed whether AI was used. External AI use can't be verified automatically."
              : template.aiUse === "generated"
              ? "AI created at least 80% of this template's blocks, based on creator disclosure and recorded Kite Flow AI edits. External AI use can't be verified automatically."
              : "AI contributed to some of this template's blocks, based on creator disclosure and recorded Kite Flow AI edits. External AI use can't be verified automatically."
          }
        >
          <SparklesIcon className="h-3.5 w-3.5" />
          {template.aiUse === "unknown"
            ? "AI use not disclosed"
            : `AI-${template.aiUse}`}
        </span>
      )}
    </div>
  );
}

// TemplateEntryLine is the first thing the template does, like its first
// command, in a line that looks like Discord's message box.
function TemplateEntryLine({ template }: { template: Template }) {
  const command = template.commands[0];
  const listener = template.eventListeners[0];
  const message = template.messages[0];
  const more =
    template.commands.length +
    template.eventListeners.length +
    template.messages.length -
    1;

  let content: React.ReactNode;
  if (command) {
    content = (
      <>
        <span className="font-mono text-foreground truncate flex-none max-w-[60%]">
          /{command.name}
        </span>
        <span className="text-xs text-muted-foreground truncate">
          {command.description}
        </span>
      </>
    );
  } else if (listener) {
    const event =
      listener.source === "discord"
        ? discordEventType(listener.type)
        : undefined;
    const Icon = event?.icon ?? eventSourceName(listener.source).icon;
    content = (
      <>
        <Icon className="h-3.5 w-3.5 text-muted-foreground flex-none" />
        <span className="text-xs text-muted-foreground truncate">
          {event
            ? `When ${event.description
                .charAt(0)
                .toLowerCase()}${event.description.slice(1)}`
            : eventSourceName(listener.source).title}
        </span>
      </>
    );
  } else if (message) {
    content = (
      <>
        <MailIcon className="h-3.5 w-3.5 text-muted-foreground flex-none" />
        <span className="text-xs text-muted-foreground truncate">
          Posts the {message.name} message
        </span>
      </>
    );
  } else {
    return null;
  }

  return (
    <div className="mt-auto rounded-md bg-muted/60 px-3 py-2 text-sm flex items-center gap-2 min-w-0">
      {content}
      {more > 0 && (
        <span className="ml-auto pl-1 text-xs text-muted-foreground flex-none">
          +{more}
        </span>
      )}
    </div>
  );
}

function TemplateAuthor({ template }: { template: Template }) {
  const { author } = template;
  if (author.kind === "kite") {
    return (
      <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <img src={logo.src} alt="" className="h-3.5 w-3.5" />
        Kite
      </span>
    );
  }
  if (author.kind === "user") {
    return (
      <span className="flex items-center gap-1.5 text-xs text-muted-foreground min-w-0">
        <img
          src={author.avatarUrl}
          alt=""
          className="h-3.5 w-3.5 rounded-full flex-none"
        />
        <span className="truncate">{author.name}</span>
      </span>
    );
  }
  return <span className="text-xs text-muted-foreground">Made by you</span>;
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

export function TemplateInstallCount({ installs }: { installs?: number }) {
  if (!installs) {
    return null;
  }

  return (
    <span className="flex items-center gap-1">
      <BotIcon className="h-3.5 w-3.5" />
      Added to {formatNumber(installs)} {installs === 1 ? "bot" : "bots"}
    </span>
  );
}
