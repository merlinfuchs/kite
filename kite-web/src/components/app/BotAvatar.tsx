import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useAppStateUserQuery } from "@/lib/api/queries";
import { useApp, useResponseData } from "@/lib/hooks/api";
import { BotIcon } from "lucide-react";

// BotAvatar is the avatar of the app's bot. Discord only tells a running bot
// its avatar, so it's a bot icon while the app is stopped.
export default function BotAvatar({
  className,
  fallbackClassName,
}: {
  className?: string;
  fallbackClassName?: string;
}) {
  const app = useApp();
  const bot = useResponseData(useAppStateUserQuery(app?.id ?? ""));

  return (
    <Avatar className={className}>
      {bot?.avatar_url && <AvatarImage src={bot.avatar_url} alt="" />}
      <AvatarFallback className={fallbackClassName}>
        <BotIcon className="size-5" />
      </AvatarFallback>
    </Avatar>
  );
}
