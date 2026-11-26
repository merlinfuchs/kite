import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useAppStateUserQuery } from "@/lib/api/queries";
import { useApp, useResponseData } from "@/lib/hooks/api";
import { BotIcon } from "lucide-react";

// BotAvatar is the avatar of the app's bot. Discord only tells a running bot
// its avatar, so it's a bot icon while the app is stopped.
//
// It shows the app from the current route unless an appId is given, which is
// how lists of apps show each app's own bot.
export default function BotAvatar({
  appId,
  className,
  fallbackClassName,
}: {
  appId?: string;
  className?: string;
  fallbackClassName?: string;
}) {
  const app = useApp();
  const bot = useResponseData(useAppStateUserQuery(appId ?? app?.id ?? ""));

  return (
    <Avatar className={className}>
      {bot?.avatar_url && <AvatarImage src={bot.avatar_url} alt="" />}
      <AvatarFallback className={fallbackClassName}>
        <BotIcon className="size-5" />
      </AvatarFallback>
    </Avatar>
  );
}
