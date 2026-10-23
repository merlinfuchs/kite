const relativeUnits: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 365 * 24 * 60 * 60],
  ["month", 30 * 24 * 60 * 60],
  ["day", 24 * 60 * 60],
  ["hour", 60 * 60],
  ["minute", 60],
  ["second", 1],
];

// Formats a Unix timestamp in seconds like Discord's relative timestamps.
export function formatRelative(unix: number, now = Date.now() / 1000) {
  const diff = unix - Math.floor(now);
  const [unit, seconds] =
    relativeUnits.find(([, s]) => Math.abs(diff) >= s) ?? relativeUnits[5];
  return new Intl.RelativeTimeFormat(undefined, { numeric: "always" }).format(
    Math.round(diff / seconds),
    unit
  );
}

// The styles of Discord timestamps, like <t:1700000000:D>, with how Discord
// formats them.
export const discordTimestampStyles: {
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

// Formats a Unix timestamp in seconds like Discord shows it in the given
// style, in the locale and time zone of the browser.
export function formatDiscordTimestampStyle(
  unix: number,
  style: string,
  now?: number
) {
  const format = (
    discordTimestampStyles.find((s) => s.style === style) ??
    discordTimestampStyles[4]
  ).format;
  // format() throws for dates past what a Date can hold.
  if (isNaN(new Date(unix * 1000).getTime())) return "Invalid Date";
  if (format === "relative") return formatRelative(unix, now);
  return new Intl.DateTimeFormat(undefined, format).format(unix * 1000);
}

// The value of a datetime-local input is in the local time zone, without one.
export function toLocalInput(unix: number) {
  const date = new Date(unix * 1000);
  const offset = date.getTimezoneOffset() * 60 * 1000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 19);
}
