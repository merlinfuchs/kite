import { useMemo, useState } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";
import { VariableIcon } from "lucide-react";
import {
  interactionPlaceholders,
  PlaceholderGroup,
} from "@/lib/flow/placeholders";
import { Input } from "../ui/input";
import { useNoPlaceholders } from "@/lib/message/state";
import {
  discordTimestampStyles,
  formatDiscordTimestampStyle,
  toLocalInput,
} from "@/tools/common/utils/time";
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
  // Without anything to fill in placeholders, only a fixed time works.
  const noPlaceholders = useNoPlaceholders();
  const [context, setContext] = useState<Context>(
    noPlaceholders ? "time" : "interaction"
  );
  // When the explorer was last opened, in seconds. The time tab previews it,
  // and isn't built before the first open.
  const [openedAt, setOpenedAt] = useState<number>();

  const [timeSource, setTimeSource] = useState<TimeSource>(
    noPlaceholders ? "custom" : "now"
  );
  // Tomorrow, to the minute so the input leaves out the seconds.
  const [customTime, setCustomTime] = useState(() =>
    toLocalInput(Math.floor(Date.now() / 60_000) * 60 + 24 * 60 * 60)
  );
  const [countdownAmount, setCountdownAmount] = useState("1");
  const [countdownUnit, setCountdownUnit] = useState<CountdownUnit>("hours");

  const placeholders = useMemo((): PlaceholderGroup[] => {
    // The same placeholders as in a flow, as messages are sent from one.
    if (context !== "time") {
      return interactionPlaceholders(
        context === "event" ? "event_discord" : "command"
      );
    }
    if (openedAt === undefined) return [];

    if (timeSource === "custom") {
      const unix = Math.floor(new Date(customTime).getTime() / 1000);
      if (isNaN(unix)) return [];
      return timePlaceholders(timeSource, unix, openedAt, `${unix}`);
    }

    if (timeSource === "countdown") {
      const amount = Math.floor(Number(countdownAmount));
      const seconds = amount * countdownUnits[countdownUnit].seconds;
      // Larger values would go into the placeholder as 1e+21 and the like.
      if (!(amount >= 1 && seconds <= maxCountdown)) return [];

      return timePlaceholders(
        timeSource,
        openedAt + seconds,
        openedAt,
        `{{now().Unix() + ${seconds}}}`
      );
    }

    return timePlaceholders(timeSource, openedAt, openedAt, "{{now().Unix()}}");
  }, [
    context,
    openedAt,
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
          ...(noPlaceholders
            ? []
            : [
                {
                  label: "Interaction",
                  value: "interaction",
                },
                {
                  label: "Event",
                  value: "event",
                },
              ]),
          {
            label: "Time",
            value: "time",
          },
        ]}
        onTabChange={(tab) => {
          setContext(tab as Context);
        }}
        onOpenChange={(open) => {
          if (open) setOpenedAt(Math.floor(Date.now() / 1000));
        }}
        header={
          context === "time" && (
            <div className="px-3 pt-2 pb-2 space-y-2">
              {!noPlaceholders && (
                <Select
                  value={timeSource}
                  onValueChange={(v) => setTimeSource(v as TimeSource)}
                >
                  <SelectTrigger className="h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.entries(timeSources).map(([source, s]) => (
                      <SelectItem value={source} key={source}>
                        {s.option}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
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
              {openedAt !== undefined && placeholders.length === 0 && (
                <div className="text-muted-foreground text-xs">
                  {timeSource === "custom"
                    ? "Pick a date and time."
                    : "Enter a number from 1 to 100 years."}
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
type TimeSource = keyof typeof timeSources;

const timeSources = {
  now: {
    option: "When the message is sent",
    group: "Current time",
    relative: "Counts up from when it's sent",
  },
  custom: {
    option: "A specific date and time",
    group: "Custom time",
    relative: "Counts down, then up",
  },
  countdown: {
    option: "Some time after it's sent",
    group: "Later",
    relative: "Counts down from when it's sent",
  },
};

type CountdownUnit = keyof typeof countdownUnits;

const maxCountdown = 100 * 365 * 24 * 60 * 60;

const countdownUnits = {
  seconds: { label: "Seconds", seconds: 1 },
  minutes: { label: "Minutes", seconds: 60 },
  hours: { label: "Hours", seconds: 60 * 60 },
  days: { label: "Days", seconds: 24 * 60 * 60 },
  weeks: { label: "Weeks", seconds: 7 * 24 * 60 * 60 },
};

// Discord timestamps, labeled with what Discord will show for them. Discord
// renders them in the locale of whoever reads the message, the browser locale
// is the closest we can get to that here.
function timePlaceholders(
  source: TimeSource,
  // The time that is previewed in the labels.
  unix: number,
  now: number,
  // What goes into <t:...>, a Unix timestamp or a placeholder for one.
  timestamp: string
): PlaceholderGroup[] {
  return [
    {
      label: timeSources[source].group,
      placeholders: discordTimestampStyles.map(({ style, name }) => {
        const relative = style === "R";
        return {
          label:
            relative && source === "now"
              ? "a few seconds ago"
              : formatDiscordTimestampStyle(unix, style, now),
          description: relative ? timeSources[source].relative : name,
          value: `<t:${timestamp}:${style}>`,
          raw: true,
          keywords: [
            name,
            "timestamp",
            ...(relative ? ["timer", "countdown", "ago"] : []),
          ],
        };
      }),
    },
  ];
}
