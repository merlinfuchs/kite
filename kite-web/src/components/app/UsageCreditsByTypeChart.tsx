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

  return (
    <Card>
      <CardHeader>
        <CardDescription>Usage by type</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="mt-4 h-[150px] sm:h-[200px]">
          <ResponsiveContainer width="100%" height="100%">
            <ChartContainer config={chartConfig}>
              <PieChart>
                <ChartTooltip
                  content={<ChartTooltipContent nameKey="type" hideLabel />}
                />
                <Pie
                  data={chartData}
                  dataKey="credits_used"
                  labelLine={false}
                  label={({ payload, ...props }) => {
                    return (
                      <text
                        cx={props.cx}
                        cy={props.cy}
                        x={props.x}
                        y={props.y}
                        textAnchor={props.textAnchor}
                        dominantBaseline={props.dominantBaseline}
                        fill="hsla(var(--foreground))"
                      >
                        {
                          chartConfig[payload.type as keyof typeof chartConfig]
                            ?.label
                        }
                      </text>
                    );
                  }}
                  nameKey="type"
                />
              </PieChart>
            </ChartContainer>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
