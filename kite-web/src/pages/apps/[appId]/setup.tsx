import { getKiteBotSetupLayout } from "@/components/app/KiteBotSetupShell";
import TemplateIcon from "@/components/app/TemplateIcon";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import BaseLayout from "@/components/common/BaseLayout";
import LoadingButton from "@/components/common/LoadingButton";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { track } from "@/lib/analytics";
import { useBundleInstallCreateMutation } from "@/lib/api/mutations";
import { installRequest, installSelection } from "@/lib/flow/templateInstall";
import { Template } from "@/lib/flow/templates";
import { useApp, useAppStateGuilds, useBundleInstalls } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useFullTemplates, useTemplates } from "@/lib/hooks/templates";
import { cn } from "@/lib/utils";
import { CheckIcon, CircleAlertIcon, CircleCheckIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

type Added = { name: string; unposted: { name: string; error: string }[] };

// The page apps on the Kite bot start on, after adding the bot. People first
// pick templates, then fill in the settings of each one on its own. The query
// can pick templates and their settings, like
// ?templates=verification,giveaway&verification.role_id=123 from Xenon.
export default function AppSetupPage() {
  const appId = useAppId();
  const app = useApp();
  const guild = useAppStateGuilds()?.[0];
  const installs = useBundleInstalls();
  // The query is only there once the router is ready, after the first render.
  const router = useRouter();
  const { query } = router;

  // Custom bots add templates from the Marketplace.
  const custom = !!app && app.bot_type !== "shared";
  useEffect(() => {
    if (custom) {
      router.replace({
        pathname: "/apps/[appId]/marketplace",
        query: { appId },
      });
    }
  }, [custom, router, appId]);

  // In the order of the Marketplace.
  const templates = useTemplates();
  const shared = useMemo(
    () => (templates ?? []).filter((t) => t.sharedBot),
    [templates]
  );
  // With their flows, to see which have settings and to add them.
  const full = useFullTemplates(shared.map((t) => t.id));
  const addedIds = useMemo(
    () => new Set((installs ?? []).flatMap((i) => (i ? [i.bundle_id] : []))),
    [installs]
  );

  const [picked, setPicked] = useState<Set<string> | null>(null);
  useEffect(() => {
    if (picked || !router.isReady || !installs) return;
    const fromLink = String(query.templates ?? "")
      .split(",")
      .filter((id) => id && !addedIds.has(id));
    setPicked(new Set(fromLink));
  }, [picked, router.isReady, installs, query.templates, addedIds]);

  const [queue, setQueue] = useState<Template[] | null>(null);
  const [index, setIndex] = useState(0);
  const [added, setAdded] = useState<Added[]>([]);
  const [starting, setStarting] = useState(false);
  const installMutation = useBundleInstallCreateMutation(appId);

  // The settings of a template in the query, by parameter key.
  function initialValues(template: Template) {
    const values: Record<string, string> = {};
    for (const p of template.parameters) {
      const value = query[`${template.id}.${p.key}`];
      if (typeof value === "string") values[p.key] = value;
    }
    return values;
  }

  // Templates without settings are added right away, the others get a step
  // each.
  async function start() {
    if (!full || !picked) return;
    setStarting(true);
    const withSettings: Template[] = [];
    const addedNow: Added[] = [];
    try {
      for (const template of full.filter((t) => picked.has(t.id))) {
        const selection = installSelection(template, undefined, new Set());
        if (selection.parameters.length || selection.customizable.length) {
          withSettings.push(template);
          continue;
        }

        const res = await installMutation.mutateAsync(
          installRequest(template, selection, {}, {})
        );
        if (!res.success) {
          toast.error(
            `Failed to add ${template.name}: ${res.error.message} (${res.error.code})`
          );
          continue;
        }
        track("template_added", {
          source: "template",
          template: template.id,
          template_version: template.version,
          count: res.data.items.length,
          first: true,
          from: "setup",
        });
        addedNow.push({ name: template.name, unposted: [] });
      }
    } finally {
      setStarting(false);
    }
    setAdded(addedNow);
    setIndex(0);
    setQueue(withSettings);
  }

  if (!router.isReady || !app || custom || !picked) {
    return (
      <BaseLayout title="Set up the Kite bot" noindex>
        {null}
      </BaseLayout>
    );
  }

  if (queue && index < queue.length) {
    const template = queue[index]!;
    const next = () => setIndex(index + 1);
    return (
      <BaseLayout title={`Set up ${template.name}`} noindex>
        <SetupStep
          key={template.id}
          template={template}
          step={index + 1}
          total={queue.length}
          initialValues={initialValues(template)}
          onInstalled={(unposted) => {
            setAdded((a) => [...a, { name: template.name, unposted }]);
            next();
          }}
          onSkip={next}
        />
      </BaseLayout>
    );
  }

  if (queue) {
    return (
      <BaseLayout title="The Kite bot is set up" noindex>
        <SetupDone added={added} onBack={() => setQueue(null)} />
      </BaseLayout>
    );
  }

  return (
    <BaseLayout title="Set up the Kite bot" noindex>
      <div className="mb-6">
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          What should the Kite bot do?
        </h1>
        <p className="text-muted-foreground text-sm">
          It&apos;s in {guild?.name ?? "your server"}. Pick the templates it
          should run there. You fill in their settings next and can add more
          later.
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        {/* The ones that can still be picked first. */}
        {[
          ...shared.filter((t) => !addedIds.has(t.id)),
          ...shared.filter((t) => addedIds.has(t.id)),
        ].map((template) => {
          const isAdded = addedIds.has(template.id);
          const selected = picked.has(template.id);
          return (
            <button
              key={template.id}
              type="button"
              disabled={isAdded}
              aria-pressed={selected}
              onClick={() => {
                const nextPicked = new Set(picked);
                if (selected) nextPicked.delete(template.id);
                else nextPicked.add(template.id);
                setPicked(nextPicked);
              }}
              className={cn(
                "flex flex-col gap-2 rounded-xl border bg-card p-4 text-left transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring",
                selected
                  ? "border-primary ring-1 ring-primary"
                  : "hover:border-foreground/30",
                isAdded && "opacity-60 cursor-default hover:border-border"
              )}
            >
              <div className="flex items-center gap-3">
                <TemplateIcon
                  name={template.icon}
                  className="h-6 w-6 flex-none text-muted-foreground"
                />
                <div className="font-medium flex-1">{template.name}</div>
                {isAdded ? (
                  <span className="text-xs text-muted-foreground">Added</span>
                ) : (
                  <div
                    className={cn(
                      "h-5 w-5 flex-none rounded-full border flex items-center justify-center",
                      selected &&
                        "border-primary bg-primary text-primary-foreground"
                    )}
                  >
                    {selected && <CheckIcon className="h-3.5 w-3.5" />}
                  </div>
                )}
              </div>
              <div className="text-sm text-muted-foreground">
                {template.description}
              </div>
            </button>
          );
        })}
      </div>

      <div className="flex flex-col sm:flex-row sm:items-center gap-3 mt-6">
        <LoadingButton
          onClick={start}
          disabled={!picked.size || !full}
          loading={starting}
        >
          {picked.size
            ? `Set up ${picked.size} ${
                picked.size === 1 ? "template" : "templates"
              }`
            : "Pick a template"}
        </LoadingButton>
        <Button variant="ghost" asChild>
          <Link href={{ pathname: "/apps/[appId]", query: { appId } }}>
            {addedIds.size ? "Go to the dashboard" : "Skip for now"}
          </Link>
        </Button>
      </div>
    </BaseLayout>
  );
}

// SetupStep is the settings of one picked template.
function SetupStep({
  template,
  step,
  total,
  initialValues,
  onInstalled,
  onSkip,
}: {
  template: Template;
  step: number;
  total: number;
  initialValues: Record<string, string>;
  onInstalled: (unposted: { name: string; error: string }[]) => void;
  onSkip: () => void;
}) {
  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center justify-between text-sm text-muted-foreground mb-2">
          <span>
            {step} of {total}
          </span>
          <button
            type="button"
            onClick={onSkip}
            className="hover:text-foreground"
          >
            Skip
          </button>
        </div>
        <Progress value={((step - 1) / total) * 100} className="h-1.5" />
      </div>

      <div className="flex items-start gap-4">
        <TemplateIcon
          name={template.icon}
          className="h-8 w-8 flex-none text-muted-foreground mt-1"
        />
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            {template.name}
          </h1>
          <p className="text-muted-foreground text-sm">
            {template.description}
          </p>
        </div>
      </div>

      {!!template.about.length && (
        <div className="space-y-3 text-sm text-muted-foreground">
          {template.about.map((paragraph, i) => (
            <p key={i}>{paragraph}</p>
          ))}
        </div>
      )}

      <TemplateInstallCard
        template={template}
        from="setup"
        initialValues={initialValues}
        title="Settings"
        actionLabel={step < total ? "Add and continue" : "Add and finish"}
        onInstalled={onInstalled}
      />
    </div>
  );
}

function SetupDone({ added, onBack }: { added: Added[]; onBack: () => void }) {
  const appId = useAppId();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          {added.length ? "The Kite bot is set up" : "Nothing was added"}
        </h1>
        <p className="text-muted-foreground text-sm">
          {added.length
            ? "It runs these templates in your server now. You can change their settings on the dashboard any time."
            : "You skipped every template. Pick some again, or add them later from the dashboard."}
        </p>
      </div>

      {!!added.length && (
        <div className="space-y-3">
          {added.map((a) => (
            <div key={a.name} className="space-y-1">
              <div className="flex items-center gap-2 font-medium">
                <CircleCheckIcon className="h-5 w-5 text-green-500" />
                {a.name}
              </div>
              {a.unposted.map((m) => (
                <div
                  key={m.name}
                  className="flex gap-2 text-sm pl-7 text-amber-600 dark:text-amber-500"
                >
                  <CircleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
                  <span>
                    Couldn&apos;t post the {m.name} message: {m.error}. Post it
                    again from the template&apos;s page on the dashboard once
                    that&apos;s fixed.
                  </span>
                </div>
              ))}
            </div>
          ))}
        </div>
      )}

      <div className="flex flex-col sm:flex-row gap-3">
        <Button asChild>
          <Link href={{ pathname: "/apps/[appId]", query: { appId } }}>
            Go to the dashboard
          </Link>
        </Button>
        {!added.length && (
          <Button variant="ghost" onClick={onBack}>
            Pick templates
          </Button>
        )}
      </div>
    </div>
  );
}

AppSetupPage.getLayout = getKiteBotSetupLayout;
