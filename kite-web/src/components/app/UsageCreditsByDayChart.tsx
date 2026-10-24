import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  XAxis,
} from "recharts";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import env from "@/lib/env/client";
import {
  useAppFeature,
  useUsageCredits,
  useUsageCreditsByDay,
} from "@/lib/hooks/api";
import { dayMs, forecastUsage } from "@/lib/usageForecast";
import {
  cn,
  formatMonthlyResetDate,
  formatNumber,
  formatUTCDay,
} from "@/lib/utils";
import { CircleHelpIcon } from "lucide-react";
import Link from "next/link";
import { useMemo } from "react";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "../ui/chart";

const chartConfig = {
  credits_used: {
    label: "Credits Used",
    color: "hsl(var(--primary))",
  },
  projected: {
    label: "Projected",
    color: "hsl(var(--primary))",
  },
} satisfies ChartConfig;

export default function UsageCreditsByDayChart() {
  const credits = useUsageCredits();

  const creditsPerMonth = useAppFeature((f) => f.usage_credits_per_month || 0);

  const creditsByDay = useUsageCreditsByDay();

  const forecast = useMemo(
    () =>
      creditsByDay && credits && creditsPerMonth !== undefined
        ? forecastUsage(
            creditsByDay.map((c) => c!),
            credits.credits_used,
            creditsPerMonth
          )
        : null,
    [creditsByDay, credits, creditsPerMonth]
  );

  const chartData = useMemo(() => {
    const now = new Date();

    return (
      creditsByDay?.map((c) => {
        const date = new Date(c!.date);
        return {
          date: date.toLocaleDateString(undefined, {
            month: "short",
            day: "numeric",
            timeZone: "UTC",
          }),
          credits_used: c!.credits_used,
          // Today only gets the hours that are left, and the bot stops once
          // its credits run out.
          projected:
            forecast &&
            dayLeft(date, now) > 0 &&
            !(forecast.runsOutAt && date > forecast.runsOutAt)
              ? Math.round(forecast.dailyRate * dayLeft(date, now))
              : undefined,
        };
      }) || []
    );
  }, [creditsByDay, forecast]);

  return (
    <Card>
      <CardHeader>
        <CardDescription className="flex items-center gap-2">
          <div>Monthly Usage</div>
          <Link
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system`}
            target="_blank"
          >
            <CircleHelpIcon className="w-5 h-5 hover:text-foreground" />
          </Link>
        </CardDescription>
        <div className="flex items-end gap-2">
          <CardTitle className="text-4xl">
            {formatNumber(credits?.credits_used)}
          </CardTitle>
          <p className="text-sm text-muted-foreground pb-1">
            of{" "}
            <span className="text-foreground">
              {formatNumber(creditsPerMonth)}
            </span>{" "}
            credits used
          </p>
        </div>
        <p className="text-sm text-muted-foreground">
          Resets on {formatMonthlyResetDate()}
          {forecast && (
            <>
              {" · "}
              <span
                className={cn(
                  forecast.runsOutAt &&
                    "text-yellow-600 dark:text-yellow-500 font-medium"
                )}
              >
                {forecast.runsOutAt
                  ? `At this rate, credits run out around ${formatUTCDay(
                      forecast.runsOutAt
                    )}`
                  : `At this rate, about ${formatNumber(
                      forecast.projected
                    )} by the end of the month`}
              </span>
            </>
          )}
        </p>
      </CardHeader>
      <CardContent>
        <div className="h-[150px] sm:h-[225px]">
          <ResponsiveContainer width="100%" height="100%">
            <ChartContainer config={chartConfig}>
              <BarChart accessibilityLayer data={chartData}>
                <ChartTooltip
                  cursor={false}
                  content={<ChartTooltipContent indicator="dashed" />}
                />
                <CartesianGrid vertical={false} />
                <XAxis
                  dataKey="date"
                  tickLine={false}
                  tickMargin={10}
                  axisLine={false}
                />
                <Bar
                  dataKey="credits_used"
                  stackId="usage"
                  radius={[4, 4, 0, 0]}
                  style={
                    {
                      fill: "hsl(var(--primary))",
                      opacity: 1,
                    } as React.CSSProperties
                  }
                />
                <Bar
                  dataKey="projected"
                  stackId="usage"
                  radius={[4, 4, 0, 0]}
                  style={
                    {
                      fill: "hsl(var(--primary))",
                      opacity: 0.25,
                    } as React.CSSProperties
                  }
                />
              </BarChart>
            </ChartContainer>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}

// dayLeft is the share of the UTC day starting at date that's still ahead.
function dayLeft(date: Date, now: Date) {
  return Math.min(1, (date.getTime() + dayMs - now.getTime()) / dayMs);
}
