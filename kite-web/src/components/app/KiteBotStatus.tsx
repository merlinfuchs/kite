import { installUrl } from "@/lib/api/client";
import { useAppStateGuildProfileQuery } from "@/lib/api/queries";
import { PermissionFlag } from "@/lib/discord/permissions";
import { Template } from "@/lib/flow/templates";
import {
  missingPermissions,
  permissionLabel,
  templatePermissions,
} from "@/lib/flow/templatePermissions";
import {
  useApp,
  useAppFeature,
  useAppStateGuildRoles,
  useAppStateGuilds,
  useLogSummary,
  useResponseData,
  useUsageCredits,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { BundleInstall } from "@/lib/types/wire.gen";
import { cn, formatMonthlyResetDate, formatNumber } from "@/lib/utils";
import {
  CircleCheckIcon,
  OctagonAlertIcon,
  TriangleAlertIcon,
} from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Progress } from "../ui/progress";
import GetCreditsButton from "./GetCreditsButton";

// Credits stand out from this share on, like in AppCreditsPopup.
const warnAt = 0.8;

const list = (items: string[]) => new Intl.ListFormat("en").format(items);

// KiteBotStatus tells whether the Kite bot can run the templates in the
// server: that it's there with what they need, recent errors, and credits.
export default function KiteBotStatus({
  added,
}: {
  added: { install: BundleInstall; template: Template }[];
}) {
  return (
    <Card className="divide-y">
      <BotRow added={added} />
      <ErrorsRow />
      <CreditsRow />
    </Card>
  );
}

function StatusRow({
  state,
  title,
  children,
  action,
}: {
  state: "ok" | "warn" | "error";
  title: string;
  children?: ReactNode;
  action?: ReactNode;
}) {
  const Icon =
    state === "ok"
      ? CircleCheckIcon
      : state === "warn"
      ? TriangleAlertIcon
      : OctagonAlertIcon;

  return (
    <div className="flex flex-col sm:flex-row sm:items-center gap-3 p-4">
      <div className="flex flex-1 min-w-0 gap-3">
        <Icon
          className={cn(
            "h-5 w-5 flex-none mt-0.5",
            state === "ok" && "text-green-500",
            state === "warn" && "text-amber-500",
            state === "error" && "text-red-500"
          )}
        />
        <div className="min-w-0 flex-1">
          <div className="font-medium">{title}</div>
          {children && (
            <div className="text-sm text-muted-foreground space-y-1 mt-0.5">
              {children}
            </div>
          )}
        </div>
      </div>
      {action && (
        <div className="flex-none sm:ml-auto pl-8 sm:pl-0">{action}</div>
      )}
    </div>
  );
}

function BotRow({
  added,
}: {
  added: { install: BundleInstall; template: Template }[];
}) {
  const appId = useAppId();
  const app = useApp();
  const guilds = useAppStateGuilds();
  const guild = guilds?.[0];
  const roles = useAppStateGuildRoles(guild?.id ?? null);
  const profile = useResponseData(
    useAppStateGuildProfileQuery(appId, guild?.id ?? "")
  );

  if (!app || !guilds) return null;

  if (!guild) {
    return (
      <StatusRow
        state="error"
        title="The Kite bot isn't in your server"
        action={
          <Button size="sm" asChild>
            <a
              href={installUrl(
                app.discord_guild_id ? `guild_id=${app.discord_guild_id}` : ""
              )}
            >
              Add it again
            </a>
          </Button>
        }
      >
        Your templates don&apos;t run until it&apos;s back.
      </StatusRow>
    );
  }

  if (!app.enabled) {
    return (
      <StatusRow
        state="error"
        title="Your app is stopped"
        action={
          <Button size="sm" variant="outline" asChild>
            <Link
              href={{ pathname: "/apps/[appId]/settings", query: { appId } }}
            >
              Settings
            </Link>
          </Button>
        }
      >
        The Kite bot doesn&apos;t run your templates until you start it again.
      </StatusRow>
    );
  }

  // What the added templates need that the bot doesn't have.
  const required = new Set<PermissionFlag>();
  for (const { install, template } of added) {
    const keys = new Set(
      install.items.filter((i) => !i.deleted).map((i) => i.key)
    );
    for (const p of templatePermissions(template, keys).permissions) {
      required.add(p);
    }
  }
  const missing = missingPermissions([...required], guild.bot_permissions);

  // Roles the templates give that are above the bot's highest role, which
  // Discord doesn't let it give.
  const botPosition = Math.max(
    0,
    ...(roles ?? []).flatMap((r) =>
      r && profile?.role_ids.includes(r.id) ? [r.position] : []
    )
  );
  const tooHigh =
    roles && profile
      ? added.flatMap(({ install, template }) =>
          template.parameters.flatMap((p) => {
            if (p.type !== "role") return [];
            // Changed values are on the items that use them.
            const ids = new Set([
              install.parameters?.[p.key],
              ...install.items.map((i) => i.parameters?.[p.key]),
            ]);
            return roles.flatMap((r) =>
              r && ids.has(r.id) && r.position > botPosition ? [r.name] : []
            );
          })
        )
      : [];

  if (!missing.length && !tooHigh.length) {
    return (
      <StatusRow
        state="ok"
        title={`The Kite bot is running in ${guild.name}`}
      />
    );
  }

  return (
    <StatusRow
      state="warn"
      title="The Kite bot can't do everything your templates need"
    >
      {!!missing.length && (
        <p>
          Give its role the {list(missing.map(permissionLabel))}{" "}
          {missing.length === 1 ? "permission" : "permissions"} in Server
          Settings, Roles.
        </p>
      )}
      {!!tooHigh.length && (
        <p>
          Move its role above {list([...new Set(tooHigh)])} in Server Settings,
          Roles, so it can give {tooHigh.length === 1 ? "it" : "them"}.
        </p>
      )}
    </StatusRow>
  );
}

function ErrorsRow() {
  const appId = useAppId();
  const summary = useLogSummary();
  if (!summary) return null;

  const errors = summary.total_errors;
  if (!errors) {
    return <StatusRow state="ok" title="No errors in the last 24 hours" />;
  }

  return (
    <StatusRow
      state="warn"
      title={`${formatNumber(errors)} ${
        errors === 1 ? "error" : "errors"
      } in the last 24 hours`}
      action={
        <Button size="sm" variant="outline" asChild>
          <Link href={{ pathname: "/apps/[appId]/logs", query: { appId } }}>
            View logs
          </Link>
        </Button>
      }
    >
      Something a template tried didn&apos;t work. The logs say what.
    </StatusRow>
  );
}

function CreditsRow() {
  const credits = useUsageCredits();
  const creditsPerMonth = useAppFeature((f) => f.usage_credits_per_month);
  if (!credits || !creditsPerMonth) return null;

  const used = credits.credits_used / creditsPerMonth;
  const low = used >= warnAt;

  return (
    <StatusRow
      state={used >= 1 ? "error" : low ? "warn" : "ok"}
      title={`${formatNumber(credits.credits_used)} of ${formatNumber(
        creditsPerMonth
      )} credits used`}
      action={low && <GetCreditsButton from="kite_bot_status" />}
    >
      <p>
        {used >= 1
          ? `Your templates are paused until the credits reset on ${formatMonthlyResetDate()}.`
          : `Sending a message or giving a role costs 1 credit. Credits reset on ${formatMonthlyResetDate()}.`}
      </p>
      <Progress
        value={Math.min(100, used * 100)}
        className={cn("h-1.5 mt-2", low && "[&>div]:bg-amber-500")}
      />
    </StatusRow>
  );
}
