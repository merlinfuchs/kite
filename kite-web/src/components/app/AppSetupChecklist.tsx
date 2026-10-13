import {
  useAppEntitiesQuery,
  useAppStateGuildsQuery,
  useUsageCreditsQuery,
} from "@/lib/api/queries";
import { getAppInviteUrl } from "@/lib/discord/invite";
import { useApp } from "@/lib/hooks/api";
import { useSetupChecklistStore } from "@/lib/state/setupChecklist";
import { App } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { useQueryClient } from "@tanstack/react-query";
import { differenceInDays } from "date-fns";
import { CheckIcon, XIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode, useEffect } from "react";
import { Button } from "../ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";

const maxAppAgeDays = 14;

export default function AppSetupChecklist() {
  const app = useApp();
  const hidden = useSetupChecklistStore((s) =>
    app ? s.hiddenAppIds.includes(app.id) : true
  );

  if (!app || hidden || !app.enabled) return null;
  if (differenceInDays(new Date(), new Date(app.created_at)) > maxAppAgeDays) {
    return null;
  }

  return <Checklist app={app} />;
}

function Checklist({ app }: { app: App }) {
  const hide = useSetupChecklistStore((s) => s.hide);
  const client = useQueryClient();

  const guilds = useAppStateGuildsQuery(app.id);
  const entities = useAppEntitiesQuery(app.id);
  const credits = useUsageCreditsQuery(app.id);

  // Steps are done in Discord or on other pages, so check again when the user
  // comes back.
  useEffect(() => {
    function refresh() {
      for (const key of [
        ["state", "guilds"],
        ["entities"],
        ["usage", "credits"],
      ]) {
        client.invalidateQueries({ queryKey: ["apps", app.id, ...key] });
      }
    }

    refresh();
    window.addEventListener("focus", refresh);
    return () => window.removeEventListener("focus", refresh);
  }, [client, app.id]);

  const invited = !!(guilds.data?.success && guilds.data.data.length);
  const hasCommand = !!(
    entities.data?.success &&
    entities.data.data.some((e) => e?.type === "command")
  );
  // Running a command uses credits.
  const triedIt = !!(
    credits.data?.success && credits.data.data.credits_used > 0
  );
  const done = invited && hasCommand && triedIt;

  useEffect(() => {
    if (done) hide(app.id);
  }, [done, hide, app.id]);

  if (guilds.isPending || entities.isPending || credits.isPending || done) {
    return null;
  }

  return (
    <Card className="mb-4 md:mb-8">
      <CardHeader className="relative">
        <Button
          size="icon"
          variant="ghost"
          className="absolute top-3 right-3 h-8 w-8 text-muted-foreground"
          onClick={() => hide(app.id)}
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
          description="Pick a server where you can test it."
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
          description="Type / in your server and pick your command."
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
          done && "bg-primary border-primary text-primary-foreground"
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
