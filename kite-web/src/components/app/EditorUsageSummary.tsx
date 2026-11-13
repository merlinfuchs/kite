import type { EditorUsage } from "@/lib/types/wire.gen";

// formatEditingTime formats a time like "2h 15m", or "<1m" for less than a
// minute.
export function formatEditingTime(seconds: number) {
  const minutes = Math.floor(seconds / 60);
  if (minutes < 1) return "<1m";
  const hours = Math.floor(minutes / 60);
  if (hours === 0) return `${minutes}m`;
  return minutes % 60 ? `${hours}h ${minutes % 60}m` : `${hours}h`;
}

export default function EditorUsageSummary({ usage }: { usage?: EditorUsage }) {
  if (!usage) return null;

  const saves = usage.save_count === 1 ? "1 save" : `${usage.save_count} saves`;
  return (
    <span className="block text-xs mt-1">
      {formatEditingTime(usage.active_seconds)} of editing and {saves} in the
      last 30 days
    </span>
  );
}
