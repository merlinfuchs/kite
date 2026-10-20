import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

// Initials avatar for the logged in user, used in the navbar and dashboard.
export default function UserAvatar({
  displayName,
  className,
}: {
  displayName?: string;
  className?: string;
}) {
  return (
    <Avatar className={cn("h-9 w-9", className)}>
      <AvatarFallback className="bg-primary/15 text-sm font-semibold uppercase text-primary">
        {abbreviateDisplayName(displayName || "")}
      </AvatarFallback>
    </Avatar>
  );
}

function abbreviateDisplayName(name: string) {
  return name
    .split(" ")
    .map((word) => word[0])
    .join("");
}
