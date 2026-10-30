import { cn } from "@/lib/utils";

// A preview of the bot's profile the way Discord shows it when someone clicks
// on the bot.
export default function BotProfilePreview({
  name,
  username,
  avatarUrl,
  bannerUrl,
  about,
  className,
}: {
  name: string;
  username?: string;
  avatarUrl?: string;
  bannerUrl?: string;
  about?: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "w-full max-w-xs overflow-hidden rounded-lg border bg-muted/40",
        className
      )}
    >
      <div
        className="h-24 bg-primary/30 bg-cover bg-center"
        style={bannerUrl ? { backgroundImage: `url(${bannerUrl})` } : undefined}
        role="img"
        aria-label="Banner"
      />
      <div className="px-4 pb-4">
        <div
          className="-mt-10 h-20 w-20 rounded-full border-[6px] border-background bg-muted bg-cover bg-center"
          style={
            avatarUrl ? { backgroundImage: `url(${avatarUrl})` } : undefined
          }
          role="img"
          aria-label="Avatar"
        />
        <div className="mt-2 flex items-center gap-1.5">
          <div className="truncate text-lg font-semibold">{name}</div>
          <span className="shrink-0 rounded bg-indigo-500 px-1 text-[10px] font-semibold text-white">
            APP
          </span>
        </div>
        {username && (
          <div className="truncate text-sm text-muted-foreground">
            {username}
          </div>
        )}
        {about && (
          <div className="mt-3 whitespace-pre-wrap break-words text-sm">
            {about}
          </div>
        )}
      </div>
    </div>
  );
}
