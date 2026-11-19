import { getKiteBotSetupLayout } from "@/components/app/KiteBotSetupShell";
import TemplateIcon from "@/components/app/TemplateIcon";
import {
  KiteBotMemberPreview,
  useKiteBotLook,
} from "@/components/app/KiteBotLook";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import BaseLayout from "@/components/common/BaseLayout";
import LoadingButton from "@/components/common/LoadingButton";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { track } from "@/lib/analytics";
import { useAppStateGuildProfileUpdateMutation } from "@/lib/api/mutations";
import { useAddTemplate } from "@/lib/hooks/addTemplate";
import { profileImageTypes, readProfileImage } from "@/lib/profileImage";
import { installSelection } from "@/lib/flow/templateInstall";
import {
  formatParameterValue,
  parameterDefaults,
  Template,
} from "@/lib/flow/templates";
import {
  useApp,
  useAppStateGuildChannels,
  useAppStateGuildRoles,
  useAppStateGuilds,
  useBundleInstalls,
} from "@/lib/hooks/api";
import { Channel, Role } from "@/lib/types/wire.gen";
import { useAppId } from "@/lib/hooks/params";
import { useFullTemplates, useTemplates } from "@/lib/hooks/templates";
import { cn } from "@/lib/utils";
import { CheckIcon, CircleAlertIcon, CircleCheckIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

type Added = { name: string; unposted: { name: string; error: string }[] };

type Step = { kind: "look" } | { kind: "template"; template: Template };

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

  // Counts coming back from Discord once, and takes the marker out of the
  // URL, so reloading doesn't count it again.
  const tracked = useRef(false);
  useEffect(() => {
    const installed = query.installed;
    if (!router.isReady || !installed || tracked.current) return;
    tracked.current = true;
    // The service only passes on a valid ref. It's the link's, which can
    // differ from the app's for an app that already existed.
    const { installed: _, ...rest } = query;
    track("kite_bot_added", {
      new: installed === "new",
      ...(typeof query.ref === "string" && { ref: query.ref }),
    });
    router.replace({ pathname: router.pathname, query: rest }, undefined, {
      shallow: true,
    });
  }, [router, query]);

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

  // The steps after picking: the bot's look, then the templates with settings.
  const [queue, setQueue] = useState<Step[] | null>(null);
  const [index, setIndex] = useState(0);
  const [added, setAdded] = useState<Added[]>([]);
  const [starting, setStarting] = useState(false);
  const { add } = useAddTemplate();
  const guildId = app?.discord_guild_id ?? null;
  const channels = useAppStateGuildChannels(guildId);
  const roles = useAppStateGuildRoles(guildId);

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

        const unposted = await add({
          template,
          selection,
          values: {},
          guildId,
          first: true,
          from: "setup",
        });
        if (unposted) addedNow.push({ name: template.name, unposted });
      }
    } finally {
      setStarting(false);
    }
    setAdded(addedNow);
    setIndex(0);
    setQueue([
      { kind: "look" },
      ...withSettings.map((template) => ({
        kind: "template" as const,
        template,
      })),
    ]);
  }

  // A link that sets up every template it picks, like from Xenon after a
  // server template, needs no steps. The templates are added at once after a
  // look at their settings, unless the user wants to change them.
  const [changing, setChanging] = useState(false);
  const linked = String(query.templates ?? "").length > 0;
  const review = useMemo(() => {
    if (!full || !channels || !roles || !picked?.size || !linked) {
      return undefined;
    }
    const reviewed = full.filter((t) => picked.has(t.id));
    const complete = reviewed.every((template) => {
      const selection = installSelection(template, undefined, new Set());
      const values = initialValues(template);
      return (
        !selection.customizable.length &&
        selection.parameters.every((p) => {
          const value = values[p.key];
          if (!value) return !p.required;
          if (p.type === "channel" || p.type === "category") {
            return channels.some((c) => c?.id === value);
          }
          if (p.type === "role") return roles.some((r) => r?.id === value);
          return true;
        })
      );
    });
    return complete ? reviewed : undefined;
    // initialValues only reads the query.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [full, channels, roles, picked, linked, query]);

  async function addAll() {
    if (!review) return;
    setStarting(true);
    const addedNow: Added[] = [];
    try {
      for (const template of review) {
        const unposted = await add({
          template,
          selection: installSelection(template, undefined, new Set()),
          values: {
            ...parameterDefaults(template.parameters),
            ...initialValues(template),
          },
          guildId,
          first: true,
          from: "setup_link",
        });
        if (unposted) addedNow.push({ name: template.name, unposted });
      }
    } finally {
      setStarting(false);
    }
    setAdded(addedNow);
    setIndex(0);
    setQueue([{ kind: "look" }]);
  }

  if (!router.isReady || !app || custom || !picked) {
    return (
      <BaseLayout title="Set up the Kite bot" noindex>
        {null}
      </BaseLayout>
    );
  }

  if (queue && index < queue.length) {
    const step = queue[index]!;
    const next = () => setIndex(index + 1);
    const progress = { step: index + 1, total: queue.length, onSkip: next };
    if (step.kind === "look") {
      return (
        <BaseLayout title="Make the Kite bot yours" noindex>
          <SetupProgress {...progress} />
          <SetupLook
            guildId={app.discord_guild_id ?? ""}
            guildName={guild?.name}
            last={queue.length === 1}
            onDone={next}
          />
        </BaseLayout>
      );
    }

    const template = step.template;
    return (
      <BaseLayout title={`Set up ${template.name}`} noindex>
        <SetupProgress {...progress} />
        <SetupStep
          key={template.id}
          template={template}
          last={index + 1 === queue.length}
          initialValues={initialValues(template)}
          onInstalled={(unposted) => {
            setAdded((a) => [...a, { name: template.name, unposted }]);
            next();
          }}
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

  if (review && !changing) {
    return (
      <BaseLayout title="Set up the Kite bot" noindex>
        <SetupReview
          templates={review}
          values={initialValues}
          guildName={guild?.name}
          channels={channels ?? []}
          roles={roles ?? []}
          adding={starting}
          onAdd={addAll}
          onChange={() => setChanging(true)}
        />
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

// SetupReview lists the templates of a link with their settings, to add them
// all at once.
function SetupReview({
  templates,
  values,
  guildName,
  channels,
  roles,
  adding,
  onAdd,
  onChange,
}: {
  templates: Template[];
  values: (template: Template) => Record<string, string>;
  guildName?: string;
  channels: (Channel | undefined)[];
  roles: (Role | undefined)[];
  adding: boolean;
  onAdd: () => void;
  onChange: () => void;
}) {
  function shown(type: string, value: string) {
    if (type === "channel" || type === "category") {
      const channel = channels.find((c) => c?.id === value);
      return channel ? `#${channel.name}` : value;
    }
    if (type === "role") {
      const role = roles.find((r) => r?.id === value);
      return role ? `@${role.name}` : value;
    }
    return formatParameterValue(type, value);
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Your templates are ready
        </h1>
        <p className="text-muted-foreground text-sm">
          They&apos;re already set up for {guildName ?? "your server"}. Check
          their settings, then add them all at once.
        </p>
      </div>

      <div className="space-y-3">
        {templates.map((template) => {
          const set = values(template);
          const settings = template.parameters.flatMap((p) =>
            set[p.key]
              ? [`${p.label || p.key}: ${shown(p.type, set[p.key]!)}`]
              : []
          );
          return (
            <Card key={template.id} className="flex items-start gap-4 p-4">
              <TemplateIcon
                name={template.icon}
                className="h-6 w-6 flex-none text-muted-foreground mt-0.5"
              />
              <div className="min-w-0">
                <div className="font-medium">{template.name}</div>
                <div className="text-sm text-muted-foreground">
                  {settings.length ? settings.join(", ") : template.description}
                </div>
              </div>
            </Card>
          );
        })}
      </div>

      <div className="flex flex-col sm:flex-row sm:items-center gap-3">
        <LoadingButton onClick={onAdd} loading={adding}>
          {templates.length === 1
            ? `Add ${templates[0]!.name}`
            : `Add all ${templates.length} templates`}
        </LoadingButton>
        <Button variant="ghost" onClick={onChange}>
          Change settings
        </Button>
      </div>
    </div>
  );
}

function SetupProgress({
  step,
  total,
  onSkip,
}: {
  step: number;
  total: number;
  onSkip: () => void;
}) {
  return (
    <div className="mb-6">
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
  );
}

// SetupLook gives the Kite bot a name and avatar in the server, which it
// shows right away.
function SetupLook({
  guildId,
  guildName,
  last,
  onDone,
}: {
  guildId: string;
  guildName?: string;
  last: boolean;
  onDone: () => void;
}) {
  const appId = useAppId();
  const current = useKiteBotLook();
  const mutation = useAppStateGuildProfileUpdateMutation(appId);
  const [nick, setNick] = useState("");
  const [avatar, setAvatar] = useState<string>();

  async function save() {
    if (!nick && !avatar) {
      onDone();
      return;
    }
    const res = await mutation.mutateAsync({
      guildId,
      nick: nick || undefined,
      avatar,
    });
    if (!res.success) {
      toast.error(
        `Failed to change the look: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    track("kite_bot_look_changed", { nick: !!nick, avatar: !!avatar });
    onDone();
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Make it yours
        </h1>
        <p className="text-muted-foreground text-sm">
          Give the Kite bot your own name and avatar in{" "}
          {guildName ?? "your server"}. Members see them in the member list and
          on everything it posts there. You can change them on the dashboard any
          time.
        </p>
      </div>

      <KiteBotMemberPreview
        name={nick || current.name}
        avatarUrl={avatar || current.avatarUrl}
      />

      <Card className="p-5 space-y-4">
        <div className="space-y-2">
          <Label htmlFor="look-nick">Name</Label>
          <Input
            id="look-nick"
            value={nick}
            placeholder={current.name}
            maxLength={32}
            onChange={(e) => setNick(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="look-avatar">Avatar</Label>
          <Input
            id="look-avatar"
            type="file"
            accept={profileImageTypes}
            onChange={async (e) => {
              const file = e.target.files?.[0];
              const image = file && (await readProfileImage(file));
              if (!image) e.target.value = "";
              setAvatar(image);
            }}
          />
        </div>
        <LoadingButton
          className="w-full"
          onClick={save}
          loading={mutation.isPending}
        >
          {last ? "Save and finish" : "Save and continue"}
        </LoadingButton>
      </Card>
    </div>
  );
}

// SetupStep is the settings of one picked template.
function SetupStep({
  template,
  last,
  initialValues,
  onInstalled,
}: {
  template: Template;
  last: boolean;
  initialValues: Record<string, string>;
  onInstalled: (unposted: { name: string; error: string }[]) => void;
}) {
  const about = template.sharedBotAbout ?? template.about;

  return (
    <div className="space-y-6">
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

      {!!about.length && (
        <div className="space-y-3 text-sm text-muted-foreground">
          {about.map((paragraph, i) => (
            <p key={i}>{paragraph}</p>
          ))}
        </div>
      )}

      <TemplateInstallCard
        template={template}
        from="setup"
        initialValues={initialValues}
        title="Settings"
        actionLabel={last ? "Add and finish" : "Add and continue"}
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
