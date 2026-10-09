import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { MarketplaceUser } from "@/lib/types/wire.gen";
import { userAvatarUrl } from "@/lib/marketplace";
import { cn } from "@/lib/utils";

export default function MarketplaceAuthor({
  user,
  className,
}: {
  user?: MarketplaceUser | null;
  className?: string;
}) {
  if (!user) {
    return (
      <span className={cn("text-sm text-muted-foreground", className)}>
        Unknown user
      </span>
    );
  }

  return (
    <div className={cn("flex items-center gap-2 min-w-0", className)}>
      <Avatar className="h-5 w-5">
        <AvatarImage src={userAvatarUrl(user)} alt="" />
        <AvatarFallback className="text-[10px]">
          {user.display_name.slice(0, 1).toUpperCase()}
        </AvatarFallback>
      </Avatar>
      <span className="text-sm text-muted-foreground truncate">
        {user.display_name}
      </span>
    </div>
  );
}
