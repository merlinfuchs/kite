import { useEditorUsage } from "@/lib/hooks/api";
import { formatDateTime, formatInterval } from "@/lib/utils";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Skeleton } from "../ui/skeleton";

export default function AppEditorUsageCard() {
  const [loadFailed, setLoadFailed] = useState(false);
  const onResponse = useCallback(
    (response: {
      success: boolean;
      error?: { message: string; code: string };
    }) => {
      setLoadFailed(!response.success);
      if (!response.success) {
        toast.error(
          `Failed to load editor activity: ${response.error?.message} (${response.error?.code})`
        );
      }
    },
    []
  );
  const usage = useEditorUsage(onResponse);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Editor activity</CardTitle>
        <CardDescription>
          Across all commands and event listeners in the last 30 days.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {loadFailed ? (
          <div className="text-sm text-destructive">
            Editor activity could not be loaded.
          </div>
        ) : !usage ? (
          <Skeleton className="h-16" />
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <UsageValue
              label="Active editor time"
              value={formatInterval(usage.total.active_seconds)}
            />
            <UsageValue label="Saves" value={usage.total.save_count} />
            <UsageValue
              label="Editor sessions"
              value={usage.total.session_count}
            />
            <UsageValue
              label="Last edited"
              value={
                usage.total.last_edited_at
                  ? formatDateTime(new Date(usage.total.last_edited_at))
                  : "No saves"
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
