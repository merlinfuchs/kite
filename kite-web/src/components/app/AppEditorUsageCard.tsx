import { useEditorUsageQuery } from "@/lib/api/queries";
import { useAppId } from "@/lib/hooks/params";
import { formatDateTime } from "@/lib/utils";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { formatEditingTime } from "./EditorUsageSummary";

export default function AppEditorUsageCard() {
  const { data } = useEditorUsageQuery(useAppId());
  // It's only extra information, so it's left out if it fails to load.
  if (data && !data.success) return null;
  const total = data?.data.total;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Editor activity</CardTitle>
        <CardDescription>
          Time spent and saves in the editor in the last 30 days, by you and
          your collaborators.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {!total ? (
          <Skeleton className="h-12" />
        ) : (
          <div className="grid gap-4 grid-cols-2 sm:grid-cols-3">
            <UsageValue
              label="Time editing"
              value={formatEditingTime(total.active_seconds)}
            />
            <UsageValue label="Saves" value={total.save_count} />
            <UsageValue
              label="Last save"
              value={
                total.last_saved_at
                  ? formatDateTime(new Date(total.last_saved_at))
                  : "None"
              }
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function UsageValue({
  label,
  value,
}: {
  label: string;
  value: string | number;
}) {
  return (
    <div>
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="text-lg font-semibold">{value}</div>
    </div>
  );
}
