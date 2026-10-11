const relativeUnits: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 365 * 24 * 60 * 60],
  ["month", 30 * 24 * 60 * 60],
  ["day", 24 * 60 * 60],
  ["hour", 60 * 60],
  ["minute", 60],
  ["second", 1],
];

// Formats a Unix timestamp in seconds like Discord's relative timestamps.
export function formatRelative(unix: number) {
  const diff = unix - Math.floor(Date.now() / 1000);
  const [unit, seconds] =
    relativeUnits.find(([, s]) => Math.abs(diff) >= s) ?? relativeUnits[5];
  return new Intl.RelativeTimeFormat(undefined, { numeric: "always" }).format(
    Math.round(diff / seconds),
    unit
  );
}
