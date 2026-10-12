import {
  useApp,
  useAppEntities,
  useAppStateGuilds,
  useUsageCredits,
} from "@/lib/hooks/api";
import { getAppInviteUrl } from "@/lib/discord/invite";
import { cn } from "@/lib/utils";
import { CheckIcon, XIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode, useEffect, useState } from "react";
import { Button } from "../ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";

// Only new apps get the checklist, older ones were set up already.
const maxAppAgeDays = 14;

export default function AppSetupChecklist() {
  const app = useApp();
  const entities = useAppEntities();
  const guilds = useAppStateGuilds();
  const credits = useUsageCredits();

  const storageKey = `kite-setup-checklist-dismissed-${app?.id}`;
  const [dismissed, setDismissed] = useState(true);
  useEffect(() => {
    if (app) setDismissed(localStorage.getItem(storageKey) === "true");
  }, [app, storageKey]);

  if (!app || !entities || dismissed) return null;

  const ageDays = (Date.now() - new Date(app.created_at).getTime()) / 86400000;
  if (ageDays > maxAppAgeDays) return null;

  const invited = !!guilds?.length;
  const hasCommand = entities.some(
    (e) => e?.type === "command" || e?.type === "event_listener"
  );
  // Running a command or event uses credits.
  const triedIt = !!credits?.credits_used;

  if (invited && hasCommand && triedIt) return null;

  function dismiss() {
    localStorage.setItem(storageKey, "true");
    setDismissed(true);
  }

  return (
    <Card className="mb-4 md:mb-8">
      <CardHeader className="relative">
        <Button
          size="icon"
          variant="ghost"
          className="absolute top-3 right-3 h-8 w-8 text-muted-foreground"
          onClick={dismiss}
          title="Hide"
          aria-label="Hide"
        >
          <XIcon className="h-4 w-4" />
        </Button>
        <CardTitle>Get your bot running</CardTitle>
        <CardDescription>
          Three steps to your first working command.
        </CardDescription>
      </CardHeader>
      <ol className="px-6 pb-6 grid gap-5 lg:grid-cols-3">
        <Step
          done={invited}
          title="Add your bot to a server"
          description="Pick a server where you can test it, ideally a small one."
        >
          <Button size="sm" variant="outline" asChild>
            <a
              href={getAppInviteUrl(app.discord_id)}
              target="_blank"
              rel="noreferrer"
            >
              Add to server
            </a>
          </Button>
        </Step>
        <Step
          done={hasCommand}
          title="Create a command"
          description="Build one from blocks, or import a template with ready-made commands."
        >
          <Button size="sm" variant="outline" asChild>
            <Link
              href={{
                pathname: "/apps/[appId]/commands",
                query: { appId: app.id },
              }}
            >
              Create command
            </Link>
          </Button>
          <Button size="sm" variant="outline" asChild>
            <Link
              href={{
                pathname: "/apps/[appId]/templates",
                query: { appId: app.id },
              }}
            >
              Browse templates
            </Link>
          </Button>
        </Step>
        <Step
          done={triedIt}
          title="Try it in Discord"
          description="Type / in your server and pick your command. Restart Discord if it doesn't show up yet."
        />
      </ol>
    </Card>
  );
}

function Step({
  done,
  title,
  description,
  children,
}: {
  done: boolean;
  title: string;
  description: string;
  children?: ReactNode;
}) {
  return (
    <li className="flex gap-3">
      <div
        className={cn(
          "flex-none h-6 w-6 rounded-full border-2 flex items-center justify-center",
          done ? "bg-primary border-primary text-primary-foreground" : ""
        )}
      >
        {done && <CheckIcon className="h-4 w-4" />}
      </div>
      <div className="space-y-2">
        <div>
          <div
            className={cn(
              "font-medium",
              done && "line-through text-muted-foreground"
            )}
          >
            {title}
          </div>
          {!done && (
            <div className="text-sm text-muted-foreground">{description}</div>
          )}
        </div>
        {!done && children && (
          <div className="flex flex-wrap gap-2">{children}</div>
        )}
      </div>
    </li>
  );
}
