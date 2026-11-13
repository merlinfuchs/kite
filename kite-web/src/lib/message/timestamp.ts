import { isToday, isYesterday } from "date-fns";

// A time like Discord shows it next to messages and in embed footers: "Today
// at 3:04 PM", "Yesterday at 3:04 PM" or "10/14/2026 3:04 PM", with the date
// and time in the viewer's locale. Discord also translates "Today at", which
// the preview leaves in English like the rest of the app.
export function formatDiscordTimestamp(date: Date, locale?: string): string {
  const time = date
    .toLocaleTimeString(locale, { hour: "numeric", minute: "2-digit" })
    // Some ICU versions put a narrow no-break space before AM/PM.
    .replace(/\u202f/g, " ");

  if (isToday(date)) return `Today at ${time}`;
  if (isYesterday(date)) return `Yesterday at ${time}`;

  const day = date.toLocaleDateString(locale, {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
  return `${day} ${time}`;
}
