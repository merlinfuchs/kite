import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import {
  useAppStateGuilds,
  useCommands,
  useEventListeners,
  useMessages,
  useUsageAnalytics,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import {
  UsageAnalyticsSourceEntry,
  UsageAnalyticsTotals,
} from "@/lib/types/wire.gen";
import { cn, formatNumber } from "@/lib/utils";
import Link from "next/link";
import { useMemo, useState } from "react";

const ranges = [
  { value: "1d", label: "1D", description: "the last 24 hours" },
  { value: "1w", label: "1W", description: "the last 7 days" },
  { value: "1m", label: "1M", description: "the last 30 days" },
  { value: "1y", label: "1Y", description: "the last 12 months" },
  { value: "all", label: "All time", description: "all time" },
];

const runsChartConfig = {
  command_executions: {
    label: "Commands",
    color: "hsl(var(--chart-3))",
  },
  event_listener_executions: {
    label: "Events",
    color: "hsl(var(--chart-4))",
  },
  message_executions: {
    label: "Messages",
    color: "hsl(var(--chart-5))",
  },
} satisfies ChartConfig;

const creditsChartConfig = {
  credits_used: {
    label: "Credits",
    color: "hsl(var(--primary))",
  },
} satisfies ChartConfig;

function formatBucket(time: string, bucket: string) {
  const date = new Date(time);
  if (bucket === "hour") {
    return date.toLocaleTimeString("en-US", { hour: "numeric" });
  }
  // Day and month buckets are UTC, so they are shown in UTC to not shift a day.
  if (bucket === "day") {
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      timeZone: "UTC",
    });
  }
  return date.toLocaleDateString("en-US", {
    month: "short",
    year: "2-digit",
    timeZone: "UTC",
  });
}

function formatBucketLong(time: string, bucket: string) {
  const date = new Date(time);
  if (bucket === "hour") {
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "numeric",
    });
  }
  if (bucket === "day") {
    return date.toLocaleDateString("en-US", {
      weekday: "short",
      month: "short",
      day: "numeric",
      timeZone: "UTC",
    });
  }
  return date.toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  });
}

export default function UsageAnalytics() {
  const [range, setRange] = useState("1m");
  const [metric, setMetric] = useState<"runs" | "credits">("runs");

  const analytics = useUsageAnalytics(range);
  const guilds = useAppStateGuilds();

  const rangeInfo = ranges.find((r) => r.value === range) ?? ranges[2];
  const totals = analytics?.totals;
  const previous = analytics?.previous_totals ?? null;

  const chartData = useMemo(
    () =>
      analytics?.series.map((entry) => ({
        label: formatBucket(entry!.time, analytics.bucket),
        long_label: formatBucketLong(entry!.time, analytics.bucket),
        credits_used: entry!.credits_used,
        command_executions: entry!.command_executions,
        event_listener_executions: entry!.event_listener_executions,
        message_executions: entry!.message_executions,
        runs:
          entry!.command_executions +
          entry!.event_listener_executions +
          entry!.message_executions,
      })) ?? [],
    [analytics]
  );

  const peak = useMemo(() => {
    let best: (typeof chartData)[number] | null = null;
    for (const entry of chartData) {
      if (entry.runs > 0 && (!best || entry.runs > best.runs)) {
        best = entry;
      }
    }
    return best;
  }, [chartData]);

  const hasData = !!totals && totals.executions > 0;

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-muted-foreground">
          Showing {rangeInfo.description}
          {analytics?.bucket === "hour" ? " in your local time" : " (UTC)"}.
        </p>
        <ToggleGroup
          type="single"
          variant="outline"
          size="sm"
          value={range}
          onValueChange={(v) => v && setRange(v)}
          className="justify-start"
        >
          {ranges.map((r) => (
            <ToggleGroupItem
              key={r.value}
              value={r.value}
              aria-label={`Show ${r.description}`}
              className="px-3"
            >
              {r.label}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      </div>

      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Credits used"
          value={totals?.credits_used}
          previous={previous?.credits_used}
        />
        <StatCard
          title="Commands run"
          value={totals?.command_executions}
          previous={previous?.command_executions}
        />
        <StatCard
          title="Events handled"
          value={totals?.event_listener_executions}
          previous={previous?.event_listener_executions}
        />
        <StatCard
          title="Message interactions"
          value={totals?.message_executions}
          previous={previous?.message_executions}
        />
      </div>

      <Card>
        <CardHeader className="flex flex-row items-start justify-between space-y-0 gap-4">
          <div className="space-y-1.5">
            <CardDescription>
              {metric === "runs" ? "Runs over time" : "Credits over time"}
            </CardDescription>
            <CardTitle className="text-3xl">
              {formatNumber(
                metric === "runs" ? totals?.executions : totals?.credits_used
              )}
              <span className="text-sm font-normal text-muted-foreground ml-2">
                {metric === "runs" ? "total runs" : "credits"}
              </span>
            </CardTitle>
          </div>
          <ToggleGroup
            type="single"
            size="sm"
            value={metric}
            onValueChange={(v) => v && setMetric(v as "runs" | "credits")}
          >
            <ToggleGroupItem value="runs">Runs</ToggleGroupItem>
            <ToggleGroupItem value="credits">Credits</ToggleGroupItem>
          </ToggleGroup>
        </CardHeader>
        <CardContent>
          {metric === "runs" ? (
            <ChartContainer
              config={runsChartConfig}
              className="h-[220px] sm:h-[280px] w-full"
            >
              <BarChart accessibilityLayer data={chartData}>
                <CartesianGrid vertical={false} />
                <XAxis
                  dataKey="label"
                  tickLine={false}
                  axisLine={false}
                  tickMargin={10}
                  minTickGap={16}
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  width={40}
                  allowDecimals={false}
                />
                <ChartTooltip
                  cursor={false}
                  content={
                    <ChartTooltipContent
                      labelFormatter={(_, payload) =>
                        payload?.[0]?.payload?.long_label
                      }
                    />
                  }
                />
                <ChartLegend content={<ChartLegendContent />} />
                <Bar
                  dataKey="command_executions"
                  stackId="runs"
                  fill="var(--color-command_executions)"
                />
                <Bar
                  dataKey="event_listener_executions"
                  stackId="runs"
                  fill="var(--color-event_listener_executions)"
                />
                <Bar
                  dataKey="message_executions"
                  stackId="runs"
                  fill="var(--color-message_executions)"
                  radius={[4, 4, 0, 0]}
                />
              </BarChart>
            </ChartContainer>
          ) : (
            <ChartContainer
              config={creditsChartConfig}
              className="h-[220px] sm:h-[280px] w-full"
            >
              <BarChart accessibilityLayer data={chartData}>
                <CartesianGrid vertical={false} />
                <XAxis
                  dataKey="label"
                  tickLine={false}
                  axisLine={false}
                  tickMargin={10}
                  minTickGap={16}
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  width={40}
                  allowDecimals={false}
                />
                <ChartTooltip
                  cursor={false}
                  content={
                    <ChartTooltipContent
                      labelFormatter={(_, payload) =>
                        payload?.[0]?.payload?.long_label
                      }
                    />
                  }
                />
                <Bar
                  dataKey="credits_used"
                  fill="var(--color-credits_used)"
                  radius={[4, 4, 0, 0]}
                />
              </BarChart>
            </ChartContainer>
          )}
          {analytics && !hasData && (
            <p className="text-sm text-muted-foreground text-center mt-2">
              Nothing ran in this period. Runs show up here once your commands,
              events or messages are used.
            </p>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-2">
        <CreditBreakdownCard totals={totals} />
        <Card>
          <CardHeader>
            <CardDescription>Health and activity</CardDescription>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-x-4 gap-y-5">
            <Fact
              label="Errors"
              value={formatNumber(analytics?.logs.errors)}
              tone={analytics?.logs.errors ? "bad" : undefined}
              href="/apps/[appId]/logs"
            />
            <Fact
              label="Warnings"
              value={formatNumber(analytics?.logs.warnings)}
              tone={analytics?.logs.warnings ? "warn" : undefined}
              href="/apps/[appId]/logs"
            />
            <Fact
              label="Credits per run"
              value={
                totals && totals.executions > 0
                  ? (totals.credits_used / totals.executions).toFixed(2)
                  : "–"
              }
            />
            <Fact
              label="Busiest"
              value={peak ? peak.long_label : "–"}
              detail={peak ? `${formatNumber(peak.runs)} runs` : undefined}
            />
            <Fact
              label="Servers"
              value={guilds ? formatNumber(guilds.length) : "–"}
              detail="right now"
              href="/apps/[appId]/guilds"
            />
            <Fact label="Total runs" value={formatNumber(totals?.executions)} />
          </CardContent>
          {analytics?.logs.partial && (
            <p className="px-6 pb-6 -mt-2 text-xs text-muted-foreground">
              Logs are kept for 30 days, so errors and warnings only cover the
              last 30 days.
            </p>
          )}
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <TopCommandsCard entries={analytics?.top_commands} />
        <TopEventListenersCard entries={analytics?.top_event_listeners} />
        <TopMessagesCard entries={analytics?.top_messages} />
      </div>
    </div>
  );
}

function StatCard({
  title,
  value,
  previous,
}: {
  title: string;
  value?: number;
  previous?: number | null;
}) {
  let change: number | null = null;
  if (value !== undefined && previous !== undefined && previous !== null) {
    change = previous === 0 ? (value === 0 ? 0 : null) : value / previous - 1;
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>{title}</CardDescription>
        <CardTitle className="text-2xl sm:text-3xl tabular-nums">
          {formatNumber(value)}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-xs text-muted-foreground">
          {previous === undefined || previous === null ? (
            "\u00a0"
          ) : change === null ? (
            <>New this period</>
          ) : (
            <>
              <span
                className={cn(
                  change > 0 && "text-green-500",
                  change < 0 && "text-red-500"
                )}
              >
                {change > 0 ? "+" : ""}
                {Math.round(change * 100)}%
              </span>{" "}
              vs previous period
            </>
          )}
        </p>
      </CardContent>
    </Card>
  );
}

function CreditBreakdownCard({ totals }: { totals?: UsageAnalyticsTotals }) {
  const rows = [
    {
      label: "Commands",
      credits: totals?.command_credits_used ?? 0,
      color: "hsl(var(--chart-3))",
    },
    {
      label: "Events",
      credits: totals?.event_listener_credits_used ?? 0,
      color: "hsl(var(--chart-4))",
    },
    {
      label: "Messages",
      credits: totals?.message_credits_used ?? 0,
      color: "hsl(var(--chart-5))",
    },
  ];
  const total = totals?.credits_used ?? 0;

  return (
    <Card>
      <CardHeader>
        <CardDescription>Where credits went</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {rows.map((row) => {
          const share = total > 0 ? row.credits / total : 0;
          return (
            <div key={row.label} className="space-y-1.5">
              <div className="flex justify-between text-sm">
                <span>{row.label}</span>
                <span className="text-muted-foreground tabular-nums">
                  {formatNumber(row.credits)} ·{" "}
                  <span className="text-foreground">
                    {Math.round(share * 100)}%
                  </span>
                </span>
              </div>
              <div className="h-2 rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full rounded-full transition-[width] duration-500 motion-reduce:transition-none"
                  style={{ width: `${share * 100}%`, background: row.color }}
                />
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}

function Fact({
  label,
  value,
  detail,
  tone,
  href,
}: {
  label: string;
  value: string;
  detail?: string;
  tone?: "bad" | "warn";
  href?: string;
}) {
  const appId = useAppId();

  const content = (
    <>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div
        className={cn(
          "text-lg font-semibold tabular-nums truncate",
          tone === "bad" && "text-red-500",
          tone === "warn" && "text-yellow-500"
        )}
      >
        {value}
      </div>
      {detail && <div className="text-xs text-muted-foreground">{detail}</div>}
    </>
  );

  if (!href) return <div className="min-w-0">{content}</div>;

  return (
    <Link
      href={{ pathname: href, query: { appId } }}
      className="min-w-0 rounded-md hover:bg-muted/50 -m-1 p-1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      {content}
    </Link>
  );
}

function TopList({
  title,
  empty,
  entries,
  resolve,
}: {
  title: string;
  empty: string;
  entries?: (UsageAnalyticsSourceEntry | undefined)[];
  resolve: (id: string) => { name: string; href?: object } | null;
}) {
  const max = Math.max(1, ...(entries ?? []).map((e) => e!.executions));

  return (
    <Card>
      <CardHeader>
        <CardDescription>{title}</CardDescription>
      </CardHeader>
      <CardContent>
        {!entries?.length ? (
          <p className="text-sm text-muted-foreground">{empty}</p>
        ) : (
          <ul className="space-y-3">
            {entries.map((entry) => {
              const resolved = resolve(entry!.id);
              const name = resolved?.name ?? "Deleted";
              return (
                <li key={entry!.id} className="space-y-1">
                  <div className="flex justify-between gap-2 text-sm">
                    {resolved?.href ? (
                      <Link
                        href={resolved.href}
                        className="truncate hover:underline"
                      >
                        {name}
                      </Link>
                    ) : (
                      <span className="truncate text-muted-foreground">
                        {name}
                      </span>
                    )}
                    <span className="text-muted-foreground tabular-nums shrink-0">
                      {formatNumber(entry!.executions)} runs
                    </span>
                  </div>
                  <div className="h-1 rounded-full bg-muted overflow-hidden">
                    <div
                      className="h-full bg-primary rounded-full"
                      style={{
                        width: `${(entry!.executions / max) * 100}%`,
                      }}
                    />
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function TopCommandsCard({
  entries,
}: {
  entries?: (UsageAnalyticsSourceEntry | undefined)[];
}) {
  const appId = useAppId();
  const commands = useCommands();

  return (
    <TopList
      title="Top commands"
      empty="No commands were run in this period."
      entries={entries}
      resolve={(id) => {
        const cmd = commands?.find((c) => c!.id === id);
        if (!cmd) return null;
        return {
          name: `/${cmd.name}`,
          href: {
            pathname: "/apps/[appId]/commands/[cmdId]",
            query: { appId, cmdId: id },
          },
        };
      }}
    />
  );
}

function TopEventListenersCard({
  entries,
}: {
  entries?: (UsageAnalyticsSourceEntry | undefined)[];
}) {
  const appId = useAppId();
  const listeners = useEventListeners();

  return (
    <TopList
      title="Top event listeners"
      empty="No events were handled in this period."
      entries={entries}
      resolve={(id) => {
        const listener = listeners?.find((l) => l!.id === id);
        if (!listener) return null;
        return {
          name: listener.description || listener.type,
          href: {
            pathname: "/apps/[appId]/events/[eventId]",
            query: { appId, eventId: id },
          },
        };
      }}
    />
  );
}

function TopMessagesCard({
  entries,
}: {
  entries?: (UsageAnalyticsSourceEntry | undefined)[];
}) {
  const appId = useAppId();
  const messages = useMessages();

  return (
    <TopList
      title="Top messages"
      empty="No message buttons or menus were used in this period."
      entries={entries}
      resolve={(id) => {
        const message = messages?.find((m) => m!.id === id);
        if (!message) return null;
        return {
          name: message.name,
          href: {
            pathname: "/apps/[appId]/messages/[messageId]",
            query: { appId, messageId: id },
          },
        };
      }}
    />
  );
}
