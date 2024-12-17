import LogSummaryCard from "./LogSummaryCard";
import { UsageCreditsByEntityChart } from "./UsageCreditsByEntityChart";

export function AppChartCards() {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <UsageCreditsByEntityChart />
      <LogSummaryCard />
    </div>
  );
}
