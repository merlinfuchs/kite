import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import CopyButton from "@/tools/common/components/CopyButton";
import Link from "next/link";
import { useMemo, useState } from "react";
import { decodeSnowflake } from "../snowflake";

export default function SnowflakeDecoder() {
  const [value, setValue] = useState("");
  const decoded = useMemo(() => decodeSnowflake(value), [value]);

  const unix = decoded && Math.floor(decoded.timestamp / 1000);
  const rows = decoded && [
    {
      label: "Created at",
      value: new Intl.DateTimeFormat(undefined, {
        dateStyle: "full",
        timeStyle: "medium",
      }).format(decoded.timestamp),
    },
    {
      label: "Created at (UTC)",
      value: new Date(decoded.timestamp).toISOString(),
    },
    { label: "Unix timestamp (ms)", value: decoded.timestamp.toString() },
    { label: "Discord timestamp", value: `<t:${unix}:F>`, copy: true },
    { label: "Worker ID", value: decoded.workerId.toString() },
    { label: "Process ID", value: decoded.processId.toString() },
    { label: "Increment", value: decoded.increment.toString() },
  ];

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="snowflake">Discord ID</Label>
        <Input
          id="snowflake"
          placeholder="1083038298513887292"
          value={value}
          onChange={(e) => setValue(e.target.value.trim())}
          className="font-mono"
        />
        {value && !decoded && (
          <p className="text-sm text-destructive">
            This isn&apos;t a Discord ID. IDs are numbers with 17 to 20 digits.
          </p>
        )}
      </div>

      {rows && (
        <div className="rounded-lg border divide-y">
          {rows.map((row) => (
            <div
              key={row.label}
              className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-3 p-4"
            >
              <div className="sm:w-48 flex-none font-medium">{row.label}</div>
              <div className="flex-auto font-mono text-sm break-all">
                {row.value}
              </div>
              {row.copy && <CopyButton value={row.value} />}
            </div>
          ))}
        </div>
      )}

      {rows && (
        <p className="text-sm text-muted-foreground">
          Want this date in another format? Use the{" "}
          <Link href="/tools/timestamp-generator" className="underline">
            timestamp generator
          </Link>
          .
        </p>
      )}
    </div>
  );
}
