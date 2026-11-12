import type { EditorUsageItem } from "@/lib/types/editorUsage";
import { formatDateTime, formatInterval } from "@/lib/utils";

export default function EditorUsageSummary({
  usage,
}: {
  usage?: EditorUsageItem | null;
}) {
  if (usage === undefined) return null;
  if (usage === null) {
    return (
      <div className="mt-2 text-xs text-muted-foreground">
        No editor activity in the last 30 days.
      </div>
    );
  }

  return (
    <div className="mt-2 text-xs text-muted-foreground">
      Last 30 days: {formatInterval(usage.active_seconds)} active ·{" "}
      {usage.session_count} sessions · {usage.save_count} saves · Last edited{" "}
      {usage.last_edited_at
        ? formatDateTime(new Date(usage.last_edited_at))
        : "not in this period"}
    </div>
  );
}
