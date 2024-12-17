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
import { useUsageCredits } from "@/lib/hooks/api";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "../ui/chart";
import { formatNumber } from "@/lib/utils";

const data = [
  {
    credits_used: 10400,
  },
  {
    credits_used: 14405,
  },
  {
    credits_used: 9400,
  },
  {
    credits_used: 8200,
  },
  {
    credits_used: 7000,
  },
  {
    credits_used: 9600,
  },
  {
    credits_used: 11244,
  },
  {
    credits_used: 26475,
  },
  {
    credits_used: 10400,
  },
  {
    credits_used: 14405,
  },
  {
    credits_used: 9400,
  },
  {
    credits_used: 8200,
  },
  {
    credits_used: 7000,
  },
  {
    credits_used: 9600,
  },
  {
    credits_used: 11244,
  },
  {
    credits_used: 26475,
  },
  {
    credits_used: 10400,
  },
  {
    credits_used: 14405,
  },
  {
    credits_used: 9400,
  },
  {
    credits_used: 8200,
  },
  {
    credits_used: 7000,
  },
  {
    credits_used: 9600,
  },
  {
    credits_used: 11244,
  },
  {
    credits_used: 26475,
  },
  {
    credits_used: 10400,
  },
  {
    credits_used: 14405,
  },
  {
    credits_used: 9400,
  },
  {
    credits_used: 8200,
  },
];

const chartConfig = {
  credits_used: {
    label: "Credits Used",
    color: "hsl(var(--primary))",
  },
} satisfies ChartConfig;

export function UsageCreditsByDayChart() {
  const credits = useUsageCredits();

  return (
    <Card>
      <CardHeader>
        <CardDescription>Monthly Usage</CardDescription>
        <div className="flex items-end gap-2">
          <CardTitle className="text-4xl">
            {formatNumber(credits?.credits_used)}
          </CardTitle>
          <p className="text-sm text-muted-foreground pb-1">
            of{" "}
            <span className="text-foreground">
              {formatNumber(credits?.total_credits)}
            </span>{" "}
            credits used
          </p>
        </div>
      </CardHeader>
      <CardContent>
        <div className="h-[150px] sm:h-[225px]">
          <ResponsiveContainer width="100%" height="100%">
            <ChartContainer config={chartConfig}>
              <BarChart accessibilityLayer data={data}>
                <ChartTooltip
                  cursor={false}
                  content={<ChartTooltipContent indicator="dashed" />}
                />
                <CartesianGrid vertical={false} />
                <Bar
                  dataKey="credits_used"
                  radius={[4, 4, 0, 0]}
                  style={
                    {
                      fill: "hsl(var(--primary))",
                      opacity: 1,
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
