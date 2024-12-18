import LogSummaryCard from "./LogSummaryCard";
import UsageCreditsByTypeChart from "./UsageCreditsByTypeChart";

export function AppChartCards() {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <UsageCreditsByTypeChart />
      <LogSummaryCard />
    </div>
  );
}
