import {
  useAppStateGuildProfileQuery,
  useAppStateUserQuery,
} from "@/lib/api/queries";
import { useApp, useAppStateGuilds, useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import { useState } from "react";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import GuildProfileDialog from "./GuildProfileDialog";

// The Kite bot's status, the same in every server, see the sharedbot package.
const kiteBotStatus = "🪁 Make a bot at kite.onl";

// useKiteBotLook is the name and avatar the Kite bot has in the app's server.
export function useKiteBotLook() {
  const appId = useAppId();
  const guildId = useApp()?.discord_guild_id ?? "";
  const user = useResponseData(useAppStateUserQuery(appId));
  const profile = useResponseData(useAppStateGuildProfileQuery(appId, guildId));
  return {
    name: profile?.nick || user?.username || "Kite",
    avatarUrl: profile?.avatar_url || user?.avatar_url || undefined,
  };
}

// KiteBotMemberPreview shows the Kite bot like the member list of a server.
export function KiteBotMemberPreview({
  name,
  avatarUrl,
}: {
  name: string;
  avatarUrl?: string;
}) {
  return (
    <div className="flex items-center gap-3 rounded-lg bg-muted/50 px-3 py-2.5">
      {avatarUrl ? (
        <img
          src={avatarUrl}
          className="h-10 w-10 flex-none rounded-full object-cover"
          alt=""
        />
      ) : (
        <div className="h-10 w-10 flex-none rounded-full bg-muted" />
      )}
      <div className="min-w-0">
        <div className="flex items-center gap-1.5">
          <span className="font-medium truncate">{name}</span>
          <span className="rounded bg-[#5865f2] px-1 text-[10px] font-semibold leading-4 text-white">
            APP
          </span>
        </div>
        <div className="text-xs text-muted-foreground truncate">
          {kiteBotStatus}
        </div>
      </div>
    </div>
  );
}

// KiteBotLook is how the Kite bot looks in the app's server, with a button to
// change it.
export default function KiteBotLook() {
  const appId = useAppId();
  const guild = useAppStateGuilds()?.[0];
  const look = useKiteBotLook();
  const [open, setOpen] = useState(false);

  if (!guild) return null;

  return (
    <Card className="flex flex-col sm:flex-row sm:items-center gap-4 p-4">
      <div className="flex-1 min-w-0 space-y-2">
        <KiteBotMemberPreview name={look.name} avatarUrl={look.avatarUrl} />
        <p className="text-xs text-muted-foreground">
          The status is the same in every server. Want your own?{" "}
          <Link
            href={{
              pathname: "/apps/[appId]/settings",
              query: { appId },
              hash: "switch",
            }}
            className="text-primary hover:underline"
          >
            Switch to Custom setup
          </Link>
        </p>
      </div>
      <div className="flex-none">
        <Button variant="outline" onClick={() => setOpen(true)}>
          Change look
        </Button>
      </div>
      <GuildProfileDialog guild={guild} open={open} onOpenChange={setOpen} />
    </Card>
  );
}
