import { useMemo, useState } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";
import { VariableIcon } from "lucide-react";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

export default function MessagePlaceholderExplorer({
  onSelect,
}: {
  onSelect: (value: string, raw?: boolean) => void;
}) {
  const [context, setContext] = useState<Context>("interaction");
  // The time tab previews the current time, so it's refreshed on every open.
  const [openedAt, setOpenedAt] = useState(() => new Date());

  const [timeSource, setTimeSource] = useState<TimeSource>("now");
  const [customTime, setCustomTime] = useState(() =>
    toDateTimeLocal(new Date(Date.now() + 24 * 60 * 60 * 1000))
  );
  const [countdownAmount, setCountdownAmount] = useState("1");
  const [countdownUnit, setCountdownUnit] = useState<CountdownUnit>("hours");

  const globalPlaceholders = useGlobalPlaceholders(context);

  const placeholders = useMemo(() => {
    if (context !== "time") return [...globalPlaceholders];

    if (timeSource === "custom") {
      const date = new Date(customTime);
      if (isNaN(date.getTime())) return [];

      return timePlaceholders({
        source: "custom",
        now: openedAt,
        date,
        timestamp: `${Math.floor(date.getTime() / 1000)}`,
      });
    }

    if (timeSource === "countdown") {
      const amount = Math.floor(Number(countdownAmount));
      if (!Number.isFinite(amount) || amount < 1) return [];

      const seconds = amount * countdownUnits[countdownUnit].seconds;
      return timePlaceholders({
        source: "countdown",
        now: openedAt,
        date: new Date(openedAt.getTime() + seconds * 1000),
        timestamp: `{{now().Unix() + ${seconds}}}`,
      });
    }

    return timePlaceholders({
      source: "now",
      now: openedAt,
      date: openedAt,
      timestamp: "{{now().Unix()}}",
    });
  }, [
    context,
    openedAt,
    globalPlaceholders,
    timeSource,
    customTime,
    countdownAmount,
    countdownUnit,
  ]);

  return (
    <div className="absolute top-10 right-1.5 z-20">
      <PlaceholderExplorer
        onSelect={onSelect}
        placeholders={placeholders}
        tab={context}
        tabs={[
          {
            label: "Interaction",
            value: "interaction",
          },
          {
            label: "Event",
            value: "event",
          },
          {
            label: "Time",
            value: "time",
          },
        ]}
        onTabChange={(tab) => {
          setContext(tab as Context);
        }}
        onOpenChange={(open) => {
          if (open) setOpenedAt(new Date());
        }}
        header={
          context === "time" && (
            <div className="px-3 pt-2 space-y-2">
              <Select
                value={timeSource}
                onValueChange={(v) => setTimeSource(v as TimeSource)}
              >
                <SelectTrigger className="h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="now">When the message is sent</SelectItem>
                  <SelectItem value="custom">
                    A specific date and time
                  </SelectItem>
                  <SelectItem value="countdown">
                    Some time after it&apos;s sent
                  </SelectItem>
                </SelectContent>
              </Select>
              {timeSource === "custom" && (
                <Input
                  type="datetime-local"
                  className="h-8 dark:[color-scheme:dark]"
                  value={customTime}
                  onChange={(e) => setCustomTime(e.target.value)}
                />
              )}
              {timeSource === "countdown" && (
                <div className="flex space-x-2">
                  <Input
                    type="number"
                    min={1}
                    className="h-8 w-24 flex-none"
                    value={countdownAmount}
                    onChange={(e) => setCountdownAmount(e.target.value)}
                  />
                  <Select
                    value={countdownUnit}
                    onValueChange={(v) => setCountdownUnit(v as CountdownUnit)}
                  >
                    <SelectTrigger className="h-8">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.entries(countdownUnits).map(([unit, u]) => (
                        <SelectItem value={unit} key={unit}>
                          {u.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
              {placeholders.length === 0 && (
                <div className="text-muted-foreground pb-2 text-xs">
                  {timeSource === "custom"
                    ? "Pick a date and time."
                    : "Enter a number of at least 1."}
                </div>
              )}
            </div>
          )
        }
      >
        <VariableIcon
          className="h-5.5 w-5.5 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
        />
      </PlaceholderExplorer>
    </div>
  );
}

type Context = "interaction" | "event" | "time";

// Where the time of a timestamp comes from: the moment the message is sent,
// a fixed date, or an offset from the moment it's sent.
type TimeSource = "now" | "custom" | "countdown";

type CountdownUnit = keyof typeof countdownUnits;

const countdownUnits = {
  seconds: { label: "Seconds", seconds: 1 },
  minutes: { label: "Minutes", seconds: 60 },
  hours: { label: "Hours", seconds: 60 * 60 },
  days: { label: "Days", seconds: 24 * 60 * 60 },
  weeks: { label: "Weeks", seconds: 7 * 24 * 60 * 60 },
};

// Formats a date as the local time value of a datetime-local input.
function toDateTimeLocal(date: Date) {
  const pad = (n: number) => `${n}`.padStart(2, "0");
  return (
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
    `T${pad(date.getHours())}:${pad(date.getMinutes())}`
  );
}

// What Discord roughly shows for a relative timestamp, like "in 2 hours".
function relativeLabel(seconds: number) {
  const steps: [Intl.RelativeTimeFormatUnit, number][] = [
    ["year", 365 * 24 * 60 * 60],
    ["month", 30 * 24 * 60 * 60],
    ["day", 24 * 60 * 60],
    ["hour", 60 * 60],
    ["minute", 60],
    ["second", 1],
  ];
  const [unit, size] =
    steps.find(([, size]) => Math.abs(seconds) >= size) ??
    steps[steps.length - 1];

  return new Intl.RelativeTimeFormat(undefined, { numeric: "always" }).format(
    Math.round(seconds / size),
    unit
  );
}

// Discord timestamps, labeled with what Discord will show for them. Discord
// renders them in the locale of whoever reads the message, the browser locale
// is the closest we can get to that here.
function timePlaceholders({
  source,
  now,
  date,
  timestamp,
}: {
  source: TimeSource;
  now: Date;
  // The time that is previewed in the labels.
  date: Date;
  // What goes into <t:...>, a Unix timestamp or a placeholder for one.
  timestamp: string;
}) {
  const time = date.toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "2-digit",
  });
  const longTime = date.toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "2-digit",
    second: "2-digit",
  });
  const shortDate = date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
  const longDate = date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
  const fullDate = date.toLocaleDateString(undefined, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const relative = {
    now: {
      label: "a few seconds ago",
      description: "Counts up from when it's sent",
    },
    custom: {
      label: relativeLabel((date.getTime() - now.getTime()) / 1000),
      description: "Counts down, then up",
    },
    countdown: {
      label: relativeLabel((date.getTime() - now.getTime()) / 1000),
      description: "Counts down from when it's sent",
    },
  }[source];

  const styles: {
    style: string;
    label: string;
    description: string;
    keywords?: string[];
  }[] = [
    {
      style: "R",
      ...relative,
      keywords: ["relative", "timer", "countdown", "ago"],
    },
    { style: "t", label: time, description: "Time" },
    { style: "T", label: longTime, description: "Time with seconds" },
    { style: "d", label: shortDate, description: "Date" },
    { style: "D", label: longDate, description: "Long date" },
    { style: "f", label: `${longDate} ${time}`, description: "Date and time" },
    {
      style: "F",
      label: `${fullDate} ${time}`,
      description: "Day, date and time",
      keywords: ["weekday"],
    },
  ];

  return [
    {
      label: { now: "Current time", custom: "Custom time", countdown: "Later" }[
        source
      ],
      placeholders: styles.map((s) => ({
        label: s.label,
        description: s.description,
        value: `<t:${timestamp}:${s.style}>`,
        raw: true,
        keywords: [
          s.label,
          s.description,
          "timestamp",
          "now",
          ...(s.keywords ?? []),
        ],
      })),
    },
  ];
}

function useGlobalPlaceholders(context: Context) {
  return useMemo(() => {
    const res = [
      {
        label: "User",
        placeholders: [
          {
            label: "User",
            value: `user`,
          },
          {
            label: "User ID",
            value: `user.id`,
          },
          {
            label: "User Mention",
            value: `user.mention`,
          },
          {
            label: "User Username",
            value: `user.username`,
          },
          {
            label: "User Discriminator",
            value: `user.discriminator`,
          },
          {
            label: "User Display Name",
            value: `user.display_name`,
          },
          {
            label: "User Avatar URL",
            value: `user.avatar_url`,
          },
          {
            label: "User Banner URL",
            value: `user.banner_url`,
          },
        ],
      },
      {
        label: "Server",
        placeholders: [
          {
            label: "Server ID",
            value: `guild.id`,
          },
        ],
      },
      {
        label: "Channel",
        placeholders: [
          {
            label: "Channel ID",
            value: `channel.id`,
          },
        ],
      },
    ];

    if (context === "event") {
      res.push({
        label: "Message",
        placeholders: [
          { label: "Message ID", value: `message.id` },
          { label: "Message Content", value: `message.content` },
        ],
      });
    }
    return res;
  }, [context]);
}
