import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import CopyButton from "@/tools/common/components/CopyButton";
import { useEffect, useState } from "react";
import { formatRelative } from "@/tools/common/utils/time";

const styles: {
  style: string;
  name: string;
  format: Intl.DateTimeFormatOptions | "relative";
}[] = [
  { style: "t", name: "Short time", format: { timeStyle: "short" } },
  { style: "T", name: "Long time", format: { timeStyle: "medium" } },
  {
    style: "d",
    name: "Short date",
    format: { year: "numeric", month: "2-digit", day: "2-digit" },
  },
  { style: "D", name: "Long date", format: { dateStyle: "long" } },
  {
    style: "f",
    name: "Short date and time",
    format: { dateStyle: "long", timeStyle: "short" },
  },
  {
    style: "F",
    name: "Long date and time",
    format: { dateStyle: "full", timeStyle: "short" },
  },
  { style: "R", name: "Relative", format: "relative" },
];

function preview(
  unix: number,
  format: Intl.DateTimeFormatOptions | "relative"
) {
  if (format === "relative") return formatRelative(unix);
  return new Intl.DateTimeFormat(undefined, format).format(unix * 1000);
}

// The latest moment a JavaScript Date can hold, in seconds.
const maxUnix = 8.64e12;

// The value of a datetime-local input is in the local time zone, without one.
function toLocalInput(unix: number) {
  const date = new Date(unix * 1000);
  const offset = date.getTimezoneOffset() * 60 * 1000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 19);
}

export default function TimestampGenerator() {
  // Starts empty so the server render doesn't depend on the current time.
  const [unix, setUnix] = useState<number>();

  useEffect(() => setUnix(Math.floor(Date.now() / 1000)), []);

  if (unix === undefined) return null;

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        <div className="space-y-2">
          <Label htmlFor="datetime">Date and time</Label>
          <div className="flex gap-3">
            <Input
              id="datetime"
              type="datetime-local"
              // The browser gives date inputs a wide minimum width.
              className="min-w-0"
              step="1"
              value={toLocalInput(unix)}
              onChange={(e) => {
                const time = new Date(e.target.value).getTime();
                if (!isNaN(time)) setUnix(Math.floor(time / 1000));
              }}
            />
            <Button
              variant="outline"
              className="flex-none"
              onClick={() => setUnix(Math.floor(Date.now() / 1000))}
            >
              Now
            </Button>
          </div>
          <p className="text-sm text-muted-foreground">
            In your time zone,{" "}
            {Intl.DateTimeFormat().resolvedOptions().timeZone}.
          </p>
        </div>
        <div className="space-y-2">
          <Label htmlFor="unix">Unix timestamp</Label>
          <Input
            id="unix"
            inputMode="numeric"
            value={unix}
            onChange={(e) => {
              const value = parseInt(e.target.value.replace(/\D/g, ""));
              if (value <= maxUnix) setUnix(value);
            }}
            className="font-mono"
          />
          <p className="text-sm text-muted-foreground">
            Seconds since January 1, 1970.
          </p>
        </div>
      </div>

      <div className="rounded-lg border divide-y">
        {styles.map(({ style, name, format }) => {
          const code = `<t:${unix}:${style}>`;
          return (
            <div
              key={style}
              className="flex flex-col sm:flex-row sm:items-center gap-3 p-4"
            >
              <div className="sm:w-48 flex-none">
                <div className="font-medium">{name}</div>
                <code className="text-sm text-muted-foreground">{code}</code>
              </div>
              <div className="flex-auto">
                {/* Discord highlights timestamps like this in messages. */}
                <span className="rounded bg-[#e3e5e8] dark:bg-[#3f4147] px-1.5 py-0.5">
                  {preview(unix, format)}
                </span>
              </div>
              <CopyButton value={code} />
            </div>
          );
        })}
      </div>
    </div>
  );
}
