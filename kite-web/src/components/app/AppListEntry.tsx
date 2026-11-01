import { App } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { ChevronRightIcon, ServerIcon, UsersIcon } from "lucide-react";
import Link from "next/link";

function appStatus(app: App) {
  if (app.token_invalid) {
    return {
      label: "Token invalid",
      dot: "bg-destructive",
      text: "text-destructive",
    };
  }
  if (!app.enabled) {
    return {
      label: "Stopped",
      dot: "bg-muted-foreground",
      text: "text-muted-foreground",
    };
  }
  if (app.no_credits) {
    return {
      label: "Out of credits",
      dot: "bg-amber-500",
      text: "text-amber-600 dark:text-amber-500",
    };
  }
  return {
    label: "Running",
    dot: "bg-emerald-500",
    text: "text-emerald-600 dark:text-emerald-500",
  };
}

function formatCount(n: number) {
  return new Intl.NumberFormat("en", { notation: "compact" }).format(n);
}

export default function AppListEntry({ app }: { app: App }) {
  const status = appStatus(app);
  const initial = app.name.trim().charAt(0).toUpperCase() || "?";

  return (
    <div className="group relative flex items-center gap-4 rounded-xl border bg-card p-4 transition-colors hover:border-primary/40 hover:bg-accent/40">
      <div className="flex h-11 w-11 flex-none items-center justify-center rounded-xl bg-primary/10 text-lg font-semibold text-primary">
        {initial}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <Link
            href={{ pathname: "/apps/[appId]", query: { appId: app.id } }}
            className="truncate font-semibold outline-none after:absolute after:inset-0 after:rounded-xl focus-visible:after:ring-2 focus-visible:after:ring-ring"
          >
            {app.name}
          </Link>
          <span
            className={cn(
              "inline-flex flex-none items-center gap-1.5 text-xs font-medium",
              status.text
            )}
          >
            <span className={cn("h-1.5 w-1.5 rounded-full", status.dot)} />
            {status.label}
          </span>
        </div>
        <p className="mt-0.5 truncate text-sm text-muted-foreground">
          {app.description || "No description"}
        </p>
        {(app.discord_guild_count !== null ||
          app.discord_user_install_count !== null) && (
          <div className="mt-2 flex items-center gap-4 text-xs text-muted-foreground">
            {app.discord_guild_count !== null && (
              <span className="inline-flex items-center gap-1">
                <ServerIcon className="h-3.5 w-3.5" />
                {formatCount(app.discord_guild_count)}{" "}
                {app.discord_guild_count === 1 ? "server" : "servers"}
              </span>
            )}
            {app.discord_user_install_count !== null && (
              <span className="inline-flex items-center gap-1">
                <UsersIcon className="h-3.5 w-3.5" />
                {formatCount(app.discord_user_install_count)}{" "}
                {app.discord_user_install_count === 1
                  ? "user install"
                  : "user installs"}
              </span>
            )}
          </div>
        )}
      </div>

      <ChevronRightIcon className="h-5 w-5 flex-none text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-foreground" />
    </div>
  );
}
