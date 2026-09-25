import { useAppFeature } from "@/lib/hooks/api";
import { describeSchedule, getSchedulePreview } from "@/lib/flow/schedule";
import { formatInterval } from "@/lib/utils";
import { useMemo } from "react";

export function ScheduleCronHelp() {
  return (
    <>
      A cron expression in UTC, e.g. */5 * * * * for every five minutes. Not
      sure how to write one? Use{" "}
      <a
        href="https://crongenerator.com"
        target="_blank"
        rel="noopener noreferrer"
        className="underline hover:text-foreground"
      >
        crongenerator.com
      </a>
      .
    </>
  );
}

export default function ScheduleCronPreview({
  cron,
  creditsPerRun,
}: {
  cron: string;
  creditsPerRun?: number;
}) {
  const minInterval = useAppFeature((f) => f.min_schedule_interval_seconds);

  const description = useMemo(() => describeSchedule(cron), [cron]);
  const preview = useMemo(() => getSchedulePreview(cron), [cron]);

  if (!preview) return null;

  return (
    <div className="text-sm text-muted-foreground space-y-1">
      {description && (
        <div className="font-medium text-foreground">{description} (UTC)</div>
      )}
      <div>Next runs:</div>
      {preview.nextRuns.map((d) => (
        <div key={d.getTime()}>
          {d.toISOString().replace("T", " ").slice(0, 19)} UTC (
          {d.toLocaleString()} local)
        </div>
      ))}
      {minInterval && preview.minGapSeconds < minInterval ? (
        <div className="text-destructive">
          Your plan allows at most one run every {formatInterval(minInterval)}.
        </div>
      ) : null}
      {creditsPerRun ? (
        <div>
          Uses about {(preview.runsPerMonth * creditsPerRun).toLocaleString()}{" "}
          credits per month.
        </div>
      ) : null}
    </div>
  );
}
