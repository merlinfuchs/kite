import { cn } from "@/lib/utils";
import { Badge } from "../ui/badge";

export default function LogLevelBadge({ level }: { level: string }) {
  return (
    <Badge
      variant="secondary"
      className={cn(
        "uppercase text-xs select-none",
        level === "info"
          ? "bg-blue-500 hover:bg-blue-500/80 text-white"
          : level === "warn"
          ? "bg-orange-500 hover:bg-orange-500/80 text-white"
          : level === "error"
          ? "bg-red-500 hover:bg-red-500/80 text-white"
          : ""
      )}
    >
      {level}
    </Badge>
  );
}
