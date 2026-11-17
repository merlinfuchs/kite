import { Pie, PieChart, ResponsiveContainer } from "recharts";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { useUsageCreditsByType } from "@/lib/hooks/api";
import { useMemo } from "react";

const chartConfig = {
  credits_used: {
    label: "Credits used",
  },
  command_flow_execution: {
    label: "Commands",
    color: "hsl(var(--chart-3))",
  },
  event_listener_flow_execution: {
    label: "Event listeners",
    color: "hsl(var(--chart-4))",
  },
  message_flow_execution: {
    label: "Messages",
    color: "hsl(var(--chart-5))",
  },
  webhook_flow_execution: {
    label: "Webhooks",
    color: "hsl(var(--chart-6))",
  },
  schedule_flow_execution: {
    label: "Scheduled tasks",
    color: "hsl(var(--chart-7))",
  },
} satisfies ChartConfig;

function formatPercent(value: number, total: number) {
  if (total <= 0 || value <= 0) return "0%";
  const pct = (value / total) * 100;
  return pct < 1 ? "<1%" : `${Math.round(pct)}%`;
}

export default function UsageCreditsByTypeChart() {
  const creditsByType = useUsageCreditsByType();

  const chartData = useMemo(
    () =>
      creditsByType?.map((credit) => ({
        type: credit!.type,
        credits_used: credit!.credits_used,
        fill: `var(--color-${credit!.type})`,
      })) || [],
    [creditsByType]
  );

  const total = useMemo(
    () => chartData.reduce((sum, d) => sum + d.credits_used, 0),
    [chartData]
  );

  return (
    <Card>
      <CardHeader>
        <CardDescription>Usage by type</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="mt-4 flex flex-col items-center gap-4 sm:flex-row sm:gap-6">
          <div className="h-[150px] w-full sm:h-[200px] sm:w-1/2">
            <ResponsiveContainer width="100%" height="100%">
              <ChartContainer config={chartConfig}>
                <PieChart>
                  <ChartTooltip
                    content={
                      <ChartTooltipContent
                        nameKey="type"
                        hideLabel
                        className="min-w-0"
                      />
                    }
                  />
                  <Pie
                    data={chartData}
                    dataKey="credits_used"
                    nameKey="type"
                    outerRadius="95%"
                    strokeWidth={0}
                  />
                </PieChart>
              </ChartContainer>
            </ResponsiveContainer>
          </div>
          <ul className="flex w-full flex-col gap-2 text-sm sm:w-1/2">
            {chartData.map((d) => (
              <li key={d.type} className="flex items-center gap-2">
                <span
                  className="h-2.5 w-2.5 shrink-0 rounded-full"
                  style={{
                    backgroundColor:
                      chartConfig[d.type as keyof typeof chartConfig] &&
                      "color" in chartConfig[d.type as keyof typeof chartConfig]
                        ? (
                            chartConfig[d.type as keyof typeof chartConfig] as {
                              color: string;
                            }
                          ).color
                        : undefined,
                  }}
                />
                <span className="truncate">
                  {chartConfig[d.type as keyof typeof chartConfig]?.label ??
                    d.type}
                </span>
                <span className="ml-auto tabular-nums text-muted-foreground">
                  {formatPercent(d.credits_used, total)}
                </span>
              </li>
            ))}
          </ul>
        </div>
      </CardContent>
    </Card>
  );
}
