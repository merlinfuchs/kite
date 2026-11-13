import AppEmptyPlaceholder from "@/components/app/AppEmptyPlaceholder";
import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import LoadingButton from "@/components/common/LoadingButton";
import { useAppAuditLogQuery } from "@/lib/api/queries";
import {
  auditActionLabel,
  auditCategories,
  canManageSettings,
} from "@/lib/collaborators";
import { useAppCollaborators, useAppRole } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { discordEventType } from "@/lib/flow/eventTypes";
import { AppAuditLogEntry } from "@/lib/types/wire.gen";
import { abbreviateName, formatDateTime } from "@/lib/utils";
import { ArrowLeftIcon, BotIcon, HistoryIcon, LockIcon } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";

const breadcrumbs = [
  {
    label: "App settings",
    href: "/apps/[appId]/settings",
  },
  {
    label: "Audit log",
  },
];

export default function AppAuditLogPage() {
  const appId = useAppId();
  const role = useAppRole();
  const canView = canManageSettings(role);
  const collaborators = useAppCollaborators();

  const [category, setCategory] = useState("all");
  const [userId, setUserId] = useState("all");

  const query = useAppAuditLogQuery(canView ? appId : "", {
    userId: userId === "all" ? undefined : userId,
    kinds: auditCategories.find((c) => c.value === category)?.kinds,
  });

  const entries = useMemo(
    () =>
      query.data?.pages.flatMap((page) =>
        page.success ? (page.data.filter(Boolean) as AppAuditLogEntry[]) : []
      ) ?? [],
    [query.data]
  );
  const filtered = category !== "all" || userId !== "all";

  const error = query.data?.pages.find((page) => !page.success);

  return (
    <AppLayout title="Audit log" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-4">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">Audit log</h1>
          <p className="text-muted-foreground text-sm">
            See who changed what in this app and when. Entries are kept for 90
            days.
          </p>
        </div>
        <Button variant="outline" size="sm" asChild>
          <Link href={`/apps/${appId}/settings`}>
            <ArrowLeftIcon className="size-4 mr-2" />
            Back to settings
          </Link>
        </Button>
      </div>
      <Separator className="my-6" />

      {role !== undefined && !canView ? (
        <AppEmptyPlaceholder
          icon={LockIcon}
          title="No access"
          description="Only admins and the owner of the app can see its audit log."
        />
      ) : (
        <div className="space-y-4">
          <div className="flex flex-wrap gap-3">
            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger className="w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All changes</SelectItem>
                {auditCategories.map((c) => (
                  <SelectItem key={c.value} value={c.value}>
                    {c.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={userId} onValueChange={setUserId}>
              <SelectTrigger className="w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Everyone</SelectItem>
                {collaborators?.map(
                  (c) =>
                    c && (
                      <SelectItem key={c.user.id} value={c.user.id}>
                        {c.user.display_name}
                      </SelectItem>
                    )
                )}
              </SelectContent>
            </Select>
          </div>

          {error && !error.success ? (
            <div className="text-sm text-destructive">
              Failed to load the audit log: {error.error.message}
            </div>
          ) : query.isPending ? (
            <Card>
              <CardContent className="p-0 divide-y">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center gap-3 p-4">
                    <Skeleton className="h-8 w-8 rounded-lg" />
                    <Skeleton className="h-4 w-72" />
                  </div>
                ))}
              </CardContent>
            </Card>
          ) : entries.length === 0 ? (
            <AppEmptyPlaceholder
              icon={HistoryIcon}
              {...(filtered
                ? {
                    title: "No matching changes",
                    description:
                      "Nothing in the last 90 days matches these filters.",
                  }
                : {
                    title: "No changes yet",
                    description:
                      "Changes collaborators make to the app show up here.",
                  })}
            />
          ) : (
            <Card>
              <CardContent className="p-0 divide-y">
                {entries.map((entry) => (
                  <AuditLogRow key={entry.id} entry={entry} appId={appId} />
                ))}
              </CardContent>
            </Card>
          )}

          {query.hasNextPage && (
            <div className="flex justify-center">
              <LoadingButton
                variant="outline"
                loading={query.isFetchingNextPage}
                onClick={() => query.fetchNextPage()}
              >
                Load more
              </LoadingButton>
            </div>
          )}
        </div>
      )}
    </AppLayout>
  );
}

AppAuditLogPage.getLayout = getAppShellLayout;

// The editor pages of the items an entry can link to, while they exist.
const targetPages: Record<string, string> = {
  command: "commands",
  event_listener: "events",
  message: "messages",
  variable: "variables",
};

function AuditLogRow({
  entry,
  appId,
}: {
  entry: AppAuditLogEntry;
  appId: string;
}) {
  const date = new Date(entry.created_at);
  // Event listeners without a description are named by their event type.
  const name =
    entry.target_type === "event_listener"
      ? discordEventType(entry.target_name)?.label ?? entry.target_name
      : entry.target_name;
  const page = targetPages[entry.target_type];
  const deleted = entry.action.endsWith(".delete");
  const href =
    page && entry.target_id && !deleted
      ? `/apps/${appId}/${page}/${entry.target_id}`
      : undefined;

  return (
    <div className="flex items-start gap-3 p-4">
      <Avatar className="rounded-lg h-8 w-8">
        <AvatarFallback className="rounded-lg text-xs">
          {entry.user ? abbreviateName(entry.user.display_name) : "?"}
        </AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0 text-sm">
        <div className="leading-relaxed">
          <span className="font-semibold">
            {entry.user?.display_name ?? "Deleted user"}
          </span>{" "}
          <span className="text-muted-foreground">
            {auditActionLabel(entry.action)}
          </span>{" "}
          {name &&
            (href ? (
              <Link
                href={href}
                className="font-medium text-primary hover:underline"
              >
                {name}
              </Link>
            ) : (
              <span className="font-medium">{name}</span>
            ))}
          {entry.via_api_token && (
            <Badge variant="outline" className="ml-2 font-normal gap-1">
              <BotIcon className="size-3" />
              via API token
            </Badge>
          )}
        </div>
        {entry.user && (
          <div className="text-xs text-muted-foreground">
            @{entry.user.discord_username}
          </div>
        )}
      </div>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="text-xs text-muted-foreground whitespace-nowrap pt-0.5">
            {formatDateTime(date)}
          </div>
        </TooltipTrigger>
        <TooltipContent>{date.toLocaleString()}</TooltipContent>
      </Tooltip>
    </div>
  );
}
