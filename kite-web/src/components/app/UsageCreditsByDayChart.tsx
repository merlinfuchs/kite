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
import { formatNumber } from "@/lib/utils";
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
} satisfies ChartConfig;

export default function UsageCreditsByDayChart() {
  const credits = useUsageCredits();

  const creditsPerMonth = useAppFeature((f) => f.usage_credits_per_month || 0);

  const creditsByDay = useUsageCreditsByDay();

  const chartData = useMemo(
    () =>
      creditsByDay?.map((c) => ({
        date: new Date(c!.date).toLocaleDateString("en-US", {
          month: "short",
          day: "numeric",
        }),
        credits_used: c!.credits_used,
      })) || [],
    [creditsByDay]
  );

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
