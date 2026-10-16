import { CheckIcon, CopyIcon, XIcon } from "lucide-react";
import { useMemo } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useAppStateGuildQuery } from "@/lib/api/queries";
import {
  decodePermissionsBitset,
  permissionBits,
} from "@/lib/discord/permissions";
import { useAppId } from "@/lib/hooks/params";
import { Guild } from "@/lib/types/wire.gen";

const administratorBit = 3;

export default function AppStateGuildDetailsDialog({
  guild,
  open,
  onOpenChange,
}: {
  guild: Guild;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  // Only loaded once the dialog is opened, it costs requests to Discord.
  const query = useAppStateGuildQuery(useAppId(), open ? guild.id : null);
  const details = query.data?.success ? query.data.data : undefined;
  const error =
    query.data && !query.data.success ? query.data.error.message : undefined;

  const granted = useMemo(
    () =>
      new Set<number>(
        decodePermissionsBitset(details?.permissions ?? "0").map((p) => p.bit)
      ),
    [details?.permissions]
  );
  const isAdministrator = granted.has(administratorBit);

  const ownerName =
    details?.owner.display_name || details?.owner.username || "Unknown user";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{guild.name}</DialogTitle>
          <DialogDescription>
            The owner of this server and what your app is allowed to do in it.
          </DialogDescription>
        </DialogHeader>

        {error ? (
          <div className="text-sm text-destructive">
            Failed to load server details: {error}
          </div>
        ) : !details ? (
          <div className="text-sm text-muted-foreground">Loading...</div>
        ) : (
          <div className="space-y-5">
            <div>
              <div className="text-sm font-medium mb-2">Owner</div>
              <div className="flex items-center gap-3">
                {details.owner.avatar_url ? (
                  <img
                    src={details.owner.avatar_url}
                    alt=""
                    className="w-9 h-9 rounded-full"
                  />
                ) : (
                  <div className="w-9 h-9 rounded-full bg-muted"></div>
                )}
                <div className="min-w-0">
                  <div className="text-sm truncate">
                    {ownerName}
                    {details.owner.username &&
                      details.owner.display_name &&
                      details.owner.username !== details.owner.display_name && (
                        <span className="text-muted-foreground">
                          {" "}
                          @{details.owner.username}
                        </span>
                      )}
                  </div>
                  <div className="text-xs text-muted-foreground font-mono">
                    {details.owner.id}
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="ml-auto h-8 w-8"
                  onClick={() => {
                    navigator.clipboard.writeText(details.owner.id);
                    toast.success("Owner ID copied");
                  }}
                >
                  <span className="sr-only">Copy owner ID</span>
                  <CopyIcon className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div>
              <div className="text-sm font-medium mb-1">Permissions</div>
              <div className="text-xs text-muted-foreground mb-3">
                {isAdministrator
                  ? "Your app is an administrator, so it can do everything in this server."
                  : "What your app's roles allow across the server. Single channels can still allow or deny more."}
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1.5">
                {permissionBits.map((permission) => {
                  const has = isAdministrator || granted.has(permission.bit);

                  return (
                    <div
                      key={permission.bit}
                      className="flex items-center gap-2 text-sm"
                    >
                      {has ? (
                        <CheckIcon className="h-4 w-4 shrink-0 text-green-500" />
                      ) : (
                        <XIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                      )}
                      <span className={has ? "" : "text-muted-foreground"}>
                        {permission.label}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
