import { useAppSetupQuery } from "@/lib/api/queries";
import { getAppInviteUrl } from "@/lib/discord/invite";
import { useApp } from "@/lib/hooks/api";
import { useSetupChecklistStore } from "@/lib/state/setupChecklist";
import { App } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { differenceInDays } from "date-fns";
import { CheckIcon, PartyPopperIcon, XIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode, useEffect, useRef, useState } from "react";
import { useTrack } from "@/lib/analytics";
import { Button } from "../ui/button";
import { Card } from "../ui/card";

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

  // Only celebrate if the user finished a step while watching, not for apps
  // that were set up already.
  const [sawIncomplete, setSawIncomplete] = useState(false);

  const query = useAppSetupQuery(app.id);
  const setup = query.data?.success ? query.data.data : null;

  const steps = setup
    ? [setup.invited, setup.has_command, setup.ran_command]
    : [];
  const doneCount = steps.filter(Boolean).length;
  const done = !!setup && doneCount === steps.length;
  const current = steps.indexOf(false);

  // Steps the user completes while the checklist is shown.
  const track = useTrack();
  const previous = useRef(setup);
  useEffect(() => {
    const before = previous.current;
    previous.current = setup;
    if (!before || !setup) return;
    for (const step of ["invited", "has_command", "ran_command"] as const) {
      if (!before[step] && setup[step]) track("setup_step_done", { step });
    }
  }, [setup, track]);

  useEffect(() => {
    if (!setup) return;
    if (!done) {
      setSawIncomplete(true);
    } else if (!sawIncomplete) {
      hide(app.id);
    }
  }, [setup, done, sawIncomplete, hide, app.id]);

  if (!setup || (done && !sawIncomplete)) return null;

  if (done) {
    return (
      <Card className="mb-4 md:mb-8 p-5 flex items-center gap-4">
        <PartyPopperIcon className="h-6 w-6 flex-none text-primary" />
        <div className="flex-auto">
          <div className="font-semibold">Your bot is up and running</div>
          <div className="text-sm text-muted-foreground">
            Someone ran your first command in Discord.
          </div>
        </div>
        <DismissButton onClick={() => hide(app.id)} />
      </Card>
    );
  }

  return (
    <Card className="mb-4 md:mb-8 overflow-hidden">
      <div className="px-5 pt-5 flex items-center gap-4">
        <div className="flex-auto">
          <div className="font-semibold">Get your bot running</div>
          <div className="text-sm text-muted-foreground">
            {doneCount} of {steps.length} done
          </div>
        </div>
        <DismissButton
          onClick={() => {
            track("setup_checklist_dismissed", { done_count: doneCount });
            hide(app.id);
          }}
        />
      </div>
      <div className="mx-5 mt-3 h-1.5 rounded-full bg-muted overflow-hidden">
        <div
          className="h-full bg-primary transition-all duration-500"
          style={{ width: `${(doneCount / steps.length) * 100}%` }}
        />
      </div>
      <ol className="p-2 pt-3">
        <Step
          index={0}
          current={current}
          done={setup.invited}
          title="Add your bot to a server"
          description="Pick a server where you can test it."
        >
          <Button size="sm" asChild>
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
          index={1}
          current={current}
          done={setup.has_command}
          title="Create a command"
          description="Build one from blocks or ask the AI to build it for you."
        >
          <Button size="sm" asChild>
            <Link
              href={{
                pathname: "/apps/[appId]/commands",
                query: { appId: app.id },
              }}
            >
              Create command
            </Link>
          </Button>
          <Button size="sm" variant="ghost" asChild>
            <Link
              href={{
                pathname: "/apps/[appId]/templates",
                query: { appId: app.id },
              }}
            >
              Start from a template
            </Link>
          </Button>
        </Step>
        <Step
          index={2}
          current={current}
          done={setup.ran_command}
          title="Try it in Discord"
          description="Type / in your server and pick your command. This ticks off once it ran."
        >
          <Button size="sm" asChild>
            <a
              href="https://discord.com/channels/@me"
              target="_blank"
              rel="noreferrer"
            >
              Open Discord
            </a>
          </Button>
        </Step>
      </ol>
    </Card>
  );
}

function DismissButton({ onClick }: { onClick: () => void }) {
  return (
    <Button
      size="icon"
      variant="ghost"
      className="h-8 w-8 flex-none text-muted-foreground"
      onClick={onClick}
      title="Hide"
      aria-label="Hide"
    >
      <XIcon className="h-4 w-4" />
    </Button>
  );
}

function Step({
  index,
  current,
  done,
  title,
  description,
  children,
}: {
  index: number;
  current: number;
  done: boolean;
  title: string;
  description: string;
  children: ReactNode;
}) {
  const isCurrent = index === current;

  return (
    <li
      className={cn(
        "flex gap-3 rounded-md px-3 py-2.5",
        isCurrent && "bg-muted/60"
      )}
    >
      <div
        className={cn(
          "flex-none h-6 w-6 rounded-full flex items-center justify-center text-xs font-semibold",
          done
            ? "bg-primary text-primary-foreground"
            : isCurrent
            ? "border-2 border-primary text-primary"
            : "border-2 text-muted-foreground"
        )}
      >
        {done ? <CheckIcon className="h-4 w-4" /> : index + 1}
      </div>
      <div className="min-w-0 pt-0.5">
        <div
          className={cn(
            "text-sm font-medium",
            !isCurrent && "text-muted-foreground"
          )}
        >
          {title}
        </div>
        {isCurrent && (
          <>
            <div className="text-sm text-muted-foreground mt-0.5">
              {description}
            </div>
            <div className="flex flex-wrap gap-2 mt-3">{children}</div>
          </>
        )}
      </div>
    </li>
  );
}
