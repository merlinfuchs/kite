import { useState } from "react";
import { toast } from "sonner";
import { TrashIcon } from "lucide-react";
import { Card } from "../ui/card";
import { Input } from "../ui/input";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import LoadingButton from "../common/LoadingButton";
import ConfirmDialog from "../common/ConfirmDialog";
import {
  useMarketplaceModeratorCreateMutation,
  useMarketplaceModeratorDeleteMutation,
  useMarketplaceModeratorsQuery,
} from "@/lib/api/marketplace";
import MarketplaceAuthor from "./MarketplaceAuthor";

export default function MarketplaceModeratorList({
  isAdmin,
}: {
  isAdmin: boolean;
}) {
  const [discordId, setDiscordId] = useState("");

  const query = useMarketplaceModeratorsQuery();
  const createMutation = useMarketplaceModeratorCreateMutation();
  const deleteMutation = useMarketplaceModeratorDeleteMutation();

  const moderators = query.data?.success
    ? query.data.data.flatMap((m) => (m ? [m] : []))
    : [];

  function onAdd() {
    const id = discordId.trim();
    if (!/^[0-9]{15,21}$/.test(id)) {
      toast.error("Enter a Discord user ID");
      return;
    }

    createMutation.mutate(
      { discord_user_id: id },
      {
        onSuccess: (res) => {
          if (res.success) {
            toast.success("Moderator added");
            setDiscordId("");
          } else {
            toast.error(
              `Failed to add moderator: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  function onRemove(id: string) {
    deleteMutation.mutate(id, {
      onSuccess: (res) => {
        if (res.success) {
          toast.success("Moderator removed");
        } else {
          toast.error(
            `Failed to remove moderator: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <div className="space-y-5 max-w-2xl">
      {isAdmin ? (
        <div className="flex gap-2">
          <Input
            value={discordId}
            onChange={(e) => setDiscordId(e.target.value)}
            placeholder="Discord user ID"
            inputMode="numeric"
            maxLength={21}
          />
          <LoadingButton onClick={onAdd} loading={createMutation.isPending}>
            Add moderator
          </LoadingButton>
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">
          Only admins can add or remove moderators.
        </p>
      )}

      <div className="flex flex-col gap-3">
        {moderators.map((m) => (
          <Card
            key={m.discord_user_id}
            className="flex items-center justify-between gap-3 px-4 py-3"
          >
            <div className="min-w-0">
              {m.user ? (
                <MarketplaceAuthor user={m.user} />
              ) : (
                <span className="text-sm text-muted-foreground">
                  Hasn&apos;t logged in yet
                </span>
              )}
              <div className="text-xs text-muted-foreground font-mono mt-1">
                {m.discord_user_id}
              </div>
            </div>
            <div className="flex items-center gap-2 flex-none">
              {m.is_admin ? (
                <Badge>Admin</Badge>
              ) : (
                <Badge variant="secondary">Moderator</Badge>
              )}
              {isAdmin && !m.is_admin && (
                <ConfirmDialog
                  title="Remove moderator?"
                  description="They will lose access to the moderation queue and reports."
                  onConfirm={() => onRemove(m.discord_user_id)}
                >
                  <Button variant="ghost" size="icon">
                    <TrashIcon className="h-4 w-4" />
                  </Button>
                </ConfirmDialog>
              )}
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
