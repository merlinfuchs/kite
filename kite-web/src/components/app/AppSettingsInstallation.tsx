import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useAppInstallationUpdateMutation } from "@/lib/api/mutations";
import { useAppInstallationQuery } from "@/lib/api/queries";
import {
  decodePermissionsBitset,
  PermissionCategory,
  permissionBits,
} from "@/lib/discord/permissions";
import { defaultGuildPermissions, getAppInviteUrl } from "@/lib/discord/invite";
import { useApp } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { AppInstallParams } from "@/lib/types/wire.gen";
import CopyButton from "@/tools/common/components/CopyButton";
import {
  ExternalLinkIcon,
  ServerIcon,
  TriangleAlertIcon,
  UserIcon,
} from "lucide-react";
import Link from "next/link";
import { ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

// Commands need this scope in both contexts, and installing to a server
// always adds the bot.
const commandsScope = "applications.commands";
const guildScopes = [commandsScope, "bot"];

const administratorBit = 3;

const categories = (
  ["General", "Voice", "Text", "Members", "Advanced"] as PermissionCategory[]
).map((name) => ({
  name,
  permissions: permissionBits.filter((p) => p.category === name),
}));

interface FormState {
  guildEnabled: boolean;
  guildPermissions: string;
  userEnabled: boolean;
}

// Discord sends null for a context that is turned off.
function toFormState(
  guild: AppInstallParams | null | undefined,
  user: AppInstallParams | null | undefined
): FormState {
  return {
    guildEnabled: !!guild,
    // Turning on server install starts with the default permissions.
    guildPermissions: guild ? guild.permissions : defaultGuildPermissions,
    userEnabled: !!user,
  };
}

export default function AppSettingsInstallation() {
  const app = useApp();
  const appId = useAppId();

  const installationQuery = useAppInstallationQuery(appId);
  const installation = installationQuery.data?.success
    ? installationQuery.data.data
    : undefined;
  const loadError =
    installationQuery.data && !installationQuery.data.success
      ? installationQuery.data.error.message
      : undefined;

  const saved = useMemo(
    () =>
      installation
        ? toFormState(installation.guild_install, installation.user_install)
        : undefined,
    [installation]
  );

  const [form, setForm] = useState<FormState>();
  useEffect(() => {
    if (saved) setForm(saved);
  }, [saved]);

  const updateMutation = useAppInstallationUpdateMutation(appId);

  // Server installs set up in the Discord Developer Portal may not add the bot
  // yet. Saving here adds it.
  const missingBot =
    !!installation?.guild_install &&
    !installation.guild_install.scopes.includes("bot");

  const dirty =
    !!form &&
    !!saved &&
    (JSON.stringify(form) !== JSON.stringify(saved) ||
      (missingBot && form.guildEnabled));
  const noneEnabled = !!form && !form.guildEnabled && !form.userEnabled;

  const update = useCallback(
    (changes: Partial<FormState>) =>
      setForm((f) => (f ? { ...f, ...changes } : f)),
    []
  );

  const save = useCallback(() => {
    if (!form) return;

    updateMutation.mutate(
      {
        guild_install: form.guildEnabled
          ? {
              scopes: guildScopes,
              permissions: form.guildPermissions,
            }
          : undefined,
        user_install: form.userEnabled
          ? { scopes: [commandsScope], permissions: "0" }
          : undefined,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Installation settings saved!");
          } else if (res.error.code === "validation_failed") {
            const reason =
              Object.values(res.error.data ?? {})[0] ?? res.error.message;
            toast.error(`Failed to save installation settings: ${reason}`);
          } else {
            toast.error(
              `Failed to save installation settings: ${res.error.message} (${res.error.code})`
            );
          }
        },
        onError(err) {
          toast.error(`Failed to save installation settings: ${err.message}`);
        },
      }
    );
  }, [form, updateMutation]);

  const permissionCount = useMemo(
    () => decodePermissionsBitset(form?.guildPermissions || "0").length,
    [form?.guildPermissions]
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Installation</CardTitle>
        <CardDescription>
          Choose where people can add your app in Discord, and what it asks for
          when they do.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {loadError ? (
          <div className="text-sm text-destructive">
            Failed to load the installation settings: {loadError}
          </div>
        ) : !form ? (
          <div className="text-sm text-muted-foreground">Loading...</div>
        ) : (
          <>
            <InstallContext
              icon={<ServerIcon className="h-5 w-5" />}
              title="Server install"
              description="Server admins can add your bot to their servers."
              enabled={form.guildEnabled}
              onEnabledChange={(v) => update({ guildEnabled: v })}
              link={
                app && installation?.guild_install
                  ? getAppInviteUrl(
                      app.discord_id,
                      "guild",
                      installation.guild_install
                    )
                  : undefined
              }
            >
              {missingBot && (
                <div className="text-sm text-muted-foreground">
                  Your server install doesn&apos;t add the bot yet. Save to add
                  it, so your bot can join servers and receive their events.
                </div>
              )}

              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <div className="text-sm font-medium">Bot permissions</div>
                  <div className="text-sm text-muted-foreground">
                    {permissionCount === 0
                      ? "No permissions"
                      : `${permissionCount} ${
                          permissionCount === 1 ? "permission" : "permissions"
                        }`}
                    {" · "}
                    <span className="font-mono">{form.guildPermissions}</span>
                  </div>
                </div>
                <PermissionsDialog
                  value={form.guildPermissions}
                  onChange={(v) => update({ guildPermissions: v })}
                />
              </div>
            </InstallContext>

            <InstallContext
              icon={<UserIcon className="h-5 w-5" />}
              title="Account install"
              description="People can add your app to their account and use its commands anywhere, even in servers your bot isn't in. Your bot can only reply there."
              enabled={form.userEnabled}
              onEnabledChange={(v) => update({ userEnabled: v })}
              link={
                app && installation?.user_install
                  ? getAppInviteUrl(
                      app.discord_id,
                      "user",
                      installation.user_install
                    )
                  : undefined
              }
            />

            {noneEnabled && (
              <div className="text-sm text-destructive">
                Turn on at least one way to install your app.
              </div>
            )}
          </>
        )}
      </CardContent>
      <CardFooter className="flex flex-wrap gap-3 border-t px-6 py-4">
        <Button
          onClick={save}
          disabled={!dirty || noneEnabled || updateMutation.isPending}
        >
          Save installation
        </Button>
        {dirty && (
          <Button variant="ghost" onClick={() => saved && setForm(saved)}>
            Reset
          </Button>
        )}
        <Button variant="outline" className="flex gap-2" asChild>
          <Link
            href={`https://discord.com/developers/applications/${app?.discord_id}/installation`}
            target="_blank"
          >
            <div>Manage on Discord</div>
            <ExternalLinkIcon className="w-4 h-4" />
          </Link>
        </Button>
      </CardFooter>
    </Card>
  );
}

function InstallContext({
  icon,
  title,
  description,
  enabled,
  onEnabledChange,
  link,
  children,
}: {
  icon: ReactNode;
  title: string;
  description: string;
  enabled: boolean;
  onEnabledChange: (enabled: boolean) => void;
  link?: string;
  children?: ReactNode;
}) {
  return (
    <div className="rounded-lg border">
      <div className="flex items-start gap-4 p-4">
        <div className="flex h-9 w-9 flex-none items-center justify-center rounded-md bg-muted text-muted-foreground">
          {icon}
        </div>
        <div className="min-w-0 flex-1">
          <div className="font-medium">{title}</div>
          <div className="text-sm text-muted-foreground">{description}</div>
        </div>
        <Switch
          checked={enabled}
          onCheckedChange={onEnabledChange}
          aria-label={title}
        />
      </div>
      {enabled && (children || link) && (
        <div className="space-y-4 border-t p-4">
          {children}
          {link && (
            <div className="space-y-2">
              <div className="text-sm font-medium">Install link</div>
              <div className="flex gap-3">
                <Input readOnly value={link} />
                <CopyButton value={link} />
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function AdministratorWarning() {
  return (
    <div className="flex gap-2 rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm">
      <TriangleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
      <div>
        Administrator gives the bot every permission and lets it ignore channel
        overwrites. Only ask for it if your bot really needs it.
      </div>
    </div>
  );
}

function PermissionsDialog({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState(value);

  useEffect(() => {
    if (open) setDraft(value);
  }, [open, value]);

  const enabled = useMemo(
    () => new Set<number>(decodePermissionsBitset(draft).map((p) => p.bit)),
    [draft]
  );

  // Changes only this bit, so bits of permissions we don't know yet survive.
  function toggle(bit: number, checked: boolean) {
    const bits = BigInt(draft || "0");
    const mask = BigInt(1) << BigInt(bit);
    setDraft((checked ? bits | mask : bits & ~mask).toString());
  }

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
        Choose permissions
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Bot permissions</DialogTitle>
            <DialogDescription>
              The permissions your bot asks for when it&apos;s added to a
              server. Server admins can still change them.
            </DialogDescription>
          </DialogHeader>

          {enabled.has(administratorBit) && <AdministratorWarning />}

          <div className="md:columns-2 gap-4">
            {categories.map((category) => (
              <div
                key={category.name}
                className="rounded-lg bg-muted/50 p-4 mb-4 break-inside-avoid"
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="font-semibold">{category.name}</div>
                  <span className="text-xs text-muted-foreground">
                    {
                      category.permissions.filter((p) => enabled.has(p.bit))
                        .length
                    }{" "}
                    of {category.permissions.length}
                  </span>
                </div>
                <div className="space-y-2">
                  {category.permissions.map((permission) => (
                    <div
                      key={permission.bit}
                      className="flex items-center gap-2"
                    >
                      <Checkbox
                        id={`install-permission-${permission.bit}`}
                        checked={enabled.has(permission.bit)}
                        onCheckedChange={(v) =>
                          toggle(permission.bit, v === true)
                        }
                      />
                      <Label
                        htmlFor={`install-permission-${permission.bit}`}
                        className="font-normal"
                      >
                        {permission.label}
                      </Label>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>

          <DialogFooter className="gap-2 sm:gap-0">
            <Button
              variant="ghost"
              onClick={() => setDraft(defaultGuildPermissions)}
            >
              Reset to Administrator
            </Button>
            <Button
              onClick={() => {
                // No permissions fall back to the default, Administrator.
                onChange(
                  draft && draft !== "0" ? draft : defaultGuildPermissions
                );
                setOpen(false);
              }}
            >
              Done
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
