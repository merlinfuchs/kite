import { setupSteps, useAppSetupQuery } from "@/lib/api/queries";
import { getAppInviteUrl } from "@/lib/discord/invite";
import { Template } from "@/lib/flow/templates";
import { useApp } from "@/lib/hooks/api";
import { useTemplates } from "@/lib/hooks/templates";
import { useSetupChecklistStore } from "@/lib/state/setupChecklist";
import { App } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { differenceInDays } from "date-fns";
import { CheckIcon, PartyPopperIcon, XIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode, useEffect, useRef, useState } from "react";
import { track } from "@/lib/analytics";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import TemplateDialog from "./TemplateDialog";
import { TemplateListEntry } from "./TemplateListEntry";

const maxAppAgeDays = 14;

// The templates offered for getting started, ones that work without an API key.
const setupTemplateIds = ["moderation", "ask-ai", "welcomer", "verification"];

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

  // Kept after it's added, when the checklist moves on to the next step.
  const [template, setTemplate] = useState<Template | null>(null);
  const [templateOpen, setTemplateOpen] = useState(false);

  const steps = setup ? setupSteps.map((s) => setup[s]) : [];
  const doneCount = steps.filter(Boolean).length;
  const done = !!setup && doneCount === steps.length;
  const current = steps.indexOf(false);

  // Steps the user completes while the checklist is shown.
  const previous = useRef(setup);
  useEffect(() => {
    const before = previous.current;
    previous.current = setup;
    if (!before || !setup) return;
    for (const step of setupSteps) {
      if (!before[step] && setup[step]) track("setup_step_done", { step });
    }
  }, [setup]);

  useEffect(() => {
    if (!setup) return;
    if (!done) {
      setSawIncomplete(true);
    } else if (!sawIncomplete) {
      hide(app.id);
    }
  }, [setup, done, sawIncomplete, hide, app.id]);

  // Outside the checklist, so it stays open when adding a template finishes it.
  const dialog = template && (
    <TemplateDialog
      template={template}
      from="setup_checklist"
      open={templateOpen}
      onOpenChange={setTemplateOpen}
    />
  );

  if (!setup || (done && !sawIncomplete)) return dialog;

  if (done) {
    return (
      <>
        {dialog}
        <Card className="mb-4 md:mb-8 p-5 flex items-center gap-4">
          <PartyPopperIcon className="h-6 w-6 flex-none text-primary" />
          <div className="flex-auto">
            <div className="font-semibold">Your bot is up and running</div>
            <div className="text-sm text-muted-foreground">
              It ran for the first time in Discord.
            </div>
          </div>
          <DismissButton onClick={() => hide(app.id)} />
        </Card>
      </>
    );
  }

  return (
    <>
      {dialog}
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
            done={setup.has_flow}
            title="Give your bot something to do"
            description="Start from a template, or build your own from blocks with the help of AI."
          >
            <SetupTemplates
              onOpen={(t) => {
                setTemplate(t);
                setTemplateOpen(true);
              }}
            />
            <Button size="sm" variant="outline" asChild>
              <Link
                href={{
                  pathname: "/apps/[appId]/commands",
                  query: { appId: app.id },
                }}
              >
                Build your own
              </Link>
            </Button>
            <Button size="sm" variant="ghost" asChild>
              <Link
                href={{
                  pathname: "/apps/[appId]/marketplace",
                  query: { appId: app.id },
                }}
              >
                Browse all templates
              </Link>
            </Button>
          </Step>
          <Step
            index={2}
            current={current}
            done={setup.ran_flow}
            title="Try it in Discord"
            description="Use what you added, like typing / and picking a command. This ticks off once it ran."
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
    </>
  );
}

// SetupTemplates are only rendered on their step, so the marketplace is only
// loaded then.
function SetupTemplates({ onOpen }: { onOpen: (template: Template) => void }) {
  const templates = useTemplates();
  const setupTemplates = setupTemplateIds.flatMap(
    (id) => templates?.find((t) => t.id === id) ?? []
  );
  if (!setupTemplates.length) return null;

  return (
    <div className="w-full grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-4">
      {setupTemplates.map((t) => (
        <TemplateListEntry key={t.id} template={t} onClick={() => onOpen(t)} />
      ))}
    </div>
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
      <div className="min-w-0 flex-1 pt-0.5">
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
