import * as React from "react";
import { format } from "date-fns";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { CalendarIcon, ClockIcon, XIcon } from "lucide-react";

function parseDate(value: string | undefined): Date | undefined {
  if (!value) return undefined;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? undefined : date;
}

export function DatePicker({
  value,
  onChange,
  now,
  onNowChange,
}: {
  value: string | undefined;
  onChange: (value: string | undefined) => void;
  // When onNowChange is set, the picker offers "Time of sending", which uses
  // the time the message is sent instead of a fixed date. The fixed date is
  // kept, so turning it off again brings the date back. onChange should turn
  // it off, so a picked or cleared date applies in one update.
  now?: boolean;
  onNowChange?: (now: boolean) => void;
}) {
  const date = parseDate(value);
  const [month, setMonth] = React.useState<Date | undefined>(date);

  // Picking a day keeps the time already set, or the current time if none is.
  function selectDay(day: Date | undefined) {
    if (!day) return onChange(undefined);
    const time = date ?? new Date();
    const next = new Date(day);
    next.setHours(time.getHours(), time.getMinutes(), 0, 0);
    onChange(next.toISOString());
  }

  function setTime(time: string) {
    const [hours, minutes] = time.split(":").map((v) => parseInt(v, 10));
    if (Number.isNaN(hours) || Number.isNaN(minutes)) return;
    const next = date ? new Date(date) : new Date();
    next.setHours(hours, minutes, 0, 0);
    onChange(next.toISOString());
  }

  function setToCurrent() {
    const current = new Date();
    current.setSeconds(0, 0);
    setMonth(current);
    onChange(current.toISOString());
  }

  function clear(e: React.MouseEvent) {
    // The icon sits inside the popover trigger, don't open it.
    e.preventDefault();
    e.stopPropagation();
    onChange(undefined);
  }

  const hasValue = !!now || !!date;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant={"outline"}
          className={cn(
            "w-full justify-start text-left font-normal",
            !hasValue && "text-muted-foreground"
          )}
        >
          {now ? (
            <ClockIcon className="mr-2 h-4 w-4" />
          ) : (
            <CalendarIcon className="mr-2 h-4 w-4" />
          )}
          {now ? (
            <span>Time of sending</span>
          ) : date ? (
            format(date, "PPP 'at' p")
          ) : (
            <span>Pick a date and time</span>
          )}
          {hasValue && (
            <XIcon
              className="ml-auto h-4 w-4 opacity-60 hover:opacity-100"
              onClick={clear}
              onPointerDown={(e) => e.stopPropagation()}
            />
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        {onNowChange && (
          <div className="flex items-center justify-between gap-4 border-b px-3 py-2.5">
            <div>
              <Label htmlFor="date-picker-now" className="text-sm">
                Time of sending
              </Label>
              <div className="text-xs text-muted-foreground">
                Use the time the message is sent
              </div>
            </div>
            <Switch
              id="date-picker-now"
              checked={!!now}
              onCheckedChange={onNowChange}
            />
          </div>
        )}
        <div
          className={cn(now && "pointer-events-none opacity-40")}
          aria-disabled={now}
        >
          <Calendar
            mode="single"
            selected={date}
            onSelect={selectDay}
            month={month}
            onMonthChange={setMonth}
            initialFocus
          />
          <div className="flex items-center gap-2 border-t px-3 py-2.5">
            <ClockIcon className="h-4 w-4 text-muted-foreground flex-none" />
            <Input
              type="time"
              className="h-8"
              value={date ? format(date, "HH:mm") : ""}
              onChange={(e) => setTime(e.target.value)}
              disabled={now}
            />
            <Button
              variant="outline"
              size="sm"
              className="h-8"
              onClick={setToCurrent}
              disabled={now}
            >
              Now
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
