import { useStatsQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";

// Rounds down, so the numbers never claim more than is true.
function formatCount(value: number) {
  if (value >= 1_000_000) {
    return `${Math.floor(value / 100_000) / 10} million`;
  }
  return `${(Math.floor(value / 1000) * 1000).toLocaleString("en-US")}+`;
}

export default function HomeStatsSection() {
  const query = useStatsQuery();
  const stats = useResponseData(query);

  // Nothing to show if the API can't be reached.
  if (query.isError || (query.isSuccess && !stats)) return null;

  const items = stats && [
    { value: stats.apps, label: "bots built" },
    { value: stats.users, label: "bot builders" },
    {
      value: stats.commands + stats.event_listeners,
      label: "commands and events",
    },
    {
      value: stats.executions_last_month,
      label: "flow runs last month",
    },
  ];

  return (
    <section className="border-y bg-muted/30">
      <div className="container grid grid-cols-2 lg:grid-cols-4 gap-x-4 gap-y-8 py-10">
        {(items ?? Array(4).fill(undefined)).map((item, i) => (
          <div key={i} className="text-center">
            {item ? (
              <>
                <div className="text-2xl sm:text-3xl md:text-4xl font-bold">
                  {formatCount(item.value)}
                </div>
                <div className="text-sm md:text-base text-muted-foreground mt-1">
                  {item.label}
                </div>
              </>
            ) : (
              // Keeps the height while loading, so the page doesn't jump.
              <div className="h-14 md:h-[4.5rem]" />
            )}
          </div>
        ))}
      </div>
    </section>
  );
}
